import { useEffect, useState } from "react";
import { collectorService, ingestServiceSimple, type CollectorConfig, type CollectorTemplate, type IngestConfig } from "@/services/collector";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Plus, Trash2, Edit, Download, Server, Package, Activity, CheckCircle, Filter } from "lucide-react";
import { toast } from "sonner";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

const typeIcons: Record<string, any> = {
  windows: Package,
  linux: Server,
  macos: Activity,
};

// 增强的数据源接口，加入高级过滤属性与场景预设
interface DataSource {
  type: string;
  path: string;
  label: string;
  enabled: boolean;
  event_ids_str?: string; 
  query?: string;         
  presets?: { name: string, ids: string }[]; 
}

export default function CollectorsPage() {
  const [configs, setConfigs] = useState<CollectorConfig[]>([]);
  const [templates, setTemplates] = useState<CollectorTemplate[]>([]);
  const [ingests, setIngests] = useState<IngestConfig[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingConfig, setEditingConfig] = useState<CollectorConfig | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [availableSources, setAvailableSources] = useState<DataSource[]>([]);
  const [activeTab, setActiveTab] = useState<string>("templates");

  const defaultStreamFields = "observer.hostname,observer.vendor,class_uid";

  const [formData, setFormData] = useState({
    name: "",
    type: "windows",
    channels: "",
    sources: "", 
    ingest_id: 0,
    stream_fields: defaultStreamFields,
    interval: 5,
  });

  const fetchData = async () => {
    try {
      const [configRes, tmplRes, ingestRes] = await Promise.all([
        collectorService.list(),
        collectorService.templates(),
        ingestServiceSimple.list(),
      ]);
      
      if (configRes.code === 200) {
        setConfigs(Array.isArray(configRes.data) ? configRes.data : []);
      }
      if (tmplRes.code === 200) {
        setTemplates(tmplRes.data || []);
      }
      if (ingestRes.code === 200) {
        setIngests(Array.isArray(ingestRes.data) ? ingestRes.data : []);
      }
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const fetchSourcesForType = async (type: string, savedSourcesStr?: string) => {
    try {
      const res = await collectorService.getSources(type);
      if (res.code === 200 && res.data) {
        const data = res.data;
        if (Array.isArray(data) && data.length > 0) {
          // 解析已保存的配置以进行回显
          const savedMap = new Map();
          if (savedSourcesStr) {
            try {
              const parsed = JSON.parse(savedSourcesStr);
              parsed.forEach((s: any) => {
                savedMap.set(s.type || s.path, {
                  enabled: s.enabled,
                  event_ids_str: s.event_ids ? s.event_ids.join(", ") : "",
                  query: s.query || ""
                });
              });
            } catch (e) {}
          }

          const sources: DataSource[] = (data as any[]).map((item: any) => {
            const key = item.type || item.Path || '';
            const savedState = savedMap.get(key) || { enabled: false, event_ids_str: "", query: "" };
            
            return {
              type: item.type || item.Type || '',
              path: item.path || item.Path || '',
              label: item.label || item.Label || item.type || '',
              enabled: savedState.enabled,
              event_ids_str: savedState.event_ids_str,
              query: savedState.query,
              presets: item.presets || [], 
            };
          });
          setAvailableSources(sources);
        }
      }
    } catch (err) {
      console.error("Failed to fetch sources:", err);
    }
  };

  useEffect(() => {
    if (formData.type) {
      fetchSourcesForType(formData.type, formData.sources);
    }
  }, [formData.type]);

  // 将 UI 状态序列化为 Agent 所需的严格 JSON 结构
  const syncSourcesToFormData = (sources: DataSource[]) => {
    const enabledSources = sources.filter(s => s.enabled).map(s => {
      let ids: number[] = [];
      if (s.event_ids_str && s.event_ids_str.trim() !== "") {
        // 将 "4624, 4625" 转换为 [4624, 4625]
        ids = s.event_ids_str.split(',')
          .map(id => parseInt(id.trim()))
          .filter(id => !isNaN(id));
      }
      
      return {
        type: s.type,
        path: s.path,
        format: formData.type === "windows" ? "windows_event" : "file",
        enabled: true,
        event_ids: ids.length > 0 ? ids : undefined,
        query: s.query && s.query.trim() !== "" ? s.query.trim() : undefined
      };
    });
    
    setFormData(prev => ({ ...prev, sources: JSON.stringify(enabledSources) }));
  };

  // 勾选/取消勾选日志源
  const toggleSource = (type: string) => {
    const updated = availableSources.map(s => 
      s.type === type ? { ...s, enabled: !s.enabled } : s
    );
    setAvailableSources(updated);
    syncSourcesToFormData(updated);
  };

  // 更新对应数据源的高级配置（EventID / XPath）
  const updateSourceConfig = (type: string, field: 'event_ids_str' | 'query', value: string) => {
    const updated = availableSources.map(s => 
      s.type === type ? { ...s, [field]: value } : s
    );
    setAvailableSources(updated);
    syncSourcesToFormData(updated);
  };

  // 处理快捷预设的点击（智能追加与去重）
  const handlePresetClick = (source: DataSource, presetIds: string) => {
    const current = source.event_ids_str ? source.event_ids_str.split(',').map(s => s.trim()).filter(s => s) : [];
    const newIds = presetIds.split(',').map(s => s.trim());
    
    // 使用 Set 进行合并去重
    const merged = Array.from(new Set([...current, ...newIds]));
    updateSourceConfig(source.type, 'event_ids_str', merged.join(', '));
  };

  const getConfigID = (c: CollectorConfig) => c.ID || c.id || 0;

  const handleOpenDialog = (config?: CollectorConfig, template?: CollectorTemplate) => {
    if (config) {
      setEditingConfig(config);
      setFormData({
        name: config.name,
        type: config.type,
        channels: config.channels || "",
        sources: config.sources || "",
        ingest_id: config.ingest_id || 0,
        stream_fields: config.stream_fields || defaultStreamFields,
        interval: config.interval || 5,
      });
      fetchSourcesForType(config.type, config.sources);
    } else if (template) {
      setEditingConfig(null);
      setFormData({
        name: template.name,
        type: template.type,
        channels: "",
        sources: "",
        ingest_id: 0,
        stream_fields: defaultStreamFields,
        interval: 5,
      });
      fetchSourcesForType(template.type);
    } else {
      setEditingConfig(null);
      setFormData({
        name: "",
        type: "windows",
        channels: "",
        sources: "",
        ingest_id: 0,
        stream_fields: defaultStreamFields,
        interval: 5,
      });
      fetchSourcesForType("windows");
    }
    setDialogOpen(true);
  };

  const handleSubmit = async () => {
    if (!formData.name || !formData.type) {
      toast.error("Name and Type are required");
      return;
    }

    setSubmitting(true);
    try {
      if (editingConfig) {
        await collectorService.update({ id: getConfigID(editingConfig), ...formData });
        toast.success("Config updated successfully");
      } else {
        await collectorService.add(formData);
        toast.success("Config created successfully");
      }
      setDialogOpen(false);
      fetchData();
    } catch (err) {
      console.error(err);
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm("Are you sure you want to delete this config?")) return;
    try {
      await collectorService.delete(id);
      toast.success("Config deleted");
      fetchData();
    } catch (err) {
      console.error(err);
    }
  };

  const handleBuild = async (config: CollectorConfig) => {
    try {
      toast.info("Compiling standalone collector binary...", {
        description: "Executing Cloud-Native compilation..."
      });
      
      const res = await collectorService.build(getConfigID(config));
      
      const url = window.URL.createObjectURL(res.data);
      const a = document.createElement('a');
      a.href = url;
      const safeName = config.name.replace(/\s+/g, '_').toLowerCase();
      const ext = config.type === "windows" ? ".exe" : "";
      a.download = `vsentry-agent-${safeName}${ext}`;
      
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
      
      toast.success("Collector compiled and downloaded!", {
        description: "Drop the binary onto your server and run it."
      });
      fetchData();
    } catch (err) {
      toast.error("Compilation failed");
    }
  };

  return (
    <div className="p-6 h-full flex flex-col">
      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col">
        <div className="flex justify-between items-center mb-4">
          <div>
            <h1 className="text-2xl font-bold flex items-center gap-2">
              <Server className="w-6 h-6" />
              Collectors
            </h1>
            <p className="text-muted-foreground text-sm">
              Deploy native, zero-dependency XDR agents tailored for OCSF.
            </p>
          </div>
          <Button onClick={() => handleOpenDialog()}>
            <Plus className="w-4 h-4 mr-2" />
            New Collector
          </Button>
        </div>

        <TabsList className="mb-4">
          <TabsTrigger value="templates">Templates</TabsTrigger>
          <TabsTrigger value="windows">Windows</TabsTrigger>
          <TabsTrigger value="linux">Linux</TabsTrigger>
          <TabsTrigger value="macos">macOS</TabsTrigger>
        </TabsList>

        <TabsContent value="templates" className="flex-1 mt-0">
          <div className="mb-6">
            <h3 className="text-sm font-medium text-muted-foreground mb-3">Available Templates</h3>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {templates.map((template) => {
                const Icon = typeIcons[template.type] || Package;
                return (
                  <Card key={template.id} className="cursor-pointer hover:border-primary transition-colors" onClick={() => handleOpenDialog(undefined, template)}>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-lg flex items-center gap-2">
                        <Icon className="w-5 h-5" />
                        {template.name}
                      </CardTitle>
                      <CardDescription>{template.description}</CardDescription>
                    </CardHeader>
                  </Card>
                );
              })}
            </div>
          </div>

          <div>
            <h3 className="text-sm font-medium text-muted-foreground mb-3">Configured Collectors</h3>
            <div className="border rounded-md bg-card">
              <Table>
                <TableHeader>
                  <TableRow className="bg-muted/5">
                    <TableHead className="w-[50px]">ID</TableHead>
                    <TableHead>Name</TableHead>
                    <TableHead>Type</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {configs.length > 0 ? (
                    configs.map((config) => {
                      const id = getConfigID(config);
                      const Icon = typeIcons[config.type] || Package;
                      return (
                        <TableRow key={id}>
                          <TableCell className="font-mono text-xs text-muted-foreground">#{id}</TableCell>
                          <TableCell className="font-medium flex items-center gap-2">
                            <Icon className="w-4 h-4 text-muted-foreground" />
                            {config.name}
                          </TableCell>
                          <TableCell>
                            <Badge variant="secondary" className="capitalize">{config.type}</Badge>
                          </TableCell>
                          <TableCell>
                            {config.build_status === "completed" ? (
                              <Badge className="bg-green-500/10 text-green-500 border-green-500/20">Ready</Badge>
                            ) : config.build_status === "building" ? (
                              <Badge className="bg-blue-500/10 text-blue-500 border-blue-500/20">Compiling...</Badge>
                            ) : (
                              <Badge variant="outline">Pending</Badge>
                            )}
                          </TableCell>
                          <TableCell className="text-right space-x-1">
                            <Button variant="outline" size="sm" onClick={() => handleBuild(config)}>
                              <Download className="w-3.5 h-3.5 mr-2" /> Build & Download
                            </Button>
                            <Button variant="ghost" size="icon" onClick={() => handleOpenDialog(config)}>
                              <Edit className="w-3.5 h-3.5" />
                            </Button>
                            <Button variant="ghost" size="icon" className="text-red-500 hover:text-red-600" onClick={() => handleDelete(id)}>
                              <Trash2 className="w-3.5 h-3.5" />
                            </Button>
                          </TableCell>
                        </TableRow>
                      );
                    })
                  ) : (
                    !loading && (
                      <TableRow>
                        <TableCell colSpan={5} className="h-32 text-center text-muted-foreground">
                          No collectors configured.
                        </TableCell>
                      </TableRow>
                    )
                  )}
                </TableBody>
              </Table>
            </div>
          </div>
        </TabsContent>

        {/* Type-specific views */}
        {["windows", "linux", "macos"].map(osType => (
          <TabsContent key={osType} value={osType} className="flex-1 mt-0">
            <div className="border rounded-md bg-card">
              <Table>
                <TableHeader>
                  <TableRow className="bg-muted/5">
                    <TableHead className="w-[50px]">ID</TableHead>
                    <TableHead>Name</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {configs.filter(c => c.type === osType).map((config) => {
                    const id = getConfigID(config);
                    return (
                      <TableRow key={id}>
                        <TableCell className="font-mono text-xs text-muted-foreground">#{id}</TableCell>
                        <TableCell className="font-medium">{config.name}</TableCell>
                        <TableCell>
                          {config.build_status === "completed" ? (
                            <Badge className="bg-green-500/10 text-green-500 border-green-500/20">
                              <CheckCircle className="w-3 h-3 mr-1"/> Ready
                            </Badge>
                          ) : (
                            <Badge variant="outline">Pending</Badge>
                          )}
                        </TableCell>
                        <TableCell className="text-right space-x-1">
                          <Button variant="ghost" size="icon" onClick={() => handleBuild(config)}>
                            <Download className="w-3.5 h-3.5" />
                          </Button>
                          <Button variant="ghost" size="icon" onClick={() => handleOpenDialog(config)}>
                            <Edit className="w-3.5 h-3.5" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    );
                  })}
                  {configs.filter(c => c.type === osType).length === 0 && (
                    <TableRow>
                      <TableCell colSpan={4} className="h-32 text-center text-muted-foreground">
                        No {osType} collectors configured.
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
          </TabsContent>
        ))}
      </Tabs>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-[650px]">
          <DialogHeader>
            <DialogTitle>
              {editingConfig ? "Edit Collector" : "Configure Collector"}
            </DialogTitle>
            <DialogDescription>
              Deploy a zero-dependency agent that normalizes logs into OCSF schema.
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4 max-h-[70vh] overflow-y-auto px-1 pr-3">
            <div className="grid gap-2">
              <Label>Probe Name</Label>
              <Input 
                value={formData.name} 
                onChange={e => setFormData({...formData, name: e.target.value})} 
                placeholder="e.g. Domain Controller Security Probe"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label>Target OS</Label>
                <Select value={formData.type} onValueChange={v => setFormData({...formData, type: v, sources: ""})}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="windows">Windows</SelectItem>
                    <SelectItem value="linux">Linux</SelectItem>
                    <SelectItem value="macos">macOS</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              
              <div className="grid gap-2">
                <Label>Target Ingest Node</Label>
                <Select value={String(formData.ingest_id)} onValueChange={v => setFormData({...formData, ingest_id: parseInt(v)})}>
                  <SelectTrigger><SelectValue placeholder="Select Receiver..." /></SelectTrigger>
                  <SelectContent>
                    {ingests.map(ing => (
                      <SelectItem key={ing.ID || ing.id} value={String(ing.ID || ing.id)}>{ing.name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>

            {/* 动态渲染的 Sources 列表与高级过滤面板 */}
            {availableSources.length > 0 && (
              <div className="border rounded-md p-4 bg-muted/10">
                <div className="flex items-center justify-between mb-3">
                  <Label className="text-sm font-semibold">Telemetry Sources & Filtering</Label>
                  <Badge variant="outline" className="text-[10px] font-mono bg-background">
                    {availableSources.filter(s => s.enabled).length} selected
                  </Badge>
                </div>
                
                <div className="grid gap-3 max-h-[300px] overflow-y-auto pr-2 custom-scrollbar">
                  {availableSources.map(source => (
                    <div 
                      key={source.type || source.path} 
                      className={`border rounded-md p-3 transition-colors ${source.enabled ? 'bg-card border-primary/50 shadow-sm' : 'bg-background hover:bg-muted/50'}`}
                    >
                      <div className="flex items-start gap-3">
                        <Checkbox 
                          id={`source-${source.type}`}
                          checked={source.enabled}
                          onCheckedChange={() => toggleSource(source.type)}
                          className="mt-1 data-[state=checked]:bg-primary"
                        />
                        <div className="grid gap-1 flex-1 cursor-pointer" onClick={() => toggleSource(source.type)}>
                          <Label className="text-sm font-medium cursor-pointer">
                            {source.label}
                          </Label>
                          <p className="text-xs text-muted-foreground font-mono">
                            {source.path}
                          </p>
                        </div>
                      </div>

                      {/* Windows 专属：高级过滤与场景预设 */}
                      {source.enabled && formData.type === "windows" && (
                        <div className="mt-3 ml-7 pl-3 border-l-2 border-primary/20 space-y-3 animate-in slide-in-from-top-1 fade-in duration-200">
                          <div className="grid gap-1.5">
                            <Label className="text-[11px] font-semibold text-muted-foreground flex items-center gap-1">
                              <Filter className="w-3 h-3" />
                              Target Event IDs (Optional)
                            </Label>
                            <Input 
                              value={source.event_ids_str || ""}
                              onChange={e => updateSourceConfig(source.type, 'event_ids_str', e.target.value)}
                              placeholder="e.g. 4624, 4625, 4688 (Leave empty for ALL)" 
                              className="h-8 text-xs font-mono bg-background"
                            />
                            
                            {/* 动态渲染场景快捷标签 */}
                            {source.presets && source.presets.length > 0 && (
                              <div className="flex flex-wrap gap-1.5 mt-1.5">
                                {source.presets.map(preset => (
                                  <Badge 
                                    key={preset.name}
                                    variant="secondary" 
                                    className="text-[10px] cursor-pointer hover:bg-primary hover:text-primary-foreground transition-colors"
                                    onClick={() => handlePresetClick(source, preset.ids)}
                                  >
                                    + {preset.name}
                                  </Badge>
                                ))}
                              </div>
                            )}
                            
                            <p className="text-[10px] text-muted-foreground leading-tight mt-1">
                              Comma-separated list. Greatly reduces CPU usage by dropping unwanted events at the OS level.
                            </p>
                          </div>
                          
                          <div className="grid gap-1.5">
                            <Label className="text-[11px] font-semibold text-muted-foreground flex items-center gap-1">
                              <Server className="w-3 h-3" />
                              Raw XPath Query (Advanced)
                            </Label>
                            <Input 
                              value={source.query || ""}
                              onChange={e => updateSourceConfig(source.type, 'query', e.target.value)}
                              placeholder="e.g. EventID=4624 and EventData/Data[@Name='TargetUserName']='admin'" 
                              className="h-8 text-xs font-mono bg-background"
                            />
                          </div>
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}

            <div className="grid gap-2">
              <Label>OCSF Stream Fields</Label>
              <Input 
                value={formData.stream_fields} 
                onChange={e => setFormData({...formData, stream_fields: e.target.value})} 
                className="font-mono text-xs bg-muted/20"
                readOnly
              />
              <p className="text-[11px] text-muted-foreground mt-0.5">
                Hardcoded for optimal time-series indexing in VictoriaLogs.
              </p>
            </div>
          </div>

          <DialogFooter className="mt-2">
            <Button variant="outline" onClick={() => setDialogOpen(false)}>Cancel</Button>
            <Button onClick={handleSubmit} disabled={submitting}>
              {submitting ? "Processing..." : editingConfig ? "Save Configuration" : "Create & Compile"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
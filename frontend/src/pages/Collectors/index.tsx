import { useEffect, useState } from "react";
import { collectorService, ingestServiceSimple, type CollectorConfig, type CollectorTemplate, type IngestConfig } from "@/services/collector";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Plus, Trash2, Edit, Download, Server, Settings, Package, Activity, CheckCircle } from "lucide-react";
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

// Data source interface for frontend
interface DataSource {
  type: string;
  path: string;
  label: string;
  enabled: boolean;
}

export default function CollectorsPage() {
  const [configs, setConfigs] = useState<CollectorConfig[]>([]);
  const [templates, setTemplates] = useState<CollectorTemplate[]>([]);
  const [ingests, setIngests] = useState<IngestConfig[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingConfig, setEditingConfig] = useState<CollectorConfig | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [availableChannels, setAvailableChannels] = useState<string[]>([]);
  const [availableSources, setAvailableSources] = useState<DataSource[]>([]);
  const [activeTab, setActiveTab] = useState<string>("templates");

  // 【核心修改】：默认 stream_fields 适配 OCSF 规范
  const defaultStreamFields = "observer.hostname,observer.vendor,class_uid";

  const [formData, setFormData] = useState({
    name: "",
    type: "windows",
    channels: "",
    sources: "", // JSON string of sources for all OS
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

  // 【核心修改】：带回显状态的 Sources 抓取逻辑
  const fetchSourcesForType = async (type: string, savedSourcesStr?: string) => {
    try {
      const res = await collectorService.getSources(type);
      if (res.code === 200 && res.data) {
        const data = res.data;
        if (Array.isArray(data) && data.length > 0) {
          if (typeof data[0] === 'string') {
            setAvailableChannels(data as string[]);
            setAvailableSources([]);
          } else {
            // 解析已保存的 JSON 以恢复 Checkbox 选中状态
            const savedSourcesMap = new Map();
            if (savedSourcesStr) {
              try {
                const parsed = JSON.parse(savedSourcesStr);
                parsed.forEach((s: any) => savedSourcesMap.set(s.type || s.path, s.enabled));
              } catch (e) {}
            }

            const sources = (data as any[]).map((item: any) => ({
              type: item.type || item.Type || '',
              path: item.path || item.Path || '',
              label: item.label || item.Label || item.type || '',
              enabled: savedSourcesMap.has(item.type || item.path) ? savedSourcesMap.get(item.type || item.path) : false,
            }));
            setAvailableSources(sources);
            setAvailableChannels([]);
          }
        }
      }
    } catch (err) {
      console.error("Failed to fetch sources:", err);
    }
  };

  // 监听 OS 类型切换，重新拉取对应平台的支持列表
  useEffect(() => {
    if (formData.type) {
      fetchSourcesForType(formData.type, formData.sources);
    }
  }, [formData.type]);

  // 切换复选框逻辑，并实时同步为 JSON string
  const toggleSource = (type: string) => {
    const updated = availableSources.map(s => {
      if (s.type === type) {
        return { ...s, enabled: !s.enabled };
      }
      return s;
    });
    setAvailableSources(updated);
    
    const enabledSources = updated.filter(s => s.enabled);
    setFormData({ ...formData, sources: JSON.stringify(enabledSources) });
  };

  const getConfigID = (c: CollectorConfig) => c.ID || c.id || 0;

  const handleOpenDialog = (config?: CollectorConfig, template?: CollectorTemplate) => {
    if (config) {
      setEditingConfig(config);
      setFormData({
        name: config.name,
        type: config.type,
        channels: config.channels || "",
        sources: config.sources || "", // 载入历史 sources
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
        channels: template.channels?.join(",") || "",
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

  const handleIngestChange = async (ingestId: string) => {
    const id = parseInt(ingestId);
    setFormData({...formData, ingest_id: id});
    
    // We don't necessarily need to fetch the token to show it here since 
    // the backend will fetch it dynamically during the build process, 
    // but fetching it for display is fine.
  };

  // 【核心修改】：抛弃 ZIP 解压，直接下发单体二进制文件
  const handleBuild = async (config: CollectorConfig) => {
    try {
      toast.info("Compiling standalone collector binary...", {
        description: "This takes 1-2 seconds with our Cloud-Native compiler."
      });
      
      const res = await collectorService.build(getConfigID(config));
      
      const url = window.URL.createObjectURL(res.data);
      const a = document.createElement('a');
      a.href = url;
      
      // 根据目标系统智能分配后缀名
      const safeName = config.name.replace(/\s+/g, '_').toLowerCase();
      const ext = config.type === "windows" ? ".exe" : "";
      a.download = `vsentry-agent-${safeName}${ext}`;
      
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
      
      toast.success("Collector compiled and downloaded!", {
        description: "Zero dependencies. Just drop it onto your server and run."
      });
      fetchData();
    } catch (err) {
      console.error(err);
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
          {/* Templates Grid */}
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
                    <CardContent>
                      <div className="flex flex-wrap gap-1">
                        {template.channels?.slice(0, 4).map((ch: string) => (
                          <Badge key={ch} variant="secondary" className="text-xs">{ch}</Badge>
                        ))}
                        {template.channels?.length > 4 && (
                          <Badge variant="outline" className="text-xs">+{template.channels.length - 4}</Badge>
                        )}
                      </div>
                    </CardContent>
                  </Card>
                );
              })}
            </div>
          </div>

          {/* Configured Collectors */}
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
                              <Badge className="bg-green-500/10 text-green-500 border-green-500/20">Ready to Deploy</Badge>
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
                            <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => handleOpenDialog(config)}>
                              <Edit className="w-3.5 h-3.5" />
                            </Button>
                            <Button variant="ghost" size="icon" className="h-8 w-8 text-red-500 hover:text-red-600" onClick={() => handleDelete(id)}>
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
                          No collectors configured. Click a template to start.
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
                          <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => handleBuild(config)}>
                            <Download className="w-3.5 h-3.5" />
                          </Button>
                          <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => handleOpenDialog(config)}>
                            <Edit className="w-3.5 h-3.5" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    );
                  })}
                  {configs.filter(c => c.type === osType).length === 0 && (
                    <TableRow>
                      <TableCell colSpan={3} className="h-32 text-center text-muted-foreground">
                        No {osType} collectors. Click Templates to create one.
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
          </TabsContent>
        ))}
      </Tabs>

      {/* Configuration Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-[550px]">
          <DialogHeader>
            <DialogTitle>
              {editingConfig ? "Edit Collector" : "Configure Collector"}
            </DialogTitle>
            <DialogDescription>
              Deploy a zero-dependency agent that normalizes logs into OCSF standard schema.
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4 max-h-[60vh] overflow-y-auto px-1">
            <div className="grid gap-2">
              <Label>Probe Name</Label>
              <Input 
                value={formData.name} 
                onChange={e => setFormData({...formData, name: e.target.value})} 
                placeholder="e.g. DMZ Web Server Probe"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label>Target OS</Label>
                <Select value={formData.type} onValueChange={v => setFormData({...formData, type: v, sources: "", channels: ""})}>
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
                <Select value={String(formData.ingest_id)} onValueChange={handleIngestChange}>
                  <SelectTrigger><SelectValue placeholder="Select Receiver..." /></SelectTrigger>
                  <SelectContent>
                    {ingests.map(ing => (
                      <SelectItem key={ing.ID || ing.id} value={String(ing.ID || ing.id)}>{ing.name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>

            {/* 【核心修改】：全平台统一的结构化 Sources JSON 选择视图 */}
            {availableSources.length > 0 ? (
              <div className="border rounded-md p-4 bg-muted/20">
                <Label className="mb-3 block text-sm font-semibold">Telemetry Sources</Label>
                <div className="grid grid-cols-2 gap-3 max-h-[220px] overflow-y-auto pr-2">
                  {availableSources.map(source => (
                    <div key={source.type || source.path} className="flex items-start gap-2">
                      <Checkbox 
                        id={`source-${source.type}`}
                        checked={source.enabled}
                        onCheckedChange={() => toggleSource(source.type)}
                        className="mt-1"
                      />
                      <div className="grid gap-1.5 leading-none">
                        <Label 
                          htmlFor={`source-${source.type}`} 
                          className="text-sm cursor-pointer font-medium"
                        >
                          {source.label}
                        </Label>
                        <p className="text-xs text-muted-foreground line-clamp-1 break-all" title={source.path}>
                          {source.path}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ) : (
              // 容错回退机制：当后端没有预设 Source List 时使用老旧输入框
              <div className="grid gap-2">
                <Label>Event Channels (Legacy Input)</Label>
                <Input 
                  value={formData.channels} 
                  onChange={e => setFormData({...formData, channels: e.target.value})}
                  placeholder="Security,System,Application"
                />
              </div>
            )}

            <div className="grid gap-2">
              <Label>OCSF Stream Fields (VictoriaLogs Indexing)</Label>
              <Input 
                value={formData.stream_fields} 
                onChange={e => setFormData({...formData, stream_fields: e.target.value})} 
                className="font-mono text-xs"
              />
              <p className="text-xs text-muted-foreground mt-1">
                Optimizes time-series indexing. Do not change unless you understand VictoriaLogs architecture.
              </p>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>Cancel</Button>
            <Button onClick={handleSubmit} disabled={submitting}>
              {submitting ? "Processing..." : editingConfig ? "Save Configuration" : "Create & Initialize"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
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

  const [formData, setFormData] = useState({
    name: "",
    type: "windows",
    channels: "",
    sources: "", // JSON string of sources for Linux
    ingest_id: 0,
    stream_fields: "channel,source,host",
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

  // Fetch available sources when type changes
  useEffect(() => {
    const fetchSources = async () => {
      try {
        const res = await collectorService.getSources(formData.type);
        if (res.code === 200 && res.data) {
          // Handle both array of strings and array of objects
          const data = res.data;
          if (Array.isArray(data) && data.length > 0) {
            if (typeof data[0] === 'string') {
              // Legacy format: array of strings
              setAvailableChannels(data as string[]);
              setAvailableSources([]);
            } else {
              // New format: array of {type, path, label}
              const sources = (data as any[]).map((item: any) => ({
                type: item.type || item.Type || '',
                path: item.path || item.Path || '',
                label: item.label || item.Label || item.type || '',
                enabled: false,
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
    
    if (formData.type) {
      fetchSources();
    }
  }, [formData.type]);

  // Toggle source selection
  const toggleSource = (type: string) => {
    const updated = availableSources.map(s => {
      if (s.type === type) {
        return { ...s, enabled: !s.enabled };
      }
      return s;
    });
    setAvailableSources(updated);
    
    // Update formData.sources as JSON
    const enabledSources = updated.filter(s => s.enabled);
    setFormData({ ...formData, sources: JSON.stringify(enabledSources) });
  };

  useEffect(() => {
    // Fetch available channels when type changes
    collectorService.channels(formData.type).then(res => {
      if (res.code === 200) {
        setAvailableChannels(res.data || []);
      }
    });
  }, [formData.type]);

  const getConfigID = (c: CollectorConfig) => c.ID || c.id || 0;

  const handleOpenDialog = (config?: CollectorConfig, template?: CollectorTemplate) => {
    if (config) {
      setEditingConfig(config);
      setFormData({
        name: config.name,
        type: config.type,
        channels: config.channels,
        ingest_id: config.ingest_id || 0,
        stream_fields: config.stream_fields || "channel,source,host",
      });
    } else if (template) {
      setEditingConfig(null);
      setFormData({
        name: template.name,
        type: template.type,
        channels: template.channels?.join(",") || "",
        ingest_id: 0,
        stream_fields: "channel,source,host",
      });
    } else {
      setEditingConfig(null);
      setFormData({
        name: "",
        type: "windows",
        channels: "",
        ingest_id: 0,
        stream_fields: "channel,source,host",
      });
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
    setFormData({...formData, ingest_id: id, token: ""});
    
    if (id && id > 0) {
      try {
        const res = await collectorService.ingestAuth(id);
        if (res.code === 200 && res.data?.token) {
          setFormData(prev => ({...prev, token: res.data.token}));
        }
      } catch (e) {
        console.error("Failed to fetch token", e);
      }
    }
  };

  const handleBuild = async (config: CollectorConfig) => {
    try {
      toast.info("Building collector package...");
      const res = await collectorService.build(getConfigID(config));
      
      // Create download link
      const url = window.URL.createObjectURL(res.data);
      const a = document.createElement('a');
      a.href = url;
      a.download = `collector_${config.name.replace(/\s+/g, '_')}.zip`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
      
      toast.success("Collector package downloaded!");
      fetchData();
    } catch (err) {
      console.error(err);
      toast.error("Build failed");
    }
  };

  const filteredTemplates = activeTab === "all" 
    ? templates 
    : templates.filter(t => t.type === activeTab);

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
              Build log collectors for different operating systems.
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
                    <TableHead>Channels</TableHead>
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
                          <TableCell className="font-mono text-xs text-muted-foreground max-w-[200px] truncate">
                            {config.channels}
                          </TableCell>
                          <TableCell>
                            {config.build_status === "completed" ? (
                              <Badge className="bg-green-500">Ready</Badge>
                            ) : config.build_status === "building" ? (
                              <Badge className="bg-blue-500">Building</Badge>
                            ) : (
                              <Badge variant="secondary">Pending</Badge>
                            )}
                          </TableCell>
                          <TableCell className="text-right space-x-1">
                            <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => handleBuild(config)}>
                              <Download className="w-3.5 h-3.5" />
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
                        <TableCell colSpan={6} className="h-32 text-center text-muted-foreground">
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
                    <TableHead>Channels</TableHead>
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
                        <TableCell className="font-mono text-xs">{config.channels}</TableCell>
                        <TableCell>
                          {config.build_status === "completed" ? (
                            <Badge className="bg-green-500"><CheckCircle className="w-3 h-3 mr-1"/> Ready</Badge>
                          ) : (
                            <Badge variant="secondary">Pending</Badge>
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
                      <TableCell colSpan={4} className="h-32 text-center text-muted-foreground">
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

      {/* Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>
              {editingConfig ? "Edit Collector" : "Configure Collector"}
            </DialogTitle>
            <DialogDescription>
              Configure the collector to select source and destination.
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4 max-h-[60vh] overflow-y-auto">
            <div className="grid gap-2">
              <Label>Name</Label>
              <Input value={formData.name} onChange={e => setFormData({...formData, name: e.target.value})} />
            </div>

            <div className="grid gap-2">
              <Label>Type</Label>
              <Select value={formData.type} onValueChange={v => setFormData({...formData, type: v})}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="windows">Windows</SelectItem>
                  <SelectItem value="linux">Linux</SelectItem>
                  <SelectItem value="macos">macOS</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="grid gap-2">
              {/* For Linux: Show source checkboxes */}
              {formData.type === "linux" && availableSources.length > 0 && (
                <div className="border rounded-md p-3 max-h-[200px] overflow-y-auto">
                  <Label className="mb-2 block">Select Data Sources to Collect:</Label>
                  <div className="grid grid-cols-2 gap-2">
                    {availableSources.map(source => (
                      <div key={source.type} className="flex items-center gap-2">
                        <Checkbox 
                          id={`source-${source.type}`}
                          checked={source.enabled}
                          onCheckedChange={() => toggleSource(source.type)}
                        />
                        <Label 
                          htmlFor={`source-${source.type}`} 
                          className="text-sm cursor-pointer font-normal"
                          title={source.path}
                        >
                          {source.label}
                        </Label>
                      </div>
                    ))}
                  </div>
                  <p className="text-xs text-muted-foreground mt-2">
                    Selected: {availableSources.filter(s => s.enabled).map(s => s.label).join(", ") || "none"}
                  </p>
                </div>
              )}

              {/* For Windows: Show old channel input */}
              {(formData.type === "windows" || availableSources.length === 0) && (
                <>
                  <Label>Channels (comma-separated)</Label>
                  <Input 
                    value={formData.channels} 
                    onChange={e => setFormData({...formData, channels: e.target.value})}
                    placeholder={availableChannels.slice(0, 3).join(", ") || "System,Application,Security"}
                  />
                  <div className="flex flex-wrap gap-1 mt-1">
                    {availableChannels.slice(0, 8).map(ch => (
                      <Badge key={ch} variant="outline" className="text-xs cursor-pointer" onClick={() => {
                        const current = formData.channels ? formData.channels.split(",").map(s => s.trim()) : [];
                        if (!current.includes(ch)) {
                          setFormData({...formData, channels: [...current, ch].join(",")});
                        }
                      }}>{ch}</Badge>
                    ))}
                  </div>
                </>
              )}

              {/* For macOS: Simple placeholder */}
              {formData.type === "macos" && (
                <div className="text-sm text-muted-foreground p-3 bg-muted rounded">
                  macOS collector is coming soon. Use the Windows or Linux option for now.
                </div>
              )}
            </div>

            <div className="grid gap-2">
              <Label>Ingest (for token)</Label>
              <Select value={String(formData.ingest_id)} onValueChange={handleIngestChange}>
                <SelectTrigger><SelectValue placeholder="Select Ingest" /></SelectTrigger>
                <SelectContent>
                  {ingests.map(ing => (
                    <SelectItem key={ing.ID || ing.id} value={String(ing.ID || ing.id)}>{ing.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {formData.token && (
                <p className="text-xs text-muted-foreground">
                  Token: {formData.token.substring(0, 16)}...
                </p>
              )}
            </div>

            <div className="grid gap-2">
              <Label>Stream Fields</Label>
              <Input value={formData.stream_fields} onChange={e => setFormData({...formData, stream_fields: e.target.value})} />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>Cancel</Button>
            <Button onClick={handleSubmit} disabled={submitting}>
              {submitting ? "Saving..." : editingConfig ? "Update" : "Create"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
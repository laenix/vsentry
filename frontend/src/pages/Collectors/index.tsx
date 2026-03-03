import { useEffect, useState } from "react";
import { collectorService, ingestServiceSimple, type CollectorConfig, type CollectorTemplate, type IngestConfig } from "@/services/collector";
import { Button } from "@/components/ui/button";
import { Card, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Plus, Server } from "lucide-react";
import { toast } from "sonner";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { CollectorTable } from "./CollectorTable";
import { CollectorDialog } from "./CollectorDialog";
import { typeIcons } from "./constants";
import type { DataSource } from "./constants";

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
    name: "", type: "windows", channels: "", sources: "", ingest_id: 0, stream_fields: defaultStreamFields, interval: 5,
  });

  const fetchData = async () => {
    try {
      const [configRes, tmplRes, ingestRes] = await Promise.all([
        collectorService.list(),
        collectorService.templates(),
        ingestServiceSimple.list(),
      ]);
      if (configRes.code === 200) setConfigs(Array.isArray(configRes.data) ? configRes.data : []);
      if (tmplRes.code === 200) setTemplates(tmplRes.data || []);
      if (ingestRes.code === 200) setIngests(Array.isArray(ingestRes.data) ? ingestRes.data : []);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchData(); }, []);

  const fetchSourcesForType = async (type: string, savedSourcesStr?: string) => {
    try {
      const res = await collectorService.getSources(type);
      if (res.code === 200 && res.data && Array.isArray(res.data) && res.data.length > 0) {
        const savedMap = new Map();
        if (savedSourcesStr) {
          try {
            JSON.parse(savedSourcesStr).forEach((s: any) => {
              // 核心修改：回显时也要把保存的自定义路径带回来
              savedMap.set(s.type || s.path, { 
                enabled: s.enabled, 
                event_ids_str: s.event_ids?.join(", ") || "", 
                query: s.query || "",
                path: s.path // 保存被用户修改过的路径
              });
            });
          } catch (e) {}
        }

        const sources: DataSource[] = res.data.map((item: any) => {
          const key = item.type || item.Path || '';
          const savedState = savedMap.get(key) || { enabled: false, event_ids_str: "", query: "", path: "" };
          return {
            type: item.type || item.Type || '',
            // 优先使用已保存的路径，如果没有则使用后端默认的模板路径
            path: savedState.path || item.path || item.Path || '',
            label: item.label || item.Label || item.type || '',
            enabled: savedState.enabled,
            event_ids_str: savedState.event_ids_str,
            query: savedState.query,
            presets: item.presets || [], 
          };
        });
        setAvailableSources(sources);
      }
    } catch (err) {
      console.error("Failed to fetch sources:", err);
    }
  };

  useEffect(() => {
    if (formData.type) fetchSourcesForType(formData.type, formData.sources);
  }, [formData.type]);

  const syncSourcesToFormData = (sources: DataSource[]) => {
    const enabledSources = sources.filter(s => s.enabled).map(s => {
      let ids: number[] = [];
      if (s.event_ids_str && s.event_ids_str.trim() !== "") {
        ids = s.event_ids_str.split(',').map(id => parseInt(id.trim())).filter(id => !isNaN(id));
      }
      return {
        type: s.type, 
        path: s.path, // 这里会把用户修改后的新路径同步进 JSON
        format: formData.type === "windows" ? "windows_event" : "file",
        enabled: true, 
        event_ids: ids.length > 0 ? ids : undefined,
        query: s.query && s.query.trim() !== "" ? s.query.trim() : undefined
      };
    });
    setFormData(prev => ({ ...prev, sources: JSON.stringify(enabledSources) }));
  };

  const toggleSource = (type: string) => {
    const updated = availableSources.map(s => s.type === type ? { ...s, enabled: !s.enabled } : s);
    setAvailableSources(updated);
    syncSourcesToFormData(updated);
  };

  const updateSourceConfig = (type: string, field: 'event_ids_str' | 'query', value: string) => {
    const updated = availableSources.map(s => s.type === type ? { ...s, [field]: value } : s);
    setAvailableSources(updated);
    syncSourcesToFormData(updated);
  };

  // 【新增功能】：更新路径
  const updateSourcePath = (type: string, newPath: string) => {
    const updated = availableSources.map(s => s.type === type ? { ...s, path: newPath } : s);
    setAvailableSources(updated);
    syncSourcesToFormData(updated);
  };

  const handlePresetClick = (source: DataSource, presetIds: string) => {
    const current = source.event_ids_str ? source.event_ids_str.split(',').map(s => s.trim()).filter(s => s) : [];
    const newIds = presetIds.split(',').map(s => s.trim());
    const merged = Array.from(new Set([...current, ...newIds]));
    updateSourceConfig(source.type, 'event_ids_str', merged.join(', '));
  };

  const getConfigID = (c: CollectorConfig) => c.ID || c.id || 0;

  const handleOpenDialog = (config?: CollectorConfig, template?: CollectorTemplate) => {
    if (config) {
      setEditingConfig(config);
      setFormData({
        name: config.name, type: config.type, channels: config.channels || "", sources: config.sources || "",
        ingest_id: config.ingest_id || 0, stream_fields: config.stream_fields || defaultStreamFields, interval: config.interval || 5,
      });
      fetchSourcesForType(config.type, config.sources);
    } else {
      const type = template ? template.type : "windows";
      setEditingConfig(null);
      setFormData({
        name: template ? template.name : "", type, channels: "", sources: "",
        ingest_id: 0, stream_fields: defaultStreamFields, interval: 5,
      });
      fetchSourcesForType(type);
    }
    setDialogOpen(true);
  };

  const handleSubmit = async () => {
    if (!formData.name || !formData.type) return toast.error("Name and Type are required");
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
    } catch (err) { console.error(err); }
  };

  const handleBuild = async (config: CollectorConfig) => {
    try {
      toast.info(`Triggering build for ${config.name}...`, { description: "Executing Cloud-Native compilation..." });
      setConfigs(configs.map(c => getConfigID(c) === getConfigID(config) ? { ...c, build_status: "building" } : c));
      await collectorService.build(getConfigID(config));
      toast.success("Compilation successful!", { description: "Binary is ready for download." });
      fetchData();
    } catch (err) {
      toast.error("Compilation failed", { description: "Check server logs for details." });
      fetchData();
    }
  };

  const handleDownload = async (config: CollectorConfig) => {
    try {
      const safeName = config.name.replace(/\s+/g, '_').toLowerCase();
      const ext = config.type === "windows" ? ".exe" : "";
      toast.info(`Downloading vsentry-agent-${safeName}${ext}...`);
      await collectorService.download(getConfigID(config), `vsentry-agent-${safeName}${ext}`);
      toast.success("Download complete");
    } catch (err: any) {
      toast.error("Download failed", { description: err.message });
      fetchData();
    }
  };

  return (
    <div className="p-6 h-full flex flex-col">
      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col">
        <div className="flex justify-between items-center mb-4">
          <div>
            <h1 className="text-2xl font-bold flex items-center gap-2">
              <Server className="w-6 h-6" /> Collectors
            </h1>
            <p className="text-muted-foreground text-sm">
              Deploy native, zero-dependency XDR agents tailored for OCSF.
            </p>
          </div>
          <Button onClick={() => handleOpenDialog()}>
            <Plus className="w-4 h-4 mr-2" /> New Collector
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
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              {templates.map((template) => {
                const Icon = typeIcons[template.type] || typeIcons.windows;
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
            <CollectorTable 
              configs={configs} loading={loading} 
              onBuild={handleBuild} onDownload={handleDownload} onEdit={handleOpenDialog} onDelete={handleDelete} 
            />
          </div>
        </TabsContent>

        {["windows", "linux", "macos"].map(osType => (
          <TabsContent key={osType} value={osType} className="flex-1 mt-0">
            <CollectorTable 
              configs={configs.filter(c => c.type === osType)} loading={loading} emptyMessage={`No ${osType} collectors configured.`}
              onBuild={handleBuild} onDownload={handleDownload} onEdit={handleOpenDialog} onDelete={handleDelete} 
            />
          </TabsContent>
        ))}
      </Tabs>

      <CollectorDialog 
        open={dialogOpen} onOpenChange={setDialogOpen} editingConfig={editingConfig}
        formData={formData} setFormData={setFormData} ingests={ingests} availableSources={availableSources}
        onToggleSource={toggleSource} onUpdateSourceConfig={updateSourceConfig} onPresetClick={handlePresetClick}
        onUpdateSourcePath={updateSourcePath} // 将新函数传入子组件
        onSubmit={handleSubmit} submitting={submitting}
      />
    </div>
  );
}
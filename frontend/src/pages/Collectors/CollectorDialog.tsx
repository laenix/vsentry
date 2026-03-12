import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Filter, Server } from "lucide-react";
import type { IngestConfig, CollectorConfig } from "@/services/collector";
import type { DataSource } from "./constants";

interface CollectorDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  editingConfig: CollectorConfig | null;
  formData: any;
  setFormData: (data: any) => void;
  ingests: IngestConfig[];
  availableSources: DataSource[];
  onToggleSource: (type: string) => void;
  onUpdateSourceConfig: (type: string, field: 'event_ids_str' | 'query', value: string) => void;
  onPresetClick: (source: DataSource, presetIds: string) => void;
  onUpdateSourcePath: (type: string, newPath: string) => void; // New增Props
  onSubmit: () => void;
  submitting: boolean;
}

export function CollectorDialog({
  open, onOpenChange, editingConfig, formData, setFormData, ingests,
  availableSources, onToggleSource, onUpdateSourceConfig, onPresetClick, onUpdateSourcePath, onSubmit, submitting
}: CollectorDialogProps) {
  
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[650px]">
        <DialogHeader>
          <DialogTitle>{editingConfig ? "Edit Collector" : "Configure Collector"}</DialogTitle>
          <DialogDescription>Deploy a zero-dependency agent that normalizes logs into OCSF schema.</DialogDescription>
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
                  <div key={source.type || source.path} className={`border rounded-md p-3 transition-colors ${source.enabled ? 'bg-card border-primary/50 shadow-sm' : 'bg-background hover:bg-muted/50'}`}>
                    <div className="flex items-start gap-3">
                      <Checkbox 
                        id={`source-${source.type}`}
                        checked={source.enabled}
                        onCheckedChange={() => onToggleSource(source.type)}
                        className="mt-1 data-[state=checked]:bg-primary"
                      />
                      <div className="grid gap-1 flex-1">
                        <Label className="text-sm font-medium cursor-pointer" onClick={() => onToggleSource(source.type)}>
                          {source.label}
                        </Label>
                        
                        {/* 【核心UI变身逻辑】：如果勾选了，且不是 Windows EventLog，则显示为可EditInput框 */}
                        {source.enabled && formData.type !== "windows" ? (
                          <Input 
                            value={source.path}
                            onChange={e => onUpdateSourcePath(source.type, e.target.value)}
                            className="h-7 text-xs font-mono mt-1 bg-background w-full"
                            placeholder="Enter absolute file path..."
                          />
                        ) : (
                          <p className="text-xs text-muted-foreground font-mono cursor-pointer" onClick={() => onToggleSource(source.type)}>
                            {source.path}
                          </p>
                        )}
                      </div>
                    </div>

                    {/* Windows 专有的 EventID / XPath FilterPanel */}
                    {source.enabled && formData.type === "windows" && (
                      <div className="mt-3 ml-7 pl-3 border-l-2 border-primary/20 space-y-3 animate-in slide-in-from-top-1 fade-in duration-200">
                        <div className="grid gap-1.5">
                          <Label className="text-[11px] font-semibold text-muted-foreground flex items-center gap-1">
                            <Filter className="w-3 h-3" />
                            Target Event IDs (Optional)
                          </Label>
                          <Input 
                            value={source.event_ids_str || ""}
                            onChange={e => onUpdateSourceConfig(source.type, 'event_ids_str', e.target.value)}
                            placeholder="e.g. 4624, 4625, 4688 (Leave empty for ALL)" 
                            className="h-8 text-xs font-mono bg-background"
                          />
                          
                          {source.presets && source.presets.length > 0 && (
                            <div className="flex flex-wrap gap-1.5 mt-1.5">
                              {source.presets.map(preset => (
                                <Badge key={preset.name} variant="secondary" 
                                  className="text-[10px] cursor-pointer hover:bg-primary hover:text-primary-foreground transition-colors"
                                  onClick={() => onPresetClick(source, preset.ids)}
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
                            onChange={e => onUpdateSourceConfig(source.type, 'query', e.target.value)}
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
          <Button variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
          <Button onClick={onSubmit} disabled={submitting}>
            {submitting ? "Processing..." : editingConfig ? "Save Configuration" : "Create"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
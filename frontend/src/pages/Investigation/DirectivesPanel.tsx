import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Search, Activity, LayoutList, X, ExternalLink, Clock, Server, AlertTriangle } from "lucide-react";
import { type InvestigationDirective } from "@/services/investigation";
import { DirectivesSelectorDialog } from "./DirectivesSelectorDialog";

interface DirectivesPanelProps {
  templates: InvestigationDirective[];
  selectedTemplates: number[];
  onChangeSelection: (ids: number[]) => void;
  contextVars: Record<string, string>;
  loading: boolean;
  onExecute: () => void;
  onRefreshTemplates: () => void;
  // 新增：时间范围和 Alert 选择
  timeRangeHours?: number;
  onTimeRangeChange?: (hours: number) => void;
  onApplyTimeRange?: () => void;
  incidentData?: any;
  selectedAlertIdx?: string;
  onAlertChange?: (val: string) => void;
}

export function DirectivesPanel({ 
  templates, 
  selectedTemplates, 
  onChangeSelection, 
  contextVars, 
  loading, 
  onExecute, 
  onRefreshTemplates,
  timeRangeHours,
  onTimeRangeChange,
  onApplyTimeRange,
  incidentData,
  selectedAlertIdx,
  onAlertChange,
}: DirectivesPanelProps) {
  const [selectorOpen, setSelectorOpen] = useState(false);
  
  // Status：用于触发Refresh
  const [refreshKey, setRefreshKey] = useState(0);

  const selectedList = templates.filter(t => selectedTemplates.includes(t.id));

  const handleRefresh = () => {
    onRefreshTemplates();
    setRefreshKey(k => k + 1);
  };

  return (
    <>
      <Card className="shadow-sm flex-none">
        <CardHeader className="pb-2 border-b">
          <div className="flex justify-between items-center gap-2">
            <div className="flex-1">
              <CardTitle className="text-base flex items-center gap-2">
                <Search className="w-4 h-4 text-primary" />
                Investigation Directives
              </CardTitle>
            </div>
            
            {/* 紧凑的时间范围和 Alert 选择器 */}
            <div className="flex items-center gap-2 flex-wrap">
              {/* Alert 选择器 */}
              {incidentData && incidentData.alerts?.length > 1 && (
                <div className="flex items-center gap-1">
                  <AlertTriangle className="w-3 h-3 text-amber-500" />
                  <Select value={selectedAlertIdx} onValueChange={onAlertChange}>
                    <SelectTrigger className="h-7 text-xs w-24">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {incidentData.alerts.map((al: any, idx: number) => {
                        const t = al.created_at || al.CreatedAt || al._time;
                        return (
                          <SelectItem key={idx} value={String(idx)} className="text-xs font-mono">
                            #{al.id}
                          </SelectItem>
                        )
                      })}
                    </SelectContent>
                  </Select>
                </div>
              )}
              
              {/* 时间范围选择器 */}
              {onTimeRangeChange && (
                <div className="flex items-center gap-1">
                  <Clock className="w-3 h-3 text-muted-foreground" />
                  <Select value={String(timeRangeHours)} onValueChange={(v) => onTimeRangeChange(Number(v))}>
                    <SelectTrigger className="h-7 text-xs w-24">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="0">Unlimited</SelectItem>
                      <SelectItem value="1">±1 hour</SelectItem>
                      <SelectItem value="2">±2 hours</SelectItem>
                      <SelectItem value="6">±6 hours</SelectItem>
                      <SelectItem value="12">±12 hours</SelectItem>
                      <SelectItem value="24">±24 hours</SelectItem>
                      <SelectItem value="72">±3 days</SelectItem>
                      <SelectItem value="168">±7 days</SelectItem>
                    </SelectContent>
                  </Select>
                  <Button size="sm" variant="outline" className="h-7 text-xs px-2" onClick={onApplyTimeRange}>
                    Apply
                  </Button>
                </div>
              )}
            </div>
          </div>
        </CardHeader>
        
        <CardContent className="p-3 bg-muted/30">
          <div className="flex justify-between items-center">
            <div className="flex items-center gap-2">
              <Button variant="outline" size="sm" onClick={() => setSelectorOpen(true)}>
                <LayoutList className="w-4 h-4 mr-2" /> Select
              </Button>
              <Button variant="outline" size="sm" title="Refresh Rules" onClick={handleRefresh}>
                <ExternalLink className="w-4 h-4 mr-1" /> Refresh
              </Button>
              <Button onClick={onExecute} disabled={loading || selectedTemplates.length === 0} className="shadow-sm ml-2">
                {loading ? <Activity className="w-4 h-4 mr-2 animate-spin" /> : <Activity className="w-4 h-4 mr-2" />}
                Execute ({selectedTemplates.length})
              </Button>
            </div>
          </div>
        </CardContent>
        <CardContent className="p-3 pt-0 bg-muted/5">
          {selectedList.length === 0 ? (
            <p className="text-xs text-muted-foreground italic text-center py-2">No directives selected. Click "Select Directives" to begin.</p>
          ) : (
            <div className="flex flex-wrap gap-2">
              {selectedList.map(tpl => (
                <Badge key={tpl.id} variant="outline" className="bg-background border-primary/30 text-primary py-1.5 px-3 flex items-center gap-2">
                  {tpl.name}
                  <X className="w-3 h-3 cursor-pointer hover:text-red-500" onClick={() => onChangeSelection(selectedTemplates.filter(id => id !== tpl.id))} />
                </Badge>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <DirectivesSelectorDialog 
        key={refreshKey}
        open={selectorOpen} 
        onOpenChange={setSelectorOpen} 
        templates={templates} 
        selectedIds={selectedTemplates} 
        onChange={onChangeSelection} 
        contextVars={contextVars} 
      />
    </>
  );
}
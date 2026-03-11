import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Plus, Trash2, Target, Server, FlaskConical } from "lucide-react";

interface ContextPanelProps {
  activeIncidentId: string | null | undefined;
  contextVars: Record<string, string>;
  newVarKey: string;
  newVarValue: string;
  setNewVarKey: (val: string) => void;
  setNewVarValue: (val: string) => void;
  handleAddVar: () => void;
  handleRemoveVar: (key: string) => void;
  // 取证上下文
  forensicsCaseId?: number;
  forensicsFileId?: number;
  forensicsFileName?: string;
}

export function ContextPanel({
  activeIncidentId,
  contextVars,
  newVarKey,
  newVarValue,
  setNewVarKey,
  setNewVarValue,
  handleAddVar,
  handleRemoveVar,
  forensicsCaseId,
  forensicsFileId,
  forensicsFileName
}: ContextPanelProps) {
  return (
    <div className="w-full md:w-80 flex flex-col gap-4 flex-none">
      <Card className="flex-1 shadow-sm border-primary/10 flex flex-col">
        <CardHeader className="pb-3 bg-primary/5 rounded-t-lg border-b border-primary/10">
          <CardTitle className="text-base flex items-center gap-2">
            <Target className="w-4 h-4 text-primary" />
            Investigation Context
          </CardTitle>
          <CardDescription className="text-xs">
            {/* 显示来源：告警 或 取证 */}
            {forensicsCaseId ? (
              <span className="flex items-center gap-1 text-purple-600 font-medium">
                <FlaskConical className="w-3 h-3" /> 
                取证案件 #{forensicsCaseId}
                {forensicsFileName && <span className="text-muted-foreground"> - {forensicsFileName}</span>}
              </span>
            ) : activeIncidentId ? (
              <span className="flex items-center gap-1 text-primary/80 font-medium">
                <Server className="w-3 h-3" /> Linked to Incident #{activeIncidentId}
              </span>
            ) : (
              "Manual Hunting Mode"
            )}
          </CardDescription>
        </CardHeader>
        <CardContent className="pt-4 flex-1 flex flex-col gap-4">
          
          <div className="space-y-2 flex-1">
            <Label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Extracted Indicators</Label>
            {Object.keys(contextVars).length === 0 ? (
              <div className="p-4 border border-dashed rounded-md bg-muted/30 text-center">
                <p className="text-xs text-muted-foreground italic">No indicators yet.</p>
                <p className="text-[10px] text-muted-foreground mt-1">Extracting from context or add manually...</p>
              </div>
            ) : (
              <div className="flex flex-wrap gap-2 mt-2">
                {Object.entries(contextVars).map(([k, v]) => {
                  const displayValue = String(v).length > 30 ? String(v).substring(0, 27) + "..." : v;
                  return (
                    <Badge key={k} variant="secondary" className="text-[11px] py-1 px-2.5 flex items-center gap-1.5 group cursor-default border-primary/10 bg-primary/5 hover:bg-primary/10 transition-colors" title={String(v)}>
                      <span className="text-muted-foreground">{k}:</span> 
                      <span className="font-mono text-foreground">{displayValue}</span>
                      <Trash2 
                        className="w-3 h-3 ml-1 text-muted-foreground hover:text-red-500 cursor-pointer opacity-0 group-hover:opacity-100 transition-opacity" 
                        onClick={() => handleRemoveVar(k)}
                      />
                    </Badge>
                  )
                })}
              </div>
            )}
          </div>

          <div className="pt-4 border-t space-y-3">
            <Label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Add Custom Indicator</Label>
            <div className="flex flex-col gap-2">
              <div className="flex gap-2">
                <Input 
                  placeholder="Key (e.g. src_ip)" 
                  value={newVarKey} onChange={e => setNewVarKey(e.target.value)} 
                  className="h-8 text-xs font-mono flex-1"
                />
                <Input 
                  placeholder="Value" 
                  value={newVarValue} onChange={e => setNewVarValue(e.target.value)} 
                  className="h-8 text-xs font-mono flex-1"
                  onKeyDown={(e) => e.key === 'Enter' && handleAddVar()}
                />
              </div>
              <Button size="sm" variant="outline" className="h-8 w-full text-xs" onClick={handleAddVar}>
                <Plus className="w-3 h-3 mr-1" /> Add Indicator
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Search, Activity, LayoutList, X, ExternalLink } from "lucide-react";
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
}

export function DirectivesPanel({ templates, selectedTemplates, onChangeSelection, contextVars, loading, onExecute, onRefreshTemplates }: DirectivesPanelProps) {
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
        <CardHeader className="pb-3 border-b">
          <div className="flex justify-between items-center">
            <div>
              <CardTitle className="text-base flex items-center gap-2">
                <Search className="w-4 h-4 text-primary" />
                Investigation Directives
              </CardTitle>
              <CardDescription className="text-xs mt-1">
                Select and execute multiple rules to build the timeline.
              </CardDescription>
            </div>
            <div className="flex items-center gap-2">
              <Button variant="outline" size="sm" onClick={() => setSelectorOpen(true)}>
                <LayoutList className="w-4 h-4 mr-2" /> Select Directives
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
        </CardHeader>
        <CardContent className="p-4 bg-muted/5">
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
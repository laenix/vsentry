import { useState, useMemo } from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";
import { Search, ChevronRight, ChevronLeft, ChevronsRight, ChevronsLeft, Target } from "lucide-react";
import { type InvestigationDirective } from "@/services/investigation";

interface DirectivesSelectorDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  templates: InvestigationDirective[];
  selectedIds: number[];
  onChange: (ids: number[]) => void;
  contextVars: Record<string, string>;
}

export function DirectivesSelectorDialog({ open, onOpenChange, templates, selectedIds, onChange, contextVars }: DirectivesSelectorDialogProps) {
  const [search, setSearch] = useState("");

  // 检查是否缺Parameter - getMissingParams = (tpl: InvestigationDirective) => {
    let requiredParams: string[] = [];
    try { requiredParams = JSON.parse(tpl.parameters || "[]"); } catch(e){}
    return requiredParams.filter(p => !contextVars[p]);
  };

  //   分离已选Sum未选，并FilterSearch词
  const availableTemplates = useMemo(() => {
    return templates.filter(t => !selectedIds.includes(t.id) && t.name.toLowerCase().includes(search.toLowerCase()));
  }, [templates, selectedIds, search]);

  const selectedTemplates = useMemo(() => {
    return templates.filter(t => selectedIds.includes(t.id));
  }, [templates, selectedIds]);

  const handleSelectAll = () => {
    // 只能全选不缺Parameter的 - validIds = availableTemplates.filter(t => getMissingParams(t).length === 0).map(t => t.id);
    onChange([...selectedIds, ...validIds]);
  };

  const handleClearAll = () => onChange([]);

  const toggleOne = (id: number, add: boolean) => {
    if (add) {
      if (selectedIds.includes(id)) return;
      onChange([...selectedIds, id]);
    } else {
      onChange(selectedIds.filter(existing => existing !== id));
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl h-[70vh] flex flex-col p-0 gap-0">
        <DialogHeader className="px-6 py-4 border-b bg-muted/10 flex-none">
          <DialogTitle className="flex items-center gap-2">
            <Target className="w-5 h-5 text-primary" />
            Manage Directives (Transfer)
          </DialogTitle>
        </DialogHeader>

        <div className="flex-1 flex overflow-hidden p-6 gap-4 bg-muted/5">
          {/* 左侧：可Option */}
          <div className="flex-1 flex flex-col border rounded-md bg-card shadow-sm overflow-hidden">
            <div className="p-3 border-b bg-muted/20 flex flex-col gap-2">
              <div className="flex justify-between items-center">
                <span className="text-sm font-semibold">Available Directives</span>
                <Badge variant="secondary">{availableTemplates.length}</Badge>
              </div>
              <div className="relative">
                <Search className="w-4 h-4 absolute left-2 top-2 text-muted-foreground" />
                <Input placeholder="Search..." className="h-8 pl-8 text-xs" value={search} onChange={e => setSearch(e.target.value)} />
              </div>
              <Button variant="outline" size="sm" className="w-full text-xs h-7 mt-1" onClick={handleSelectAll}>
                Select All Valid <ChevronsRight className="w-3 h-3 ml-1" />
              </Button>
            </div>
            <ScrollArea className="flex-1 p-2">
              <div className="flex flex-col gap-2">
                {availableTemplates.map(tpl => {
                  const missing = getMissingParams(tpl);
                  const isDisabled = missing.length > 0;
                  return (
                    <div key={tpl.id} 
                      className={`p-2 border rounded text-xs flex justify-between items-center transition-colors ${isDisabled ? 'opacity-50 cursor-not-allowed bg-muted/30' : 'hover:bg-primary/5 hover:border-primary/30 cursor-pointer'}`}
                      onClick={() => !isDisabled && toggleOne(tpl.id, true)}
                    >
                      <div className="flex flex-col gap-1 min-w-0">
                        <span className="font-semibold truncate">{tpl.name}</span>
                        {isDisabled && <span className="text-[9px] text-red-500">Missing: {missing.join(', ')}</span>}
                      </div>
                      {!isDisabled && <ChevronRight className="w-4 h-4 text-muted-foreground" />}
                    </div>
                  )
                })}
              </div>
            </ScrollArea>
          </div>

          {/* 右侧：已Option */}
          <div className="flex-1 flex flex-col border rounded-md bg-card shadow-sm overflow-hidden">
            <div className="p-3 border-b bg-primary/5 flex flex-col gap-2">
              <div className="flex justify-between items-center">
                <span className="text-sm font-semibold text-primary">Selected Directives</span>
                <Badge className="bg-primary">{selectedTemplates.length}</Badge>
              </div>
              {/* 占位对齐 */}
              <div className="h-8"></div>
              <Button variant="outline" size="sm" className="w-full text-xs h-7 mt-1 text-red-500 hover:text-red-600 hover:bg-red-50" onClick={handleClearAll}>
                <ChevronsLeft className="w-3 h-3 mr-1" /> Clear All
              </Button>
            </div>
            <ScrollArea className="flex-1 p-2">
              <div className="flex flex-col gap-2">
                {selectedTemplates.map(tpl => (
                  <div key={tpl.id} 
                    className="p-2 border border-primary/20 bg-primary/5 rounded text-xs flex justify-between items-center cursor-pointer hover:bg-red-50 hover:border-red-200 transition-colors group"
                    onClick={() => toggleOne(tpl.id, false)}
                  >
                    <div className="flex flex-col gap-1 min-w-0">
                      <span className="font-semibold truncate text-primary group-hover:text-red-600">{tpl.name}</span>
                    </div>
                    <ChevronLeft className="w-4 h-4 text-muted-foreground group-hover:text-red-500" />
                  </div>
                ))}
              </div>
            </ScrollArea>
          </div>
        </div>

        <DialogFooter className="px-6 py-4 border-t bg-muted/10 flex-none">
          <Button onClick={() => onOpenChange(false)}>Confirm Selection</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
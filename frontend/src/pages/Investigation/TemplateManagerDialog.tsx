import { useState } from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Textarea } from "@/components/ui/textarea";
import { Plus, Edit, Trash2, Settings, ArrowLeft } from "lucide-react";
import { investigationService, type InvestigationTemplate } from "@/services/investigation";
import { toast } from "sonner";

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  templates: InvestigationTemplate[];
  onRefresh: () => void; // 增删改后刷新父组件列表
}

export function TemplateManagerDialog({ open, onOpenChange, templates, onRefresh }: Props) {
  const [editing, setEditing] = useState<Partial<InvestigationTemplate> | null>(null);

  const handleSave = async () => {
    if (!editing?.name || !editing?.logsql) {
      toast.error("Name and LogSQL are required");
      return;
    }
    
    // 智能提取 LogSQL 中的参数 ${xxx}
    const paramRegex = /\$\{([^}]+)\}/g;
    const params: string[] = [];
    let match;
    while ((match = paramRegex.exec(editing.logsql || "")) !== null) {
      if (!params.includes(match[1])) params.push(match[1]);
    }
    const finalData = { ...editing, parameters: JSON.stringify(params) };

    try {
      if (finalData.id) {
        await investigationService.updateTemplate(finalData);
        toast.success("Template updated");
      } else {
        await investigationService.addTemplate(finalData);
        toast.success("Template created");
      }
      setEditing(null);
      onRefresh();
    } catch (e) {
      toast.error("Failed to save template");
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm("Are you sure you want to delete this template?")) return;
    try {
      await investigationService.deleteTemplate(id);
      toast.success("Template deleted");
      onRefresh();
    } catch (e) {
      toast.error("Failed to delete template");
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl h-[80vh] flex flex-col p-0">
        <DialogHeader className="px-6 py-4 border-b bg-muted/10 flex-none">
          <div className="flex items-center gap-2">
            {editing && (
              <Button variant="ghost" size="icon" className="h-6 w-6 mr-2" onClick={() => setEditing(null)}>
                <ArrowLeft className="w-4 h-4" />
              </Button>
            )}
            <Settings className="w-5 h-5 text-primary" />
            <DialogTitle>{editing ? (editing.id ? "Edit Directive" : "New Directive") : "Manage Investigation Directives"}</DialogTitle>
          </div>
        </DialogHeader>

        <div className="flex-1 overflow-auto p-6">
          {editing ? (
            <div className="space-y-4 max-w-2xl mx-auto">
              <div className="grid gap-2">
                <Label>Directive Name</Label>
                <Input value={editing.name || ""} onChange={e => setEditing({...editing, name: e.target.value})} placeholder="e.g. Brute Force Hunt" />
              </div>
              <div className="grid gap-2">
                <Label>Description</Label>
                <Input value={editing.description || ""} onChange={e => setEditing({...editing, description: e.target.value})} placeholder="What does this hunt for?" />
              </div>
              <div className="grid gap-2">
                <Label>LogSQL Query (Use ${`{var}`} for parameters)</Label>
                <Textarea 
                  className="font-mono text-xs h-32" 
                  value={editing.logsql || ""} 
                  onChange={e => setEditing({...editing, logsql: e.target.value})} 
                  placeholder='src_endpoint.ip="${src_ip}" AND activity_name="Logon Failed"' 
                />
                <p className="text-[10px] text-muted-foreground">
                  Parameters will be automatically extracted and required before execution. e.g. ${`{src_ip}`} or ${`{hostname}`}
                </p>
              </div>
              <div className="flex justify-end gap-2 pt-4">
                <Button variant="outline" onClick={() => setEditing(null)}>Cancel</Button>
                <Button onClick={handleSave}>Save Directive</Button>
              </div>
            </div>
          ) : (
            <div className="space-y-4">
              <div className="flex justify-end">
                <Button size="sm" onClick={() => setEditing({ name: "", logsql: "" })}>
                  <Plus className="w-4 h-4 mr-2" /> New Directive
                </Button>
              </div>
              <div className="border rounded-md">
                <Table>
                  <TableHeader className="bg-muted/5">
                    <TableRow>
                      <TableHead>Name</TableHead>
                      <TableHead>Required Params</TableHead>
                      <TableHead className="w-[300px]">LogSQL Snippet</TableHead>
                      <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {templates.map(tpl => {
                      let params: string[] = [];
                      try { params = JSON.parse(tpl.parameters || "[]"); } catch(e){}
                      return (
                        <TableRow key={tpl.id}>
                          <TableCell className="font-medium">{tpl.name}</TableCell>
                          <TableCell>
                            <div className="flex flex-wrap gap-1">
                              {params.length > 0 ? params.map(p => (
                                <Badge key={p} variant="secondary" className="text-[10px] font-mono">{p}</Badge>
                              )) : <span className="text-xs text-muted-foreground">None</span>}
                            </div>
                          </TableCell>
                          <TableCell className="font-mono text-[10px] text-muted-foreground truncate max-w-[300px]">
                            {tpl.logsql}
                          </TableCell>
                          <TableCell className="text-right">
                            <Button variant="ghost" size="icon" onClick={() => setEditing(tpl)}><Edit className="w-4 h-4" /></Button>
                            <Button variant="ghost" size="icon" className="text-red-500" onClick={() => handleDelete(tpl.id)}><Trash2 className="w-4 h-4" /></Button>
                          </TableCell>
                        </TableRow>
                      )
                    })}
                  </TableBody>
                </Table>
              </div>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
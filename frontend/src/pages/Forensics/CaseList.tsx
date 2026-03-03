import { useEffect, useState } from "react";
import { forensicsService, type ForensicTask } from "@/services/forensics";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, CardFooter } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Briefcase, Plus, FolderOpen, Clock, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { ScrollArea } from "@/components/ui/scroll-area";

interface CaseListProps {
  onOpenCase: (id: number) => void;
}

export function CaseList({ onOpenCase }: CaseListProps) {
  const [tasks, setTasks] = useState<ForensicTask[]>([]);
  const [loading, setLoading] = useState(true);
  
  // 创建案件弹窗状态
  const [createOpen, setCreateOpen] = useState(false);
  const [newName, setNewName] = useState("");
  const [newDesc, setNewDesc] = useState("");

  const fetchTasks = async () => {
    setLoading(true);
    try {
      const res = await forensicsService.listTasks();
      if (res.code === 200 && res.data) {
        setTasks(res.data);
      }
    } catch (e) {
      toast.error("Failed to load forensic cases");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTasks();
  }, []);

  const handleCreate = async () => {
    if (!newName.trim()) {
      toast.error("Case name is required");
      return;
    }
    try {
      await forensicsService.createTask({ name: newName, description: newDesc });
      toast.success("Case created successfully");
      setCreateOpen(false);
      setNewName("");
      setNewDesc("");
      fetchTasks();
    } catch (e) {
      toast.error("Failed to create case");
    }
  };

  const handleDelete = async (e: React.MouseEvent, id: number) => {
    e.stopPropagation(); // 阻止触发打开案件的点击事件
    if (!confirm("Delete this case and all its files permanently?")) return;
    try {
      await forensicsService.deleteTask(id);
      toast.success("Case deleted");
      fetchTasks();
    } catch (err) {
      toast.error("Failed to delete case");
    }
  };

  return (
    <div className="p-6 h-full flex flex-col gap-6">
      <div className="flex justify-between items-end flex-none">
        <div>
          <h1 className="text-2xl font-bold tracking-tight flex items-center gap-2">
            <Briefcase className="w-6 h-6 text-primary" /> Forensics Sandbox
          </h1>
          <p className="text-muted-foreground text-sm mt-1">
            Isolated environment for parsing and analyzing PCAP, EVTX, and raw logs.
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="w-4 h-4 mr-2" /> New Case
        </Button>
      </div>

      <ScrollArea className="flex-1 -mx-6 px-6">
        {tasks.length === 0 && !loading ? (
          <div className="h-64 flex flex-col items-center justify-center text-muted-foreground border-2 border-dashed rounded-lg mt-4">
            <FolderOpen className="w-10 h-10 mb-2 opacity-20" />
            <p>No active forensic cases.</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 pb-6">
            {tasks.map(task => (
              <Card 
                key={task.id} 
                className="group cursor-pointer hover:border-primary/50 transition-all hover:shadow-md flex flex-col h-48"
                onClick={() => onOpenCase(task.id)}
              >
                <CardHeader className="pb-2">
                  <div className="flex justify-between items-start">
                    <CardTitle className="text-base truncate pr-4" title={task.name}>
                      {task.name}
                    </CardTitle>
                    <Badge variant={task.status === 'open' ? 'default' : 'secondary'} className="text-[10px]">
                      {task.status}
                    </Badge>
                  </div>
                  <CardDescription className="text-xs line-clamp-2 mt-1 h-8">
                    {task.description || "No description provided."}
                  </CardDescription>
                </CardHeader>
                <CardContent className="flex-1">
                  <div className="text-xs text-muted-foreground flex items-center gap-2 bg-muted/30 p-2 rounded-md">
                    <FolderOpen className="w-3.5 h-3.5" />
                    <span>{task.files?.length || 0} Evidences Collected</span>
                  </div>
                </CardContent>
                <CardFooter className="pt-0 border-t flex justify-between items-center py-3 bg-muted/5">
                  <span className="text-[10px] text-muted-foreground flex items-center gap-1">
                    <Clock className="w-3 h-3" /> {new Date(task.created_at).toLocaleDateString()}
                  </span>
                  <Button variant="ghost" size="icon" className="h-6 w-6 text-muted-foreground hover:text-red-500 opacity-0 group-hover:opacity-100 transition-opacity" onClick={(e) => handleDelete(e, task.id)}>
                    <Trash2 className="w-3.5 h-3.5" />
                  </Button>
                </CardFooter>
              </Card>
            ))}
          </div>
        )}
      </ScrollArea>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create New Forensic Case</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label>Case Name</Label>
              <Input placeholder="e.g. Operation Phishing Hunt 2026" value={newName} onChange={e => setNewName(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label>Description (Optional)</Label>
              <Input placeholder="Scope and objective of this sandbox..." value={newDesc} onChange={e => setNewDesc(e.target.value)} />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>Cancel</Button>
            <Button onClick={handleCreate}>Initialize Sandbox</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
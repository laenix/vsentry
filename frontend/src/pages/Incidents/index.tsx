import { useEffect, useState } from "react";
import { incidentService } from "@/services/incidents";
import type { Incident } from "@/services/incidents";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { 
  CheckCircle, Eye, RotateCw, ShieldAlert, CheckCheck, 
  Archive, Fingerprint, User, Tag, ShieldCheck 
} from "lucide-react";
import { toast } from "sonner";
import { IncidentDetailDialog } from "./IncidentDetailDialog"; 
import { IncidentResolveDialog } from "./IncidentResolveDialog";
import { IncidentAssignDialog } from "./IncidentAssignDialog";

export default function IncidentsPage() {
  const [incidents, setIncidents] = useState<Incident[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState("all"); 
  
  // 弹窗状态管理
  const [selectedIncident, setSelectedIncident] = useState<Incident | null>(null);
  const [detailOpen, setDetailOpen] = useState(false);
  const [resolveTargetId, setResolveTargetId] = useState<number | null>(null);
  const [resolveDialogOpen, setResolveDialogOpen] = useState(false);
  const [assignTarget, setAssignTarget] = useState<Incident | null>(null);
  const [assignDialogOpen, setAssignDialogOpen] = useState(false);

  // 获取真实 ID (兼容后端 GORM 默认的大写 ID)
  const getIncidentID = (i: Incident) => i.ID || (i as any).id || 0;

  // 1. 加载事件列表
  const fetchIncidents = async () => {
    setLoading(true);
    try {
      const res = await incidentService.list();
      if (res.code === 200) {
        let list = res.data || [];
        // 按照最后活跃时间 (last_seen) 倒序排列
        list.sort((a, b) => {
           const tA = new Date(a.last_seen || a.CreatedAt).getTime();
           const tB = new Date(b.last_seen || b.CreatedAt).getTime();
           return tB - tA;
        });
        setIncidents(list);
      }
    } catch (err) {
      console.error(err);
      toast.error("Failed to load incidents");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchIncidents();
    const timer = setInterval(fetchIncidents, 30000); 
    return () => clearInterval(timer);
  }, []);

  // 2. 处理认领 (Acknowledge)
  const handleAcknowledge = async (id: number) => {
    try {
      await incidentService.acknowledge(id);
      toast.success(`Incident #${id} acknowledged`);
      setIncidents(prev => prev.map(i => getIncidentID(i) === id ? { 
        ...i, status: "acknowledged", assignee: 1 
      } : i));
    } catch (e) { console.error(e); }
  };

  // 3. 处理关闭 (Resolve - 带有分类和评论输入)
  const handleConfirmResolve = async (classification: string, comment: string) => {
    if (!resolveTargetId) return;
    try {
      await incidentService.resolve({ id: resolveTargetId, classification, comment });
      toast.success(`Incident #${resolveTargetId} resolved`);
      setIncidents(prev => prev.map(i => getIncidentID(i) === resolveTargetId ? { 
        ...i, status: "resolved" 
      } : i));
    } catch (e) { toast.error("Failed to resolve incident"); }
  };

  // 前端过滤逻辑
  const filteredData = incidents.filter(i => {
    if (filter === "all") return true;
    return i.status === filter || (filter === "ack" && i.status === "acknowledged");
  });

  return (
    <div className="p-6 h-full flex flex-col bg-background text-foreground">
      {/* 头部标题区 */}
      <div className="flex justify-between items-start mb-6">
        <div className="flex items-center gap-3">
           <div className="p-2 bg-primary/10 rounded-lg text-primary">
             <ShieldAlert className="w-6 h-6" />
           </div>
           <div>
            <h1 className="text-2xl font-bold tracking-tight">Incident Center</h1>
            <p className="text-muted-foreground text-sm">Investigate and manage aggregated security incidents.</p>
           </div>
        </div>
        <Button variant="outline" size="sm" onClick={fetchIncidents} disabled={loading}>
          <RotateCw className={`w-4 h-4 mr-2 ${loading ? "animate-spin" : ""}`} />
          Refresh
        </Button>
      </div>

      {/* 状态切换 Tabs */}
      <div className="flex items-center gap-4 mb-4">
        <Tabs value={filter} onValueChange={setFilter} className="w-[500px]">
          <TabsList>
            <TabsTrigger value="all">All</TabsTrigger>
            <TabsTrigger value="new" className="gap-2">
              New
              {incidents.filter(i => i.status === 'new').length > 0 && (
                <Badge variant="destructive" className="h-5 px-1.5 text-[10px] rounded-full">
                  {incidents.filter(i => i.status === 'new').length}
                </Badge>
              )}
            </TabsTrigger>
            <TabsTrigger value="ack">Acknowledged</TabsTrigger>
            <TabsTrigger value="resolved">Resolved</TabsTrigger>
          </TabsList>
        </Tabs>
      </div>

      {/* 事件列表表格 */}
      <div className="border rounded-md bg-card flex-1 overflow-auto shadow-sm">
        <Table>
          <TableHeader>
            <TableRow className="bg-muted/5">
              <TableHead className="w-[60px]">ID</TableHead>
              <TableHead className="w-[100px]">Status</TableHead>
              <TableHead className="w-[100px]">Severity</TableHead>
              <TableHead className="min-w-[200px]">Incident Name / Fingerprint</TableHead>
              <TableHead className="w-[100px]">Alerts</TableHead>
              <TableHead className="w-[120px]">Assignee</TableHead>
              <TableHead className="w-[180px]">Last Seen</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredData.length > 0 ? (
              filteredData.map((incident) => {
                const id = getIncidentID(incident);
                return (
                  <TableRow key={id} className="group h-[65px]">
                    <TableCell className="font-mono text-xs text-muted-foreground">#{id}</TableCell>
                    
                    <TableCell>
                      {incident.status === 'new' && <Badge variant="destructive">New</Badge>}
                      {incident.status === 'acknowledged' && <Badge variant="secondary" className="bg-blue-100 text-blue-700">Ack</Badge>}
                      {incident.status === 'resolved' && <Badge variant="outline" className="text-emerald-600 border-emerald-200 bg-emerald-50"><ShieldCheck className="w-3 h-3 mr-1"/> Closed</Badge>}
                    </TableCell>

                    <TableCell>
                      <Badge variant="outline" className={`font-bold uppercase text-[10px] ${
                        incident.severity === 'critical' ? 'border-red-500 text-red-600' : 
                        incident.severity === 'high' ? 'border-orange-500 text-orange-600' : 'border-blue-500 text-blue-600'
                      }`}>
                        {incident.severity}
                      </Badge>
                    </TableCell>

                    <TableCell>
                      <div className="flex flex-col gap-0.5">
                        <div className="font-semibold text-sm truncate max-w-[300px]">{incident.name}</div>
                        <div className="flex items-center gap-1 text-[10px] text-muted-foreground font-mono">
                          <Fingerprint className="w-3 h-3" /> {incident.fingerprint?.substring(0, 12)}...
                        </div>
                      </div>
                    </TableCell>

                    <TableCell>
                        <Badge variant="secondary" className="font-mono">
                          {incident.alert_count} evidence
                        </Badge>
                    </TableCell>

                    <TableCell>
                       {incident.assignee ? (
                         <div className="flex items-center gap-2 text-xs">
                           <User className="w-3 h-3 text-primary" />
                           <span>{incident.assignee === 1 ? "Me" : `#${incident.assignee}`}</span>
                         </div>
                       ) : <span className="text-muted-foreground text-xs italic">Unassigned</span>}
                    </TableCell>

                    <TableCell className="text-xs text-muted-foreground">
                      {new Date(incident.last_seen || incident.CreatedAt).toLocaleString()}
                    </TableCell>

                    <TableCell className="text-right">
                      <div className="flex justify-end gap-1">
                        <Button variant="ghost" size="icon" onClick={() => { setSelectedIncident(incident); setDetailOpen(true); }}>
                          <Eye className="w-4 h-4" />
                        </Button>
                        {incident.status === 'new' && (
                          <Button variant="ghost" size="icon" className="text-blue-500" onClick={() => handleAcknowledge(id)}>
                            <CheckCircle className="w-4 h-4" />
                          </Button>
                        )}
                        {incident.status !== 'resolved' && (
                          <Button variant="ghost" size="icon" className="text-emerald-500" onClick={() => { setResolveTargetId(id); setResolveDialogOpen(true); }}>
                            <CheckCheck className="w-4 h-4" />
                          </Button>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                )
              })
            ) : (
              <TableRow>
                <TableCell colSpan={8} className="h-64 text-center text-muted-foreground">
                  No incidents found matching current filter.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {/* 弹窗组件挂载 */}
      <IncidentDetailDialog 
        open={detailOpen} 
        onOpenChange={setDetailOpen} 
        alertId={selectedIncident ? getIncidentID(selectedIncident) : null} // ✅ 传递 ID 以便触发详情 API
        onAcknowledge={handleAcknowledge}
        onResolve={(id: number) => { setResolveTargetId(id); setResolveDialogOpen(true); }}
      />

      <IncidentResolveDialog 
        open={resolveDialogOpen} 
        onOpenChange={setResolveDialogOpen} 
        onConfirm={handleConfirmResolve} 
      />
    </div>
  );
}
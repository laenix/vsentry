import React, { useEffect, useState } from "react";
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { 
  ShieldAlert, Clock, Database, 
  User, Activity, Info, ExternalLink 
} from "lucide-react";
import { incidentService } from "@/services/incidents";
import type { Incident } from "@/services/incidents";
import { ReadOnlyJsonViewer } from "@/components/editor/ReadOnlyJsonViewer";
import { toast } from "sonner";

// 1. 接口定义：对齐 Go 后端结构
interface Alert {
  id: number;
  // 兼容多种时间字段格式 (Go CreatedAt 或 JSON _time)
  created_at?: string; 
  CreatedAt?: string;
  _time?: string;
  
  content: string;     // 原始 JSON
  fingerprint: string; 
}

interface IncidentDetail extends Incident {
  first_seen: string; 
  last_seen: string; 
  alerts: Alert[];    
}

interface IncidentDetailDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  alertId: number | null;
  onAcknowledge: (id: number) => void;
  onResolve: (id: number) => void;
}

export function IncidentDetailDialog({ 
  open, 
  onOpenChange, 
  alertId, 
  onAcknowledge, 
  onResolve 
}: IncidentDetailDialogProps) {
  const [data, setData] = useState<IncidentDetail | null>(null);
  const [loading, setLoading] = useState(false);

  // 1. 加载详情
  useEffect(() => {
    if (open && alertId) {
      const fetchDetail = async () => {
        setLoading(true);
        try {
          const res = await incidentService.detail(alertId); 
          if (res.code === 200) {
            setData(res.data);
          }
        } catch (error) {
          console.error("Investigation failed:", error);
          toast.error("Failed to load incident details");
        } finally {
          setLoading(false);
        }
      };
      fetchDetail();
    }
  }, [open, alertId]);

  // ✅ 2. 核心：在新窗口打开纯时间范围查询
  const handleInvestigateNewWindow = (alert: Alert) => {
    // A. 智能获取时间
    const timeStr = alert.created_at || alert.CreatedAt || alert._time || new Date().toISOString();
    const eventTime = new Date(timeStr).getTime();
    
    if (isNaN(eventTime)) {
      toast.error("Invalid timestamp in evidence");
      return;
    }

    // B. 计算前后 5 分钟 (Context Buffer)
    const BUFFER_MS = 5 * 60 * 1000; 
    const start = new Date(eventTime - BUFFER_MS).toISOString();
    const end = new Date(eventTime + BUFFER_MS).toISOString();

    // C. 构造 VictoriaLogs 标准时间范围查询
    // 语法: _time:[start_iso, end_iso]
    // 不包含 fingerprint 或 content，只看这段时间发生了什么
    const vlQuery = `_time:[${start}, ${end}]`;

    // D. 构造 URL (只带 q 参数)
    const url = `/logs?q=${encodeURIComponent(vlQuery)}`;

    // E. 在新标签页打开
    window.open(url, '_blank');
    
    toast.success("Context investigation opened in new tab");
  };

  if (!data) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      {/* 布局修复: 固定高度 + Flex Column 确保滚动条正常工作 */}
      <DialogContent className="max-w-5xl h-[90vh] flex flex-col p-0 gap-0 overflow-hidden">
        
        {/* Header: 仅展示信息 */}
        <DialogHeader className="px-6 py-4 border-b bg-muted/10 flex-none">
          <div className="flex justify-between items-start">
            <div className="space-y-1">
              <div className="flex items-center gap-3 mb-1">
                <ShieldAlert className={data.severity === 'critical' ? 'text-red-600' : 'text-orange-500'} />
                <DialogTitle className="text-xl font-bold">{data.name}</DialogTitle>
                <Badge variant={data.status === 'new' ? 'destructive' : 'outline'}>{data.status}</Badge>
              </div>
              <div className="flex items-center gap-4 text-xs text-muted-foreground">
                <span className="font-mono">ID: #{data.ID}</span>
                <span className="flex items-center gap-1">
                  <User className="w-3 h-3"/> {data.assignee ? `Assignee: #${data.assignee}` : 'Unassigned'}
                </span>
                <span className="flex items-center gap-1">
                   <Clock className="w-3 h-3"/> First Seen: {new Date(data.first_seen || Date.now()).toLocaleTimeString()}
                </span>
              </div>
            </div>
            {/* 顶部的全局 Investigate Logs 按钮已移除 */}
          </div>
        </DialogHeader>

        {/* Scrollable Content */}
        <ScrollArea className="flex-1 bg-muted/5">
          <div className="p-6 space-y-8 max-w-4xl mx-auto">
            
            {/* 统计概览卡片 */}
            <div className="grid grid-cols-3 gap-4">
              <SummaryCard 
                icon={<Activity className="w-4 h-4 text-blue-500" />} 
                label="Alert Count" 
                value={`${data.alert_count} Evidence Items`} 
              />
              <SummaryCard 
                icon={<Clock className="w-4 h-4 text-orange-500" />} 
                label="Last Seen" 
                value={new Date(data.last_seen || Date.now()).toLocaleString()} 
              />
              <SummaryCard 
                icon={<Info className="w-4 h-4 text-purple-500" />} 
                label="Type" 
                value={data.label || "Security Alert"} 
              />
            </div>

            {/* 证据时间轴 */}
            <div className="space-y-6">
              <h3 className="text-sm font-bold flex items-center gap-2 text-muted-foreground uppercase tracking-wider">
                <Database className="w-4 h-4" /> Evidence Timeline
              </h3>
              
              <div className="space-y-8 relative before:absolute before:inset-0 before:ml-5 before:-translate-x-px before:h-full before:w-0.5 before:bg-gradient-to-b before:from-transparent before:via-slate-300 before:to-transparent">
                {data.alerts && data.alerts.length > 0 ? (
                  data.alerts.map((alert, index) => {
                    const timeStr = alert.created_at || alert.CreatedAt || alert._time;
                    const displayTime = timeStr ? new Date(timeStr).toLocaleString() : "Unknown Time";

                    return (
                      <div key={alert.id || index} className="relative flex items-start gap-6 group">
                        {/* 左侧圆点 */}
                        <div className="absolute left-5 -translate-x-1/2 w-3 h-3 rounded-full border-2 border-primary bg-background group-hover:scale-125 transition-transform z-10" />
                        
                        <div className="flex-1 ml-10 space-y-2">
                          {/* Alert Header: 包含行内操作按钮 */}
                          <div className="flex items-center justify-between bg-card border rounded-t-md p-2 px-3 shadow-sm group-hover:border-primary/30 transition-colors">
                            <div className="flex items-center gap-3">
                                <Badge variant="secondary" className="font-mono text-[10px] h-5">
                                  #{alert.id || index + 1}
                                </Badge>
                                <span className="text-xs text-muted-foreground flex items-center gap-1">
                                  <Clock className="w-3 h-3" /> {displayTime}
                                </span>
                            </div>

                            {/* ✅ 行内按钮：打开新窗口查询上下文 */}
                            <Button 
                              variant="ghost" 
                              size="sm" 
                              className="h-6 text-[10px] gap-1.5 hover:bg-blue-50 hover:text-blue-600"
                              onClick={() => handleInvestigateNewWindow(alert)}
                            >
                              <ExternalLink className="w-3 h-3" />
                              Open Context
                            </Button>
                          </div>
                          
                          {/* JSON Viewer */}
                          <div className="border border-t-0 rounded-b-md overflow-hidden shadow-sm">
                            <ReadOnlyJsonViewer 
                              value={alert.content} 
                              height="200px" 
                              className="border-none"
                            />
                          </div>
                        </div>
                      </div>
                    );
                  })
                ) : (
                  <div className="text-center py-10 text-muted-foreground border-2 border-dashed rounded-lg">
                    No evidence details available.
                  </div>
                )}
              </div>
            </div>
          </div>
        </ScrollArea>

        {/* Footer */}
        <DialogFooter className="px-6 py-4 border-t bg-muted/10 flex-none gap-2">
          <Button variant="outline" onClick={() => onOpenChange(false)}>Close Review</Button>
          {data.status === 'new' && (
            <Button onClick={() => onAcknowledge(data.ID)}>
              Acknowledge Incident
            </Button>
          )}
          {data.status !== 'resolved' && (
            <Button className="bg-emerald-600 hover:bg-emerald-700 text-white" onClick={() => onResolve(data.ID)}>
              Resolve Incident
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

// 辅助组件：信息卡片
function SummaryCard({ icon, label, value }: { icon: any; label: string; value: string }) {
  return (
    <div className="bg-card border rounded-lg p-3 space-y-1 shadow-sm">
      <div className="flex items-center gap-2 text-[10px] font-bold text-muted-foreground uppercase tracking-tight">
        {icon} {label}
      </div>
      <div className="text-sm font-semibold truncate">{value}</div>
    </div>
  );
}
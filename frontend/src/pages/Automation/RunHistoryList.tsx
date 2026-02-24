import { useEffect, useState } from "react";
import { 
    Table, TableBody, TableCell, TableHead, TableHeader, TableRow 
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { ExternalLink, CheckCircle2, XCircle, Clock, Loader2, AlertCircle } from "lucide-react";
import { automationService} from "@/services/automation";
import type { PlaybookExecution } from "@/services/automation";
import { toast } from "sonner";
import { format } from "date-fns";

export default function RunHistoryList() {
  const [data, setData] = useState<PlaybookExecution[]>([]);
  const [loading, setLoading] = useState(true);

  // 简单的轮询，每 5 秒刷新一次列表，看是否有新任务完成
  useEffect(() => {
    const fetchHistory = async () => {
      try {
        const res = await automationService.getGlobalExecutions({ limit: 20 });
        if (res.code === 200) {
            setData(res.data || []);
        }
      } catch (error) {
        console.error("Failed to load history", error);
        // 不弹出 toast 避免轮询时骚扰用户
      } finally {
        setLoading(false);
      }
    };

    fetchHistory();
    const interval = setInterval(fetchHistory, 5000); // 5s 轮询

    return () => clearInterval(interval);
  }, []);

  if (loading && data.length === 0) {
    return (
        <div className="flex justify-center items-center h-64 border rounded-md bg-card">
          <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
        </div>
    );
  }

  if (data.length === 0) {
    return (
        <div className="flex flex-col justify-center items-center h-64 border rounded-md bg-card text-muted-foreground gap-2">
            <Clock className="w-8 h-8 opacity-20" />
            <p>No execution history found yet.</p>
        </div>
    );
  }

  return (
    <div className="border rounded-md bg-card">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Status</TableHead>
            <TableHead>Playbook</TableHead>
            <TableHead>Context</TableHead>
            <TableHead>Duration</TableHead>
            <TableHead>Executed At</TableHead>
            <TableHead className="text-right">Run ID</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((run) => (
            <TableRow key={run.id} className="hover:bg-muted/50">
              <TableCell>
                <div className="flex items-center gap-2">
                  {run.status === 'success' ? (
                      <CheckCircle2 className="w-4 h-4 text-emerald-500" />
                  ) : run.status === 'failed' ? (
                      <XCircle className="w-4 h-4 text-red-500" />
                  ) : (
                      <Loader2 className="w-4 h-4 text-blue-500 animate-spin" />
                  )}
                  <span className={`text-xs font-medium capitalize ${
                      run.status === 'success' ? 'text-emerald-700' : 
                      run.status === 'failed' ? 'text-red-700' : 'text-blue-700'
                  }`}>
                      {run.status}
                  </span>
                </div>
              </TableCell>
              <TableCell className="font-medium text-sm">
                  {/* 如果后端返回了 playbook_name 则显示，否则显示 ID */}
                  {run.playbook_name || `Playbook #${run.playbook_id}`}
              </TableCell>
              <TableCell>
                  {run.trigger_context_id ? (
                      <div className="flex items-center gap-1 text-xs text-blue-600 hover:underline cursor-pointer">
                          Incident #{run.trigger_context_id} <ExternalLink className="w-3 h-3" />
                      </div>
                  ) : (
                      <span className="text-xs text-muted-foreground">Manual / Test</span>
                  )}
              </TableCell>
              <TableCell className="text-xs text-muted-foreground font-mono">
                  {run.duration_ms}ms
              </TableCell>
              <TableCell className="text-xs text-muted-foreground">
                  {run.start_time ? format(new Date(run.start_time), "MMM d, HH:mm:ss") : "-"}
              </TableCell>
              <TableCell className="text-right text-xs font-mono text-muted-foreground">
                  #{run.id}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
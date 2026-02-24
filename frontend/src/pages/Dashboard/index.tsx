import { useEffect, useState } from "react";
import { dashboardService } from "@/services/dashboard";
import type { DashboardStats } from "@/services/dashboard";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Activity, AlertTriangle, ShieldCheck } from "lucide-react";

export default function DashboardPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const res = await dashboardService.getStats();
        if (res.code === 200) {
          setStats(res.data);
        }
      } catch (error) {
        console.error("Failed to fetch dashboard stats", error);
      }
    };
    fetchStats();
    
    // 可选：每30秒轮询一次
    const timer = setInterval(fetchStats, 30000);
    return () => clearInterval(timer);
  }, []);

  if (!stats) return <div className="p-8">Loading dashboard...</div>;

  return (
    <div className="p-6 space-y-6">
      <h1 className="text-2xl font-bold tracking-tight">Security Overview</h1>
      
      <div className="grid gap-4 md:grid-cols-3">
        {/* Total Alerts Card */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Alerts</CardTitle>
            <ShieldCheck className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.total_alerts}</div>
            <p className="text-xs text-muted-foreground">All time detections</p>
          </CardContent>
        </Card>

        {/* New Alerts Card */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">New Alerts</CardTitle>
            <AlertTriangle className="h-4 w-4 text-orange-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-orange-600">{stats.new_alerts}</div>
            <p className="text-xs text-muted-foreground">Require attention</p>
          </CardContent>
        </Card>

        {/* System Status (Mocked or from vlogs_metrics) */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">System Status</CardTitle>
            <Activity className="h-4 w-4 text-emerald-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-emerald-600">Healthy</div>
            {/* 这里可以展示 stats.vlogs_metrics 里的具体数据 */}
            <p className="text-xs text-muted-foreground">VictoriaLogs Active</p>
          </CardContent>
        </Card>
      </div>

      {/* Severity Breakdown */}
      <div className="grid gap-4 md:grid-cols-2">
         {/* 可以放一个 Chart 展示 stats.severity_counts */}
         <Card>
            <CardHeader>
                <CardTitle>Severity Distribution</CardTitle>
            </CardHeader>
            <CardContent>
                {Object.entries(stats.severity_counts || {}).map(([sev, count]) => (
                    <div key={sev} className="flex justify-between py-2 border-b last:border-0">
                        <span className="capitalize">{sev}</span>
                        <span className="font-mono font-bold">{count}</span>
                    </div>
                ))}
            </CardContent>
         </Card>
      </div>
    </div>
  );
}
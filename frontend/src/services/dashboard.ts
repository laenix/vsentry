import { apiClient } from "@/lib/api/vsentry-client";
import type { APIResponse } from "@/lib/api/vsentry-client";

export interface DashboardStats {
  total_alerts: number;
  new_alerts: number;
  severity_counts: Record<string, number>;
  vlogs_metrics: any; // 具体结构看后端返回
}

export const dashboardService = {
  getStats: () => apiClient.get<any, APIResponse<DashboardStats>>("/dashboard"),
};
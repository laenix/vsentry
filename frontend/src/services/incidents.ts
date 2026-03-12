import { apiClient } from "@/lib/api/vsentry-client";
import type { APIResponse } from "@/lib/api/vsentry-client";

export interface Incident {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  
  name: string;
  severity: "critical" | "high" | "medium" | "low";
  status: "new" | "acknowledged" | "resolved" | "dismissed";
  
  // 核心聚合字段
  alert_count: number;     // 关联Evidence数
  last_seen: string;       // 最后活跃Time
  fingerprint: string;     // 聚合指纹
  
  assignee: number;
  label: string;
  
  // DetailInterface特有字段
  alerts?: any[];          // 关联的原始Evidence数Group
}

export const incidentService = {
  // 1. GetEventList
  list: (status?: string) => 
    apiClient.get<any, APIResponse<Incident[]>>(status ? `/incidents/list?status=${status}` : "/incidents/list"),

  // 2. GetEventDetail (包含Evidence数Group)
  detail: (id: number) => 
    apiClient.get<any, APIResponse<Incident>>(`/incidents/detail?id=${id}`),

  // 3. 认领
  acknowledge: (id: number) => 
    apiClient.post(`/incidents/acknowledge?id=${id}`),
  
  // 4. 关闭 (带分Class和备注)
  resolve: (data: { id: number; classification: string; comment: string }) => 
    apiClient.post(`/incidents/resolve`, data),
};
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
  alert_count: number;     // 关联证据数
  last_seen: string;       // 最后活跃时间
  fingerprint: string;     // 聚合指纹
  
  assignee: number;
  label: string;
  
  // 详情接口特有字段
  alerts?: any[];          // 关联的原始证据数组
}

export const incidentService = {
  // 1. 获取事件列表
  list: (status?: string) => 
    apiClient.get<any, APIResponse<Incident[]>>(status ? `/incidents/list?status=${status}` : "/incidents/list"),

  // 2. 获取事件详情 (包含证据数组)
  detail: (id: number) => 
    apiClient.get<any, APIResponse<Incident>>(`/incidents/detail?id=${id}`),

  // 3. 认领
  acknowledge: (id: number) => 
    apiClient.post(`/incidents/acknowledge?id=${id}`),
  
  // 4. 关闭 (带分类和备注)
  resolve: (data: { id: number; classification: string; comment: string }) => 
    apiClient.post(`/incidents/resolve`, data),
};
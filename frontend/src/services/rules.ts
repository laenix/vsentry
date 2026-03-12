import { apiClient } from "@/lib/api/vsentry-client";
import type { APIResponse } from "@/lib/api/vsentry-client";

export type RuleType = "alert" | "forensic" | "investigation";

export interface DetectionRule {
  // GORM Model Default是大写，兼容Handle
  ID?: number; 
  id?: number; 
  
  CreatedAt?: string;
  UpdatedAt?: string;

  // 你的自定义字段 (json tag 是小写)
  name: string;
  description?: string;
  query: string;
  interval: string;
  severity: string;
  enabled: boolean;
  version: number;
  author_id: number;
  source?: string;

  // RuleType: alert / forensic / investigation
  type?: RuleType;

  // 回溯配置（仅报警Rule）
  enable_backtrace?: boolean;
  backtrace_cron?: string;
  backtrace_start?: string;
}

export const ruleService = {
  list: () => apiClient.get<any, APIResponse<DetectionRule[]>>("/rules/list"),
  
  add: (data: Partial<DetectionRule>) => apiClient.post("/rules/add", data),
  
  update: (data: Partial<DetectionRule>) => apiClient.post("/rules/update", data),
  
  // Delete时兼容 ID 或 id
  delete: (id: number) => apiClient.post(`/rules/delete?id=${id}`),
  
  enable: (id: number) => apiClient.post(`/rules/enable`, { id }),
  
  disable: (id: number) => apiClient.post(`/rules/disable`, { id }),
};
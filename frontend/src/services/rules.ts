import { apiClient } from "@/lib/api/vsentry-client";
import type { APIResponse } from "@/lib/api/vsentry-client";

export interface DetectionRule {
  // GORM Model 默认是大写，兼容处理
  ID?: number; 
  id?: number; 
  
  CreatedAt?: string;
  UpdatedAt?: string;

  // 你的自定义字段 (json tag 是小写)
  name: string;
  description?: string; // 新增描述
  query: string;
  interval: string;
  severity: string;
  enabled: boolean;
  version: number;     // 新增版本
  author_id: number;   // 新增作者ID
  source?: string;
}

export const ruleService = {
  list: () => apiClient.get<any, APIResponse<DetectionRule[]>>("/rules/list"),
  
  add: (data: Partial<DetectionRule>) => apiClient.post("/rules/add", data),
  
  update: (data: Partial<DetectionRule>) => apiClient.post("/rules/update", data),
  
  // 删除时兼容 ID 或 id
  delete: (id: number) => apiClient.post(`/rules/delete?id=${id}`),
  
  enable: (id: number) => apiClient.post(`/rules/enable`, { id }),
  
  disable: (id: number) => apiClient.post(`/rules/disable`, { id }),
};
import { apiClient } from "@/lib/api/vsentry-client";
import type { APIResponse } from "@/lib/api/vsentry-client";
import type { Node, Edge, Viewport } from "@xyflow/react";

// --- Types Definition ---

export interface Playbook {
  ID: number;
  name: string;
  description: string;
  is_active: boolean;
  trigger_type: string; 
  
  // React Flow 的完整数据
  definition: {
    nodes: Node[];
    edges: Edge[];
    viewport: Viewport;
  };
  
  // 统计数据 (可选)
  last_run_at?: string;
  run_count?: number;
  success_rate?: number;

  created_at: string;
  updated_at: string;
}

export interface PlaybookExecution {
  id: number;
  playbook_id: number;
  playbook_name?: string;
  status: "running" | "success" | "failed";
  trigger_context_id: number;
  
  start_time: string;
  end_time: string;
  duration_ms: number;

  // 节点执行日志: Key = NodeID, Value = Result
  // 这是实现“动态补全”的关键数据源！
  logs: Record<string, StepResult>;
}

export interface StepResult {
  status: "success" | "failed" | "skipped";
  output: any;  // JSON 数据 (HTTP Response, Calculation Result etc.)
  error?: string;
  start_time?: string;
  end_time?: string;
}

// --- Service Implementation ---

// 基础路径对应后端的 r.Group("/playbooks")
const BASE_URL = "/playbooks";

export const automationService = {
  // 1. 获取剧本列表 [GET /playbooks]
  getList: (params?: { page?: number; keyword?: string }) => 
    apiClient.get<any, APIResponse<Playbook[]>>(BASE_URL, { params }),

  // 2. 获取详情 [GET /playbooks/:id]
  getDetail: (id: string | number) => 
    apiClient.get<any, APIResponse<Playbook>>(`${BASE_URL}/${id}`),

  // 3. 创建 [POST /playbooks]
  create: (data: Partial<Playbook>) => 
    apiClient.post<any, APIResponse<Playbook>>(BASE_URL, data),

  // 4. 更新 [PUT /playbooks/:id]
  update: (id: string | number, data: Partial<Playbook>) => 
    apiClient.put<any, APIResponse<Playbook>>(`${BASE_URL}/${id}`, data),

  // 5. 删除 [DELETE /playbooks/:id]
  delete: (id: string | number) => 
    apiClient.delete<any, APIResponse<null>>(`${BASE_URL}/${id}`),

  // 6. 运行测试 [POST /playbooks/:id/run]
  runTest: (id: string | number, payload: { incident_id: number; dry_run: boolean }) => 
    apiClient.post<any, APIResponse<{ execution_id: number }>>(`${BASE_URL}/${id}/run`, payload),

  // 7. 获取单个剧本的执行历史列表 [GET /playbooks/:id/executions]
  getExecutions: (id: string | number) => 
    apiClient.get<any, APIResponse<PlaybookExecution[]>>(`${BASE_URL}/${id}/executions`),

  // 8. 获取单次执行详情 [GET /playbooks/executions/:exec_id]
  // 用于获取详细 logs，支撑动态补全
  getExecutionDetail: (execId: string | number) => 
    apiClient.get<any, APIResponse<PlaybookExecution>>(`${BASE_URL}/executions/${execId}`),

  // 9. 获取全局执行历史 [GET /playbooks/executions]
  // 对应后端补充的 automation.GET("/executions", ...)
  getGlobalExecutions: (params?: { page?: number; limit?: number }) =>
    apiClient.get<any, APIResponse<PlaybookExecution[]>>(`${BASE_URL}/executions`, { params }),

  // 1. 绑定规则到剧本 [POST /playbooks/:id/bind-rules]
  bindRules: (playbookId: number | string, ruleIds: number[]) =>
    apiClient.post<any, APIResponse<null>>(`${BASE_URL}/${playbookId}/bind-rules`, { rule_ids: ruleIds }),

  // 2. 获取已绑定的规则 [GET /playbooks/:id/rules]
  getBoundRules: (playbookId: number | string) =>
    apiClient.get<any, APIResponse<any[]>>(`${BASE_URL}/${playbookId}/rules`),

  // 3. 解绑规则 [DELETE /playbooks/:id/rules/:rule_id]
  unbindRule: (playbookId: number | string, ruleId: number | string) =>
    apiClient.delete<any, APIResponse<null>>(`${BASE_URL}/${playbookId}/rules/${ruleId}`),
  
};
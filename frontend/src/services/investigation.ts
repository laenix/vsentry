import { apiClient } from "@/lib/api/vsentry-client";
import type { APIResponse } from "@/lib/api/vsentry-client";

export interface InvestigationTemplate {
  id: number;
  name: string;
  description: string;
  logsql: string;
  parameters: string; // 这是一个 JSON 字符串数组，如 '["src_ip", "hostname"]'
}

export interface ExecuteParams {
  template_id: number;
  incident_id?: number;
  params?: Record<string, string>;
}

export interface ExecuteResult {
  logsql: string;
  events: any[];
  count: number;
  context_used: Record<string, string>;
}

export const investigationService = {
  listTemplates: () => 
    apiClient.get<any, APIResponse<InvestigationTemplate[]>>("/investigation/templates"),
  
  // ✅ 新增：模板的 CRUD 管理
  addTemplate: (data: Partial<InvestigationTemplate>) => 
    apiClient.post<any, APIResponse<any>>("/investigation/templates", data),
    
  updateTemplate: (data: Partial<InvestigationTemplate>) => 
    apiClient.put<any, APIResponse<any>>("/investigation/templates", data),
    
  deleteTemplate: (id: number) => 
    apiClient.delete<any, APIResponse<any>>(`/investigation/templates?id=${id}`),

  execute: (data: ExecuteParams) => 
    apiClient.post<any, APIResponse<ExecuteResult>>("/investigation/execute", data),
};
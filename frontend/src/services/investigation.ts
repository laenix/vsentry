import { apiClient } from "@/lib/api/vsentry-client";
import type { APIResponse } from "@/lib/api/vsentry-client";

// InvestigationRule: 来自 Rule Center 的InvestigationRule
export interface InvestigationRule {
  id: number;
  name: string;
  description: string;
  query: string;
  type: "investigation";
  enabled: boolean;
}

// InvestigationPage 使用的指令Type
export interface InvestigationDirective {
  id: number;
  name: string;
  description: string;
  logsql: string;
  parameters: string; // JSON 数Group字符串，如 '["src_ip", "hostname"]'
}

export interface ExecuteParams {
  rule_id: number;
  incident_id?: number;
  params?: Record<string, string>;
}

export interface ExecuteResult {
  logsql: string;
  template_name?: string;
  events: any[];
  count: number;
  context_used: Record<string, string>;
}

// 自动从 LogSQL/Query Medium提取参数 ${xxx}
export function extractParameters(query: string): string[] {
  const paramRegex = /\$\{([^}]+)\}/g;
  const params: string[] = [];
  let match;
  while ((match = paramRegex.exec(query)) !== null) {
    if (!params.includes(match[1])) params.push(match[1]);
  }
  return params;
}

export const investigationService = {
  // 从 Rule Center Get type="investigation" 的Rule
  listRules: () => 
    apiClient.get<any, APIResponse<{ rules: InvestigationRule[] }>>("/rules/list"),

  // ExecuteInvestigation
  execute: (data: ExecuteParams) => 
    apiClient.post<any, APIResponse<ExecuteResult>>("/investigation/execute", data),
};
import { apiClient } from "@/lib/api/vsentry-client";
import type { APIResponse } from "@/lib/api/vsentry-client";

export interface Connector {
  id?: number;
  ID?: number;
  created_at?: string;
  updated_at?: string;
  name: string;
  type: string;
  protocol: string;
  host?: string;
  port?: number;
  username?: string;
  password?: string;
  api_key?: string;
  endpoint?: string;
  is_enabled?: boolean;
  description?: string;
  config?: string;
}

export interface ConnectorTemplate {
  id: string;
  name: string;
  type: string;
  protocol: string;
  default_port: number;
  description: string;
  icon: string;
}

export const connectorService = {
  list: () => apiClient.get<any, APIResponse<Connector[]>>("/connectors/list"),
  
  templates: () => apiClient.get<any, APIResponse<ConnectorTemplate[]>>("/connectors/templates"),
  
  add: (data: Partial<Connector>) => apiClient.post("/connectors/add", data),
  
  update: (data: Partial<Connector>) => apiClient.post("/connectors/update", data),
  
  delete: (id: number) => apiClient.post(`/connectors/delete?id=${id}`),
  
  test: (data: Partial<Connector>) => apiClient.post("/connectors/test", data),
};
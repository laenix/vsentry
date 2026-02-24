import { apiClient } from "@/lib/api/vsentry-client";
import type { APIResponse } from "@/lib/api/vsentry-client";

export interface IngestConfig {
  id?: number;
  ID?: number;
  created_at?: string;
  updated_at?: string;
  name: string;
  endpoint: string;
  type: string;
  source: string;
  _stream_fields?: string;
}

export interface SystemConfig {
  external_url: string;
}

export const ingestService = {
  list: () => apiClient.get<any, APIResponse<IngestConfig[]>>("/ingestmanager/list"),
  
  add: (data: Partial<IngestConfig>) => apiClient.post("/ingestmanager/add", data),
  
  update: (data: Partial<IngestConfig>) => apiClient.post("/ingestmanager/update", data),
  
  delete: (id: number) => apiClient.post(`/ingestmanager/delete?id=${id}`),
  
  getAuth: (id: number) => apiClient.get<any, APIResponse<{token: string}>>(`/ingestmanager/auth/${id}`),
};

export const configService = {
  get: () => apiClient.get<any, APIResponse<SystemConfig>>("/config"),
};
import { apiClient } from "@/lib/api/vsentry-client";
import type { APIResponse } from "@/lib/api/vsentry-client";

export interface CollectorConfig {
  id?: number;
  ID?: number;
  created_at?: string;
  updated_at?: string;
  name: string;
  template_id?: number;
  type: string;
  channels: string;
  sources?: string; // JSON string of sources for Linux
  interval?: number;
  ingest_id?: number;
  endpoint?: string;
  token?: string;
  stream_fields?: string;
  is_enabled?: boolean;
  build_status?: string;
  build_output?: string;
}

export interface CollectorTemplate {
  id: string;
  name: string;
  type: string;
  description: string;
  icon: string;
  channels: string[];
}

export interface IngestConfig {
  ID?: number;
  id?: number;
  name: string;
}

export const collectorService = {
  list: () => apiClient.get<any, APIResponse<CollectorConfig[]>>("/collectors/list"),
  
  templates: () => apiClient.get<any, APIResponse<CollectorTemplate[]>>("/collectors/templates"),
  
  channels: (type: string) => apiClient.get<any, APIResponse<string[]>>(`/collectors/channels?type=${type}`),
  
  // Get available data sources for a collector type (returns array of {type, path, label})
  getSources: (type: string) => apiClient.get<any, APIResponse<any[]>>(`/collectors/channels?type=${type}`),
  
  add: (data: Partial<CollectorConfig>) => apiClient.post("/collectors/add", data),
  
  update: (data: Partial<CollectorConfig>) => apiClient.post("/collectors/update", data),
  
  delete: (id: number) => apiClient.post(`/collectors/delete?id=${id}`),
  
  // 核心修复：绕过 apiClient 的 JSON 拦截器，使用原生 Fetch 处理纯二进制流
  build: async (id: number) => {
    // 如果你的系统有 JWT Token，请从 localStorage 或 Cookie 中获取
    const token = localStorage.getItem("vsentry_token");
    
    const response = await fetch(`/api/collectors/build?id=${id}`, {
      method: 'POST',
      headers: {
        // 必须带上认证头，否则后端会报 401
        'Authorization': `Bearer ${token}` 
      }
    });

    if (!response.ok) {
      throw new Error(`Build failed with status: ${response.status}`);
    }

    // 将响应体作为纯二进制 Blob 提取
    const blob = await response.blob();
    
    // 包装成 { data: blob } 以兼容你 UI 组件中 const res = await ...; window.URL.createObjectURL(res.data) 的写法
    return { data: blob };
  },
  
  ingestAuth: (id: number) => apiClient.get<any, APIResponse<{token: string}>>(`/ingestmanager/auth/${id}`),
};

export const ingestServiceSimple = {
  list: () => apiClient.get<any, APIResponse<IngestConfig[]>>("/ingestmanager/list"),
  getAuth: (id: number) => apiClient.get<any, APIResponse<{token: string}>>(`/ingestmanager/auth/${id}`),
};
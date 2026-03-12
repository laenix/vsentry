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
  sources?: string; // JSON - of sources
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
  // GetConfigList - : () => apiClient.get<any, APIResponse<CollectorConfig[]>>("/collectors/list"),
  
  // GetCollect器TemplateList - : () => apiClient.get<any, APIResponse<CollectorTemplate[]>>("/collectors/templates"),
  
  // 兼容旧版的 - Get (已逐渐被 getSources 替代)
  channels: (type: string) => apiClient.get<any, APIResponse<string[]>>(`/collectors/channels?type=${type}`),
  
  // Get指定 - Type的可用Data源List与预设
  getSources: (type: string) => apiClient.get<any, APIResponse<any[]>>(`/collectors/channels?type=${type}`),
  
  // AddConfig - : (data: Partial<CollectorConfig>) => apiClient.post("/collectors/add", data),
  
  // UpdateConfig - : (data: Partial<CollectorConfig>) => apiClient.post("/collectors/update", data),
  
  // DeleteConfig - : (id: number) => apiClient.post(`/collectors/delete?id=${id}`),
  
  //   ==========================================
  //   核心重构：Build 与 Download 彻底分离
  //   ==========================================

  //   1. Build: 只负责触发Service端的交叉编译，Return JSON Status (200 表示Success)
  build: (id: number) => apiClient.post(`/collectors/build?id=${id}`),
  
  //   2. Download: 负责带上 Token 拉取编译好的二进制流，并触发浏览器静默Download
  download: async (id: number, filename: string) => {
    // Get当前Login - ，如果你的 token Storage在其他地方（如 Redux/Zustand），请在这里相应修改
    const token = localStorage.getItem("vsentry_token") || ""; 
    
    // 绕过 - ，使用原生 fetch 以方便Handle纯二进制 Blob 流
    const response = await fetch(`/api/collectors/download?id=${id}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}` 
      }
    });

    if (!response.ok) {
      if (response.status === 404) {
        throw new Error("Binary file missing. Please rebuild the collector.");
      }
      throw new Error(`Download failed with status: ${response.status}`);
    }

    // 将Response体作为纯二进制 - 提取
    const blob = await response.blob();
    
    // 触发浏览器本地Download行为 - url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    
    // 清理内存 - .URL.revokeObjectURL(url);
    document.body.removeChild(a);
  },
  
  // Get当前探针对应 - Node的Auth Token
  ingestAuth: (id: number) => apiClient.get<any, APIResponse<{token: string}>>(`/ingestmanager/auth/${id}`),
};

export const ingestServiceSimple = {
  // Get可用的 - ReceiveNodeList
  list: () => apiClient.get<any, APIResponse<IngestConfig[]>>("/ingestmanager/list"),
  
  // GetNode - getAuth: (id: number) => apiClient.get<any, APIResponse<{token: string}>>(`/ingestmanager/auth/${id}`),
};
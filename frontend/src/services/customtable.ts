import { apiClient } from "@/lib/api/vsentry-client";
import type { APIResponse } from "@/lib/api/vsentry-client";

export interface CustomTable {
  id?: number;
  ID?: number;
  created_at?: string;
  updated_at?: string;
  name: string;
  stream_fields: string;
  description?: string;
  query?: string;
  is_active?: boolean;
}

export const customTableService = {
  list: () => apiClient.get<any, APIResponse<CustomTable[]>>("/customtables/list"),
  
  add: (data: Partial<CustomTable>) => apiClient.post("/customtables/add", data),
  
  update: (data: Partial<CustomTable>) => apiClient.post("/customtables/update", data),
  
  delete: (id: number) => apiClient.post(`/customtables/delete?id=${id}`),
};
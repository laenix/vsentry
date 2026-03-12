import { apiClient } from "@/lib/api/vsentry-client";
import type { APIResponse } from "@/lib/api/vsentry-client";

// ==================== Data模型Type定义 ====================

export type ParseStatus = 'pending' | 'parsing' | 'completed' | 'failed';

export interface ForensicFile {
  id: number;
  task_id: number;
  file_name: string;
  original_name: string;
  file_type: string;
  file_size: number;
  file_path: string;
  parse_status: ParseStatus;
  parse_message: string;
  event_count: number;
  created_at: string;
  updated_at: string;
}

export interface ForensicTask {
  id: number;
  name: string;
  description: string;
  status: 'open' | 'closed';
  created_at: string;
  updated_at: string;
  // 后端 Preload 关联带出的FileList
  files?: ForensicFile[]; 
}

// ==================== API 交互层 ====================

export const forensicsService = {
  // 1. Case (Task) Manage
  listTasks: () => 
    apiClient.get<any, APIResponse<ForensicTask[]>>("/forensics/tasks"),
    
  createTask: (data: { name: string; description?: string }) => 
    apiClient.post<any, APIResponse<ForensicTask>>("/forensics/tasks", data),
    
  getTask: (id: number | string) => 
    apiClient.get<any, APIResponse<ForensicTask>>(`/forensics/tasks/${id}`),
    
  deleteTask: (id: number | string) => 
    apiClient.delete<any, APIResponse<any>>(`/forensics/tasks/${id}`),

  // 2. EvidenceFile (File) Manage
  uploadFile: (taskId: number | string, file: File) => {
    const formData = new FormData();
    formData.append("task_id", String(taskId));
    formData.append("file", file);

    const timeout = file.size > 50 * 1024 * 1024 ? 300000 : 60000;

    return apiClient.post<any, APIResponse<ForensicFile>>("/forensics/upload", formData, {
      timeout,
      headers: {
        "Content-Type": "multipart/form-data",
      },
    });
  },

  deleteFile: (id: number | string) => 
    apiClient.delete<any, APIResponse<any>>(`/forensics/files/${id}`),

  // 3. ExecuteForensicsRule
  executeRules: (caseId: number, fileId: number, ruleIds: number[]) => 
    apiClient.post<any, APIResponse<any>>("/forensics/execute-rules", {
      case_id: caseId,
      file_id: fileId,
      rule_ids: ruleIds,
    }),
};
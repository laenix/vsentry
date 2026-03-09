import { apiClient } from "@/lib/api/vsentry-client";
import type { APIResponse } from "@/lib/api/vsentry-client";

// ==================== 数据模型类型定义 ====================

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
  // 后端 Preload 关联带出的文件列表
  files?: ForensicFile[]; 
}

// ==================== API 交互层 ====================

export const forensicsService = {
  // 1. 案件 (Task) 管理
  listTasks: () => 
    apiClient.get<any, APIResponse<ForensicTask[]>>("/forensics/tasks"),
    
  createTask: (data: { name: string; description?: string }) => 
    apiClient.post<any, APIResponse<ForensicTask>>("/forensics/tasks", data),
    
  getTask: (id: number | string) => 
    apiClient.get<any, APIResponse<ForensicTask>>(`/forensics/tasks/${id}`),
    
  deleteTask: (id: number | string) => 
    apiClient.delete<any, APIResponse<any>>(`/forensics/tasks/${id}`),

  // 2. 证据文件 (File) 管理
  /**
   * 上传证据文件 (支持大文件，使用 FormData)
   */
  uploadFile: (taskId: number | string, file: File) => {
    const formData = new FormData();
    formData.append("task_id", String(taskId));
    formData.append("file", file);

    // 根据文件大小动态调整超时，100MB 文件大概需要 2-3 分钟
    const timeout = file.size > 50 * 1024 * 1024 ? 300000 : 60000;

    // Axios 在收到 FormData 时，会自动设置界限 (boundary) 并修改 Content-Type
    // 但为了严谨，我们显式声明它
    return apiClient.post<any, APIResponse<ForensicFile>>("/forensics/upload", formData, {
      timeout,
      headers: {
        "Content-Type": "multipart/form-data",
      },
      // 如果你的 apiClient 支持，这里未来甚至可以加上 onUploadProgress 做上传进度条
    });
  },

  deleteFile: (id: number | string) => 
    apiClient.delete<any, APIResponse<any>>(`/forensics/files/${id}`),
};
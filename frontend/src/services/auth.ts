import { apiClient } from "@/lib/api/vsentry-client";
import type { APIResponse } from "@/lib/api/vsentry-client";

export const authService = {
  // 登录接口 (FormData)
  login: async (username: string, password: string) => {
    const formData = new FormData();
    formData.append("name", username);
    formData.append("password", password);

    // 注意：axios 会自动处理 FormData 的 Content-Type
    const res = await apiClient.post<any, APIResponse<{ token: string }>>("/login", formData);
    return res.data; 
  },

  // 获取用户信息
  getUserInfo: async () => {
    return apiClient.post("/user/userinfo");
  }
};
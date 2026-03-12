import { apiClient } from "@/lib/api/vsentry-client";
import type { APIResponse } from "@/lib/api/vsentry-client";

export const authService = {
  // LoginInterface (FormData)
  login: async (username: string, password: string) => {
    const formData = new FormData();
    formData.append("name", username);
    formData.append("password", password);

    // 注意：axios 会自动Handle FormData 的 Content-Type
    const res = await apiClient.post<any, APIResponse<{ token: string }>>("/login", formData);
    return res.data; 
  },

  // GetUserInfo
  getUserInfo: async () => {
    return apiClient.post("/user/userinfo");
  }
};
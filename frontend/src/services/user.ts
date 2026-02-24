import { apiClient } from "@/lib/api/vsentry-client";
import type { APIResponse } from "@/lib/api/vsentry-client";

export interface User {
  ID?: number;
  id?: number;
  user_name?: string;
  username?: string;
}

export interface UserFormData {
  name: string;
  password: string;
}

export const userService = {
  list: () => apiClient.get<any, APIResponse<User[]>>("/users/list"),
  
  add: (data: UserFormData) => apiClient.post("/users/add", null, {
    params: data,
  }),
  
  delete: (id: number) => apiClient.post(`/users/delete?id=${id}`),
  
  updatePassword: (oldPassword: string, newPassword: string) => 
    apiClient.post("/user/changepassword", { old_password: oldPassword, new_password: newPassword }),
};
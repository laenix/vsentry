import axios from "axios";
import { toast } from "sonner";

// 1. 配置 Base URL
const API_BASE_URL = "/api";

export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
});

// 2. 请求拦截器：自动携带 Token
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem("vsentry_token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// 3. 响应拦截器：统一处理错误
apiClient.interceptors.response.use(
  (response) => {
    // 假设后端返回 {code: 200, data: ..., msg: ...}
    // 如果 code 不是 200，视为业务错误
    if (response.data && response.data.code !== 200) {
      toast.error(`Error ${response.data.code}: ${response.data.msg}`);
      return Promise.reject(new Error(response.data.msg));
    }
    return response.data; // 直接返回 data 包装层，或者 response.data.data
  },
  (error) => {
    // 处理 HTTP 状态码错误 (401, 500 等)
    if (error.response?.status === 401) {
      toast.error("Session expired. Please login again.");
      localStorage.removeItem("vsentry_token");
      window.location.href = "/login"; // 强制跳转登录
    } else {
      toast.error(error.message || "Network Error");
    }
    return Promise.reject(error);
  }
);

// 定义通用响应结构
export interface APIResponse<T = any> {
  code: number;
  data: T;
  msg: string;
}
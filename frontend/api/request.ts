// frontend/api/request.ts
import axios, {
  AxiosInstance,
  AxiosError,
  InternalAxiosRequestConfig,
} from "axios";
import type { ApiResponse } from "./types";

// ✅ 深度转换对象键名（递归）
const transformKeys = (obj: any, transformer: (key: string) => string): any => {
  if (obj === null || typeof obj !== "object") return obj;
  if (obj instanceof Date) return obj;
  if (Array.isArray(obj))
    return obj.map((item) => transformKeys(item, transformer));

  const result: any = {};
  for (const key in obj) {
    if (Object.prototype.hasOwnProperty.call(obj, key)) {
      const newKey = transformer(key);
      result[newKey] = transformKeys(obj[key], transformer);
    }
  }
  return result;
};

// 蛇形转驼峰（响应数据）
const toCamelCase = (obj: any): any => {
  return transformKeys(obj, (key) => {
    // 将 snake_case 转换为 camelCase
    return key.replace(/_([a-z])/g, (_, letter) => letter.toUpperCase());
  });
};

const apiBase = (import.meta.env.VITE_API_BASE_URL as string) || "/api/v1";

const request: AxiosInstance = axios.create({
  baseURL: apiBase,
  timeout: 30000,
  headers: {
    "Content-Type": "application/json",
  },
});

// ✅ 请求拦截器：添加 token + 转换请求数据
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // 从 sessionStorage 获取 token
    const token = sessionStorage.getItem("token");
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }

    // 当前后端 JSON/查询参数使用 camelCase 时，不需要转换请求字段名
    return config;
  },
  (error: AxiosError) => {
    return Promise.reject(error);
  },
);

// ✅ 响应拦截器：转换响应数据 + 处理 401 + 提取 data.data
let isRefreshing = false;
let failedQueue: Array<{ resolve: Function; reject: Function }> = [];

const processQueue = (error: any | null, token: string | null = null) => {
  failedQueue.forEach((promise) => {
    if (error) {
      promise.reject(error);
    } else {
      promise.resolve(token);
    }
  });
  failedQueue = [];
};

// ========== ✅ 新增：中文错误消息映射函数 ==========
const getChineseErrorMessage = (
  error: AxiosError<ApiResponse<any>>,
): string => {
  // 网络错误
  if (error.message === "Network Error") {
    return "网络连接失败，请检查网络";
  }
  if (error.code === "ECONNABORTED" || error.message.includes("timeout")) {
    return "请求超时，请稍后重试";
  }

  if (error.response) {
    const status = error.response.status;
    const data = error.response.data;

    // 优先使用后端返回的消息，并尝试翻译
    if (data?.message) {
      const msg = data.message;
      const translations: Record<string, string> = {
        "Permission denied": "权限不足",
        "Invalid credentials": "用户名或密码错误",
        "Token expired": "登录已过期，请重新登录",
        "Invalid token": "无效的登录凭证",
        "Missing authorization token": "缺少授权令牌",
        "Account disabled": "账号已被禁用",
        "Device not found": "设备不存在",
        "Recording not found": "录制文件不存在",
        "Screenshot not found": "截图不存在",
        "User not found": "用户不存在",
        "Role not found": "角色不存在",
        "Username already exists": "用户名已存在",
        "Invalid request": "请求参数错误",
        "Upload session not found": "上传会话不存在",
        "File not found": "文件不存在",
        "Rate limit exceeded": "请求过于频繁，请稍后再试",
        success: "操作成功",
        created: "创建成功",
        updated: "更新成功",
        deleted: "删除成功",
      };

      for (const [eng, chn] of Object.entries(translations)) {
        if (msg.includes(eng)) {
          return chn;
        }
      }
      return msg;
    }

    // 根据状态码返回中文消息
    const statusMessages: Record<number, string> = {
      400: "请求参数错误",
      401: "未授权，请重新登录",
      403: "权限不足，无法执行此操作",
      404: "请求的资源不存在",
      429: "请求过于频繁，请稍后再试",
      500: "服务器内部错误，请稍后重试",
      502: "网关错误",
      503: "服务不可用，请稍后再试",
    };

    return statusMessages[status] || `请求失败 (${status})`;
  }

  return error.message || "未知错误，请稍后重试";
};

request.interceptors.response.use(
  (response) => {
    // 转换响应数据为 camelCase
    let responseData = response.data;
    if (responseData && typeof responseData === "object") {
      responseData = toCamelCase(responseData);
    }

    // 检查业务错误码
    const res = responseData as ApiResponse<any>;

    if (res.code !== 0 && res.code !== undefined) {
      // ✅ 翻译业务错误消息
      const errorMsg = res.message || "请求失败";
      const translations: Record<string, string> = {
        "Permission denied": "权限不足",
        "Invalid credentials": "用户名或密码错误",
        "Token expired": "登录已过期，请重新登录",
      };
      const chineseMsg = translations[errorMsg] || errorMsg;
      return Promise.reject(new Error(chineseMsg));
    }

    // ✅ 如果是文件下载，返回完整响应对象
    if (response.config?.responseType === "blob") {
      return response;
    }

    // ✅ 保持兼容：统一返回 Axios response 对象，业务层通过 response.data 访问真实数据
    if (res.data !== undefined) {
      response.data = res.data;
      return response;
    }

    response.data = responseData;
    return response;
  },
  async (error: AxiosError<ApiResponse<any>>) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & {
      _retry?: boolean;
    };

    // 401 处理：尝试刷新 token
    if (error.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        // 等待刷新完成
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        })
          .then(() => {
            return request(originalRequest);
          })
          .catch((err) => {
            return Promise.reject(err);
          });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      const refreshToken = sessionStorage.getItem("refreshToken");
      if (!refreshToken) {
        clearSessionAndRedirect();
        return Promise.reject(new Error("请重新登录"));
      }

      try {
        const response = await axios.post(`${apiBase}/auth/refresh`, {
          refreshToken,
        });
        let responseData = response.data;

        // 转换响应数据
        if (responseData && typeof responseData === "object") {
          responseData = toCamelCase(responseData);
        }

        if (responseData.code === 0 && responseData.data) {
          const newToken = responseData.data.token;
          const newExpiresAt = responseData.data.expiresAt;

          if (newToken) {
            sessionStorage.setItem("token", newToken);
            sessionStorage.setItem("expiresAt", newExpiresAt);
            processQueue(null, newToken);

            if (originalRequest.headers) {
              originalRequest.headers.Authorization = `Bearer ${newToken}`;
            }
            return request(originalRequest);
          }
        }
        throw new Error("刷新令牌失败");
      } catch (refreshError) {
        processQueue(refreshError, null);
        clearSessionAndRedirect();
        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }

    // ✅ 修改：使用中文错误消息
    const chineseMessage = getChineseErrorMessage(error);
    return Promise.reject(new Error(chineseMessage));
  },
);

const clearSessionAndRedirect = () => {
  sessionStorage.clear();
  // 使用 window.location 跳转，避免循环依赖
  if (window.location.pathname !== "/login") {
    window.location.href = "/login";
  }
};

export default request;

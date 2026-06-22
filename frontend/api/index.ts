import request from "./request";
import type {
  LoginRequest,
  LoginResponse,
  User,
  Role,
} from "./types";

import {
  formatBytes,
  formatDateTime,
  formatDate,
  formatTime,
  getRelativeTime,
  getStatusType,
  getStatusText,
  formatDuration,
  truncateText,
  formatDeviceId,
  formatPercent,
  formatNumber,
  formatISOString,
} from "./format";

export {
  formatBytes,
  formatDateTime,
  formatDate,
  formatTime,
  getRelativeTime,
  getStatusType,
  getStatusText,
  formatDuration,
  truncateText,
  formatDeviceId,
  formatPercent,
  formatNumber,
  formatISOString,
};

export const formatUTC = (date: Date = new Date()): string => date.toISOString();
export const parseUTC = (dateStr: string): Date => new Date(dateStr);

export const authApi = {
  login: async (data: LoginRequest): Promise<{ data: LoginResponse }> => {
    const response = await request.post<any>("/auth/login", data);
    if (
      response.data &&
      typeof response.data === "object" &&
      "data" in response.data
    ) {
      return { data: response.data.data as LoginResponse };
    }
    return { data: response.data as LoginResponse };
  },

  refresh: async (
    refreshToken: string,
  ): Promise<{
    data: { token: string; refreshToken: string; expiresAt: string };
  }> => {
    const response = await request.post<any>("/auth/refresh", { refreshToken });
    if (
      response.data &&
      typeof response.data === "object" &&
      "data" in response.data
    ) {
      return {
        data: response.data.data as {
          token: string;
          refreshToken: string;
          expiresAt: string;
        },
      };
    }
    return {
      data: response.data as {
        token: string;
        refreshToken: string;
        expiresAt: string;
      },
    };
  },

  logout: () => request.post("/auth/logout"),
};

export const userApi = {
  list: (params?: { page?: number; pageSize?: number; keyword?: string }) =>
    request.get<{ items: User[]; total: number }>("/users", { params }),

  create: (data: {
    username: string;
    password: string;
    realName?: string;
    email?: string;
    roleIds?: string[];
  }) => request.post<User>("/users", data),

  get: (id: string) => request.get<User>(`/users/${id}`),

  update: (
    id: string,
    data: {
      realName?: string;
      email?: string;
      status?: number;
      roleIds?: string[];
    },
  ) => request.put<User>(`/users/${id}`, data),

  delete: (id: string) => request.delete(`/users/${id}`),

  resetPassword: (id: string, password: string) =>
    request.post(`/users/${id}/reset-password`, { password }),

  getProfile: () => request.get<User>("/user/profile"),

  updateProfile: (data: { realName?: string; email?: string }) =>
    request.put<User>("/user/profile", data),

  changeOwnPassword: (data: { oldPassword: string; newPassword: string }) =>
    request.post("/user/change-password", data),
};

export const roleApi = {
  list: () => request.get<Role[]>("/roles"),

  create: (data: {
    name: string;
    description?: string;
    permissions: string[];
  }) => request.post<Role>("/roles", data),

  update: (
    id: string,
    data: { name?: string; description?: string; permissions?: string[] },
  ) => request.put<Role>(`/roles/${id}`, data),

  delete: (id: string) => request.delete(`/roles/${id}`),
};

// frontend/store/user.ts
import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { authApi } from "@api/index";

export const useUserStore = defineStore(
  "user",
  () => {
    const token = ref<string>("");
    const refreshToken = ref<string>("");
    const expiresAt = ref<string>("");
    const username = ref<string>("");
    const role = ref<string>("auditor");
    const permissions = ref<string[]>([]);

    const normalizeRoleName = (roleName: string): string => {
      const normalized = roleName.trim().toLowerCase();
      switch (normalized) {
        case "管理员":
        case "administrator":
        case "superadmin":
        case "admin":
          return "admin";
        case "审计员":
        case "auditor":
          return "auditor";
        case "操作员":
        case "operator":
        case "ops":
          return "operator";
        default:
          return "auditor";
      }
    };

    const parseRoleFromToken = (tokenStr: string): string => {
      try {
        if (!tokenStr) return "auditor";
        const parts = tokenStr.split(".");
        if (parts.length !== 3) return "auditor";
        const base64 = parts[1].replace(/-/g, "+").replace(/_/g, "/");
        const padded = base64.padEnd(
          base64.length + ((4 - (base64.length % 4)) % 4),
          "=",
        );
        const bytes = Uint8Array.from(atob(padded), (c) => c.charCodeAt(0));
        const payload = JSON.parse(new TextDecoder().decode(bytes));
        return normalizeRoleName(payload.role || "auditor");
      } catch {
        return "auditor";
      }
    };

    const parsePermissionsFromToken = (tokenStr: string): string[] => {
      try {
        if (!tokenStr) return [];
        const parts = tokenStr.split(".");
        if (parts.length !== 3) return [];
        const base64 = parts[1].replace(/-/g, "+").replace(/_/g, "/");
        const padded = base64.padEnd(
          base64.length + ((4 - (base64.length % 4)) % 4),
          "=",
        );
        const bytes = Uint8Array.from(atob(padded), (c) => c.charCodeAt(0));
        const payload = JSON.parse(new TextDecoder().decode(bytes));
        return payload.permissions || [];
      } catch {
        return [];
      }
    };

    const isAuthenticated = computed(() => {
      if (!token.value) return false;
      if (!expiresAt.value) return true;
      try {
        const expiryDate = new Date(expiresAt.value);
        if (isNaN(expiryDate.getTime())) return true;
        return expiryDate > new Date();
      } catch {
        return true;
      }
    });

    const login = async (usernameVal: string, password: string) => {
      const { data } = await authApi.login({ username: usernameVal, password });

      token.value = data.token;
      refreshToken.value = data.refreshToken;
      expiresAt.value = data.expiresAt;
      username.value = usernameVal;

      const perms = parsePermissionsFromToken(data.token);
      permissions.value = perms;
      role.value = parseRoleFromToken(data.token);

      sessionStorage.setItem("token", data.token);
      sessionStorage.setItem("refreshToken", data.refreshToken);
      sessionStorage.setItem("expiresAt", data.expiresAt);
      sessionStorage.setItem("username", usernameVal);
      sessionStorage.setItem("permissions", JSON.stringify(perms));
      sessionStorage.setItem("role", role.value);
    };

    const restoreSession = () => {
      const storedToken = sessionStorage.getItem("token");
      const storedRefreshToken = sessionStorage.getItem("refreshToken");
      const storedExpiresAt = sessionStorage.getItem("expiresAt");
      const storedUsername = sessionStorage.getItem("username");
      const storedPermissions = sessionStorage.getItem("permissions");

      if (storedToken && storedExpiresAt) {
        const expiryDate = new Date(storedExpiresAt);
        if (expiryDate > new Date()) {
          token.value = storedToken;
          refreshToken.value = storedRefreshToken || "";
          expiresAt.value = storedExpiresAt;
          username.value = storedUsername || "";
          role.value = parseRoleFromToken(storedToken);
          permissions.value = storedPermissions
            ? JSON.parse(storedPermissions)
            : parsePermissionsFromToken(storedToken);
          return true;
        }
      }
      return false;
    };

    const logout = async () => {
      try {
        await authApi.logout();
      } catch {
        // ignore
      }
      token.value = "";
      refreshToken.value = "";
      expiresAt.value = "";
      username.value = "";
      role.value = "auditor";
      permissions.value = [];
      sessionStorage.clear();
    };

    const refresh = async () => {
      if (!refreshToken.value) {
        await logout();
        return false;
      }

      try {
        const { data } = await authApi.refresh(refreshToken.value);
        token.value = data.token;
        expiresAt.value = data.expiresAt;
        permissions.value = parsePermissionsFromToken(data.token);
        role.value = parseRoleFromToken(data.token);
        sessionStorage.setItem("token", data.token);
        sessionStorage.setItem("expiresAt", data.expiresAt);
        sessionStorage.setItem("permissions", JSON.stringify(permissions.value));
        return true;
      } catch {
        await logout();
        return false;
      }
    };

    const hasPermission = (permission: string): boolean => {
      if (role.value === "admin") return true;
      return permissions.value.includes(permission);
    };

    return {
      token,
      refreshToken,
      expiresAt,
      username,
      role,
      permissions,
      isAuthenticated,
      login,
      logout,
      refresh,
      restoreSession,
      hasPermission,
    };
  },
  { persist: false },
);

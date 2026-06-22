// frontend/src/store/permission.ts
import { Directive, App } from "vue";
import { useUserStore } from "@store/user";

/**
 * 检查用户是否拥有指定权限
 */
function hasPermission(permission: string): boolean {
  const userStore = useUserStore();

  if (userStore.role === "admin") {
    return true;
  }

  const userPermissions = userStore.permissions || [];
  return userPermissions.includes(permission);
}

export function hasAnyPermission(permissions: string[]): boolean {
  const userStore = useUserStore();

  if (userStore.role === "admin") {
    return true;
  }

  const userPermissions = userStore.permissions || [];
  return permissions.some((p) => userPermissions.includes(p));
}

export function hasAllPermissions(permissions: string[]): boolean {
  const userStore = useUserStore();

  if (userStore.role === "admin") {
    return true;
  }

  const userPermissions = userStore.permissions || [];
  return permissions.every((p) => userPermissions.includes(p));
}

export const permissionDirective: Directive = {
  mounted(el, binding) {
    // ✅ 关键修复：检查 el 本身是否存在
    if (!el || !el.parentNode) {
      return;
    }

    const { value } = binding;

    if (!value) {
      return;
    }

    let hasPerm = false;
    if (typeof value === "string") {
      hasPerm = hasPermission(value);
    } else if (Array.isArray(value)) {
      hasPerm = hasAnyPermission(value);
    }

    if (!hasPerm) {
      el.parentNode.removeChild(el);
    }
  },

  // ✅ 添加 updated 钩子，处理动态权限变化
  updated(el, binding) {
    if (!el || !el.parentNode) {
      return;
    }

    const { value } = binding;

    if (!value) {
      return;
    }

    let hasPerm = false;
    if (typeof value === "string") {
      hasPerm = hasPermission(value);
    } else if (Array.isArray(value)) {
      hasPerm = hasAnyPermission(value);
    }

    if (!hasPerm) {
      el.parentNode.removeChild(el);
    }
  },
};

export function setupPermissionDirective(app: App) {
  app.directive("permission", permissionDirective);
}

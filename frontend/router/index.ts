import { createRouter, createWebHistory } from "vue-router";
import { useUserStore } from "@store/user";

const APP_TITLE = "DHPG管理后台";

const routes = [
  {
    path: "/login",
    name: "Login",
    component: () => import("@views/Login.vue"),
    meta: { requiresAuth: false, title: "登录" },
  },
  {
    path: "/",
    component: () => import("@layouts/MainLayout.vue"),
    meta: { requiresAuth: true },
    children: [
      {
        path: "",
        redirect: "/site-management",
      },
      {
        path: "site-management",
        name: "SiteManagement",
        component: () => import("@views/SiteManagement.vue"),
        meta: {
          title: "出款站点",
          icon: "OfficeBuilding",
          permission: "site:view",
        },
      },
      {
        path: "site-stats",
        name: "SiteStats",
        component: () => import("@views/SiteStats.vue"),
        meta: {
          title: "出款统计",
          icon: "TrendCharts",
          permission: "stats:view",
        },
      },
      {
        path: "monthly-stats",
        name: "MonthlyStats",
        component: () => import("@views/MonthlyStats.vue"),
        meta: {
          title: "出款汇总",
          icon: "DataAnalysis",
          permission: "stats:view",
        },
      },
      {
        path: "attendance",
        redirect: "/attendance/records",
      },
      {
        path: "attendance/records",
        name: "AttendanceRecords",
        component: () => import("@views/attendance/AttendanceRecords.vue"),
        meta: {
          title: "考勤查看",
          icon: "Calendar",
          permission: "attendance:view",
        },
      },
      {
        path: "attendance/performance",
        name: "PerformanceReview",
        component: () => import("@views/attendance/PerformanceReview.vue"),
        meta: {
          title: "绩效考核",
          icon: "TrendCharts",
          permission: "attendance:view",
        },
      },
      {
        path: "attendance/penalty",
        name: "PenaltyDetails",
        component: () => import("@views/attendance/PenaltyDetails.vue"),
        meta: {
          title: "罚款详情",
          icon: "Warning",
          permission: "attendance:view",
        },
      },
      {
        path: "users",
        name: "Users",
        component: () => import("@views/Users.vue"),
        meta: { title: "用户管理", icon: "User", permission: "user:view" },
      },
      {
        path: "roles",
        name: "Roles",
        component: () => import("@views/Roles.vue"),
        meta: { title: "角色管理", icon: "Avatar", permission: "role:view" },
      },
      {
        path: "profile",
        name: "Profile",
        component: () => import("@views/Profile.vue"),
        meta: { title: "个人资料", icon: "UserFilled" },
      },
      {
        path: "operation-logs",
        name: "OperationLogs",
        component: () => import("@views/OperationLogs.vue"),
        meta: {
          title: "操作日志",
          icon: "Document",
          permission: "operation_log:view",
        },
      },
    ],
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

function checkPermission(permission: string): boolean {
  const userStore = useUserStore();
  if (!userStore.isAuthenticated) return false;
  if (userStore.role === "admin") return true;
  return (userStore.permissions || []).includes(permission);
}

const setPageTitle = (to: { meta?: { title?: string } }) => {
  const pageTitle = to.meta?.title as string;
  document.title = pageTitle ? `${pageTitle} - ${APP_TITLE}` : APP_TITLE;
};

router.beforeEach((to, _from, next) => {
  const userStore = useUserStore();

  if (to.meta.requiresAuth !== false && !userStore.isAuthenticated) {
    if (!userStore.restoreSession()) {
      next("/login");
      return;
    }
  }

  if (to.path === "/login" && userStore.isAuthenticated) {
    next("/site-management");
    return;
  }

  const requiredPermission = to.meta.permission as string;
  if (requiredPermission && userStore.isAuthenticated) {
    if (!checkPermission(requiredPermission)) {
      next("/profile");
      return;
    }
  }

  setPageTitle(to);
  next();
});

export default router;

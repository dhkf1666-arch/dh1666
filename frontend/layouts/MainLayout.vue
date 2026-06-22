<template>
  <el-container class="layout-container">
    <el-aside :width="isCollapse ? '64px' : '220px'" class="sidebar">
      <div class="logo">
        <img src="/dh.png" alt="Logo" class="logo-img" />
        <span v-if="!isCollapse">DHPG管理后台</span>
      </div>

      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapse"
        :collapse-transition="false"
        router
        unique-opened
        :default-openeds="defaultOpeneds"
        @open="handleMenuOpen"
        @close="handleMenuClose"
        background-color="#001529"
        text-color="#fff"
        active-text-color="#409eff"
      >
        <el-sub-menu index="payment-group" v-if="hasPaymentPermission">
          <template #title>
            <el-icon><Money /></el-icon>
            <span>出款管理</span>
          </template>
          <el-menu-item index="/site-management" v-permission="'site:view'">
            <el-icon><OfficeBuilding /></el-icon>
            <span>出款站点</span>
          </el-menu-item>
          <el-menu-item index="/site-stats" v-permission="'stats:view'">
            <el-icon><TrendCharts /></el-icon>
            <span>出款统计</span>
          </el-menu-item>
          <el-menu-item index="/monthly-stats" v-permission="'stats:view'">
            <el-icon><DataAnalysis /></el-icon>
            <span>出款汇总</span>
          </el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="attendance-group" v-if="hasAttendancePermission">
          <template #title>
            <el-icon><Calendar /></el-icon>
            <span>考勤绩效</span>
          </template>
          <el-menu-item index="/attendance/records" v-permission="'attendance:view'">
            <el-icon><Calendar /></el-icon>
            <span>考勤查看</span>
          </el-menu-item>
          <el-menu-item index="/attendance/performance" v-permission="'attendance:view'">
            <el-icon><TrendCharts /></el-icon>
            <span>绩效考核</span>
          </el-menu-item>
          <el-menu-item index="/attendance/penalty" v-permission="'attendance:view'">
            <el-icon><Warning /></el-icon>
            <span>罚款详情</span>
          </el-menu-item>
        </el-sub-menu>

        <el-menu-item index="/operation-logs" v-permission="'operation_log:view'">
          <el-icon><Document /></el-icon>
          <span>操作日志</span>
        </el-menu-item>

        <el-sub-menu index="system" v-if="hasSystemPermission">
          <template #title>
            <el-icon><Setting /></el-icon>
            <span>系统管理</span>
          </template>
          <el-menu-item index="/users" v-permission="'user:view'">
            <el-icon><User /></el-icon>
            <span>用户管理</span>
          </el-menu-item>
          <el-menu-item index="/roles" v-permission="'role:view'">
            <el-icon><Avatar /></el-icon>
            <span>角色管理</span>
          </el-menu-item>
          <el-menu-item index="/profile">
            <el-icon><UserFilled /></el-icon>
            <span>个人资料</span>
          </el-menu-item>
        </el-sub-menu>

        <el-menu-item index="/profile" v-if="!hasSystemPermission">
          <el-icon><UserFilled /></el-icon>
          <span>个人资料</span>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="header">
        <div class="header-left">
          <el-icon class="collapse-btn" @click="toggleCollapse">
            <Fold v-if="!isCollapse" />
            <Expand v-else />
          </el-icon>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/site-management' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item
              v-for="(item, index) in breadcrumbs"
              :key="index"
              :to="index === breadcrumbs.length - 1 ? undefined : item.path"
            >
              {{ item.title }}
            </el-breadcrumb-item>
          </el-breadcrumb>
        </div>

        <div class="header-right">
          <el-dropdown @command="handleCommand" class="user-dropdown">
            <div class="user-info">
              <el-avatar :size="36" :src="userAvatar" class="user-avatar" />
              <span class="username">{{ userStore.username || "Admin" }}</span>
              <el-icon class="arrow-icon"><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">
                  <el-icon><User /></el-icon>
                  个人资料
                </el-dropdown-item>
                <el-dropdown-item command="changePassword">
                  <el-icon><Lock /></el-icon>
                  修改密码
                </el-dropdown-item>
                <el-dropdown-item command="logout" divided>
                  <el-icon><SwitchButton /></el-icon>
                  退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <div class="tabs-section" v-if="visitedViews.length > 0">
        <div class="tabs-container">
          <div class="tabs-scroll">
            <div
              v-for="item in visitedViews"
              :key="item.path"
              class="tab-item"
              :class="{ active: activeTab === item.path }"
              @click="handleTabClick(item.path)"
              @contextmenu.prevent="(e) => handleContextMenu(e, item.path)"
              @mouseenter="hoveredTab = item.path"
              @mouseleave="hoveredTab = null"
            >
              <span class="tab-icon">
                <el-icon><component :is="getTabIcon(item.path)" /></el-icon>
              </span>
              <span class="tab-title">{{ item.title }}</span>
              <span
                v-if="hoveredTab === item.path || activeTab === item.path"
                class="tab-close"
                @click.stop="removeTab(item.path)"
                title="关闭"
              >
                <el-icon><Close /></el-icon>
              </span>
            </div>
          </div>
          <div class="tabs-actions" v-if="visitedViews.length > 0">
            <el-dropdown @command="handleTabsCommand" trigger="click" placement="bottom-end">
              <el-button :icon="ArrowDown" size="small" circle />
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="close-current" :disabled="visitedViews.length === 1">
                    <el-icon><Close /></el-icon>
                    关闭当前
                  </el-dropdown-item>
                  <el-dropdown-item command="close-other" :disabled="visitedViews.length === 1">
                    <el-icon><Remove /></el-icon>
                    关闭其他
                  </el-dropdown-item>
                  <el-dropdown-item command="close-all">
                    <el-icon><FolderDelete /></el-icon>
                    关闭所有
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
      </div>

      <div
        v-if="contextMenuVisible"
        class="tab-context-menu"
        :style="{ left: contextMenuX + 'px', top: contextMenuY + 'px' }"
        @click.stop
      >
        <div class="context-menu-item" @click="closeCurrentTab">
          <el-icon><Close /></el-icon>
          <span>关闭当前</span>
        </div>
        <div class="context-menu-item" @click="closeOtherTabs">
          <el-icon><Remove /></el-icon>
          <span>关闭其他</span>
        </div>
        <div class="context-menu-item" @click="closeAllTabs">
          <el-icon><FolderDelete /></el-icon>
          <span>关闭所有</span>
        </div>
      </div>

      <el-main class="main-content">
        <router-view v-slot="{ Component, route: currentRoute }">
          <transition name="fade-transform" mode="out-in">
            <keep-alive :include="cachedViews">
              <component :is="Component" :key="currentRoute.path" />
            </keep-alive>
          </transition>
        </router-view>
      </el-main>
    </el-container>

    <div v-if="isMobile && !isCollapse" class="mobile-mask" @click="toggleCollapse"></div>

    <el-dialog v-model="changePasswordVisible" title="修改密码" width="450px" class="animate-dialog">
      <el-form :model="passwordForm" :rules="passwordRules" ref="passwordFormRef" label-width="100px">
        <el-form-item label="当前密码" prop="oldPassword">
          <el-input v-model="passwordForm.oldPassword" type="password" show-password placeholder="请输入当前密码" />
        </el-form-item>
        <el-form-item label="新密码" prop="newPassword">
          <el-input v-model="passwordForm.newPassword" type="password" show-password placeholder="请输入新密码（至少6位）" />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input v-model="passwordForm.confirmPassword" type="password" show-password placeholder="请再次输入新密码" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="changePasswordVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmChangePassword" :loading="changePasswordLoading">确定</el-button>
      </template>
    </el-dialog>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed, watch, reactive } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import {
  Fold,
  Expand,
  ArrowDown,
  Close,
  Remove,
  FolderDelete,
  Document,
  Setting,
  User,
  Avatar,
  UserFilled,
  Lock,
  SwitchButton,
  TrendCharts,
  Warning,
  Money,
  Calendar,
  OfficeBuilding,
  DataAnalysis,
} from "@element-plus/icons-vue";
import { useUserStore } from "@store/user";
import { userApi } from "@api/index";

const route = useRoute();
const router = useRouter();
const userStore = useUserStore();
const isCollapse = ref(false);

const roleAvatarMap: Record<string, string> = {
  admin: "/1.png",
  operator: "/2.jpg",
  auditor: "/2.jpg",
  default: "1.png",
};

const userAvatar = computed(
  () => roleAvatarMap[userStore.role] || roleAvatarMap.default,
);

const changePasswordVisible = ref(false);
const changePasswordLoading = ref(false);
const passwordFormRef = ref();
const passwordForm = reactive({
  oldPassword: "",
  newPassword: "",
  confirmPassword: "",
});

const passwordRules = {
  oldPassword: [{ required: true, message: "请输入当前密码", trigger: "blur" }],
  newPassword: [
    { required: true, message: "请输入新密码", trigger: "blur" },
    { min: 6, message: "密码长度不能小于6位", trigger: "blur" },
  ],
  confirmPassword: [
    { required: true, message: "请确认密码", trigger: "blur" },
    {
      validator: (_: unknown, value: string, callback: (err?: Error) => void) => {
        if (value !== passwordForm.newPassword) callback(new Error("两次输入的密码不一致"));
        else callback();
      },
      trigger: "blur",
    },
  ],
};

const activeMenu = computed(() => route.path);

const getDefaultOpeneds = (): string[] => {
  const saved = localStorage.getItem("openedMenus");
  if (saved) {
    try {
      return JSON.parse(saved);
    } catch {
      return ["payment-group"];
    }
  }
  return ["payment-group"];
};

const defaultOpeneds = ref<string[]>(getDefaultOpeneds());

const saveOpenedMenus = (openeds: string[]) => {
  localStorage.setItem("openedMenus", JSON.stringify(openeds));
};

const handleMenuOpen = (index: string) => {
  defaultOpeneds.value = [index];
  saveOpenedMenus(defaultOpeneds.value);
};

const handleMenuClose = (index: string) => {
  if (defaultOpeneds.value[0] === index) {
    defaultOpeneds.value = [];
    saveOpenedMenus(defaultOpeneds.value);
  }
};

const breadcrumbs = computed(() => {
  const items: { path: string; title: string }[] = [];
  for (let i = 1; i < route.matched.length; i++) {
    const title = route.matched[i].meta?.title as string;
    if (title) items.push({ path: route.matched[i].path, title });
  }
  if (items.length === 0 && route.meta?.title) {
    items.push({ path: route.path, title: route.meta.title as string });
  }
  return items;
});

const hasSystemPermission = computed(
  () => userStore.hasPermission("user:view") || userStore.hasPermission("role:view"),
);
const hasAttendancePermission = computed(() => userStore.hasPermission("attendance:view"));
const hasPaymentPermission = computed(
  () => userStore.hasPermission("site:view") || userStore.hasPermission("stats:view"),
);

interface VisitedView {
  path: string;
  title: string;
  name?: string;
}

const visitedViews = ref<VisitedView[]>([]);
const activeTab = ref("");
const cachedViews = ref<string[]>([]);
const hoveredTab = ref<string | null>(null);
const contextMenuVisible = ref(false);
const contextMenuX = ref(0);
const contextMenuY = ref(0);
const currentContextMenuPath = ref("");
const isMobile = ref(false);

const iconMap: Record<string, unknown> = {
  "/site-management": OfficeBuilding,
  "/site-stats": TrendCharts,
  "/monthly-stats": DataAnalysis,
  "/attendance/records": Calendar,
  "/attendance/performance": TrendCharts,
  "/attendance/penalty": Warning,
  "/users": User,
  "/roles": Avatar,
  "/profile": UserFilled,
  "/operation-logs": Document,
};

const getTabIcon = (path: string) => iconMap[path] || Document;

const addView = (view: VisitedView) => {
  if (!visitedViews.value.some((v) => v.path === view.path)) {
    visitedViews.value.push(view);
  }
};

const removeTab = (targetPath: string) => {
  const index = visitedViews.value.findIndex((v) => v.path === targetPath);
  if (index !== -1) {
    visitedViews.value.splice(index, 1);
    if (targetPath === route.path && visitedViews.value.length > 0) {
      const nextPath =
        visitedViews.value[index]?.path || visitedViews.value[index - 1]?.path;
      if (nextPath) router.push(nextPath);
    }
  }
};

const closeCurrentTab = () => {
  if (currentContextMenuPath.value) removeTab(currentContextMenuPath.value);
  else if (route.path) removeTab(route.path);
  contextMenuVisible.value = false;
};

const closeOtherTabs = () => {
  const currentPath = currentContextMenuPath.value || route.path;
  const current = visitedViews.value.find((v) => v.path === currentPath);
  visitedViews.value = current ? [current] : [];
  contextMenuVisible.value = false;
};

const closeAllTabs = () => {
  visitedViews.value = [];
  router.push("/site-management");
  contextMenuVisible.value = false;
};

const handleTabsCommand = (command: string) => {
  if (command === "close-current") removeTab(route.path);
  else if (command === "close-other") {
    const current = visitedViews.value.find((v) => v.path === route.path);
    visitedViews.value = current ? [current] : [];
  } else if (command === "close-all") {
    visitedViews.value = [];
    router.push("/site-management");
  }
};

const handleTabClick = (path: string) => router.push(path);

const handleContextMenu = (e: MouseEvent, path: string) => {
  e.preventDefault();
  currentContextMenuPath.value = path;
  contextMenuX.value = e.clientX;
  contextMenuY.value = e.clientY;
  contextMenuVisible.value = true;
  const closeMenu = () => {
    contextMenuVisible.value = false;
    document.removeEventListener("click", closeMenu);
  };
  setTimeout(() => document.addEventListener("click", closeMenu), 10);
};

const addCacheView = (name: string) => {
  if (name && !cachedViews.value.includes(name)) cachedViews.value.push(name);
};

const toggleCollapse = () => {
  isCollapse.value = !isCollapse.value;
};

const handleCommand = async (command: string) => {
  if (command === "logout") {
    try {
      await userStore.logout();
      sessionStorage.clear();
      localStorage.removeItem("openedMenus");
      ElMessage.success("已退出登录");
      window.location.href = "/login";
    } catch {
      sessionStorage.clear();
      window.location.href = "/login";
    }
  } else if (command === "profile") {
    router.push("/profile");
  } else if (command === "changePassword") {
    changePasswordVisible.value = true;
  }
};

const confirmChangePassword = async () => {
  if (!passwordFormRef.value) return;
  try {
    await passwordFormRef.value.validate();
  } catch {
    return;
  }
  changePasswordLoading.value = true;
  try {
    await userApi.changeOwnPassword({
      oldPassword: passwordForm.oldPassword,
      newPassword: passwordForm.newPassword,
    });
    ElMessage.success("密码修改成功，请重新登录");
    changePasswordVisible.value = false;
    setTimeout(() => {
      userStore.logout();
      router.push("/login");
    }, 1500);
  } catch (error: unknown) {
    const message = error instanceof Error ? error.message : "修改密码失败";
    ElMessage.error(message);
  } finally {
    changePasswordLoading.value = false;
  }
};

watch(
  route,
  (to) => {
    const title = (to.meta.title as string) || (to.name as string) || "未知";
    addView({ path: to.path, title, name: to.name as string });
    activeTab.value = to.path;
    if (to.meta.keepAlive && to.name) addCacheView(to.name as string);
  },
  { immediate: true },
);
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.sidebar {
  transition: width 0.3s;
  overflow-x: hidden;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%) !important;
}

.sidebar .el-menu,
.sidebar .el-menu--inline {
  background: transparent !important;
}

.sidebar .el-menu-item,
.sidebar .el-sub-menu__title,
.sidebar .el-menu--inline .el-menu-item {
  background-color: transparent !important;
}

.sidebar :deep(.el-menu-item):hover,
.sidebar :deep(.el-sub-menu__title):hover,
.sidebar :deep(.el-menu--inline .el-menu-item):hover {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%) !important;
  color: white !important;
  border-radius: 8px !important;
  margin: 4px 8px !important;
  width: calc(100% - 16px) !important;
}

.sidebar :deep(.el-menu-item.is-active),
.sidebar :deep(.el-menu--inline .el-menu-item.is-active) {
  background: transparent !important;
  color: #409eff !important;
}

.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: #fff;
  font-size: 18px;
  font-weight: bold;
  border-bottom: 1px solid #021423;
}

.header {
  background-color: #fff;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 20px;
}

.collapse-btn {
  font-size: 20px;
  cursor: pointer;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 20px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 24px;
}

.user-info:hover {
  background: #f5f5f5;
}

.username {
  font-size: 14px;
  color: #333;
  max-width: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tabs-section {
  margin-top: 1px;
  background: white;
  border-bottom: 1px solid #eef2f6;
}

.tabs-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 20px 0;
}

.tabs-scroll {
  display: flex;
  align-items: center;
  gap: 6px;
  overflow-x: auto;
  flex: 1;
}

.tab-item {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: #f9fafc;
  border-radius: 12px 12px 8px 8px;
  cursor: pointer;
  font-size: 13px;
  color: #5a6874;
  white-space: nowrap;
  border: 1px solid #edf2f7;
}

.tab-item.active {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-color: transparent;
}

.tab-close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border-radius: 6px;
  font-size: 12px;
}

.main-content {
  background-color: #f5f7fb;
  padding: 20px;
  overflow-y: auto;
}

.fade-transform-leave-active,
.fade-transform-enter-active {
  transition: all 0.3s ease;
}

.fade-transform-enter-from {
  opacity: 0;
  transform: translateX(24px);
}

.fade-transform-leave-to {
  opacity: 0;
  transform: translateX(-24px);
}

.mobile-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 999;
}

.tab-context-menu {
  position: fixed;
  background: white;
  border-radius: 14px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  padding: 8px 0;
  min-width: 150px;
  z-index: 10000;
}

.context-menu-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 18px;
  cursor: pointer;
  font-size: 13px;
}

.context-menu-item:hover {
  background: #f0f7ff;
  color: #409eff;
}
</style>

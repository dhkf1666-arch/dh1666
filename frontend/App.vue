<template>
  <router-view v-slot="{ Component }">
    <transition name="fade" mode="out-in">
      <component :is="Component" />
    </transition>
  </router-view>
</template>

<script setup lang="ts">
import { onMounted } from "vue";
import { useUserStore } from "@store/user";
import { ElMessage } from "element-plus";

const userStore = useUserStore();

onMounted(() => {
  console.log("App.vue mounted, isAuthenticated:", userStore.isAuthenticated);
  userStore.restoreSession();
});

// 全局错误处理
window.addEventListener("error", (event: ErrorEvent) => {
  console.error("全局错误捕获:", event.error);
  if (event.error instanceof Error) {
    ElMessage.error(`应用错误: ${event.error.message}`);
  }
});

// 处理未捕获的 Promise 拒绝
window.addEventListener(
  "unhandledrejection",
  (event: PromiseRejectionEvent) => {
    console.error("未处理的 Promise 拒绝:", event.reason);
    ElMessage.error(`请求失败: ${event.reason?.message || event.reason}`);
  },
);
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family:
    -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue",
    Arial, sans-serif;
  background-color: #f0f2f5;
}

/* 页面切换动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* 卡片悬停动画 */
.card-hover {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.card-hover:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

/* 按钮过渡动画 */
.el-button {
  transition: all 0.2s ease !important;
}

.el-button:hover {
  transform: translateY(-1px);
}

/* 链接按钮悬停效果 */
.el-button.is-link {
  transition: all 0.2s ease !important;
}

.el-button.is-link:hover {
  transform: translateX(2px);
}

/* 表格行悬停效果 */
.el-table__row {
  transition: all 0.2s ease;
}

.el-table__row:hover {
  transform: scale(1.01);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

/* 卡片阴影优化 */
.el-card {
  transition: all 0.3s ease;
}

/* 对话框动画 */
.el-dialog {
  animation: dialogFadeIn 0.3s ease;
}

@keyframes dialogFadeIn {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

/* 加载动画优化 */
.el-loading-mask {
  border-radius: 8px;
}

/* 标签过渡动画 */
.el-tag {
  transition: all 0.2s ease;
}

.el-tag:hover {
  transform: scale(1.02);
}

/* 下拉菜单动画 */
.el-dropdown-menu {
  animation: dropdownFadeIn 0.2s ease;
}

@keyframes dropdownFadeIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 消息提示动画 */
.el-message {
  animation: messageSlideIn 0.3s ease !important;
}

@keyframes messageSlideIn {
  from {
    opacity: 0;
    transform: translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 滚动条样式 */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 4px;
}

::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 4px;
  transition: background 0.2s;
}

::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

/* 移动端适配 */
@media (max-width: 768px) {
  ::-webkit-scrollbar {
    width: 4px;
    height: 4px;
  }

  .el-dialog {
    width: 95% !important;
    margin: 20px auto !important;
  }

  .el-message {
    top: 20px !important;
  }
}

/* 平板适配 */
@media (min-width: 769px) and (max-width: 1024px) {
  .el-dialog {
    width: 80% !important;
  }
}
</style>

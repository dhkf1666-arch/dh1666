<template>
  <div class="profile-page">
    <!-- 背景装饰 -->
    <div class="bg-decoration">
      <div class="blob blob-1"></div>
      <div class="blob blob-2"></div>
      <div class="blob blob-3"></div>
    </div>

    <div class="profile-container">
      <!-- 左侧：用户信息卡片 -->
      <div class="profile-sidebar">
        <el-card class="user-card" shadow="never">
          <!-- 用户头像区域 -->
          <div class="user-header">
            <div class="avatar-container">
              <el-avatar :size="100" :src="userAvatar" class="user-avatar">
                <el-icon><User /></el-icon>
              </el-avatar>
              <div class="avatar-ring"></div>
            </div>

            <div class="user-name-wrapper">
              <h2 class="user-name">{{ profile.username }}</h2>
              <div class="status-badge" :class="getStatusClass">
                <span class="status-dot"></span>
                <span class="status-text">{{ getStatusText }}</span>
              </div>
            </div>

            <div class="user-role-badge" :class="userStore.role">
              <el-icon><Star /></el-icon>
              <span>{{ roleText }}</span>
            </div>
            <p class="user-desc">{{ roleDescription }}</p>
          </div>

          <!-- 用户统计信息 -->
          <div class="user-stats">
            <div class="stat-item">
              <div class="stat-value">{{ stats.loginCount || 1 }}</div>
              <div class="stat-label">登录次数</div>
            </div>
            <div class="stat-divider"></div>
            <div class="stat-item">
              <div class="stat-value">{{ stats.roleCount || 1 }}</div>
              <div class="stat-label">角色数量</div>
            </div>
            <div class="stat-divider"></div>
            <div class="stat-item">
              <div class="stat-value">{{ stats.permissionCount || 0 }}</div>
              <div class="stat-label">权限数量</div>
            </div>
          </div>

          <el-divider />

          <!-- 用户基本信息 -->
          <div class="info-list">
            <div class="info-item">
              <div class="info-icon">
                <el-icon><Calendar /></el-icon>
              </div>
              <div class="info-content">
                <div class="info-label">注册时间</div>
                <div class="info-value">
                  {{ formatDate(profile.createdAt) }}
                </div>
              </div>
            </div>
            <div class="info-item">
              <div class="info-icon">
                <el-icon><Clock /></el-icon>
              </div>
              <div class="info-content">
                <div class="info-label">最后登录</div>
                <div class="info-value">
                  {{ formatDate(profile.lastLogin) || "首次登录" }}
                </div>
              </div>
            </div>
            <div class="info-item">
              <div class="info-icon">
                <el-icon><Star /></el-icon>
              </div>
              <div class="info-content">
                <div class="info-label">权限数量</div>
                <div class="info-value">{{ stats.permissionCount || 0 }} 项权限</div>
              </div>
            </div>
          </div>
        </el-card>
      </div>

      <!-- 右侧：编辑区域 -->
      <div class="profile-main">
        <!-- 个人信息编辑卡片 -->
        <el-card class="edit-card" shadow="never">
          <template #header>
            <div class="card-header">
              <div class="header-left">
                <div class="header-icon">
                  <el-icon><UserFilled /></el-icon>
                </div>
                <div>
                  <h3 class="card-title">个人信息</h3>
                  <p class="card-subtitle">管理您的基本信息</p>
                </div>
              </div>
            </div>
          </template>

          <el-form :model="profile" label-width="100px" class="modern-form">
            <el-form-item label="用户名">
              <div class="form-field">
                <el-input
                  v-model="profile.username"
                  disabled
                  class="modern-input"
                >
                  <template #prefix>
                    <el-icon><User /></el-icon>
                  </template>
                </el-input>
              </div>
            </el-form-item>

            <el-form-item label="姓名">
              <div class="form-field">
                <el-input
                  v-model="profile.realName"
                  :disabled="!isAdmin"
                  placeholder="请输入您的姓名"
                  class="modern-input"
                >
                  <template #prefix>
                    <el-icon><Edit /></el-icon>
                  </template>
                </el-input>
              </div>
            </el-form-item>

            <el-form-item label="邮箱">
              <div class="form-field">
                <el-input
                  v-model="profile.email"
                  :disabled="!isAdmin"
                  placeholder="请输入邮箱地址"
                  class="modern-input"
                >
                  <template #prefix>
                    <el-icon><Message /></el-icon>
                  </template>
                </el-input>
              </div>
            </el-form-item>

            <!-- <el-form-item v-if="!isAdmin">
              <div class="permission-alert">
                <el-alert type="info" :closable="false" show-icon>
                  <template #title>
                    <span
                      >💡
                      个人信息修改权限仅限管理员，如需修改请联系系统管理员</span
                    >
                  </template>
                </el-alert>
              </div>
            </el-form-item> -->

            <el-form-item>
              <el-button
                type="primary"
                @click="updateProfile"
                :loading="updateLoading"
                :disabled="!isAdmin"
                class="submit-btn"
              >
                <el-icon><Check /></el-icon>
                保存更改
              </el-button>
              <!-- <el-button v-if="!isAdmin" disabled class="disabled-submit">
                <el-icon><Lock /></el-icon>
                无权限编辑
              </el-button> -->
              <el-form-item v-if="!isAdmin">
                <div class="permission-alert">
                  <el-alert type="info" :closable="false" show-icon>
                    <template #title>
                      <span
                        >💡
                        个人信息修改权限仅限管理员，如需修改请联系系统管理员</span
                      >
                    </template>
                  </el-alert>
                </div>
              </el-form-item>
            </el-form-item>
          </el-form>
        </el-card>

        <!-- 修改密码卡片 -->
        <el-card class="password-card" shadow="never">
          <template #header>
            <div class="card-header">
              <div class="header-left">
                <div class="header-icon">
                  <el-icon><Lock /></el-icon>
                </div>
                <div>
                  <h3 class="card-title">修改密码</h3>
                  <p class="card-subtitle">定期更换密码以保护账户安全</p>
                </div>
              </div>
              <div class="header-right">
                <el-tag type="warning" effect="plain" size="small"
                  >安全建议</el-tag
                >
              </div>
            </div>
          </template>

          <el-form
            :model="passwordForm"
            :rules="passwordRules"
            ref="passwordFormRef"
            label-width="100px"
            class="modern-form"
          >
            <el-form-item label="当前密码" prop="oldPassword">
              <div class="form-field">
                <el-input
                  v-model="passwordForm.oldPassword"
                  type="password"
                  show-password
                  placeholder="请输入当前密码"
                  class="modern-input"
                >
                  <template #prefix>
                    <el-icon><Key /></el-icon>
                  </template>
                </el-input>
              </div>
            </el-form-item>

            <el-form-item label="新密码" prop="newPassword">
              <div class="form-field">
                <el-input
                  v-model="passwordForm.newPassword"
                  type="password"
                  show-password
                  placeholder="请输入新密码（至少6位）"
                  class="modern-input"
                >
                  <template #prefix>
                    <el-icon><Lock /></el-icon>
                  </template>
                </el-input>
              </div>
            </el-form-item>

            <!-- 密码强度指示器 -->
            <el-form-item v-if="passwordForm.newPassword" label=" ">
              <div class="password-strength">
                <div class="strength-header">
                  <span class="strength-label">密码强度</span>
                  <span class="strength-value" :class="strengthClass">{{
                    strengthText
                  }}</span>
                </div>
                <div class="strength-bar-container">
                  <div
                    class="strength-bar"
                    :style="{ width: strengthPercent + '%' }"
                    :class="strengthClass"
                  ></div>
                </div>
                <div class="strength-hints" v-if="strengthScore < 3">
                  <el-icon><Warning /></el-icon>
                  <span>建议包含大小写字母、数字和特殊字符</span>
                </div>
              </div>
            </el-form-item>

            <el-form-item label="确认密码" prop="confirmPassword">
              <div class="form-field">
                <el-input
                  v-model="passwordForm.confirmPassword"
                  type="password"
                  show-password
                  placeholder="请再次输入新密码"
                  class="modern-input"
                >
                  <template #prefix>
                    <el-icon><CircleCheck /></el-icon>
                  </template>
                </el-input>
              </div>
            </el-form-item>

            <el-form-item>
              <div class="form-actions">
                <el-button
                  type="primary"
                  @click="changePassword"
                  :loading="passwordLoading"
                  class="submit-btn"
                >
                  <el-icon><Refresh /></el-icon>
                  确认修改密码
                </el-button>
                <el-button @click="resetPasswordForm" class="reset-btn">
                  <el-icon><RefreshLeft /></el-icon>
                  重置
                </el-button>
              </div>
            </el-form-item>
          </el-form>

          <!-- 安全提示 -->
          <div class="security-tips">
            <div class="tips-title">
              <el-icon><InfoFilled /></el-icon>
              <span>密码安全最佳实践</span>
            </div>
            <ul class="tips-list">
              <li>使用至少8位字符的密码</li>
              <li>包含大小写字母、数字和特殊字符的组合</li>
              <li>避免使用生日、手机号等个人信息</li>
              <li>不要在多个平台使用相同的密码</li>
              <li>建议每90天更换一次密码</li>
            </ul>
          </div>
        </el-card>

        <!-- 最近活动卡片 -->
        <el-card
          class="activity-card"
          shadow="never"
          v-if="recentActivities.length > 0"
        >
          <template #header>
            <div class="card-header">
              <div class="header-left">
                <div class="header-icon">
                  <el-icon><List /></el-icon>
                </div>
                <div>
                  <h3 class="card-title">最近活动</h3>
                  <p class="card-subtitle">您近期的操作记录</p>
                </div>
              </div>
              <div class="header-right">
                <el-button link type="primary" @click="viewAllActivities">
                  查看全部
                  <el-icon><ArrowRight /></el-icon>
                </el-button>
              </div>
            </div>
          </template>
          <div class="activity-timeline">
            <div
              v-for="(activity, index) in recentActivities"
              :key="index"
              class="activity-item"
            >
              <div class="activity-icon" :class="activity.type">
                <el-icon><component :is="activity.icon" /></el-icon>
              </div>
              <div class="activity-content">
                <div class="activity-title">{{ activity.title }}</div>
                <div class="activity-time">{{ activity.time }}</div>
              </div>
              <div class="activity-status">
                <el-tag
                  :type="activity.status === 'success' ? 'success' : 'info'"
                  size="small"
                  effect="light"
                >
                  {{ activity.status === "success" ? "成功" : "进行中" }}
                </el-tag>
              </div>
            </div>
          </div>
        </el-card>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import {
  User,
  Edit,
  Message,
  Lock,
  Key,
  Check,
  CircleCheck,
  Refresh,
  RefreshLeft,
  InfoFilled,
  Star,
  Calendar,
  Clock,
  UserFilled,
  List,
  ArrowRight,
  Warning,
} from "@element-plus/icons-vue";
import { userApi } from "@api/index";
import { useUserStore } from "@store/user";

const router = useRouter();
const userStore = useUserStore();
const updateLoading = ref(false);
const passwordLoading = ref(false);
const passwordFormRef = ref();

// 统计信息
const stats = reactive({
  loginCount: 0,
  roleCount: 1,
  permissionCount: 0,
});

// 最近活动
const recentActivities = ref<any[]>([]);

// 判断是否是管理员
const isAdmin = computed(() => userStore.role === "admin");

// 角色显示文本
const roleText = computed(() => {
  const roleMap: Record<string, string> = {
    admin: "系统管理员",
    operator: "运维操作员",
    auditor: "安全审计员",
    viewer: "查看员",
  };
  return roleMap[userStore.role] || "系统用户";
});

const roleDescription = computed(() => {
  const descMap: Record<string, string> = {
    admin: "拥有系统的所有管理权限，可管理用户、角色与业务数据",
    operator: "可管理出款站点、考勤绩效等业务数据",
    auditor: "只读权限，可查看业务数据与操作日志",
    viewer: "基础查看权限，仅可查看指定数据",
  };
  return descMap[userStore.role] || "系统用户";
});

// 在线状态
const getStatusClass = computed(() => {
  return userStore.isAuthenticated ? "online" : "offline";
});

const getStatusText = computed(() => {
  return userStore.isAuthenticated ? "在线" : "离线";
});

// 用户头像
const userAvatar = computed(() => {
  const name = profile.username?.charAt(0)?.toUpperCase() || "U";
  const colorMap: Record<string, string> = {
    admin: "667eea",
    operator: "43e97b",
    auditor: "f093fb",
    viewer: "4facfe",
  };
  const color = colorMap[userStore.role] || "667eea";
  return `https://ui-avatars.com/api/?background=${color}&color=fff&size=100&rounded=true&bold=true&length=1&name=${name}`;
});

// 密码强度计算
const strengthScore = ref(0);
const strengthPercent = computed(() => (strengthScore.value / 4) * 100);
const strengthText = computed(() => {
  if (strengthScore.value === 0) return "未设置";
  if (strengthScore.value === 1) return "弱";
  if (strengthScore.value === 2) return "中";
  if (strengthScore.value === 3) return "强";
  return "非常强";
});

const strengthClass = computed(() => {
  if (strengthScore.value <= 1) return "weak";
  if (strengthScore.value === 2) return "medium";
  return "strong";
});

const calculateStrength = (password: string) => {
  let score = 0;
  if (password.length >= 8) score++;
  if (/[a-z]/.test(password) && /[A-Z]/.test(password)) score++;
  if (/[0-9]/.test(password)) score++;
  if (/[^a-zA-Z0-9]/.test(password)) score++;
  strengthScore.value = score;
};

const profile = reactive({
  username: "",
  realName: "",
  email: "",
  createdAt: "",
  lastLogin: "" as string | null | undefined,
});

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
    {
      validator: (_: any, value: string, callback: Function) => {
        calculateStrength(value);
        if (value && value.length < 6) {
          callback(new Error("密码长度不能小于6位"));
        } else {
          callback();
        }
      },
      trigger: "blur",
    },
  ],
  confirmPassword: [
    { required: true, message: "请确认密码", trigger: "blur" },
    {
      validator: (_: any, value: string, callback: Function) => {
        if (value !== passwordForm.newPassword) {
          callback(new Error("两次输入的密码不一致"));
        } else {
          callback();
        }
      },
      trigger: "blur",
    },
  ],
};

const formatDate = (date: string | null | undefined): string => {
  if (!date) return "-";
  const d = new Date(date);
  if (isNaN(d.getTime())) return "-";
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")} ${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}:${String(d.getSeconds()).padStart(2, "0")}`;
};

// Profile.vue
const loadProfile = async () => {
  try {
    const { data } = await userApi.getProfile();
    profile.username = data.username;
    profile.realName = data.realName || "";
    profile.email = data.email || "";
    profile.createdAt = data.createdAt;
    profile.lastLogin = data.lastLogin;

    stats.permissionCount = userStore.permissions?.length || 0;
  } catch (error: any) {
    ElMessage.error(error.message || "加载个人信息失败");
  }
};

const updateProfile = async () => {
  if (!isAdmin.value) {
    ElMessage.warning("只有管理员可以修改个人信息");
    return;
  }

  updateLoading.value = true;
  try {
    await userApi.updateProfile({
      realName: profile.realName,
      email: profile.email,
    });
    ElMessage.success("个人信息更新成功");
  } catch (error: any) {
    ElMessage.error(error.message || "更新失败");
  } finally {
    updateLoading.value = false;
  }
};

const changePassword = async () => {
  if (!passwordFormRef.value) return;

  try {
    await passwordFormRef.value.validate();
  } catch {
    return;
  }

  passwordLoading.value = true;
  try {
    await userApi.changeOwnPassword({
      oldPassword: passwordForm.oldPassword,
      newPassword: passwordForm.newPassword,
    });
    ElMessage.success("密码修改成功，请重新登录");
    setTimeout(() => {
      userStore.logout();
      window.location.href = "/login";
    }, 1500);
  } catch (error: any) {
    ElMessage.error(error.message || "修改密码失败");
  } finally {
    passwordLoading.value = false;
  }
};

const resetPasswordForm = () => {
  passwordForm.oldPassword = "";
  passwordForm.newPassword = "";
  passwordForm.confirmPassword = "";
  strengthScore.value = 0;
  passwordFormRef.value?.clearValidate();
};

const viewAllActivities = () => {
  if (userStore.hasPermission("operation_log:view")) {
    router.push("/operation-logs");
  }
};

onMounted(() => {
  loadProfile();
});
</script>

<style scoped>
.profile-page {
  padding: 0px;
  min-height: 100vh;
  background: #f0f2f6;
  position: relative;
}

/* ========== 背景装饰 ========== */
.bg-decoration {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
  z-index: 0;
  overflow: hidden;
}

.blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.4;
  animation: float 20s ease-in-out infinite;
}

.blob-1 {
  width: 500px;
  height: 500px;
  background: linear-gradient(135deg, #667eea, #764ba2);
  top: -200px;
  right: -100px;
  animation-delay: 0s;
}

.blob-2 {
  width: 600px;
  height: 600px;
  background: linear-gradient(135deg, #f093fb, #f5576c);
  bottom: -250px;
  left: -150px;
  animation-delay: -5s;
}

.blob-3 {
  width: 400px;
  height: 400px;
  background: linear-gradient(135deg, #4facfe, #00f2fe);
  top: 40%;
  left: 30%;
  animation-delay: -10s;
}

@keyframes float {
  0%,
  100% {
    transform: translate(0, 0) scale(1);
  }
  33% {
    transform: translate(30px, -30px) scale(1.05);
  }
  66% {
    transform: translate(-20px, 20px) scale(0.95);
  }
}

/* 确保内容在背景之上 */
.profile-container {
  display: flex;
  gap: 24px;
  max-width: 1400px;
  margin: 0 auto;
  position: relative;
  z-index: 1;
}

/* ========== 左侧边栏 ========== */
.profile-sidebar {
  width: 360px;
  flex-shrink: 0;
}

.user-card {
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.3);
  overflow: hidden;
  transition: all 0.3s ease;
}

.user-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
}

.user-header {
  text-align: center;
  padding: 32px 24px 24px;
}

.avatar-container {
  position: relative;
  display: inline-block;
}

.user-avatar {
  border: 4px solid #fff;
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.15);
  transition: all 0.3s ease;
  z-index: 2;
  position: relative;
}

.avatar-ring {
  position: absolute;
  top: -4px;
  left: -4px;
  right: -4px;
  bottom: -4px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea, #764ba2);
  opacity: 0.3;
  animation: pulse 2s infinite;
  z-index: 1;
}

@keyframes pulse {
  0% {
    opacity: 0.3;
    transform: scale(1);
  }
  50% {
    opacity: 0.6;
    transform: scale(1.02);
  }
  100% {
    opacity: 0.3;
    transform: scale(1);
  }
}

.status-badge {
  background: #fff;
  border-radius: 20px;
  padding: 4px 10px;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  font-weight: 500;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.status-badge.online {
  color: #67c23a;
  background: #f0f9f0;
}

.status-badge.offline {
  color: #909399;
  background: #f5f7fa;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: currentColor;
}

.status-text {
  font-size: 12px;
}

/* 用户名和状态行 */
.user-name-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin: 16px 0 8px;
  flex-wrap: wrap;
}

.user-name {
  margin: 0;
  font-size: 22px;
  font-weight: 700;
  color: #1a1a2e;
}

.user-role-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 16px;
  border-radius: 30px;
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 12px;
}

.user-role-badge.admin {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.user-role-badge.operator {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
  color: white;
}

.user-role-badge.auditor {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  color: white;
}

.user-role-badge.viewer {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  color: white;
}

.user-desc {
  font-size: 13px;
  color: #909399;
  line-height: 1.5;
  margin: 0;
}

.user-stats {
  display: flex;
  justify-content: space-around;
  padding: 20px 0;
  background: rgba(245, 247, 250, 0.8);
  margin: 0 -20px;
}

.stat-item {
  text-align: center;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  background: linear-gradient(135deg, #667eea, #764ba2);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.stat-label {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.stat-divider {
  width: 1px;
  background: #e4e7ed;
}

.info-list {
  padding: 8px 0;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 0;
}

.info-icon {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea15, #764ba215);
  border-radius: 12px;
  color: #667eea;
}

.info-content {
  flex: 1;
}

.info-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.info-value {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

/* ========== 右侧主内容 ========== */
.profile-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.edit-card,
.password-card,
.activity-card {
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.3);
  transition: all 0.3s ease;
}

.edit-card:hover,
.password-card:hover,
.activity-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.12);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-icon {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea15 0%, #764ba215 100%);
  border-radius: 16px;
  color: #667eea;
}

.card-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1a1a2e;
}

.card-subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: #909399;
}

.modern-form {
  padding: 8px 0;
}

.form-field {
  width: 100%;
}

.modern-input :deep(.el-input__wrapper) {
  border-radius: 14px;
  border: 1px solid #e4e7ed;
  transition: all 0.3s;
  padding: 8px 16px;
  background: rgba(255, 255, 255, 0.9);
}

.modern-input :deep(.el-input__wrapper:hover) {
  border-color: #667eea;
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.1);
}

.modern-input :deep(.el-input__wrapper.is-focus) {
  border-color: #667eea;
  box-shadow: 0 0 0 2px rgba(102, 126, 234, 0.2);
}

.modern-input :deep(.el-input__prefix) {
  color: #909399;
  font-size: 16px;
}

.submit-btn {
  border-radius: 14px;
  padding: 12px 28px;
  font-size: 15px;
  background: linear-gradient(135deg, #667eea, #764ba2);
  border: none;
  transition: all 0.3s ease;
}

.submit-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(102, 126, 234, 0.4);
}

.disabled-submit {
  border-radius: 14px;
  padding: 12px 28px;
  font-size: 15px;
  cursor: not-allowed;
  opacity: 0.6;
}

.permission-alert {
  margin: 8px 0 8px 10px;
}

.permission-alert :deep(.el-alert) {
  border-radius: 14px;
  background: rgba(240, 249, 255, 0.9);
  border: 1px solid #b3d8ff;
}

/* 密码强度 */
.password-strength {
  background: rgba(245, 247, 250, 0.8);
  border-radius: 14px;
  padding: 16px;
  width: 100%;
}

.strength-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 10px;
}

.strength-label {
  font-size: 13px;
  color: #606266;
}

.strength-value {
  font-size: 13px;
  font-weight: 600;
}

.strength-value.weak {
  color: #f56c6c;
}
.strength-value.medium {
  color: #e6a23c;
}
.strength-value.strong {
  color: #67c23a;
}

.strength-bar-container {
  background: #e4e7ed;
  border-radius: 10px;
  overflow: hidden;
  height: 8px;
}

.strength-bar {
  height: 100%;
  border-radius: 10px;
  transition: width 0.3s ease;
}

.strength-bar.weak {
  background: linear-gradient(90deg, #f56c6c, #f89898);
}
.strength-bar.medium {
  background: linear-gradient(90deg, #e6a23c, #f5c16a);
}
.strength-bar.strong {
  background: linear-gradient(90deg, #67c23a, #95d475);
}

.strength-hints {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 10px;
  font-size: 12px;
  color: #e6a23c;
}

.form-actions {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
}

.reset-btn {
  border-radius: 14px;
  padding: 12px 24px;
  transition: all 0.3s ease;
}

.reset-btn:hover {
  transform: translateY(-2px);
}

/* 安全提示 */
.security-tips {
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid rgba(0, 0, 0, 0.05);
}

.tips-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 12px;
}

.tips-list {
  margin: 0;
  padding-left: 20px;
}

.tips-list li {
  font-size: 13px;
  color: #909399;
  line-height: 1.8;
}

/* 活动时间线 */
.activity-timeline {
  padding: 8px 0;
}

.activity-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 0;
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
  transition: all 0.2s ease;
}

.activity-item:hover {
  transform: translateX(4px);
  background: rgba(245, 247, 250, 0.5);
}

.activity-item:last-child {
  border-bottom: none;
}

.activity-icon {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
}

.activity-icon.login {
  background: linear-gradient(135deg, #e6f7ff, #bae7ff);
  color: #1890ff;
}
.activity-icon.edit {
  background: linear-gradient(135deg, #fff7e6, #ffe7ba);
  color: #faad14;
}
.activity-icon.security {
  background: linear-gradient(135deg, #f6ffed, #d9f7be);
  color: #52c41a;
}

.activity-content {
  flex: 1;
}

.activity-title {
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
}

.activity-time {
  font-size: 12px;
  color: #909399;
}

/* 滚动条美化 */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 4px;
}

::-webkit-scrollbar-thumb {
  background: linear-gradient(135deg, #667eea, #764ba2);
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(135deg, #5a67d8, #6b46c1);
}

/* 响应式 */
@media (max-width: 1200px) {
  .profile-page {
    padding: 0;
  }

  .blob-1,
  .blob-2,
  .blob-3 {
    opacity: 0.2;
  }
}

@media (max-width: 1024px) {
  .profile-sidebar {
    width: 320px;
  }
}

@media (max-width: 768px) {
  .profile-page {
    padding: 1px;
  }

  .blob-1,
  .blob-2,
  .blob-3 {
    display: none;
  }

  .profile-container {
    flex-direction: column;
  }

  .profile-sidebar {
    width: 100%;
  }

  .user-stats {
    margin: 0 -16px;
  }

  .form-actions {
    flex-direction: column;
  }

  .submit-btn,
  .reset-btn {
    width: 100%;
  }

  .card-header {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>

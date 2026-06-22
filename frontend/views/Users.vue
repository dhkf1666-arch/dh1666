<template>
  <div class="users-page">
    <!-- 背景装饰 -->
    <div class="bg-decoration">
      <div class="blob blob-1"></div>
      <div class="blob blob-2"></div>
      <div class="blob blob-3"></div>
    </div>

    <el-card class="filter-card card-hover">
      <el-form :inline="true" :model="filters" class="filter-form">
        <el-form-item label="关键词">
          <el-input
            v-model="filters.keyword"
            placeholder="用户名/姓名"
            clearable
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
          <el-button
            type="success"
            @click="handleCreate"
            v-permission="'user:create'"
          >
            <el-icon><Plus /></el-icon>
            新建用户
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card card-hover">
      <el-table :data="users" v-loading="loading" stripe>
        <el-table-column prop="username" label="用户名" width="150" />
        <el-table-column prop="realName" label="姓名" width="120" />
        <el-table-column prop="email" label="邮箱" min-width="180" />
        <el-table-column prop="roles" label="角色" width="150">
          <template #default="{ row }">
            <el-tag
              v-for="role in row.roles"
              :key="role.id"
              size="small"
              style="margin-right: 4px; margin-bottom: 2px"
            >
              {{ role.name }}
            </el-tag>
            <span v-if="!row.roles?.length" class="text-muted">未分配</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? "启用" : "禁用" }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="170">
          <template #default="{ row }">
            {{ formatTime(row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column prop="lastLogin" label="最后登录" width="170">
          <template #default="{ row }">
            {{ formatTime(row.lastLogin) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="320" fixed="right" align="center">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button
                link
                type="primary"
                size="small"
                @click="handleEdit(row)"
                v-permission="'user:update'"
                class="action-btn edit-btn"
              >
                <el-icon><Edit /></el-icon>
                编辑
              </el-button>
              <el-button
                link
                type="warning"
                size="small"
                @click="handleResetPassword(row)"
                v-permission="'user:manage'"
                class="action-btn reset-btn"
              >
                <el-icon><Key /></el-icon>
                重置密码
              </el-button>
              <el-button
                link
                type="danger"
                size="small"
                @click="handleDelete(row)"
                v-permission="'user:delete'"
                :disabled="row.username === 'admin'"
                class="action-btn delete-btn"
              >
                <el-icon><Delete /></el-icon>
                删除
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadUsers"
          @current-change="loadUsers"
        />
      </div>
    </el-card>

    <!-- 创建/编辑用户对话框 -->
    <el-dialog
      :title="dialogTitle"
      v-model="dialogVisible"
      width="550px"
      @close="resetForm"
      class="animate-dialog modern-dialog"
    >
      <el-form :model="form" :rules="rules" ref="formRef" label-width="80px">
        <el-form-item label="用户名" prop="username" v-if="!isEdit">
          <el-input v-model="form.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="!isEdit">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码（至少6位）"
            show-password
          />
          <div class="form-tip">密码长度至少6位，建议包含字母和数字</div>
        </el-form-item>
        <el-form-item label="姓名" prop="realName">
          <el-input v-model="form.realName" placeholder="请输入姓名" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="角色" prop="roleIds">
          <el-select
            v-model="form.roleIds"
            multiple
            placeholder="请选择角色"
            style="width: 100%"
          >
            <el-option
              v-for="role in roles"
              :key="role.id"
              :label="role.name"
              :value="role.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-switch
            v-model="form.status"
            :active-value="1"
            :inactive-value="0"
            active-text="启用"
            inactive-text="禁用"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm" :loading="submitLoading"
          >确定</el-button
        >
      </template>
    </el-dialog>

    <!-- 重置密码对话框 -->
    <el-dialog
      title="重置密码"
      v-model="passwordDialogVisible"
      width="400px"
      @close="resetPasswordForm"
      class="animate-dialog modern-dialog"
    >
      <el-form
        :model="passwordForm"
        :rules="passwordRules"
        ref="passwordFormRef"
        label-width="80px"
      >
        <el-form-item label="新密码" prop="password">
          <el-input
            v-model="passwordForm.password"
            type="password"
            placeholder="请输入新密码"
            show-password
          />
          <div class="form-tip">密码长度至少6位，建议包含字母和数字</div>
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input
            v-model="passwordForm.confirmPassword"
            type="password"
            placeholder="请确认新密码"
            show-password
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="passwordDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          @click="confirmResetPassword"
          :loading="resetLoading"
          >确定</el-button
        >
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Plus, Edit, Key, Delete } from "@element-plus/icons-vue";
import { userApi, roleApi } from "@api/index";
import type { User, Role } from "@api/types";

const loading = ref(false);
const submitLoading = ref(false);
const resetLoading = ref(false);
const dialogVisible = ref(false);
const passwordDialogVisible = ref(false);
const isEdit = ref(false);
const formRef = ref();
const passwordFormRef = ref();
const currentUserId = ref("");

const users = ref<User[]>([]);
const roles = ref<Role[]>([]);

const filters = reactive({
  keyword: "",
});

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
});

const form = reactive({
  username: "",
  password: "",
  realName: "",
  email: "",
  roleIds: [] as string[],
  status: 1,
});

const passwordForm = reactive({
  password: "",
  confirmPassword: "",
});

const dialogTitle = computed(() => (isEdit.value ? "编辑用户" : "新建用户"));

// 密码强度验证
const validatePasswordStrength = (
  _: any,
  value: string,
  callback: Function,
) => {
  if (!value) {
    callback();
    return;
  }
  if (value.length < 6) {
    callback(new Error("密码长度不能小于6位"));
    return;
  }
  const hasLetter = /[a-zA-Z]/.test(value);
  const hasNumber = /\d/.test(value);
  if (!hasLetter || !hasNumber) {
    callback(new Error("建议包含字母和数字组合"));
  }
  callback();
};

const rules = computed(() => ({
  username: [
    { required: true, message: "请输入用户名", trigger: "blur" },
    { min: 3, max: 20, message: "用户名长度应在3-20位之间", trigger: "blur" },
    {
      pattern: /^[a-zA-Z0-9_]+$/,
      message: "用户名只能包含字母、数字和下划线",
      trigger: "blur",
    },
  ],
  password: [
    { required: !isEdit.value, message: "请输入密码", trigger: "blur" },
    { validator: validatePasswordStrength, trigger: "blur" },
  ],
  email: [{ type: "email", message: "请输入有效的邮箱", trigger: "blur" }],
  roleIds: [{ type: "array", message: "请选择角色" }],
}));

const passwordRules = {
  password: [
    { required: true, message: "请输入新密码", trigger: "blur" },
    { min: 6, message: "密码长度不能小于6位", trigger: "blur" },
    {
      pattern: /^(?=.*[A-Za-z])(?=.*\d)[A-Za-z\d]{6,}$/,
      message: "密码必须包含字母和数字",
      trigger: "blur",
    },
  ],
  confirmPassword: [
    { required: true, message: "请确认密码", trigger: "blur" },
    {
      validator: (_: any, value: string, callback: Function) => {
        if (value !== passwordForm.password) {
          callback(new Error("两次输入的密码不一致"));
        } else {
          callback();
        }
      },
      trigger: "blur",
    },
  ],
};

const formatTime = (time: string | null | undefined) => {
  if (!time) return "-";
  return new Date(time).toLocaleString("zh-CN");
};

const loadRoles = async () => {
  try {
    const { data } = await roleApi.list();
    roles.value = data;
  } catch (error: any) {
    console.error("加载角色列表失败:", error);
  }
};

const loadUsers = async () => {
  loading.value = true;
  try {
    const { data } = await userApi.list({
      page: pagination.page,
      pageSize: pagination.pageSize,
      keyword: filters.keyword || undefined,
    });
    users.value = data.items || [];
    pagination.total = data.total;
  } catch (error: any) {
    ElMessage.error(error.message || "加载用户列表失败");
  } finally {
    loading.value = false;
  }
};

const handleSearch = () => {
  pagination.page = 1;
  loadUsers();
};

const handleReset = () => {
  filters.keyword = "";
  handleSearch();
};

const resetForm = () => {
  form.username = "";
  form.password = "";
  form.realName = "";
  form.email = "";
  form.roleIds = [];
  form.status = 1;
  formRef.value?.clearValidate();
};

const resetPasswordForm = () => {
  passwordForm.password = "";
  passwordForm.confirmPassword = "";
  passwordFormRef.value?.clearValidate();
};

const handleCreate = () => {
  isEdit.value = false;
  resetForm();
  dialogVisible.value = true;
};

const handleEdit = async (row: User) => {
  isEdit.value = true;
  currentUserId.value = row.id;
  form.username = row.username;
  form.realName = row.realName || "";
  form.email = row.email || "";
  form.status = row.status;
  form.roleIds = row.roles?.map((r) => r.id) || [];

  if (roles.value.length === 0) {
    await loadRoles();
  }

  dialogVisible.value = true;
};

const submitForm = async () => {
  if (!formRef.value) return;

  try {
    await formRef.value.validate();
  } catch {
    return;
  }

  submitLoading.value = true;
  try {
    if (isEdit.value) {
      await userApi.update(currentUserId.value, {
        realName: form.realName,
        email: form.email,
        status: form.status,
        roleIds: form.roleIds,
      });
      ElMessage.success("用户更新成功");
    } else {
      await userApi.create({
        username: form.username,
        password: form.password,
        realName: form.realName,
        email: form.email,
        roleIds: form.roleIds,
      });
      ElMessage.success("用户创建成功");
    }
    dialogVisible.value = false;
    loadUsers();
  } catch (error: any) {
    ElMessage.error(error.message || "操作失败");
  } finally {
    submitLoading.value = false;
  }
};

const handleResetPassword = (row: User) => {
  currentUserId.value = row.id;
  resetPasswordForm();
  passwordDialogVisible.value = true;
};

const confirmResetPassword = async () => {
  if (!passwordFormRef.value) return;

  try {
    await passwordFormRef.value.validate();
  } catch {
    return;
  }

  resetLoading.value = true;
  try {
    await userApi.resetPassword(currentUserId.value, passwordForm.password);
    ElMessage.success("密码重置成功");
    passwordDialogVisible.value = false;
  } catch (error: any) {
    ElMessage.error(error.message || "重置密码失败");
  } finally {
    resetLoading.value = false;
  }
};

const handleDelete = (row: User) => {
  if (row.username === "admin") {
    ElMessage.warning("不能删除系统管理员账号");
    return;
  }

  ElMessageBox.confirm(
    `确定要删除用户 "${row.username}" 吗？此操作不可恢复。`,
    "提示",
    {
      confirmButtonText: "确定",
      cancelButtonText: "取消",
      type: "warning",
    },
  )
    .then(async () => {
      try {
        await userApi.delete(row.id);
        ElMessage.success("删除成功");
        loadUsers();
      } catch (error: any) {
        ElMessage.error(error.message || "删除失败");
      }
    })
    .catch(() => {});
};

onMounted(() => {
  loadRoles();
  loadUsers();
});
</script>

<style scoped>
.users-page {
  padding: 0;
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
.filter-card,
.table-card {
  position: relative;
  z-index: 1;
}

.filter-card {
  margin-bottom: 20px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.3);
  transition: all 0.3s ease;
}

.table-card {
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.3);
  transition: all 0.3s ease;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  line-height: 1.2;
}

.text-muted {
  color: #909399;
  font-size: 12px;
}

/* 卡片悬停动画 */
.card-hover {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.card-hover:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.12);
}

/* 筛选表单样式 */
.filter-form {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 16px;
}

.filter-form .el-form-item {
  margin-bottom: 0;
  display: inline-flex;
  align-items: center;
}

/* 按钮样式 */
:deep(.el-button--primary) {
  background: linear-gradient(135deg, #667eea, #764ba2);
  border: none;
  transition: all 0.3s ease;
}

:deep(.el-button--primary):hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

:deep(.el-button--success) {
  background: linear-gradient(135deg, #43e97b, #38f9d7);
  border: none;
  color: #1a1a2e;
}

:deep(.el-button--success):hover {
  transform: translateY(-2px);
}

:deep(.el-button--warning) {
  background: linear-gradient(135deg, #f093fb, #f5576c);
  border: none;
  color: white;
}

:deep(.el-button--warning):hover {
  transform: translateY(-2px);
}

:deep(.el-button--danger) {
  background: linear-gradient(135deg, #fa709a, #fee140);
  border: none;
  color: white;
}

:deep(.el-button--danger):hover {
  transform: translateY(-2px);
}

/* 表格样式优化 */
:deep(.el-table) {
  background: transparent;
}

:deep(.el-table__header-wrapper) {
  background: rgba(248, 250, 252, 0.9);
}

:deep(.el-table__header th) {
  background: transparent;
  font-weight: 600;
  color: #1e293b;
}

:deep(.el-table__row) {
  transition: all 0.2s ease;
}

:deep(.el-table__row:hover) {
  transform: translateX(2px);
  background: linear-gradient(90deg, rgba(102, 126, 234, 0.05), transparent);
}

/* 链接按钮样式 */
.el-button.is-link {
  transition: all 0.2s ease;
}

.el-button.is-link:hover {
  transform: translateX(2px);
}

/* 标签样式 */
.el-tag {
  transition: all 0.2s ease;
}

.el-tag:hover {
  transform: scale(1.02);
}

/* 现代化对话框 */
.modern-dialog :deep(.el-dialog) {
  border-radius: 20px;
  overflow: hidden;
}

.modern-dialog :deep(.el-dialog__header) {
  background: linear-gradient(135deg, #667eea, #764ba2);
  padding: 16px 20px;
  margin: 0;
}

.modern-dialog :deep(.el-dialog__title) {
  color: white;
  font-weight: 600;
  font-size: 18px;
}

.modern-dialog :deep(.el-dialog__headerbtn .el-dialog__close) {
  color: white;
  font-size: 18px;
}

.modern-dialog :deep(.el-dialog__body) {
  padding: 20px;
  max-height: 60vh;
  overflow-y: auto;
}

.modern-dialog :deep(.el-dialog__footer) {
  padding: 16px 20px;
  border-top: 1px solid rgba(0, 0, 0, 0.05);
}

/* 对话框动画 */
.animate-dialog :deep(.el-dialog) {
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

/* 分页样式 */
.pagination :deep(.el-pagination .btn-prev),
.pagination :deep(.el-pagination .btn-next),
.pagination :deep(.el-pager li) {
  border-radius: 10px;
  margin: 0 3px;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  font-weight: 500;
  min-width: 36px;
  height: 36px;
}

.pagination :deep(.el-pager li.active) {
  background: linear-gradient(135deg, #667eea, #764ba2);
  border-color: transparent;
  color: #ffffff;
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
  .users-page {
    padding: 1px;
  }

  .blob-1,
  .blob-2,
  .blob-3 {
    opacity: 0.2;
  }
}

@media (max-width: 768px) {
  .users-page {
    padding: 2px;
  }

  .blob-1,
  .blob-2,
  .blob-3 {
    display: none;
  }

  .filter-form {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }

  .filter-form .el-form-item {
    margin-bottom: 0;
    width: 100%;
  }

  .filter-form .el-form-item .el-input,
  .filter-form .el-form-item .el-select {
    width: 100% !important;
  }

  .table-card :deep(.el-table) {
    overflow-x: auto;
  }

  .pagination {
    overflow-x: auto;
  }

  .pagination :deep(.el-pagination) {
    flex-wrap: wrap;
    justify-content: center;
  }

  .el-dialog {
    width: 95% !important;
    margin: 20px auto !important;
  }
}

/* 操作按钮组 */
.action-buttons {
  display: flex;
  gap: 8px;
  justify-content: center;
  flex-wrap: wrap;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
  transition: all 0.2s ease;
}

/* 编辑按钮 */
.edit-btn {
  background: rgba(64, 158, 255, 0.1);
  color: #409eff;
}

.edit-btn:hover {
  background: rgba(64, 158, 255, 0.2);
  transform: translateY(-1px);
}

/* 重置密码按钮 */
.reset-btn {
  background: rgba(230, 162, 60, 0.1);
  color: #e6a23c;
}

.reset-btn:hover {
  background: rgba(230, 162, 60, 0.2);
  transform: translateY(-1px);
}

/* 删除按钮 */
.delete-btn {
  background: rgba(245, 108, 108, 0.1);
  color: #f56c6c;
}

.delete-btn:hover {
  background: rgba(245, 108, 108, 0.2);
  transform: translateY(-1px);
}

/* 按钮图标样式 */
.action-btn .el-icon {
  font-size: 13px;
}

.action-btn:hover .el-icon {
  transform: scale(1.05);
}

/* 禁用按钮样式 */
.delete-btn.is-disabled {
  background: rgba(245, 108, 108, 0.05);
  color: #c0c4cc;
  cursor: not-allowed;
}

.delete-btn.is-disabled:hover {
  transform: none;
}
</style>

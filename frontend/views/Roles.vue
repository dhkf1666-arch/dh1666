<template>
  <div class="roles-page">
    <!-- 背景装饰 -->
    <div class="bg-decoration">
      <div class="blob blob-1"></div>
      <div class="blob blob-2"></div>
      <div class="blob blob-3"></div>
    </div>

    <el-card class="card-hover">
      <template #header>
        <div class="card-header">
          <span>角色管理</span>
          <el-button
            type="primary"
            @click="handleCreate"
            v-permission="'role:create'"
          >
            <el-icon><Plus /></el-icon> 新建角色
          </el-button>
        </div>
      </template>

      <el-table :data="roles" v-loading="loading" stripe>
        <el-table-column prop="name" label="角色名称" width="150" />
        <el-table-column prop="description" label="描述" min-width="200" />
        <el-table-column prop="permissions" label="权限数量" width="100">
          <template #default="{ row }"
            >{{ getPermissionsCount(row) }} 个权限</template
          >
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="180">
          <template #default="{ row }">{{
            formatTime(row.createdAt)
          }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right" align="center">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button
                link
                type="primary"
                @click="handleEdit(row)"
                v-permission="'role:update'"
                class="action-btn edit-btn"
              >
                <el-icon><Edit /></el-icon>
                编辑
              </el-button>
              <el-button
                link
                type="danger"
                @click="handleDelete(row)"
                v-permission="'role:delete'"
                class="action-btn delete-btn"
              >
                <el-icon><Delete /></el-icon>
                删除
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 角色编辑对话框 -->
    <el-dialog
      :title="dialogTitle"
      v-model="dialogVisible"
      width="850px"
      class="role-dialog modern-dialog"
    >
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="2"
            placeholder="请输入角色描述"
          />
        </el-form-item>

        <!-- 细粒度权限树 -->
        <el-form-item label="权限" prop="permissions">
          <div class="permission-tree">
            <div class="permission-actions">
              <el-button size="small" @click="expandAll">全部展开</el-button>
              <el-button size="small" @click="collapseAll">全部收起</el-button>
              <el-button size="small" type="primary" @click="selectAll"
                >全选所有</el-button
              >
              <el-button size="small" type="danger" @click="deselectAll"
                >取消全选</el-button
              >
            </div>

            <div
              v-for="group in permissionGroups"
              :key="group.id"
              class="permission-group"
            >
              <div class="group-header">
                <el-checkbox
                  :model-value="isGroupAllSelected(group)"
                  :indeterminate="isGroupIndeterminate(group)"
                  @change="() => toggleGroup(group)"
                >
                  <strong>{{ group.name }}</strong>
                  <span class="group-desc" v-if="group.description"
                    >（{{ group.description }}）</span
                  >
                </el-checkbox>
              </div>
              <div class="group-permissions">
                <el-checkbox
                  v-for="perm in group.permissions"
                  :key="perm.code"
                  :model-value="selectedPermissions[perm.code]"
                  @change="(val: boolean) => updatePermission(perm.code, val)"
                  class="permission-item"
                >
                  <span class="perm-name">{{ perm.name }}</span>
                  <span v-if="perm.description" class="perm-desc"
                    >（{{ perm.description }}）</span
                  >
                </el-checkbox>
              </div>
            </div>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm" :loading="submitLoading"
          >确定</el-button
        >
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Plus, Edit, Delete } from "@element-plus/icons-vue";
import { roleApi } from "@api/index";
import type { Role } from "@api/types";

const loading = ref(false);
const submitLoading = ref(false);
const dialogVisible = ref(false);
const isEdit = ref(false);
const formRef = ref();
const roles = ref<Role[]>([]);
const currentRoleId = ref("");

const form = reactive({
  name: "",
  description: "",
});

// ✅ 完整权限定义 - 与后端 models.go 中的 DefaultRoles 完全一致
interface Permission {
  code: string;
  name: string;
  description?: string;
}

interface PermissionGroup {
  id: string;
  name: string;
  description?: string;
  permissions: Permission[];
}

const permissionGroups: PermissionGroup[] = [
  {
    id: "user",
    name: "用户管理",
    description: "管理后台用户账号",
    permissions: [
      { code: "user:view", name: "查看用户", description: "查看用户列表和详情" },
      { code: "user:create", name: "创建用户", description: "添加新用户" },
      { code: "user:update", name: "更新用户", description: "修改用户信息" },
      { code: "user:delete", name: "删除用户", description: "删除用户" },
      { code: "user:manage", name: "管理用户", description: "重置密码、分配角色" },
    ],
  },
  {
    id: "role",
    name: "角色管理",
    description: "管理角色和权限",
    permissions: [
      { code: "role:view", name: "查看角色", description: "查看角色列表" },
      { code: "role:create", name: "创建角色", description: "添加新角色" },
      { code: "role:update", name: "更新角色", description: "修改角色信息" },
      { code: "role:delete", name: "删除角色", description: "删除角色" },
    ],
  },
  {
    id: "attendance",
    name: "考勤管理",
    description: "管理员工考勤、绩效、罚款",
    permissions: [
      { code: "attendance:view", name: "查看考勤", description: "查看考勤记录、绩效、罚款" },
      { code: "attendance:edit", name: "编辑考勤", description: "修改考勤、绩效、罚款" },
    ],
  },
  {
    id: "site",
    name: "出款站点管理",
    description: "管理出款站点和员工账号",
    permissions: [
      { code: "site:view", name: "查看站点", description: "查看站点列表和员工账号" },
      { code: "site:manage", name: "管理站点", description: "增删改站点和员工账号" },
    ],
  },
  {
    id: "stats",
    name: "出款统计",
    description: "管理出款数据统计",
    permissions: [
      { code: "stats:view", name: "查看统计", description: "查看出款统计数据" },
      { code: "stats:edit", name: "编辑统计", description: "上传数据、清除数据" },
    ],
  },
  {
    id: "operation_log",
    name: "操作日志管理",
    description: "管理系统操作日志",
    permissions: [
      { code: "operation_log:view", name: "查看日志", description: "查看操作日志列表" },
      { code: "operation_log:export", name: "导出日志", description: "导出操作日志" },
      { code: "operation_log:delete", name: "删除日志", description: "删除操作日志" },
    ],
  },
];

// 选中的权限
const selectedPermissions = ref<Record<string, boolean>>({});

const dialogTitle = computed(() => (isEdit.value ? "编辑角色" : "新建角色"));

const rules = {
  name: [{ required: true, message: "请输入角色名称", trigger: "blur" }],
};

const formatTime = (time: string) => {
  if (!time) return "-";
  return new Date(time).toLocaleString("zh-CN");
};

// 获取权限数量
const getPermissionsCount = (role: Role) => {
  const perms = role.permissions;
  if (Array.isArray(perms)) {
    return perms.length;
  }
  return 0;
};

// 检查组是否全选
const isGroupAllSelected = (group: PermissionGroup) => {
  if (group.permissions.length === 0) return false;
  return group.permissions.every((p) => selectedPermissions.value[p.code]);
};

// 检查组是否部分选中
const isGroupIndeterminate = (group: PermissionGroup) => {
  const selected = group.permissions.filter(
    (p) => selectedPermissions.value[p.code],
  );
  return selected.length > 0 && selected.length < group.permissions.length;
};

// 切换组全选
const toggleGroup = (group: PermissionGroup) => {
  const allSelected = isGroupAllSelected(group);
  group.permissions.forEach((p) => {
    selectedPermissions.value[p.code] = !allSelected;
  });
};

// 更新单个权限
const updatePermission = (code: string, value: boolean) => {
  selectedPermissions.value[code] = value;
};

// 获取选中的权限列表
const getSelectedPermissionsList = (): string[] => {
  return Object.keys(selectedPermissions.value).filter(
    (k) => selectedPermissions.value[k],
  );
};

// 全选所有权限
const selectAll = () => {
  permissionGroups.forEach((group) => {
    group.permissions.forEach((p) => {
      selectedPermissions.value[p.code] = true;
    });
  });
};

// 取消全选
const deselectAll = () => {
  selectedPermissions.value = {};
};

// 展开所有分组
const expandAll = () => {
  const groups = document.querySelectorAll(".permission-group");
  groups.forEach((group) => {
    (group as HTMLElement).style.maxHeight = "none";
  });
};

// 收起所有分组
const collapseAll = () => {
  const groups = document.querySelectorAll(".permission-group");
  groups.forEach((group) => {
    (group as HTMLElement).style.maxHeight = "300px";
  });
};

const loadRoles = async () => {
  loading.value = true;
  try {
    const { data } = await roleApi.list();
    roles.value = data;
  } catch (error: any) {
    ElMessage.error(error.message || "加载角色列表失败");
  } finally {
    loading.value = false;
  }
};

const handleCreate = () => {
  isEdit.value = false;
  form.name = "";
  form.description = "";
  selectedPermissions.value = {};
  dialogVisible.value = true;
};

const handleEdit = (row: Role) => {
  isEdit.value = true;
  currentRoleId.value = row.id;
  form.name = row.name;
  form.description = row.description || "";

  // 清空之前的选中状态
  selectedPermissions.value = {};

  // 获取权限列表
  let permsList: string[] = [];
  const rolePerms = row.permissions;

  if (Array.isArray(rolePerms)) {
    permsList = rolePerms;
  } else if (rolePerms && typeof rolePerms === "object") {
    // 兼容旧格式
    const permObj = rolePerms as any;
    if (permObj.type === "all") {
      selectAll();
      dialogVisible.value = true;
      return;
    } else if (
      permObj.type === "custom" &&
      Array.isArray(permObj.permissions)
    ) {
      permsList = permObj.permissions;
    }
  }

  // 设置选中的权限
  permsList.forEach((code) => {
    selectedPermissions.value[code] = true;
  });

  dialogVisible.value = true;
};

const submitForm = async () => {
  if (!formRef.value) return;

  try {
    await formRef.value.validate();
  } catch {
    return;
  }

  // 直接发送权限字符串数组
  const permissions = getSelectedPermissionsList();

  submitLoading.value = true;
  try {
    if (isEdit.value) {
      await roleApi.update(currentRoleId.value, {
        name: form.name,
        description: form.description,
        permissions,
      });
      ElMessage.success("角色更新成功");
    } else {
      await roleApi.create({
        name: form.name,
        description: form.description,
        permissions,
      });
      ElMessage.success("角色创建成功");
    }
    dialogVisible.value = false;
    await loadRoles();
  } catch (error: any) {
    console.error("Submit error:", error);
    ElMessage.error(error.message || "操作失败");
  } finally {
    submitLoading.value = false;
  }
};

const handleDelete = (row: Role) => {
  ElMessageBox.confirm(`确定要删除角色 "${row.name}" 吗？`, "提示", {
    confirmButtonText: "确定",
    cancelButtonText: "取消",
    type: "warning",
  })
    .then(async () => {
      await roleApi.delete(row.id);
      ElMessage.success("删除成功");
      await loadRoles();
    })
    .catch(() => {});
};

onMounted(() => {
  loadRoles();
});
</script>

<style scoped>
.roles-page {
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
.card-hover {
  position: relative;
  z-index: 1;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-hover {
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.3);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.card-hover:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.12);
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

/* 权限树样式 */
.permission-tree {
  max-height: 450px;
  overflow-y: auto;
  border: 1px solid rgba(0, 0, 0, 0.05);
  border-radius: 12px;
  padding: 12px;
  background: rgba(250, 250, 250, 0.8);
}

.permission-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
}

.permission-group {
  margin-bottom: 16px;
  border: 1px solid rgba(0, 0, 0, 0.05);
  border-radius: 12px;
  overflow: hidden;
  transition: all 0.2s;
  background: white;
}

.permission-group:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  transform: translateX(2px);
}

.group-header {
  padding: 12px 16px;
  background: linear-gradient(135deg, #f5f7fa 0%, #f0f2f5 100%);
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
}

.group-desc {
  font-size: 12px;
  color: #909399;
  margin-left: 8px;
  font-weight: normal;
}

.group-permissions {
  padding: 12px 16px;
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
}

.permission-item {
  min-width: 140px;
  transition: all 0.2s ease;
}

.permission-item:hover {
  transform: translateX(2px);
}

.perm-name {
  font-weight: 500;
}

.perm-desc {
  font-size: 12px;
  color: #909399;
  margin-left: 4px;
}

/* 复选框样式优化 */
:deep(.el-checkbox__input.is-checked .el-checkbox__inner) {
  background-color: #667eea;
  border-color: #667eea;
}

:deep(.el-checkbox__input.is-indeterminate .el-checkbox__inner) {
  background-color: #667eea;
  border-color: #667eea;
}

:deep(.el-checkbox__inner:hover) {
  border-color: #667eea;
}

/* 滚动条美化 */
.permission-tree::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

.permission-tree::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.permission-tree::-webkit-scrollbar-thumb {
  background: linear-gradient(135deg, #667eea, #764ba2);
  border-radius: 3px;
}

.permission-tree::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(135deg, #5a67d8, #6b46c1);
}

/* 全局滚动条 */
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

/* 对话框动画 */
.role-dialog :deep(.el-dialog) {
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

/* 按钮通用悬停 */
.el-button {
  transition: all 0.2s ease !important;
}

.el-button:hover {
  transform: translateY(-1px);
}

/* 响应式 */
@media (max-width: 1200px) {
  .roles-page {
    padding: 1px;
  }

  .blob-1,
  .blob-2,
  .blob-3 {
    opacity: 0.2;
  }
}

@media (max-width: 768px) {
  .roles-page {
    padding: 2px;
  }

  .blob-1,
  .blob-2,
  .blob-3 {
    display: none;
  }

  .role-dialog :deep(.el-dialog) {
    width: 95% !important;
    margin: 20px auto !important;
  }

  .permission-actions {
    justify-content: center;
  }

  .permission-actions .el-button {
    padding: 5px 10px;
    font-size: 12px;
  }

  .group-permissions {
    gap: 12px;
  }

  .permission-item {
    min-width: 120px;
  }

  .perm-desc {
    display: none;
  }

  .card-header {
    flex-direction: column;
    gap: 12px;
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
</style>

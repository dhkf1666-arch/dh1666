<template>
  <div class="operation-logs">
    <!-- 背景装饰 -->
    <div class="bg-decoration">
      <div class="blob blob-1"></div>
      <div class="blob blob-2"></div>
      <div class="blob blob-3"></div>
    </div>

    <!-- 统计卡片 -->
    <!-- <div class="stats-grid">
      <div class="stat-card-modern">
        <div class="stat-card-inner">
          <div
            class="stat-icon-wrapper"
            style="
              background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            "
          >
            <el-icon :size="28"><Document /></el-icon>
          </div>
          <div class="stat-details">
            <div class="stat-value">{{ formatNumber(pagination.total) }}</div>
            <div class="stat-label">总操作数</div>
          </div>
        </div>
      </div>
      <div class="stat-card-modern">
        <div class="stat-card-inner">
          <div
            class="stat-icon-wrapper"
            style="
              background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
            "
          >
            <el-icon :size="28"><User /></el-icon>
          </div>
          <div class="stat-details">
            <div class="stat-value">{{ uniqueOperators }}</div>
            <div class="stat-label">操作人数</div>
          </div>
        </div>
      </div>
      <div class="stat-card-modern">
        <div class="stat-card-inner">
          <div
            class="stat-icon-wrapper"
            style="
              background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
            "
          >
            <el-icon :size="28"><TrendCharts /></el-icon>
          </div>
          <div class="stat-details">
            <div class="stat-value">{{ todayCount }}</div>
            <div class="stat-label">今日操作</div>
          </div>
        </div>
      </div>
      <div class="stat-card-modern">
        <div class="stat-card-inner">
          <div
            class="stat-icon-wrapper"
            style="
              background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
            "
          >
            <el-icon :size="28"><Clock /></el-icon>
          </div>
          <div class="stat-details">
            <div class="stat-value">{{ formatNumber(avgExecutionTime) }}ms</div>
            <div class="stat-label">平均耗时</div>
          </div>
        </div>
      </div>
    </div> -->

    <!-- 筛选栏 -->
    <el-card class="filter-card" shadow="hover">
      <el-form :inline="true" :model="filters" class="filter-form">
        <el-form-item label="操作人">
          <el-input
            v-model="filters.operatorName"
            placeholder="请输入姓名"
            clearable
            style="width: 150px"
          />
        </el-form-item>
        <el-form-item label="操作模块">
          <el-select
            v-model="filters.operationModule"
            placeholder="全部"
            clearable
            style="width: 140px"
          >
            <el-option
              v-for="mod in modules"
              :key="mod"
              :label="getModuleName(mod)"
              :value="mod"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="操作类型">
          <el-select
            v-model="filters.operationType"
            placeholder="全部"
            clearable
            style="width: 120px"
          >
            <el-option
              v-for="option in operationTypeOptions"
              :key="option.value"
              :label="option.label"
              :value="option.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            :shortcuts="dateShortcuts"
            style="width: 280px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch" :loading="loading">
            <el-icon><Search /></el-icon>搜索
          </el-button>
          <el-button @click="handleReset">
            <el-icon><Refresh /></el-icon>重置
          </el-button>
          <el-button type="success" @click="handleExport" :loading="exporting">
            <el-icon><Download /></el-icon>导出
          </el-button>
          <el-button
            type="danger"
            @click="showDeleteDialog"
            v-permission="'operation_log:delete'"
          >
            <el-icon><Delete /></el-icon>删除
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 日志表格 -->
    <el-card class="table-card" shadow="hover">
      <el-table
        :data="logs"
        v-loading="loading"
        stripe
        border
        row-key="id"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="50" />
        <el-table-column type="index" width="50" label="序号" />
        <el-table-column prop="operatorName" label="操作人" width="100" />
        <el-table-column prop="operatorRole" label="角色" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="getRoleType(row.operatorRole)">{{
              row.operatorRole || "-"
            }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="operationType" label="操作类型" width="90">
          <template #default="{ row }">
            <el-tag size="small" :type="getTypeTag(row.operationType)">{{
              getTypeName(row.operationType)
            }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="operationModule" label="操作模块" width="120">
          <template #default="{ row }">{{
            getModuleName(row.operationModule)
          }}</template>
        </el-table-column>
        <el-table-column
          label="操作内容"
          min-width="320"
        >
          <template #default="{ row }">
            <div class="operation-content">
              <div class="operation-desc">{{ formatOperationContent(row) }}</div>
              <div
                v-if="formatOperationTarget(row)"
                class="operation-target"
              >
                对象：{{ formatOperationTarget(row) }}
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="ipAddress" label="IP地址" width="140" />
        <el-table-column
          prop="executionTimeMs"
          label="耗时"
          width="80"
          align="center"
        >
          <template #default="{ row }">
            <el-tag size="small" :type="getTimeTag(row.executionTimeMs)"
              >{{ row.executionTimeMs }}ms</el-tag
            >
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="操作时间" width="170">
          <template #default="{ row }">{{
            formatDateTime(row.createdAt)
          }}</template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[20, 50, 100, 200]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadLogs"
          @current-change="loadLogs"
        />
      </div>
    </el-card>

    <!-- 删除对话框 -->
    <el-dialog v-model="deleteDialogVisible" title="删除操作日志" width="450px">
      <el-form label-width="100px">
        <el-form-item label="删除方式">
          <el-radio-group v-model="deleteMode">
            <el-radio value="days">删除 N 天前的日志</el-radio>
            <el-radio value="all">清空所有日志</el-radio>
            <el-radio value="selected">删除选中项</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="deleteMode === 'days'" label="保留天数">
          <el-input-number v-model="deleteDays" :min="1" :max="365" />
          <span class="form-tip">删除 {{ deleteDays }} 天前的日志</span>
        </el-form-item>
        <el-alert v-if="deleteMode === 'all'" type="warning" show-icon>
          此操作将清空所有操作日志，不可恢复！
        </el-alert>
        <el-alert
          v-else-if="deleteMode === 'selected' && selectedRows.length === 0"
          type="info"
          show-icon
        >
          请先在表格中勾选要删除的记录
        </el-alert>
        <el-alert
          v-else-if="deleteMode === 'selected'"
          type="warning"
          show-icon
        >
          将删除选中的 {{ selectedRows.length }} 条记录，不可恢复！
        </el-alert>
      </el-form>
      <template #footer>
        <el-button @click="deleteDialogVisible = false">取消</el-button>
        <el-button type="danger" @click="confirmDelete" :loading="deleting"
          >确认删除</el-button
        >
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { ElMessage } from "element-plus";
import { Search, Refresh, Download, Delete } from "@element-plus/icons-vue";
import adminApi from "@api/admin_api";

const loading = ref(false);
const exporting = ref(false);
const deleting = ref(false);
const logs = ref<any[]>([]);
const modules = ref<string[]>([]);
const types = ref<string[]>([]);
const selectedRows = ref<any[]>([]);
const deleteDialogVisible = ref(false);
const deleteMode = ref("selected");
const deleteDays = ref(90);
const dateRange = ref<[string, string] | null>(null);

const filters = ref({
  operatorName: "",
  operationModule: "",
  operationType: "",
  startDate: "",
  endDate: "",
});

// 修改操作类型选项（只显示修改相关）
const operationTypeOptions = [
  { value: "UPDATE", label: "修改" },
  { value: "CREATE", label: "创建" },
  { value: "DELETE", label: "删除" },
  { value: "UPLOAD", label: "上传" },
];

const pagination = ref({
  page: 1,
  pageSize: 20,
  total: 0,
});

const dateShortcuts = [
  {
    text: "今日",
    value: () => {
      const d = new Date();
      return [d, d];
    },
  },
  {
    text: "昨日",
    value: () => {
      const d = new Date();
      d.setDate(d.getDate() - 1);
      return [d, d];
    },
  },
  {
    text: "近7天",
    value: () => {
      const end = new Date();
      const start = new Date();
      start.setDate(start.getDate() - 6);
      return [start, end];
    },
  },
  {
    text: "近30天",
    value: () => {
      const end = new Date();
      const start = new Date();
      start.setDate(start.getDate() - 29);
      return [start, end];
    },
  },
];

function formatDateTime(time: string) {
  if (!time) return "-";
  return new Date(time).toLocaleString("zh-CN");
}

function getModuleName(module: string) {
  const map: Record<string, string> = {
    USER: "用户管理",
    ROLE: "角色管理",
    DEVICE: "设备管理",
    RECORDING: "录屏管理",
    SCREENSHOT: "截图管理",
    ACTIVITY: "活动日志",
    POLICY: "策略管理",
    ATTENDANCE: "考勤管理",
    PERFORMANCE: "绩效管理",
    PENALTY: "罚款管理",
    SITE_STATS: "出款统计",
    SALARY: "工资管理",
    EMPLOYEE: "员工管理",
    TASK: "任务管理",
    OTHER: "其他",
  };
  return map[module] || module;
}

function formatOperationContent(row: any): string {
  return row.operationDesc || row.operation_desc || "-";
}

function formatOperationTarget(row: any): string {
  const desc = formatOperationContent(row);
  const target = (row.targetName || row.target_name || "").trim();
  if (!target || desc.includes(target)) {
    return "";
  }
  return target;
}

function getTypeName(type: string) {
  const map: Record<string, string> = {
    QUERY: "查询",
    CREATE: "创建",
    UPDATE: "更新",
    DELETE: "删除",
    LOGIN: "登录",
    LOGOUT: "登出",
    UPLOAD: "上传",
    EXPORT: "导出",
    VIEW: "查看",
    DOWNLOAD: "下载",
    OTHER: "其他",
  };
  return map[type] || type;
}

function getTypeTag(type: string) {
  const map: Record<string, string> = {
    QUERY: "info",
    CREATE: "success",
    UPDATE: "warning",
    DELETE: "danger",
    UPLOAD: "primary",
    EXPORT: "success",
  };
  return map[type] || "info";
}

function getRoleType(role: string) {
  const map: Record<string, string> = {
    admin: "danger",
    operator: "warning",
    auditor: "info",
  };
  return map[role] || "info";
}

function getTimeTag(ms: number) {
  if (ms < 100) return "success";
  if (ms < 500) return "warning";
  return "danger";
}

async function loadModulesAndTypes() {
  try {
    const [modRes, typeRes] = await Promise.all([
      adminApi.getOperationModules(),
      adminApi.getOperationTypes(),
    ]);
    modules.value = modRes.data || [];
    types.value = typeRes.data || [];
  } catch (error) {
    console.error("加载筛选选项失败:", error);
  }
}

async function loadLogs() {
  loading.value = true;
  try {
    const params: any = {
      page: pagination.value.page,
      pageSize: pagination.value.pageSize,
    };
    if (filters.value.operatorName)
      params.operatorName = filters.value.operatorName;
    if (filters.value.operationModule)
      params.operationModule = filters.value.operationModule;
    if (filters.value.operationType)
      params.operationType = filters.value.operationType;
    if (dateRange.value && dateRange.value[0])
      params.startDate = dateRange.value[0];
    if (dateRange.value && dateRange.value[1])
      params.endDate = dateRange.value[1];

    const response = await adminApi.getOperationLogs(params);
    logs.value = response.data?.items || [];
    pagination.value.total = response.data?.total || 0;
  } catch (error: any) {
    ElMessage.error(error.message || "加载日志失败");
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.value.page = 1;
  loadLogs();
}

function handleReset() {
  filters.value = {
    operatorName: "",
    operationModule: "",
    operationType: "",
    startDate: "",
    endDate: "",
  };
  dateRange.value = null;
  handleSearch();
}

function handleSelectionChange(rows: any[]) {
  selectedRows.value = rows;
}

async function handleExport() {
  exporting.value = true;
  try {
    const params: any = {};
    if (filters.value.operatorName)
      params.operatorName = filters.value.operatorName;
    if (filters.value.operationModule)
      params.operationModule = filters.value.operationModule;
    if (dateRange.value && dateRange.value[0])
      params.startDate = dateRange.value[0];
    if (dateRange.value && dateRange.value[1])
      params.endDate = dateRange.value[1];

    const response = await adminApi.exportOperationLogs(params);
    const blob = response.data as Blob;
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `operation_logs_${new Date().toISOString().slice(0, 19).replace(/:/g, "-")}.csv`;
    a.click();
    URL.revokeObjectURL(url);
    ElMessage.success("导出成功");
  } catch (error: any) {
    ElMessage.error(error.message || "导出失败");
  } finally {
    exporting.value = false;
  }
}

function showDeleteDialog() {
  deleteDialogVisible.value = true;
}

async function confirmDelete() {
  deleting.value = true;
  try {
    let data: any = {};
    if (deleteMode.value === "days") {
      data.daysOld = deleteDays.value;
    } else if (deleteMode.value === "all") {
      data.deleteAll = true;
    } else {
      if (selectedRows.value.length === 0) {
        ElMessage.warning("请先选择要删除的记录");
        return;
      }
      data.ids = selectedRows.value.map((r) => r.id);
    }
    await adminApi.deleteOperationLogs(data);
    ElMessage.success("删除成功");
    deleteDialogVisible.value = false;
    selectedRows.value = [];
    await loadLogs();
  } catch (error: any) {
    ElMessage.error(error.message || "删除失败");
  } finally {
    deleting.value = false;
  }
}

onMounted(() => {
  loadModulesAndTypes();
  loadLogs();
});
</script>

<style scoped>
.operation-logs {
  padding: 0px;
  min-height: 100vh;
  background: #f0f2f6;
  position: relative;
}

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
}
.blob-2 {
  width: 600px;
  height: 600px;
  background: linear-gradient(135deg, #f093fb, #f5576c);
  bottom: -250px;
  left: -150px;
}
.blob-3 {
  width: 400px;
  height: 400px;
  background: linear-gradient(135deg, #4facfe, #00f2fe);
  top: 40%;
  left: 30%;
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

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 24px;
  position: relative;
  z-index: 1;
}

.stat-card-modern {
  background: #ffffff;
  border-radius: 20px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.06);
  transition: all 0.3s ease;
}
.stat-card-modern:hover {
  transform: translateY(-6px);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.12);
}
.stat-card-inner {
  padding: 20px 24px;
  display: flex;
  align-items: center;
  gap: 16px;
}
.stat-icon-wrapper {
  width: 56px;
  height: 56px;
  border-radius: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.15);
}
.stat-details {
  flex: 1;
}
.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #1a1a2e;
  line-height: 1.2;
  margin-bottom: 6px;
}
.stat-label {
  font-size: 13px;
  color: #666;
  font-weight: 500;
}

.filter-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.3);
  margin-bottom: 20px;
  position: relative;
  z-index: 1;
}
.filter-card :deep(.el-card__body) {
  padding: 16px 20px;
}
.filter-form {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 16px;
}
.filter-form .el-form-item {
  margin-bottom: 0;
}

.table-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.3);
  position: relative;
  z-index: 1;
}
.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
.form-tip {
  margin-left: 12px;
  font-size: 12px;
  color: #909399;
}

.operation-content {
  line-height: 1.5;
}

.operation-desc {
  color: #1e293b;
  font-size: 13px;
  font-weight: 500;
}

.operation-target {
  margin-top: 4px;
  color: #64748b;
  font-size: 12px;
}

@media (max-width: 1200px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  .filter-form {
    flex-direction: column;
    align-items: stretch;
  }
  .filter-form .el-form-item {
    width: 100%;
  }
}
</style>

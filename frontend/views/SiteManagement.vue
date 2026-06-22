<!-- admin_SiteManagement.vue -->
<template>
  <div class="site-management">
    <!-- 背景装饰 -->
    <div class="bg-decoration">
      <div class="blob blob-1"></div>
      <div class="blob blob-2"></div>
      <div class="blob blob-3"></div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div
        class="stat-card-modern"
        v-for="stat in dashboardStats"
        :key="stat.label"
      >
        <div class="stat-card-inner">
          <div class="stat-icon-wrapper" :style="{ background: stat.gradient }">
            <el-icon :size="28"><component :is="stat.icon" /></el-icon>
          </div>
          <div class="stat-details">
            <div class="stat-value">{{ stat.value }}</div>
            <div class="stat-label">{{ stat.label }}</div>
          </div>
        </div>
        <div class="stat-bg-pattern"></div>
      </div>
    </div>

    <el-tabs v-model="activeTab" class="modern-tabs">
      <el-tab-pane label="🏢 出款管理" name="sites">
        <div class="site-tab">
          <el-card class="toolbar-card" shadow="never">
            <div class="toolbar-container">
              <div class="toolbar-left">
                <div class="filter-switch">
                  <el-switch
                    v-model="filterActive"
                    active-text="显示启用"
                    inactive-text="全部"
                    @change="loadSites"
                    size="large"
                  />
                </div>
              </div>
              <div class="toolbar-right">
                <el-button
                  class="action-btn primary-btn"
                  @click="showAddSiteDialog"
                >
                  <el-icon><Plus /></el-icon>新增站点
                </el-button>
                <el-button class="action-btn" @click="loadSites">
                  <el-icon><Refresh /></el-icon>刷新
                </el-button>
              </div>
            </div>
          </el-card>

          <el-card class="table-card-modern" shadow="never">
            <el-table
              v-loading="siteLoading"
              :data="sites"
              stripe
              style="width: 100%"
              :header-cell-style="headerCellStyle"
              :row-class-name="tableRowClassName"
            >
              <el-table-column
                type="index"
                width="60"
                label="序号"
                align="center"
              >
                <template #default="{ $index }">
                  <div class="index-cell">{{ $index + 1 }}</div>
                </template>
              </el-table-column>
              <el-table-column
                prop="code"
                label="站点代码"
                width="120"
                align="center"
              >
                <template #default="{ row }">
                  <el-tag
                    :type="row.isActive ? 'success' : 'info'"
                    effect="light"
                    size="large"
                  >
                    {{ row.code }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="name" label="站点名称" min-width="150" />
              <el-table-column label="员工数" width="100" align="center">
                <template #default="{ row }">
                  <div class="stat-badge employee-badge">
                    <el-icon><User /></el-icon>
                    {{ row.accountCount || 0 }}
                  </div>
                </template>
              </el-table-column>
              <el-table-column label="数据条数" width="100" align="center">
                <template #default="{ row }">
                  <div class="stat-badge data-badge">
                    <el-icon><Document /></el-icon>
                    {{ row.dataCount || 0 }}
                  </div>
                </template>
              </el-table-column>
              <el-table-column
                prop="sortOrder"
                label="排序"
                width="80"
                align="center"
              >
                <template #default="{ row }">
                  <div class="sort-order">{{ row.sortOrder }}</div>
                </template>
              </el-table-column>
              <el-table-column label="状态" width="80" align="center">
                <template #default="{ row }">
                  <el-tag
                    :type="row.isActive ? 'success' : 'info'"
                    size="small"
                  >
                    {{ row.isActive ? "启用" : "禁用" }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column
                label="操作"
                width="150"
                fixed="right"
                align="center"
              >
                <template #default="{ row }">
                  <div class="action-buttons">
                    <el-button link type="primary" @click="editSite(row)">
                      <el-icon><Edit /></el-icon>
                    </el-button>
                    <el-button link type="danger" @click="deleteSite(row)">
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </div>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </div>
      </el-tab-pane>

      <el-tab-pane label="👥 员工账号" name="accounts">
        <div class="accounts-tab">
          <el-card class="toolbar-card" shadow="never">
            <div class="toolbar-container">
              <div class="toolbar-filters">
                <div class="filter-item">
                  <el-select
                    v-model="filterSiteId"
                    placeholder="选择站点"
                    clearable
                    filterable
                    @change="handleSiteFilterChange"
                    size="large"
                    class="filter-select"
                  >
                    <el-option
                      v-for="site in sites"
                      :key="site.id"
                      :label="`${site.code} - ${site.name}`"
                      :value="site.id"
                    />
                  </el-select>
                </div>
                <div class="filter-item">
                  <el-input
                    v-model="filterAccountName"
                    placeholder="搜索账号"
                    clearable
                    @input="handleSearch"
                    size="large"
                    class="filter-input"
                  >
                    <template #prefix>
                      <el-icon><Search /></el-icon>
                    </template>
                  </el-input>
                </div>
                <div class="filter-item">
                  <el-select
                    v-model="filterShift"
                    placeholder="班次筛选"
                    clearable
                    @change="handleFilterChange"
                    size="large"
                    class="filter-select"
                  >
                    <el-option label="🌞 A班" value="day" />
                    <el-option label="🌙 B班" value="night" />
                  </el-select>
                </div>
              </div>
              <div class="toolbar-actions">
                <el-button
                  class="action-btn primary-btn"
                  @click="showAddAccountDialog"
                >
                  <el-icon><Plus /></el-icon>添加员工
                </el-button>
                <el-button class="action-btn" @click="loadAccounts">
                  <el-icon><Refresh /></el-icon>刷新
                </el-button>
              </div>
            </div>
          </el-card>

          <el-card class="table-card-modern" shadow="never">
            <el-table
              v-loading="accountLoading"
              :data="accounts"
              stripe
              style="width: 100%"
              :header-cell-style="headerCellStyle"
              :row-class-name="tableRowClassName"
            >
              <el-table-column
                type="index"
                width="60"
                label="序号"
                align="center"
              >
                <template #default="{ $index }">
                  <div class="index-cell">{{ $index + 1 }}</div>
                </template>
              </el-table-column>
              <el-table-column
                prop="siteCode"
                label="站点"
                width="120"
                align="center"
              >
                <template #default="{ row }">
                  <el-tag size="small" effect="plain">{{
                    row.siteCode
                  }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="name" label="员工姓名" min-width="120">
                <template #default="{ row }">
                  <div class="employee-name-cell">
                    <el-avatar
                      :size="32"
                      :style="{ backgroundColor: getAvatarColor(row.name) }"
                    >
                      {{ row.name?.charAt(0) }}
                    </el-avatar>
                    <span>{{ row.name }}</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column
                prop="accountName"
                label="后台账号"
                min-width="150"
              >
                <template #default="{ row }">
                  <el-tag type="primary" effect="light" size="large">
                    <el-icon><UserFilled /></el-icon>
                    {{ row.accountName }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="班次" width="80" align="center">
                <template #default="{ row }">
                  <el-tag
                    :type="row.shift === 'day' ? 'success' : 'warning'"
                    effect="dark"
                    size="small"
                  >
                    {{ row.shift === "day" ? "🌞 A班" : "🌙 B班" }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="数据条数" width="100" align="center">
                <template #default="{ row }">
                  <div class="stat-badge data-badge">
                    <el-icon><Document /></el-icon>
                    {{ row.dataCount || 0 }}
                  </div>
                </template>
              </el-table-column>
              <el-table-column label="状态" width="80" align="center">
                <template #default="{ row }">
                  <el-tag
                    :type="row.isActive ? 'success' : 'info'"
                    size="small"
                  >
                    {{ row.isActive ? "启用" : "禁用" }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column
                label="操作"
                width="150"
                fixed="right"
                align="center"
              >
                <template #default="{ row }">
                  <div class="action-buttons">
                    <el-button link type="primary" @click="editAccount(row)">
                      <el-icon><Edit /></el-icon>
                    </el-button>
                    <el-button link type="danger" @click="deleteAccount(row)">
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </div>
                </template>
              </el-table-column>
            </el-table>
            <div class="pagination-modern" v-if="accountTotal > 0">
              <el-pagination
                v-model:current-page="accountPage"
                v-model:page-size="accountPageSize"
                :page-sizes="[10, 20, 50, 100]"
                :total="accountTotal"
                layout="total, sizes, prev, pager, next, jumper"
                @size-change="handleAccountSizeChange"
                @current-change="handleAccountCurrentChange"
                background
              />
            </div>
          </el-card>
        </div>
      </el-tab-pane>
    </el-tabs>

    <!-- 站点对话框 -->
    <el-dialog
      v-model="siteDialogVisible"
      :title="siteDialogTitle"
      width="500px"
      class="modern-dialog"
    >
      <el-form
        ref="siteFormRef"
        :model="siteFormData"
        :rules="siteFormRules"
        label-width="80px"
      >
        <el-form-item label="站点代码" prop="code">
          <el-input
            v-model="siteFormData.code"
            placeholder="如: 25S"
            :disabled="isEditSite"
            size="large"
          />
        </el-form-item>
        <el-form-item label="站点名称" prop="name">
          <el-input
            v-model="siteFormData.name"
            placeholder="如: 25S站点"
            size="large"
          />
        </el-form-item>
        <el-form-item label="排序" prop="sort_order">
          <el-input-number
            v-model="siteFormData.sort_order"
            :min="0"
            :max="100"
            size="large"
          />
        </el-form-item>
        <el-form-item label="状态" prop="is_active">
          <el-switch
            v-model="siteFormData.is_active"
            active-text="启用"
            inactive-text="禁用"
            size="large"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="siteDialogVisible = false" size="large"
          >取消</el-button
        >
        <el-button
          type="primary"
          @click="submitSite"
          :loading="siteSubmitting"
          size="large"
        >
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 员工账号对话框 -->
    <el-dialog
      v-model="accountDialogVisible"
      :title="accountDialogTitle"
      width="500px"
      class="modern-dialog"
    >
      <el-form
        ref="accountFormRef"
        :model="accountFormData"
        :rules="accountFormRules"
        label-width="80px"
      >
        <el-form-item label="站点" prop="site_id">
          <el-select
            v-model="accountFormData.site_id"
            placeholder="请选择站点"
            style="width: 100%"
            size="large"
          >
            <el-option
              v-for="site in sites"
              :key="site.id"
              :label="`${site.code} - ${site.name}`"
              :value="site.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="员工姓名" prop="name">
          <el-input
            v-model="accountFormData.name"
            placeholder="请输入员工姓名"
            size="large"
          />
        </el-form-item>
        <el-form-item label="后台账号" prop="account_name">
          <el-input
            v-model="accountFormData.account_name"
            placeholder="如: dhcs1919"
            size="large"
          />
        </el-form-item>
        <el-form-item label="班次" prop="shift">
          <el-radio-group v-model="accountFormData.shift" size="large">
            <el-radio value="day">🌞 A班</el-radio>
            <el-radio value="night">🌙 B班</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="状态" prop="is_active">
          <el-switch
            v-model="accountFormData.is_active"
            active-text="启用"
            inactive-text="禁用"
            size="large"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="accountDialogVisible = false" size="large"
          >取消</el-button
        >
        <el-button
          type="primary"
          @click="submitAccount"
          :loading="accountSubmitting"
          size="large"
        >
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onActivated } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import {
  Plus,
  Refresh,
  Edit,
  Delete,
  User,
  Document,
  UserFilled,
  Search,
} from "@element-plus/icons-vue";
import adminApi from "@api/admin_api";

// 选项卡
const activeTab = ref("sites");

// 站点数据
const siteLoading = ref(false);
const sites = ref([]);
const filterActive = ref(true);

// 员工账号数据
const accountLoading = ref(false);
const accounts = ref([]);
const filterSiteId = ref("");
const filterAccountName = ref("");
const filterShift = ref("");

// 仪表盘统计数据
const dashboardStats = computed(() => {
  const totalSites = sites.value.length;
  const totalAccounts = accounts.value.length;
  const activeSites = sites.value.filter((s) => s.isActive).length;
  const activeAccounts = accounts.value.filter((a) => a.isActive).length;

  return [
    {
      icon: User,
      label: "总站点数",
      value: totalSites,
      gradient: "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
    },
    {
      icon: Document,
      label: "总员工数",
      value: totalAccounts,
      gradient: "linear-gradient(135deg, #f093fb 0%, #f5576c 100%)",
    },
    {
      icon: UserFilled,
      label: "启用站点",
      value: activeSites,
      gradient: "linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)",
    },
    {
      icon: Search,
      label: "启用员工",
      value: activeAccounts,
      gradient: "linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)",
    },
  ];
});

// 站点对话框
const siteDialogVisible = ref(false);
const siteDialogTitle = ref("新增站点");
const isEditSite = ref(false);
const siteSubmitting = ref(false);
const siteFormRef = ref(null);
const siteFormData = ref({
  id: null,
  code: "",
  name: "",
  sort_order: 0,
  is_active: true,
});
const siteFormRules = {
  code: [{ required: true, message: "请输入站点代码", trigger: "blur" }],
  name: [{ required: true, message: "请输入站点名称", trigger: "blur" }],
};

// 员工账号对话框
const accountDialogVisible = ref(false);
const accountDialogTitle = ref("添加员工");
const isEditAccount = ref(false);
const accountSubmitting = ref(false);
const accountFormRef = ref(null);
const accountFormData = ref({
  id: null,
  site_id: "",
  name: "",
  account_name: "",
  shift: "day",
  is_active: true,
});
const accountFormRules = {
  site_id: [{ required: true, message: "请选择站点", trigger: "change" }],
  name: [{ required: true, message: "请输入员工姓名", trigger: "blur" }],
  account_name: [
    { required: true, message: "请输入后台账号", trigger: "blur" },
    { min: 2, message: "账号至少2个字符", trigger: "blur" },
  ],
  shift: [{ required: true, message: "请选择班次", trigger: "change" }],
};

// 表头样式
const headerCellStyle = {
  background: "linear-gradient(135deg, #1a1a2e 0%, #16213e 100%)",
  color: "#ffffff",
  fontWeight: "600",
  fontSize: "14px",
  textAlign: "center",
  borderBottom: "none",
};

// 表格行样式
const tableRowClassName = ({ rowIndex }) => {
  if (rowIndex % 2 === 0) return "even-row";
  return "";
};

// 获取头像颜色
const getAvatarColor = (name) => {
  const colors = [
    "#667eea",
    "#764ba2",
    "#f093fb",
    "#f5576c",
    "#4facfe",
    "#00f2fe",
    "#43e97b",
    "#38f9d7",
    "#fa709a",
    "#fee140",
  ];
  const index = (name?.charCodeAt(0) || 0) % colors.length;
  return colors[index];
};

// 加载站点列表
const loadSites = async () => {
  siteLoading.value = true;
  try {
    const params = {};
    if (filterActive.value) params.is_active = true;
    const response = await adminApi.getSites(params);
    sites.value = response.data?.items || [];
  } catch (error) {
    console.error("加载站点失败:", error);
    ElMessage.error("加载站点失败");
  } finally {
    siteLoading.value = false;
  }
};

// 加载员工账号列表（支持分页）
const loadAccounts = async () => {
  accountLoading.value = true;
  try {
    const params = {
      skip: (accountPage.value - 1) * accountPageSize.value,
      limit: accountPageSize.value,
    };
    if (filterSiteId.value) params.site_id = filterSiteId.value;
    if (filterAccountName.value) params.account_name = filterAccountName.value;
    if (filterShift.value) params.shift = filterShift.value;

    const response = await adminApi.getEmployeeAccounts(params);
    accounts.value = response.data?.items || [];
    accountTotal.value = response.data?.total || 0;
  } catch (error) {
    console.error("加载员工账号失败:", error);
    ElMessage.error("加载员工账号失败");
  } finally {
    accountLoading.value = false;
  }
};

const accountPage = ref(1);
const accountPageSize = ref(20);
const accountTotal = ref(0);

// 分页大小变化
const handleAccountSizeChange = (val) => {
  accountPageSize.value = val;
  accountPage.value = 1;
  loadAccounts();
};

// 当前页变化
const handleAccountCurrentChange = (val) => {
  accountPage.value = val;
  loadAccounts();
};

// 站点筛选变化时重置分页
const handleSiteFilterChange = () => {
  accountPage.value = 1;
  loadAccounts();
};

// 搜索防抖
let searchTimer;
const handleSearch = () => {
  clearTimeout(searchTimer);
  searchTimer = setTimeout(() => {
    accountPage.value = 1;
    loadAccounts();
  }, 300);
};

const handleFilterChange = () => {
  accountPage.value = 1;
  loadAccounts();
};

// ==================== 站点操作 ====================
const showAddSiteDialog = () => {
  isEditSite.value = false;
  siteDialogTitle.value = "新增站点";
  siteFormData.value = {
    id: null,
    code: "",
    name: "",
    sort_order: 0,
    is_active: true,
  };
  siteDialogVisible.value = true;
};

const editSite = (row) => {
  isEditSite.value = true;
  siteDialogTitle.value = "编辑站点";
  siteFormData.value = {
    id: row.id,
    code: row.code,
    name: row.name,
    sort_order: row.sortOrder,
    is_active: row.isActive,
  };
  siteDialogVisible.value = true;
};

const deleteSite = (row) => {
  ElMessageBox.confirm(`确定要删除站点 "${row.code}" 吗？`, "警告", {
    confirmButtonText: "确定",
    cancelButtonText: "取消",
    type: "warning",
  }).then(async () => {
    try {
      await adminApi.deleteSite(row.id);
      ElMessage.success("删除成功");
      loadSites();
      loadAccounts();
    } catch (error) {
      console.error("删除失败:", error);
      ElMessage.error(error.response?.data?.detail || "删除失败");
    }
  });
};

const submitSite = async () => {
  if (!siteFormRef.value) return;
  await siteFormRef.value.validate(async (valid) => {
    if (valid) {
      siteSubmitting.value = true;
      try {
        const submitData = {
          code: siteFormData.value.code,
          name: siteFormData.value.name,
          sort_order: siteFormData.value.sort_order,
          is_active: siteFormData.value.is_active,
        };

        if (isEditSite.value) {
          await adminApi.updateSite(siteFormData.value.id, submitData);
          ElMessage.success("站点更新成功");
        } else {
          await adminApi.createSite(submitData);
          ElMessage.success("站点创建成功");
        }
        siteDialogVisible.value = false;
        await loadSites();
        await loadAccounts();
      } catch (error) {
        console.error("提交失败:", error);
        ElMessage.error(error.response?.data?.detail || "提交失败");
      } finally {
        siteSubmitting.value = false;
      }
    }
  });
};

// ==================== 员工账号操作 ====================
const showAddAccountDialog = () => {
  isEditAccount.value = false;
  accountDialogTitle.value = "添加员工";
  accountFormData.value = {
    id: null,
    site_id: "",
    name: "",
    account_name: "",
    shift: "day",
    is_active: true,
  };
  accountDialogVisible.value = true;
};

const editAccount = (row) => {
  isEditAccount.value = true;
  accountDialogTitle.value = "编辑员工";
  accountFormData.value = {
    id: row.id,
    site_id: row.siteId,
    name: row.name,
    account_name: row.accountName,
    shift: row.shift,
    is_active: row.isActive,
  };
  accountDialogVisible.value = true;
};

const deleteAccount = (row) => {
  ElMessageBox.confirm(
    `确定要删除员工 "${row.name}" (账号: ${row.account_name}) 吗？`,
    "警告",
    {
      confirmButtonText: "确定",
      cancelButtonText: "取消",
      type: "warning",
    },
  ).then(async () => {
    try {
      await adminApi.deleteEmployeeAccount(row.id);
      ElMessage.success("删除成功");
      loadAccounts();
    } catch (error) {
      console.error("删除失败:", error);
      ElMessage.error(error.response?.data?.detail || "删除失败");
    }
  });
};

const submitAccount = async () => {
  if (!accountFormRef.value) return;
  await accountFormRef.value.validate(async (valid) => {
    if (valid) {
      accountSubmitting.value = true;
      try {
        if (isEditAccount.value) {
          await adminApi.updateEmployeeAccount(accountFormData.value.id, {
            name: accountFormData.value.name,
            account_name: accountFormData.value.account_name,
            shift: accountFormData.value.shift,
            is_active: accountFormData.value.is_active,
          });
          ElMessage.success("员工更新成功");
        } else {
          await adminApi.createEmployeeAccount({
            site_id: accountFormData.value.site_id,
            name: accountFormData.value.name,
            account_name: accountFormData.value.account_name,
            shift: accountFormData.value.shift,
          });
          ElMessage.success("员工添加成功");
        }
        accountDialogVisible.value = false;
        loadAccounts();
      } catch (error) {
        console.error("提交失败:", error);
        ElMessage.error(error.response?.data?.detail || "提交失败");
      } finally {
        accountSubmitting.value = false;
      }
    }
  });
};

const refreshData = () => {
  loadSites();
  loadAccounts();
};

onMounted(refreshData);

onActivated(refreshData);
</script>

<style scoped>
.site-management {
  padding: 0px;
  min-height: 100vh;
  background: #f0f2f6;
  position: relative;
}

/* 背景装饰 */
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

/* 统计卡片网格 */
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
  position: relative;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.06);
  transition: all 0.3s ease;
  cursor: pointer;
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
  position: relative;
  z-index: 2;
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

.stat-bg-pattern {
  position: absolute;
  bottom: -20px;
  right: -20px;
  width: 100px;
  height: 100px;
  background: repeating-linear-gradient(
    45deg,
    rgba(0, 0, 0, 0.02) 0px,
    rgba(0, 0, 0, 0.02) 2px,
    transparent 2px,
    transparent 8px
  );
  border-radius: 50%;
  pointer-events: none;
}

/* 现代化选项卡 */
.modern-tabs {
  position: relative;
  z-index: 1;
}

.modern-tabs :deep(.el-tabs__header) {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  margin-bottom: 20px;
  padding: 8px 20px;
  border: 1px solid rgba(255, 255, 255, 0.3);
}

.modern-tabs :deep(.el-tabs__item) {
  font-size: 15px;
  font-weight: 500;
  padding: 0 28px;
  height: 48px;
  line-height: 48px;
}

.modern-tabs :deep(.el-tabs__item.is-active) {
  color: #667eea;
  font-weight: 600;
}

.modern-tabs :deep(.el-tabs__active-bar) {
  background: linear-gradient(90deg, #667eea, #764ba2);
  height: 3px;
}

/* 工具栏卡片 */
.toolbar-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.3);
  margin-bottom: 20px;
  transition: all 0.3s ease;
}

.toolbar-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
}

.toolbar-card :deep(.el-card__body) {
  padding: 16px 20px;
}

.toolbar-container {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}

.toolbar-filters {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.filter-item {
  min-width: 160px;
}

.filter-select,
.filter-input {
  width: 100%;
}

.toolbar-actions,
.toolbar-right {
  display: flex;
  gap: 12px;
}

/* 现代化表格卡片 */
.table-card-modern {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.3);
  overflow: hidden;
  transition: all 0.3s ease;
}

.table-card-modern:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
}

/* 按钮样式 */
.action-btn {
  border-radius: 12px;
  padding: 10px 20px;
  font-weight: 500;
  transition: all 0.3s ease;
  background: #f5f5f5;
  border-color: #e0e0e0;
  color: #666;
}

.action-btn:hover {
  transform: translateY(-2px);
  background: #eee;
}

.primary-btn {
  background: linear-gradient(135deg, #667eea, #764ba2);
  border: none;
  color: white;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.primary-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(102, 126, 234, 0.4);
  color: white;
}

/* 表格样式增强 */
:deep(.el-table) {
  background: transparent;
  font-size: 13px;
}

:deep(.el-table th) {
  font-weight: 600;
  padding: 14px 0;
}

:deep(.el-table td) {
  padding: 12px 0;
}

:deep(.even-row) {
  background-color: #fafafa;
}

:deep(.el-table__row:hover) {
  background-color: #e6f7ff !important;
}

.index-cell {
  font-weight: 600;
  color: #667eea;
}

.stat-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 500;
}

.employee-badge {
  background: #e6f7ff;
  color: #1890ff;
}

.data-badge {
  background: #f6ffed;
  color: #52c41a;
}

.sort-order {
  display: inline-block;
  padding: 4px 10px;
  background: #f5f5f5;
  border-radius: 12px;
  font-weight: 500;
  color: #666;
}

.action-buttons {
  display: flex;
  gap: 8px;
  justify-content: center;
}

.action-buttons .el-button {
  padding: 6px;
  font-size: 16px;
}

.employee-name-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* 现代化分页 */
.pagination-modern {
  margin-top: 20px;
  padding: 16px 20px;
  display: flex;
  justify-content: flex-end;
  border-top: 1px solid #e4e7ed;
}

.pagination-modern :deep(.el-pagination) {
  --el-pagination-bg-color: transparent;
}

.pagination-modern
  :deep(.el-pagination.is-background .el-pager li:not(.is-disabled).is-active) {
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: white;
}

/* 现代化对话框 */
.modern-dialog :deep(.el-dialog) {
  border-radius: 20px;
  overflow: hidden;
}

.modern-dialog :deep(.el-dialog__header) {
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: white;
  padding: 18px 24px;
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
  padding: 24px;
}

.modern-dialog :deep(.el-dialog__footer) {
  padding: 16px 24px;
  border-top: 1px solid #e4e7ed;
}

/* 表单样式优化 */
:deep(.el-form-item__label) {
  font-weight: 500;
  color: #1a1a2e;
}

:deep(.el-input__wrapper) {
  border-radius: 10px;
  transition: all 0.3s ease;
}

:deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px #667eea;
}

:deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 2px rgba(102, 126, 234, 0.2);
}

:deep(.el-select .el-input__wrapper) {
  border-radius: 10px;
}

/* 响应式 */
@media (max-width: 1200px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 16px;
  }

  .site-management {
    padding: 0px;
  }

  .blob-1,
  .blob-2,
  .blob-3 {
    opacity: 0.2;
  }
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }

  .site-management {
    padding: 0px;
  }

  .blob-1,
  .blob-2,
  .blob-3 {
    display: none;
  }

  .toolbar-container {
    flex-direction: column;
    align-items: stretch;
  }

  .toolbar-filters {
    flex-direction: column;
  }

  .filter-item {
    width: 100%;
  }

  .toolbar-actions,
  .toolbar-right {
    justify-content: flex-end;
  }

  .modern-tabs :deep(.el-tabs__item) {
    padding: 0 16px;
    font-size: 13px;
  }
}
</style>

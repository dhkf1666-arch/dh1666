<template>
  <div class="monthly-stats">
    <!-- 背景装饰 -->
    <div class="bg-decoration">
      <div class="blob blob-1"></div>
      <div class="blob blob-2"></div>
      <div class="blob blob-3"></div>
    </div>

    <!-- 筛选栏 - 玻璃态 -->
    <el-card class="filter-card" shadow="never">
      <div class="filter-container">
        <div class="filter-group">
          <div class="filter-item">
            <label class="filter-label">
              <el-icon><Monitor /></el-icon>
              班次
            </label>
            <el-select
              v-model="filters.shift"
              placeholder="全部班次"
              clearable
              @change="loadData"
              size="large"
              class="filter-select"
            >
              <el-option label="白班" value="day" />
              <el-option label="夜班" value="night" />
            </el-select>
          </div>
          <div class="filter-item">
            <label class="filter-label">
              <el-icon><Calendar /></el-icon>
              月份
            </label>
            <el-date-picker
              v-model="filters.month"
              type="month"
              placeholder="选择月份"
              format="YYYY年MM月"
              value-format="YYYY-MM"
              @change="loadData"
              size="large"
              class="filter-date"
            />
          </div>
          <div class="filter-item">
            <label class="filter-label">
              <el-icon><Rank /></el-icon>
              排行依据
            </label>
            <el-select
              v-model="filters.rankBy"
              placeholder="排行依据"
              @change="loadData"
              size="large"
              class="filter-select"
            >
              <el-option label="处理订单总数" value="total_value" />
              <el-option label="平均处理时间" value="avg_time" />
              <el-option label="最快平均时间" value="fastest" />
            </el-select>
          </div>
        </div>
        <div class="filter-actions">
          <el-button @click="resetFilters" class="action-btn reset-btn">
            <el-icon><Refresh /></el-icon>重置
          </el-button>
          <el-button
            type="primary"
            @click="exportData"
            class="action-btn export-btn"
          >
            <el-icon><Download /></el-icon>导出报表
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 统计卡片 - 现代风格 -->
    <div class="stats-grid">
      <div
        class="stat-card-modern"
        v-for="stat in modernStats"
        :key="stat.label"
      >
        <div class="stat-card-inner">
          <div class="stat-icon-wrapper" :style="{ background: stat.gradient }">
            <el-icon :size="28"><component :is="stat.icon" /></el-icon>
          </div>
          <div class="stat-details">
            <div class="stat-value">
              <span>{{ stat.displayValue }}</span>
            </div>
            <div class="stat-label">{{ stat.label }}</div>
          </div>
        </div>
        <div class="stat-bg-pattern"></div>
      </div>
    </div>

    <!-- 排行榜区域 - 三栏布局 -->
    <div class="rankings-container">
      <div
        class="rank-card-modern"
        v-for="(rank, idx) in rankConfigs"
        :key="idx"
      >
        <div class="rank-header" :class="`rank-header-${idx}`">
          <div class="rank-header-left">
            <div class="rank-icon" :class="`rank-icon-${idx}`">
              <el-icon><component :is="rank.icon" /></el-icon>
            </div>
            <span class="rank-title">{{ rank.title }}</span>
          </div>
          <div class="rank-badge">{{ rank.badge }}</div>
        </div>
        <div class="rank-list-modern">
          <div
            v-for="(item, index) in rank.data"
            :key="item.employee_name"
            class="rank-item-modern"
            :class="{ 'is-top': index < 3 }"
          >
            <div class="rank-number-wrap">
              <div class="rank-medal" v-if="index === 0">🥇</div>
              <div class="rank-medal" v-else-if="index === 1">🥈</div>
              <div class="rank-medal" v-else-if="index === 2">🥉</div>
              <div class="rank-number-plain" v-else>{{ index + 1 }}</div>
            </div>
            <div class="rank-info-main">
              <div class="rank-name">{{ item.employee_name }}</div>
              <div class="rank-stats">
                <span class="stat-badge">{{ item.total_value }} 笔</span>
                <span
                  class="time-badge"
                  :class="getTimeBadgeClass(item.total_avg_seconds)"
                >
                  {{ item.total_avg_time }}
                </span>
              </div>
            </div>
            <div class="rank-value-bar">
              <div
                class="progress-bar"
                :style="{
                  width: getRankProgress(item.total_value, rank.maxValue) + '%',
                }"
              ></div>
            </div>
          </div>
          <div v-if="rank.data.length === 0" class="empty-rank-modern">
            <el-icon><Document /></el-icon>
            <span>暂无数据</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 详细数据表格 - 高级表格 -->
    <el-card class="table-card-modern" shadow="never">
      <template #header>
        <div class="table-header-modern">
          <div class="header-left">
            <el-icon class="header-icon"><DataAnalysis /></el-icon>
            <span class="header-title">月度详细统计</span>
            <div class="header-tags" v-if="filters.month || filters.shift">
              <el-tag
                size="small"
                v-if="filters.month"
                type="info"
                effect="plain"
                >{{ filters.month }}</el-tag
              >
              <el-tag
                size="small"
                v-if="filters.shift === 'day'"
                type="info"
                effect="plain"
                >白班</el-tag
              >
              <el-tag
                size="small"
                v-else-if="filters.shift === 'night'"
                type="info"
                effect="plain"
                >夜班</el-tag
              >
            </div>
          </div>
          <div class="header-right">
            <el-tag type="success" effect="dark" size="small"
              >{{ summaryData.length }} 名员工</el-tag
            >
          </div>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="sortedData"
        stripe
        style="width: 100%"
        border
        :default-sort="{ prop: 'total_value', order: 'descending' }"
        @sort-change="handleSortChange"
        :header-cell-style="headerCellStyle"
        :row-class-name="tableRowClassName"
      >
        <el-table-column
          type="index"
          width="70"
          label="排名"
          fixed="left"
          align="center"
        >
          <template #default="{ $index }">
            <div class="rank-cell">
              <span v-if="$index === 0" class="rank-medal-small">🥇</span>
              <span v-else-if="$index === 1" class="rank-medal-small">🥈</span>
              <span v-else-if="$index === 2" class="rank-medal-small">🥉</span>
              <span v-else class="rank-number-small">{{ $index + 1 }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column
          prop="employee_name"
          label="处理人"
          min-width="120"
          fixed="left"
          align="center"
        >
          <template #default="{ row }">
            <div class="employee-cell">
              <el-avatar
                :size="32"
                :style="{ backgroundColor: getAvatarColor(row.employee_name) }"
              >
                {{ row.employee_name?.charAt(0) }}
              </el-avatar>
              <span class="employee-name">{{ row.employee_name }}</span>
            </div>
          </template>
        </el-table-column>

        <!-- 动态站点列 -->
        <el-table-column
          v-for="site in siteColumns"
          :key="site"
          :prop="`sites.${site}.value`"
          :label="site"
          min-width="100"
          align="center"
          sortable="custom"
        >
          <template #default="{ row }">
            <div v-if="row.sites && row.sites[site]" class="site-cell">
              <span class="site-value">{{ row.sites[site].value || 0 }}</span>
              <span
                class="site-time"
                v-if="row.sites[site].avg_time_str !== '-'"
              >
                {{ row.sites[site].avg_time_str }}
              </span>
            </div>
            <span v-else class="na-cell">—</span>
          </template>
        </el-table-column>

        <el-table-column
          prop="total_value"
          label="处理订单总数"
          min-width="140"
          align="center"
          sortable="custom"
          fixed="right"
        >
          <template #default="{ row }">
            <div class="total-value-cell">
              <span class="total-number">{{ row.total_value || 0 }}</span>
              <div
                class="total-bar"
                :style="{ width: getTotalPercent(row.total_value) + '%' }"
              ></div>
            </div>
          </template>
        </el-table-column>

        <el-table-column
          prop="total_avg_time"
          label="平均处理时间"
          min-width="130"
          align="center"
          sortable="custom"
          fixed="right"
        >
          <template #default="{ row }">
            <el-tag
              :type="getAvgTimeType(row.total_avg_seconds)"
              effect="light"
              size="large"
              class="time-tag"
            >
              <el-icon><Timer /></el-icon>
              {{ row.total_avg_time || "-" }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { ElMessage } from "element-plus";
import {
  Refresh,
  Download,
  User,
  Document,
  TrendCharts,
  Timer,
  Trophy,
  Warning,
  DataAnalysis,
  Monitor,
  Calendar,
  Rank,
  Clock,
} from "@element-plus/icons-vue";
import adminApi from "@api/admin_api";

const loading = ref(false);
const summaryData = ref([]);
const siteColumns = ref([]);
const sortField = ref("total_value");
const sortOrder = ref("descending");

const filters = ref({
  shift: "",
  month: getCurrentMonth(),
  rankBy: "total_value",
});

// 获取当前月份
function getCurrentMonth() {
  const now = new Date();
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}`;
}

// 格式化秒数为时间字符串
const formatSecondsToTime = (seconds) => {
  if (!seconds || seconds <= 0) return "0秒";
  const minutes = Math.floor(seconds / 60);
  const secs = seconds % 60;
  if (minutes > 0 && secs > 0) return `${minutes}分${secs}秒`;
  if (minutes > 0) return `${minutes}分`;
  return `${secs}秒`;
};

// 统计汇总 - 移到前面定义
const summaryStats = computed(() => {
  let totalEmployees = summaryData.value.length;
  let totalRecords = 0;
  let totalValue = 0;
  let totalWeightedTime = 0;
  let totalWeight = 0;

  summaryData.value.forEach((row) => {
    totalRecords += row.total_value || 0;
    totalValue += row.total_value || 0;

    if (row.total_avg_seconds && row.total_value) {
      totalWeightedTime += row.total_avg_seconds * row.total_value;
      totalWeight += row.total_value;
    }
  });

  const avgTime =
    totalWeight > 0 ? Math.round(totalWeightedTime / totalWeight) : 0;

  return {
    totalEmployees,
    totalRecords,
    totalValue,
    avgTimeStr: formatSecondsToTime(avgTime),
  };
});

// 现代化统计卡片数据
const modernStats = computed(() => [
  {
    icon: User,
    label: "出款人数",
    value: summaryStats.value.totalEmployees,
    displayValue: summaryStats.value.totalEmployees,
    gradient: "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
  },
  {
    icon: Document,
    label: "订单总数",
    value: summaryStats.value.totalRecords,
    displayValue: summaryStats.value.totalRecords,
    gradient: "linear-gradient(135deg, #f093fb 0%, #f5576c 100%)",
  },
  {
    icon: TrendCharts,
    label: "处理订单总数",
    value: summaryStats.value.totalValue,
    displayValue: summaryStats.value.totalValue,
    gradient: "linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)",
  },
  {
    icon: Clock,
    label: "整体平均时间",
    value: summaryStats.value.avgTimeStr,
    displayValue: summaryStats.value.avgTimeStr,
    gradient: "linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)",
  },
]);

// 排行榜配置
const topByValue = computed(() => {
  return [...summaryData.value]
    .sort((a, b) => (b.total_value || 0) - (a.total_value || 0))
    .slice(0, 10);
});

const topByFastest = computed(() => {
  return [...summaryData.value]
    .filter((item) => item.total_avg_seconds > 0)
    .sort(
      (a, b) => (a.total_avg_seconds || 9999) - (b.total_avg_seconds || 9999),
    )
    .slice(0, 10);
});

const topBySlowest = computed(() => {
  return [...summaryData.value]
    .filter((item) => item.total_avg_seconds > 0)
    .sort((a, b) => (b.total_avg_seconds || 0) - (a.total_avg_seconds || 0))
    .slice(0, 10);
});

const rankConfigs = computed(() => {
  const maxTotal = Math.max(
    ...summaryData.value.map((i) => i.total_value || 0),
    1,
  );
  return [
    {
      title: "处理订单排行榜",
      icon: Trophy,
      badge: "TOP 10",
      data: topByValue.value,
      maxValue: maxTotal,
    },
    {
      title: "最快平均时间",
      icon: Timer,
      badge: "效率之星",
      data: topByFastest.value,
      maxValue: maxTotal,
    },
    {
      title: "待提升榜单",
      icon: Warning,
      badge: "关注",
      data: topBySlowest.value,
      maxValue: maxTotal,
    },
  ];
});

// 获取排名进度百分比
const getRankProgress = (value, maxValue) => {
  if (!maxValue || maxValue === 0) return 0;
  return (value / maxValue) * 100;
};

// 获取总订单百分比（用于表格进度条）
const getTotalPercent = (value) => {
  const maxTotal = Math.max(
    ...summaryData.value.map((i) => i.total_value || 0),
    1,
  );
  return (value / maxTotal) * 100;
};

// 时间标签样式
const getTimeBadgeClass = (seconds) => {
  if (!seconds) return "";
  if (seconds < 60) return "fast";
  if (seconds < 120) return "medium";
  return "slow";
};

// 排序后的数据
const sortedData = computed(() => {
  let data = [...summaryData.value];

  if (sortField.value === "total_value") {
    data.sort((a, b) => {
      const valA = a.total_value || 0;
      const valB = b.total_value || 0;
      return sortOrder.value === "descending" ? valB - valA : valA - valB;
    });
  } else if (sortField.value === "total_avg_time") {
    data.sort((a, b) => {
      const valA = a.total_avg_seconds || 0;
      const valB = b.total_avg_seconds || 0;
      return sortOrder.value === "descending" ? valB - valA : valA - valB;
    });
  } else if (sortField.value?.startsWith("sites.")) {
    const siteCode = sortField.value.split(".")[1];
    data.sort((a, b) => {
      const valA = a.sites?.[siteCode]?.value || 0;
      const valB = b.sites?.[siteCode]?.value || 0;
      return sortOrder.value === "descending" ? valB - valA : valA - valB;
    });
  }

  return data;
});

// 表格行样式
const tableRowClassName = ({ rowIndex }) => {
  if (rowIndex === 0) return "gold-row";
  if (rowIndex === 1) return "silver-row";
  if (rowIndex === 2) return "bronze-row";
  return "";
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

// 获取平均时间标签类型
const getAvgTimeType = (seconds) => {
  if (!seconds) return "info";
  if (seconds < 60) return "success";
  if (seconds < 120) return "warning";
  return "danger";
};

// 加载数据
const loadData = async () => {
  loading.value = true;
  try {
    const params = {};
    if (filters.value.shift) params.shift = filters.value.shift;
    if (filters.value.month) {
      params.start_date = `${filters.value.month}-01`;
      const lastDay = new Date(filters.value.month + "-01");
      lastDay.setMonth(lastDay.getMonth() + 1);
      lastDay.setDate(0);
      params.end_date = `${filters.value.month}-${String(lastDay.getDate()).padStart(2, "0")}`;
    }

    const response = await adminApi.getSiteStatsSummary(params);
    const data = response.data || {};

    const rawItems = data.items || [];
    summaryData.value = rawItems.map((item) => {
      const convertedSites = {};
      if (item.sites) {
        for (const [siteCode, siteData] of Object.entries(item.sites)) {
          convertedSites[siteCode] = {
            value: siteData.value || 0,
            avg_time_seconds: siteData.avgTimeSeconds || 0,
            avg_time_str: siteData.avgTimeStr || "-",
          };
        }
      }

      return {
        employee_name: item.employeeName,
        account_name: item.accountName,
        total_value: item.totalValue,
        total_avg_seconds: item.totalAvgSeconds,
        total_avg_time: item.totalAvgTime,
        sites: convertedSites,
      };
    });

    siteColumns.value = data.siteColumns || data.site_columns || [];

    console.log("加载月度数据成功:", {
      employees: summaryData.value.length,
      sites: siteColumns.value,
    });
  } catch (error) {
    console.error("加载数据失败:", error);
    ElMessage.error("加载数据失败");
  } finally {
    loading.value = false;
  }
};

// 排序变化处理
const handleSortChange = ({ prop, order }) => {
  if (prop) {
    sortField.value = prop;
    sortOrder.value = order;
  }
};

// 重置筛选
const resetFilters = () => {
  filters.value = {
    shift: "",
    month: getCurrentMonth(),
    rankBy: "total_value",
  };
  loadData();
};

// 导出数据
const exportData = () => {
  if (summaryData.value.length === 0) {
    ElMessage.warning("暂无数据可导出");
    return;
  }

  const exportRows = summaryData.value.map((row, index) => {
    const rowData = {
      排名: index + 1,
      处理人: row.employee_name,
    };
    siteColumns.value.forEach((site) => {
      const siteData = row.sites[site];
      rowData[site] = siteData?.value || "-";
    });
    rowData["处理订单总数"] = row.total_value;
    rowData["平均处理时间"] = row.total_avg_time;
    return rowData;
  });

  const headers = Object.keys(exportRows[0]);
  const csvRows = [
    headers.join(","),
    ...exportRows.map((row) =>
      headers
        .map((h) => {
          const val = row[h] || "";
          return `"${String(val).replace(/"/g, '""')}"`;
        })
        .join(","),
    ),
  ];

  const blob = new Blob(["\uFEFF" + csvRows.join("\n")], {
    type: "text/csv;charset=utf-8;",
  });
  const link = document.createElement("a");
  const url = URL.createObjectURL(blob);
  link.href = url;

  const shiftText =
    filters.value.shift === "day"
      ? "白班"
      : filters.value.shift === "night"
        ? "夜班"
        : "全部";
  link.setAttribute(
    "download",
    `${filters.value.month}_${shiftText}_月度汇总统计.csv`,
  );
  link.click();
  URL.revokeObjectURL(url);
};

onMounted(() => {
  loadData();
});
</script>

<style scoped>
.monthly-stats {
  /* padding: 28px 32px; */
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

/* 筛选卡片 */
.filter-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 20px;
  border: 1px solid rgba(255, 255, 255, 0.3);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.08);
  margin-bottom: 28px;
  position: relative;
  z-index: 1;
}

.filter-card :deep(.el-card__body) {
  padding: 20px 24px;
}

.filter-container {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}

.filter-group {
  display: flex;
  gap: 24px;
  flex-wrap: wrap;
}

.filter-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.filter-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  font-weight: 500;
  color: #1a1a2e;
}

.filter-select,
.filter-date {
  width: 160px;
}

.filter-actions {
  display: flex;
  gap: 12px;
}

.action-btn {
  border-radius: 12px;
  padding: 10px 20px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.reset-btn {
  background: #f5f5f5;
  border-color: #e0e0e0;
  color: #666;
}

.reset-btn:hover {
  background: #eee;
  transform: translateY(-2px);
}

.export-btn {
  background: linear-gradient(135deg, #667eea, #764ba2);
  border: none;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.export-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(102, 126, 234, 0.4);
}

/* 统计卡片网格 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 24px;
  margin-bottom: 32px;
  position: relative;
  z-index: 1;
}

.stat-card-modern {
  background: #ffffff;
  border-radius: 24px;
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
  gap: 18px;
  position: relative;
  z-index: 2;
}

.stat-icon-wrapper {
  width: 60px;
  height: 60px;
  border-radius: 20px;
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
  font-size: 32px;
  font-weight: 700;
  color: #1a1a2e;
  line-height: 1.2;
  margin-bottom: 6px;
}

.stat-value span {
  display: inline-block;
}

.stat-label {
  font-size: 14px;
  color: #666;
  font-weight: 500;
}

.stat-bg-pattern {
  position: absolute;
  bottom: -20px;
  right: -20px;
  width: 120px;
  height: 120px;
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

/* 排行榜容器 */
.rankings-container {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 24px;
  margin-bottom: 32px;
  position: relative;
  z-index: 1;
}

.rank-card-modern {
  background: #ffffff;
  border-radius: 24px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.06);
  transition: all 0.3s ease;
}

.rank-card-modern:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.1);
}

.rank-header {
  padding: 16px 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #f0f0f0;
}

.rank-header-0 {
  background: linear-gradient(135deg, #fff5e6, #fff);
}

.rank-header-1 {
  background: linear-gradient(135deg, #e6f7ff, #fff);
}

.rank-header-2 {
  background: linear-gradient(135deg, #fff0f0, #fff);
}

.rank-header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.rank-icon {
  width: 36px;
  height: 36px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
}

.rank-icon-0 {
  background: linear-gradient(135deg, #ffd700, #ffb347);
  color: white;
}

.rank-icon-1 {
  background: linear-gradient(135deg, #4facfe, #00f2fe);
  color: white;
}

.rank-icon-2 {
  background: linear-gradient(135deg, #f5576c, #f093fb);
  color: white;
}

.rank-title {
  font-size: 16px;
  font-weight: 600;
  color: #1a1a2e;
}

.rank-badge {
  font-size: 12px;
  padding: 4px 10px;
  border-radius: 20px;
  background: rgba(0, 0, 0, 0.05);
  color: #666;
  font-weight: 500;
}

.rank-list-modern {
  max-height: 380px;
  overflow-y: auto;
  padding: 4px 0;
}

.rank-item-modern {
  padding: 12px 20px;
  display: flex;
  align-items: center;
  gap: 14px;
  border-bottom: 1px solid #f5f5f5;
  transition: background 0.2s;
  position: relative;
}

.rank-item-modern:hover {
  background: #fafafa;
}

.rank-item-modern.is-top {
  background: linear-gradient(90deg, #fffbf0, transparent);
}

.rank-number-wrap {
  width: 40px;
  text-align: center;
}

.rank-medal {
  font-size: 24px;
}

.rank-number-plain {
  font-size: 16px;
  font-weight: 600;
  color: #999;
}

.rank-info-main {
  flex: 1;
}

.rank-name {
  font-weight: 600;
  color: #1a1a2e;
  margin-bottom: 6px;
}

.rank-stats {
  display: flex;
  gap: 12px;
  align-items: center;
}

.stat-badge {
  font-size: 12px;
  color: #667eea;
  font-weight: 500;
  background: #f0f0ff;
  padding: 2px 8px;
  border-radius: 12px;
}

.time-badge {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 12px;
  font-weight: 500;
}

.time-badge.fast {
  background: #e6f7e6;
  color: #52c41a;
}

.time-badge.medium {
  background: #fff7e6;
  color: #fa8c16;
}

.time-badge.slow {
  background: #fff0f0;
  color: #f5222d;
}

.rank-value-bar {
  width: 80px;
}

.progress-bar {
  height: 6px;
  background: linear-gradient(90deg, #667eea, #764ba2);
  border-radius: 3px;
  transition: width 0.5s ease;
}

.empty-rank-modern {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 20px;
  color: #999;
  gap: 10px;
}

.empty-rank-modern .el-icon {
  font-size: 48px;
}

/* 表格卡片 */
.table-card-modern {
  background: #ffffff;
  border-radius: 24px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.06);
  position: relative;
  z-index: 1;
}

.table-card-modern :deep(.el-card__header) {
  background: #ffffff;
  border-bottom: 1px solid #f0f0f0;
  padding: 18px 24px;
}

.table-header-modern {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-icon {
  font-size: 22px;
  color: #667eea;
}

.header-title {
  font-size: 18px;
  font-weight: 600;
  color: #1a1a2e;
}

.header-tags {
  display: flex;
  gap: 8px;
  margin-left: 12px;
}

/* 表格样式增强 */
:deep(.el-table) {
  border-radius: 0;
  font-size: 13px;
}

:deep(.el-table th) {
  font-weight: 600;
  padding: 14px 0;
}

:deep(.el-table td) {
  padding: 12px 0;
}

:deep(.gold-row) {
  background: linear-gradient(90deg, rgba(255, 215, 0, 0.08), transparent);
}

:deep(.silver-row) {
  background: linear-gradient(90deg, rgba(192, 192, 192, 0.08), transparent);
}

:deep(.bronze-row) {
  background: linear-gradient(90deg, rgba(205, 127, 50, 0.08), transparent);
}

.rank-cell {
  display: flex;
  justify-content: center;
  align-items: center;
}

.rank-medal-small {
  font-size: 20px;
}

.rank-number-small {
  font-weight: 600;
  color: #999;
}

.employee-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

.employee-name {
  font-weight: 500;
  color: #1a1a2e;
}

.site-cell {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.site-value {
  font-weight: 600;
  color: #1a1a2e;
  font-size: 14px;
}

.site-time {
  font-size: 11px;
  color: #999;
}

.na-cell {
  color: #ccc;
}

.total-value-cell {
  position: relative;
  padding: 4px 8px;
}

.total-number {
  font-weight: 700;
  color: #667eea;
  font-size: 16px;
  position: relative;
  z-index: 1;
}

.total-bar {
  position: absolute;
  left: 0;
  top: 0;
  height: 100%;
  background: linear-gradient(
    90deg,
    rgba(102, 126, 234, 0.15),
    rgba(118, 75, 162, 0.08)
  );
  border-radius: 4px;
  transition: width 0.3s ease;
  pointer-events: none;
}

.time-tag {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 20px;
  font-size: 13px;
}

/* 响应式 */
@media (max-width: 1400px) {
  .stats-grid {
    gap: 18px;
  }

  .rankings-container {
    gap: 18px;
  }

  .monthly-stats {
    padding: 20px;
  }
}

@media (max-width: 1200px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .rankings-container {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }

  .filter-container {
    flex-direction: column;
    align-items: stretch;
  }

  .filter-group {
    flex-direction: column;
  }

  .filter-item {
    width: 100%;
  }

  .filter-select,
  .filter-date {
    flex: 1;
  }

  .filter-actions {
    justify-content: flex-end;
  }
}
</style>

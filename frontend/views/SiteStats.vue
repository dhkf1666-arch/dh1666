<template>
  <div class="site-stats">
    <!-- 背景装饰 -->
    <div class="bg-decoration">
      <div class="blob blob-1"></div>
      <div class="blob blob-2"></div>
      <div class="blob blob-3"></div>
    </div>

    <!-- 筛选栏 -->
    <el-card class="filter-bar" shadow="hover">
      <!-- 原有内容保持不变 -->
      <el-row :gutter="16" align="middle">
        <el-col :span="3">
          <el-input
            v-model="filters.employeeName"
            placeholder="搜索员工姓名"
            clearable
            @clear="handleSearch"
            @keyup.enter="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-col>

        <el-col :span="4">
          <el-select
            v-model="filters.siteId"
            placeholder="选择站点"
            clearable
            filterable
            @change="handleSiteFilterChange"
            style="width: 100%"
          >
            <template #prefix>
              <el-icon><Location /></el-icon>
            </template>
            <el-option
              v-for="site in sites"
              :key="site.id"
              :label="`${site.code} - ${site.name}`"
              :value="site.id"
            />
          </el-select>
        </el-col>

        <el-col :span="5">
          <el-date-picker
            v-model="filters.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            @change="handleDateRangeChange"
            style="width: 100%"
          />
        </el-col>

        <el-col :span="4">
          <el-radio-group
            v-model="activeShift"
            @change="loadData"
            size="default"
          >
            <el-radio-button value="day">🌞 A班</el-radio-button>
            <el-radio-button value="night">🌙 B班</el-radio-button>
          </el-radio-group>
        </el-col>

        <el-col :span="4">
          <el-button-group>
            <el-button
              :type="displayMode === 'site' ? 'primary' : 'default'"
              @click="switchMode('site')"
            >
              站点汇总
            </el-button>
            <el-button
              :type="displayMode === 'stacked' ? 'primary' : 'default'"
              @click="switchMode('stacked')"
            >
              日期堆叠
            </el-button>
          </el-button-group>
        </el-col>
      </el-row>
    </el-card>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: #e6f7ff; color: #1890ff">
              <el-icon><Document /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ summaryStats.totalRecords }}</div>
              <div class="stat-label">总记录数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div
              class="stat-icon"
              :style="
                hoverCardStats.visible
                  ? 'background: #e6f7ff; color: #1890ff'
                  : 'background: #f6ffed; color: #52c41a'
              "
            >
              <el-icon v-if="hoverCardStats.visible"><Calendar /></el-icon>
              <el-icon v-else><TrendCharts /></el-icon>
            </div>
            <div class="stat-info">
              <div
                class="stat-value"
                :class="{ 'hover-value': hoverCardStats.visible }"
              >
                {{
                  hoverCardStats.visible
                    ? hoverCardStats.totalValue
                    : summaryStats.totalValue
                }}
              </div>
              <div class="stat-label">
                <span v-if="hoverCardStats.visible">
                  📅 {{ hoverCardStats.date }} · 👥
                  {{ hoverCardStats.employeeCount }}人出款
                </span>
                <span v-else> 总处理笔数 </span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: #fff7e6; color: #fa8c16">
              <el-icon><Timer /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ summaryStats.avgTimeStr }}</div>
              <div class="stat-label">平均处理时间</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: #f9f0ff; color: #722ed1">
              <el-icon><User /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ summaryStats.activeEmployees }}</div>
              <div class="stat-label">出款人员</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 动态列表格 -->
    <!-- ========== 模式1：站点汇总模式（原有表格） ========== -->
    <el-card v-if="displayMode === 'site'" class="summary-card" shadow="hover">
      <template #header>
        <div
          class="card-header"
          style="
            display: flex;
            justify-content: space-between;
            align-items: center;
          "
        >
          <div>
            <span>
              <el-icon v-if="activeShift === 'day'"><Sunny /></el-icon>
              <el-icon v-else><Moon /></el-icon>
              {{ activeShift === "day" ? "A班" : "B班" }} - 出款统计
            </span>
            <span class="header-tip" v-if="selectedSiteName">
              (当前站点: {{ selectedSiteName }})
            </span>
          </div>
          <div>
            <el-button type="primary" @click="showUploadDialog">
              <el-icon><Upload /></el-icon>上传数据
            </el-button>
            <el-button @click="resetFilters">
              <el-icon><Refresh /></el-icon>重置筛选
            </el-button>
            <el-button @click="exportData" :disabled="summaryData.length === 0">
              <el-icon><Download /></el-icon>导出报表
            </el-button>
            <el-button type="danger" plain @click="showClearDialog">
              <el-icon><Delete /></el-icon>清除指定数据
            </el-button>
            <el-button type="danger" @click="clearAllData" plain>
              <el-icon><Delete /></el-icon>清空所有数据
            </el-button>
          </div>
        </div>
      </template>

      <!-- 原有表格内容保持不变 -->
      <el-table
        v-loading="loading"
        :data="displayData"
        stripe
        style="width: 100%"
        border
        height="calc(100vh - 280px)"
        @sort-change="handleSortChange"
        :header-cell-style="{
          background:
            'linear-gradient(180deg, #A6CDF7 0%, #62ADF0 50%, #0971E0 100%)',
          color: '#1a2a3a',
          fontWeight: 'bold',
          fontSize: '14px',
          borderTop: '1px solid rgba(255,255,255,0.6)',
          borderBottom: '2px solid #7cb3e0',
          boxShadow:
            'inset 0 1px 0 rgba(255,255,255,0.4), 0 2px 6px rgba(0,0,0,0.08)',
          textShadow: '0 1px 0 rgba(255,255,255,0.3)',
        }"
      >
        <el-table-column
          type="index"
          width="60"
          fixed="left"
          label="序号"
          align="center"
        />
        <el-table-column
          label="统计日期"
          width="120"
          fixed="left"
          align="center"
        >
          <template #default>
            <div class="date-cell">
              <span class="display-date">
                {{
                  filters.dateRange && filters.dateRange.length === 2
                    ? `${filters.dateRange[0]} 至 ${filters.dateRange[1]}`
                    : selectedDate
                }}
              </span>
            </div>
          </template>
        </el-table-column>
        <el-table-column
          prop="employee_name"
          label="姓名"
          min-width="70"
          fixed="left"
          align="center"
        />

        <!-- 动态站点列 -->
        <el-table-column
          v-for="site in siteColumns"
          :key="site"
          :label="site"
          min-width="140"
          align="center"
          sortable="custom"
          :prop="`sites.${site}.value`"
        >
          <template #default="{ row }">
            <div v-if="row.sites && row.sites[site]" class="site-cell">
              <div class="value">笔数: {{ row.sites[site].value || 0 }}</div>
              <div class="time">
                平均: {{ row.sites[site].avg_time_str || "-" }}
              </div>
            </div>
            <div v-else class="site-cell empty">
              <div class="value">-</div>
            </div>
          </template>
        </el-table-column>

        <el-table-column
          label="总计"
          min-width="140"
          align="center"
          fixed="right"
          sortable="custom"
          prop="total_value"
        >
          <template #default="{ row }">
            <div class="total-cell">
              <div class="value">笔数: {{ row.total_value || 0 }}</div>
              <div class="time">平均: {{ row.total_avg_time || "-" }}</div>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- ✅ 分页组件 -->
      <div class="pagination" v-if="totalItems > 0">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :total="totalItems"
          :page-sizes="[20, 50, 100, 200]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handlePageSizeChange"
          @current-change="handleCurrentPageChange"
        />
      </div>

      <div v-if="!loading && summaryData.length === 0" class="empty-data">
        <el-empty description="暂无数据，请先上传数据文件" />
      </div>
    </el-card>

    <!-- ========== 模式2：日期堆叠模式 ========== -->
    <el-card
      v-if="displayMode === 'stacked'"
      class="summary-card"
      shadow="hover"
    >
      <template #header>
        <div
          class="card-header"
          style="
            display: flex;
            justify-content: space-between;
            align-items: center;
          "
        >
          <div>
            <span>
              <el-icon v-if="activeShift === 'day'"><Sunny /></el-icon>
              <el-icon v-else><Moon /></el-icon>
              {{ activeShift === "day" ? "A班" : "B班" }} - 出款统计
            </span>
            <span
              class="header-tip"
              v-if="filters.dateRange && filters.dateRange.length === 2"
            >
              ({{ filters.dateRange[0] }} 至 {{ filters.dateRange[1] }})
            </span>
          </div>

          <div>
            <el-button type="primary" @click="showUploadDialog">
              <el-icon><Upload /></el-icon>上传数据
            </el-button>
            <el-button @click="resetFilters">
              <el-icon><Refresh /></el-icon>重置筛选
            </el-button>
            <el-button
              @click="exportStackedData"
              :disabled="stackedDisplayData.length === 0"
            >
              <el-icon><Download /></el-icon>导出报表
            </el-button>
            <el-button type="danger" plain @click="showClearDialog">
              <el-icon><Delete /></el-icon>清除指定数据
            </el-button>
            <el-button type="danger" plain @click="showDeleteDateDialog">
              <el-icon><Delete /></el-icon>删除指定日期数据
            </el-button>
          </div>
        </div>
      </template>

      <!-- 全局固定表头（Grid布局） -->
      <div class="global-table-header">
        <div
          class="grid-header"
          :style="{
            gridTemplateColumns: `60px 140px minmax(100px, 1fr) repeat(${siteColumns.length}, minmax(100px, 1fr)) minmax(100px, 1fr)`,
          }"
        >
          <div>序号</div>
          <div>统计日期</div>
          <div>姓名</div>
          <div v-for="site in siteColumns" :key="site">{{ site }}</div>
          <div>总计</div>
        </div>
      </div>

      <!-- 每日卡片容器 -->
      <div v-loading="stackedLoading" class="daily-cards-container">
        <div
          v-if="dailyCardData.length === 0 && !stackedLoading"
          class="empty-data"
        >
          <el-empty description="暂无数据" />
        </div>

        <div
          v-for="dateGroup in paginatedDailyCards"
          :key="dateGroup.date"
          class="daily-card"
          @mouseenter="onCardMouseEnter(dateGroup)"
          @mouseleave="onCardMouseLeave"
        >
          <!-- 数据行 -->
          <div
            v-for="(employee, idx) in dateGroup.employees"
            :key="idx"
            class="grid-row"
            :class="{ 'row-striped': idx % 2 === 1 }"
            :style="{
              gridTemplateColumns: `60px 140px minmax(100px, 1fr) repeat(${siteColumns.length}, minmax(100px, 1fr)) minmax(100px, 1fr)`,
            }"
          >
            <!-- 序号 -->
            <div class="col-index">{{ idx + 1 }}</div>

            <!-- 统计日期（带 tooltip 提示） -->
            <div class="col-date">
              <el-tooltip
                :content="`${dateGroup.date}\n${dateGroup.employeeCount}人出款\n当日总:${dateGroup.totalValue}笔`"
                placement="top"
                effect="dark"
                :show-after="300"
              >
                <div class="date-cell-simple">
                  {{ dateGroup.date }}
                </div>
              </el-tooltip>
            </div>

            <!-- 姓名 -->
            <div
              class="col-name"
              :class="{
                'search-highlight': isSearchMatch(employee.employee_name),
              }"
            >
              {{ employee.employee_name }}
            </div>

            <!-- 站点列 -->
            <div v-for="site in siteColumns" :key="site" class="col-site">
              <template v-if="employee.sites && employee.sites[site]">
                <div class="site-value">
                  {{ employee.sites[site].value || 0 }}笔
                </div>
                <div class="site-time">
                  {{ employee.sites[site].avg_time_str || "-" }}
                </div>
              </template>
              <span v-else>-</span>
            </div>

            <!-- 总计 -->
            <div class="col-total">
              <div class="total-value">{{ employee.total_value || 0 }}笔</div>
              <div class="total-time">{{ employee.total_avg_time || "-" }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- 分页组件 -->
      <div class="pagination" v-if="datePagesTotal > 0">
        <el-pagination
          v-model:current-page="dateCurrentPage"
          v-model:page-size="datePageSize"
          :total="datePagesTotal"
          :page-sizes="[5, 10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleDatePageSizeChange"
          @current-change="handleDatePageChange"
        />
      </div>
    </el-card>

    <!-- 上传对话框 -->
    <el-dialog v-model="uploadVisible" title="上传站点数据" width="820px">
      <el-form :model="uploadForm" label-width="110px">
        <el-form-item label="出款模式" required>
          <el-radio-group v-model="uploadForm.mode" @change="handleUploadModeChange">
            <el-radio value="BX">🌐 BX模式</el-radio>
            <el-radio value="NN">🆕 NN模式</el-radio>
          </el-radio-group>
          <div class="form-tip">
            选择文件后可在下方手动指定列；系统也会根据表头自动推荐。
          </div>
        </el-form-item>

        <el-form-item label="选择站点" required>
          <el-select
            v-model="uploadForm.siteId"
            placeholder="请选择站点"
            style="width: 100%"
            @change="handleSiteChange"
          >
            <el-option
              v-for="site in sites"
              :key="site.id"
              :label="`${site.code} - ${site.name}`"
              :value="site.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="选择班次" required>
          <el-radio-group v-model="uploadForm.shift">
            <el-radio value="day">🌞 A班</el-radio>
            <el-radio value="night">🌙 B班</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="统计日期" required>
          <el-date-picker
            v-model="uploadForm.statDate"
            type="date"
            placeholder="选择统计日期"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            style="width: 100%"
          />
        </el-form-item>

        <el-form-item label="上传文件" required>
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :on-change="handleFileChange"
            :limit="1"
            accept=".xlsx,.xls,.csv"
          >
            <el-button type="primary"
              ><el-icon><Upload /></el-icon>选择文件</el-button
            >
            <template #tip>
              <div class="el-upload__tip">
                支持 .xlsx / .xls / .csv。上传后请确认下方三列映射是否正确。
              </div>
            </template>
          </el-upload>
        </el-form-item>

        <el-form-item
          v-if="columnOptions.length > 0"
          label="列映射"
          required
        >
          <el-row :gutter="12" class="column-mapping-row">
            <el-col :span="8">
              <div class="column-mapping-label">开始时间</div>
              <el-select
                v-model="uploadForm.startCol"
                placeholder="选择列"
                filterable
                style="width: 100%"
              >
                <el-option
                  v-for="col in columnOptions"
                  :key="`start-${col.index}`"
                  :label="col.label"
                  :value="col.index"
                />
              </el-select>
            </el-col>
            <el-col :span="8">
              <div class="column-mapping-label">完成时间</div>
              <el-select
                v-model="uploadForm.endCol"
                placeholder="选择列"
                filterable
                style="width: 100%"
              >
                <el-option
                  v-for="col in columnOptions"
                  :key="`end-${col.index}`"
                  :label="col.label"
                  :value="col.index"
                />
              </el-select>
            </el-col>
            <el-col :span="8">
              <div class="column-mapping-label">操作人</div>
              <el-select
                v-model="uploadForm.accountCol"
                placeholder="选择列"
                filterable
                style="width: 100%"
              >
                <el-option
                  v-for="col in columnOptions"
                  :key="`account-${col.index}`"
                  :label="col.label"
                  :value="col.index"
                />
              </el-select>
            </el-col>
          </el-row>
          <div v-if="columnMappingHint" class="form-tip">{{ columnMappingHint }}</div>
        </el-form-item>

        <el-form-item v-else-if="columnPreviewLoading" label="列映射">
          <span class="form-tip">正在读取文件表头...</span>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="uploadVisible = false">取消</el-button>
        <el-button type="primary" @click="uploadData" :loading="uploading"
          >开始上传</el-button
        >
      </template>
    </el-dialog>

    <!-- 清除数据对话框 -->
    <el-dialog
      v-model="clearDialogVisible"
      title="清除站点数据"
      width="450px"
      @close="resetClearForm"
    >
      <el-form :model="clearForm" label-width="80px">
        <el-form-item label="选择站点" required>
          <el-select
            v-model="clearForm.siteId"
            placeholder="请选择站点"
            filterable
            style="width: 100%"
          >
            <el-option
              v-for="site in sites"
              :key="site.id"
              :label="`${site.code} - ${site.name}`"
              :value="site.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="选择班次" required>
          <el-radio-group v-model="clearForm.shift">
            <el-radio value="day">🌞 A班</el-radio>
            <el-radio value="night">🌙 B班</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="选择日期" required>
          <el-date-picker
            v-model="clearForm.date"
            type="date"
            placeholder="选择要清除的日期"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            style="width: 100%"
            :disabled-date="disabledFutureDate"
          />
        </el-form-item>

        <el-alert title="⚠️ 警告" type="warning" :closable="false" show-icon>
          <template #default>
            此操作将删除该站点 {{ clearForm.shift === "day" ? "A班" : "B班" }}
            {{ clearForm.date || "所选日期" }} 的所有数据，不可恢复！
          </template>
        </el-alert>
      </el-form>

      <template #footer>
        <el-button @click="clearDialogVisible = false">取消</el-button>
        <el-button
          type="danger"
          @click="confirmClearData"
          :loading="clearingData"
        >
          确认清除
        </el-button>
      </template>
    </el-dialog>

    <!-- 删除指定日期数据对话框 -->
    <el-dialog
      v-model="deleteDateDialogVisible"
      title="删除指定日期数据"
      width="450px"
      @close="resetDeleteDateForm"
    >
      <el-form :model="deleteDateForm" label-width="80px">
        <el-form-item label="选择日期" required>
          <el-date-picker
            v-model="deleteDateForm.date"
            type="date"
            placeholder="选择要删除的日期"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            style="width: 100%"
            :disabled-date="disabledFutureDate"
          />
        </el-form-item>

        <el-alert title="⚠️ 警告" type="warning" :closable="false" show-icon>
          <template #default>
            此操作将删除
            {{ deleteDateForm.date || "所选日期" }}
            的<b>所有站点、所有班次</b>的数据，不可恢复！
          </template>
        </el-alert>
      </el-form>

      <template #footer>
        <el-button @click="deleteDateDialogVisible = false">取消</el-button>
        <el-button
          type="danger"
          @click="confirmDeleteDateData"
          :loading="deletingDateData"
        >
          确认删除
        </el-button>
      </template>
    </el-dialog>

    <div
      v-show="tooltipVisible"
      class="card-tooltip"
      :style="{ top: tooltipY + 'px', left: tooltipX + 'px' }"
    >
      <div class="tooltip-date">{{ tooltipDate }}</div>
      <div class="tooltip-stats">{{ tooltipEmployeeCount }}人出款</div>
      <div class="tooltip-stats">当日总: {{ tooltipTotalValue }}笔</div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from "vue";
import { ElMessage, ElMessageBox, ElLoading } from "element-plus";
import {
  Sunny,
  Moon,
  Upload,
  Refresh,
  Search,
  Document,
  TrendCharts,
  Timer,
  User,
  Download,
  Delete,
  Calendar,
} from "@element-plus/icons-vue";
import adminApi from "@api/admin_api";
import { getCurrentBeijingTime } from "@api/format";

// ==================== 辅助函数（先定义，避免初始化顺序问题） ====================
const getYesterdayDate = () => {
  const date = new Date();
  date.setDate(date.getDate() - 1);
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
};

function getTodayDate() {
  const beijingTimeStr = getCurrentBeijingTime();
  return beijingTimeStr.split(" ")[0];
}

const formatSecondsToTime = (seconds) => {
  if (!seconds || seconds <= 0) return "0秒";
  const minutes = Math.floor(seconds / 60);
  const secs = seconds % 60;
  if (minutes > 0 && secs > 0) return `${minutes}分${secs}秒`;
  if (minutes > 0) return `${minutes}分`;
  return `${secs}秒`;
};

const formatDateDisplay = (dateStr) => {
  if (!dateStr) return "";
  const date = new Date(dateStr);
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}/${month}/${day}`;
};

const isSearchMatch = (employeeName) => {
  if (!filters?.value?.employeeName) return false;
  return employeeName
    ?.toLowerCase()
    .includes(filters.value.employeeName.toLowerCase());
};

const getCurrentMonthRange = () => {
  const now = new Date();
  const year = now.getFullYear();
  const month = now.getMonth();
  const firstDay = new Date(year, month, 1);
  let lastDay = new Date(year, month, 30);
  if (lastDay.getMonth() !== month) {
    lastDay = new Date(year, month + 1, 0);
  }
  const formatDate = (date) => {
    const y = date.getFullYear();
    const m = String(date.getMonth() + 1).padStart(2, "0");
    const d = String(date.getDate()).padStart(2, "0");
    return `${y}-${m}-${d}`;
  };
  return [formatDate(firstDay), formatDate(lastDay)];
};

const initStackedDateRange = () => {
  const [startDate, endDate] = getCurrentMonthRange();
  stackedDateRange.value = [startDate, endDate];
};

const setDateRangeByMode = () => {
  if (displayMode.value === "site") {
    filters.value.dateRange = [siteSummaryDate.value, siteSummaryDate.value];
  } else {
    if (stackedDateRange.value.length === 2) {
      filters.value.dateRange = [...stackedDateRange.value];
    } else {
      initStackedDateRange();
      filters.value.dateRange = [...stackedDateRange.value];
    }
  }
};

const disabledFutureDate = (time) => {
  return time.getTime() > Date.now();
};

// ==================== 响应式变量 ====================
const loading = ref(false);
const sites = ref([]);
const allEmployees = ref([]);
const activeShift = ref("day");
const summaryData = ref([]);
const siteColumns = ref([]);
const selectedDate = ref(getTodayDate());

const uploadVisible = ref(false);
const uploading = ref(false);
const uploadPreview = ref(null);

// 删除指定日期数据对话框
const deleteDateDialogVisible = ref(false);
const deletingDateData = ref(false);
const deleteDateForm = ref({
  date: "",
});

const uploadForm = ref({
  mode: "BX",
  siteId: null,
  shift: "day",
  statDate: getYesterdayDate(),
  file: null,
  accountCol: null,
  startCol: null,
  endCol: null,
});
const uploadRef = ref(null);
const columnOptions = ref([]);
const columnPreviewLoading = ref(false);
const columnMappingHint = ref("");

const formatColumnLabel = (colNum, letter, header) => {
  const label = `${letter}列(${colNum})`;
  if (!header) return `${label} - (空)`;
  const text = header.length > 24 ? `${header.slice(0, 24)}...` : header;
  return `${label} - ${text}`;
};

const applySuggestedColumns = (suggested) => {
  if (!suggested) return;
  uploadForm.value.accountCol = suggested.accountCol ?? null;
  uploadForm.value.startCol = suggested.startCol ?? null;
  uploadForm.value.endCol = suggested.endCol ?? null;
  const sourceMap = {
    manual: "手动指定",
    "bx-default": "BX默认列",
    "nn-default": "NN默认列",
    "header-all": "表头自动识别",
    "header-account": "表头自动识别(账号)",
    "header-partial": "表头部分识别",
  };
  const sourceText = sourceMap[suggested.source] || "自动推荐";
  columnMappingHint.value = `已按「${sourceText}」预选列，如有变化请手动调整。`;
};

const loadColumnPreview = async (file) => {
  if (!file) return;
  columnPreviewLoading.value = true;
  columnOptions.value = [];
  columnMappingHint.value = "";
  try {
    const formData = new FormData();
    formData.append("file", file);
    formData.append("mode", uploadForm.value.mode);
    const response = await adminApi.previewSiteStatsUpload(formData);
    const data = response.data || {};
    const columns = data.columns || [];
    columnOptions.value = columns.map((col) => ({
      index: col.index,
      letter: col.letter,
      header: col.header,
      label: col.label || formatColumnLabel(col.index, col.letter, col.header),
    }));
    applySuggestedColumns(data.suggested);
    if (columnOptions.value.length === 0) {
      ElMessage.warning("未能读取文件表头，请检查文件格式");
    }
  } catch (error) {
    console.error("读取列预览失败:", error);
    ElMessage.error(error.message || "读取文件表头失败");
  } finally {
    columnPreviewLoading.value = false;
  }
};

// 当前悬停卡片的统计数据
const hoverCardStats = ref({
  visible: false,
  date: "",
  employeeCount: 0,
  totalValue: 0,
});

// 站点汇总模式分页
const currentPage = ref(1);
const pageSize = ref(50);
const totalItems = ref(0);

// 筛选条件
const filters = ref({
  employeeName: "",
  dateRange: [],
  siteId: "",
  employeeId: "",
});

// 清除数据对话框
const clearDialogVisible = ref(false);
const clearingData = ref(false);
const clearForm = ref({
  siteId: null,
  shift: "day",
  date: "",
});

// 日期堆叠模式
const displayMode = ref("stacked");
const stackedLoading = ref(false);
const stackedDisplayData = ref([]);

// 日期分页
const dateCurrentPage = ref(1);
const datePageSize = ref(10);

// 站点汇总模式的日期
const siteSummaryDate = ref(getTodayDate());

// 堆叠模式的日期范围
const stackedDateRange = ref([]);

// 排序相关
const sortField = ref("employee_name");
const sortOrder = ref("ascending");

// ==================== 计算属性 ====================
const displayData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  const end = start + pageSize.value;
  return sortedSummaryData.value.slice(start, end);
});

const selectedSiteName = computed(() => {
  if (!filters.value.siteId) return "";
  const site = sites.value.find((s) => s.id === filters.value.siteId);
  return site ? `${site.code} - ${site.name}` : "";
});

const filteredEmployees = computed(() => {
  if (!filters.value.siteId) return [];
  return allEmployees.value.filter(
    (emp) => emp.site_id === filters.value.siteId,
  );
});

const summaryStats = computed(() => {
  if (displayMode.value === "site") {
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
      totalRecords,
      totalValue,
      avgTimeStr: formatSecondsToTime(avgTime),
      activeEmployees: summaryData.value.length,
    };
  } else {
    let totalRecords = 0;
    let totalValue = 0;
    let totalWeightedTime = 0;
    let totalWeight = 0;
    const uniqueEmployees = new Set();

    // ✅ 添加平均时间计算
    stackedDisplayData.value.forEach((row) => {
      if (row.employee_name) uniqueEmployees.add(row.employee_name);
      totalRecords += row.total_value || 0;
      totalValue += row.total_value || 0;

      // 计算加权平均时间
      if (row.total_avg_seconds && row.total_value) {
        totalWeightedTime += row.total_avg_seconds * row.total_value;
        totalWeight += row.total_value;
      }
    });

    const avgTime =
      totalWeight > 0 ? Math.round(totalWeightedTime / totalWeight) : 0;

    return {
      totalRecords,
      totalValue,
      avgTimeStr: formatSecondsToTime(avgTime), // ✅ 修改这里
      activeEmployees: uniqueEmployees.size,
    };
  }
});

// 按日期分组数据
const groupedByDate = computed(() => {
  const groups = {};
  for (const item of stackedDisplayData.value) {
    const date = item.date;
    if (!groups[date]) {
      groups[date] = { date, employees: [], totalValue: 0 };
    }
    groups[date].employees.push(item);
    groups[date].totalValue += item.total_value || 0;
  }
  return groups;
});

// 转换为每日卡片数据
// 转换为每日卡片数据
const dailyCardData = computed(() => {
  const result = [];
  const sortedDates = Object.keys(groupedByDate.value).sort(
    (a, b) => new Date(b) - new Date(a),
  );
  for (const date of sortedDates) {
    const dateGroup = groupedByDate.value[date];
    result.push({
      date: formatDateDisplay(date),
      employees: dateGroup.employees,
      employeeCount: dateGroup.employees.length,
      totalValue: dateGroup.totalValue,
    });
  }
  return result;
});

const datePagesTotal = computed(() => dailyCardData.value.length);

const paginatedDailyCards = computed(() => {
  const start = (dateCurrentPage.value - 1) * datePageSize.value;
  const end = start + datePageSize.value;
  return dailyCardData.value.slice(start, end);
});

const sortedSummaryData = computed(() => {
  let data = [...summaryData.value];
  if (!sortField.value) return data;
  data.sort((a, b) => {
    let valA, valB;
    if (sortField.value.startsWith("sites.")) {
      const siteCode = sortField.value.split(".")[1];
      valA = a.sites?.[siteCode]?.value || 0;
      valB = b.sites?.[siteCode]?.value || 0;
    } else if (sortField.value === "total_value") {
      valA = a.total_value || 0;
      valB = b.total_value || 0;
    } else if (sortField.value === "stat_date") {
      valA = selectedDate.value;
      valB = selectedDate.value;
      return sortOrder.value === "ascending" ? 0 : 0;
    } else {
      valA = a[sortField.value] || "";
      valB = b[sortField.value] || "";
    }
    if (typeof valA === "number" && typeof valB === "number") {
      return sortOrder.value === "ascending" ? valA - valB : valB - valA;
    } else {
      const strA = String(valA);
      const strB = String(valB);
      return sortOrder.value === "ascending"
        ? strA.localeCompare(strB, "zh-CN")
        : strB.localeCompare(strA, "zh-CN");
    }
  });
  return data;
});

const maxTotalValue = computed(() => {
  const values = summaryData.value.map((item) => item.total_value || 0);
  return values.length > 0 ? Math.max(...values) : 0;
});

const minTotalValue = computed(() => {
  const values = summaryData.value.map((item) => item.total_value || 0);
  return values.length > 0 ? Math.min(...values) : 0;
});

// ==================== 数据加载函数 ====================
const loadSites = async () => {
  try {
    const response = await adminApi.getSites({ is_active: true });
    sites.value = response.data?.items || [];
  } catch (error) {
    console.error("加载站点失败:", error);
    ElMessage.error("加载站点失败");
  }
};

const loadEmployees = async () => {
  try {
    const response = await adminApi.getEmployeeAccounts({ limit: 1000 });
    allEmployees.value = response.data?.items || [];
  } catch (error) {
    console.error("加载员工失败:", error);
    ElMessage.error("加载员工失败");
  }
};

const loadData = async () => {
  loading.value = true;
  try {
    const params = {};
    if (filters.value.employeeName)
      params.employee_name = filters.value.employeeName;
    if (filters.value.siteId)
      // ✅ 添加站点筛选参数
      params.site_id = filters.value.siteId;
    if (filters.value.dateRange && filters.value.dateRange.length === 2) {
      params.start_date = filters.value.dateRange[0];
      params.end_date = filters.value.dateRange[1];
    }
    params.shift = activeShift.value;

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
    totalItems.value = summaryData.value.length;
    currentPage.value = 1;
  } catch (error) {
    console.error("加载数据失败:", error);
    ElMessage.error("加载数据失败");
  } finally {
    loading.value = false;
  }
};

const loadStackedData = async () => {
  stackedLoading.value = true;
  try {
    const params = {};
    if (filters.value.employeeName)
      params.employee_name = filters.value.employeeName;
    if (filters.value.siteId) params.site_id = filters.value.siteId;
    if (filters.value.employeeId)
      params.employee_account_id = filters.value.employeeId;
    if (filters.value.dateRange && filters.value.dateRange.length === 2) {
      params.start_date = `${filters.value.dateRange[0]} 00:00:00`;
      params.end_date = `${filters.value.dateRange[1]} 23:59:59`;
    }
    params.shift = activeShift.value;

    const response = await adminApi.getSiteStatsStacked(params);
    const data = response.data || {};
    const rawItems = data.items || [];

    // ✅ 修复1：从 rawItems 中收集站点列（因为后端返回的是驼峰 siteCode，且没有返回 site_columns）
    const siteCodeSet = new Set();
    for (const item of rawItems) {
      if (item.siteCode) {
        siteCodeSet.add(item.siteCode);
      }
    }
    siteColumns.value = Array.from(siteCodeSet).sort();

    console.log("=== 后端返回的原始数据 ===");
    console.log("siteColumns 从数据中提取:", siteColumns.value);
    console.log("rawItems 数量:", rawItems.length);
    if (rawItems.length > 0) {
      console.log("第一条数据完整结构:", JSON.stringify(rawItems[0], null, 2));
      console.log("第一条数据的所有字段:", Object.keys(rawItems[0]));
    }

    if (rawItems.length === 0) {
      stackedDisplayData.value = [];
      return;
    }

    // ✅ 修改：按日期+员工+班次 分组聚合（防止 day 和 night 数据合并）
    const map = new Map();
    for (const item of rawItems) {
      // ✅ 修复2：后端返回的是驼峰命名，直接使用
      const employeeName = item.employeeName;
      const shiftValue = item.shift || "";
      if (!employeeName) continue;

      // ✅ 修复3：日期处理 - 去掉时间部分，只保留 YYYY-MM-DD
      const dateStr = item.date ? item.date.split("T")[0] : "";

      // 分组键包含班次
      const key = `${dateStr}|${employeeName}|${shiftValue}`;

      if (!map.has(key)) {
        map.set(key, {
          date: dateStr,
          employee_name: employeeName,
          account_name: item.accountName,
          shift: shiftValue,
          sites: {},
          total_value: 0,
          total_weighted_time: 0,
          total_weight: 0,
        });
      }

      const record = map.get(key);
      // ✅ 修复4：站点代码是 siteCode（驼峰）
      const siteCode = item.siteCode;

      // 获取平均时间（秒数）
      let avgSeconds = 0;
      if (item.avgTimeSeconds !== undefined && item.avgTimeSeconds !== null) {
        avgSeconds = item.avgTimeSeconds;
      } else if (item.avg_time_str) {
        avgSeconds = parseTimeToSeconds(item.avg_time_str);
      }

      const itemValue = item.value || 0;

      // 累加同一个站点的数据
      if (!record.sites[siteCode]) {
        record.sites[siteCode] = {
          value: 0,
          total_seconds: 0,
          avg_time_seconds: 0,
          avg_time_str: "-",
        };
      }

      const site = record.sites[siteCode];

      // 累加笔数和总时间
      site.value += itemValue;
      site.total_seconds += itemValue * avgSeconds;

      // 重新计算平均时间
      if (site.value > 0) {
        site.avg_time_seconds = Math.round(site.total_seconds / site.value);
        site.avg_time_str = formatSecondsToTime(site.avg_time_seconds);
      }

      // 累加总计（用于计算员工整体的平均时间）
      record.total_value += itemValue;
      record.total_weighted_time += itemValue * avgSeconds;
      record.total_weight += itemValue;
    }

    // 计算每条记录（每个员工每天每个班次）的总平均时间
    const dataArray = Array.from(map.values()).map((item) => {
      if (item.total_weight > 0) {
        const avgSeconds = Math.round(
          item.total_weighted_time / item.total_weight,
        );
        item.total_avg_time = formatSecondsToTime(avgSeconds);
        item.total_avg_seconds = avgSeconds;
      } else {
        item.total_avg_time = "-";
        item.total_avg_seconds = 0;
      }
      // 删除不再需要的临时字段
      delete item.total_weighted_time;
      delete item.total_weight;
      return item;
    });

    // 排序：按日期倒序，同日期按姓名正序
    dataArray.sort((a, b) => {
      if (a.date !== b.date) return new Date(b.date) - new Date(a.date);
      return (a.employee_name || "").localeCompare(
        b.employee_name || "",
        "zh-CN",
      );
    });

    console.log("处理后数据示例:", dataArray[0]);
    console.log("处理后数据总数:", dataArray.length);

    stackedDisplayData.value = dataArray;
    dateCurrentPage.value = 1;
  } catch (error) {
    console.error("加载堆叠数据失败:", error);
    ElMessage.error(
      "加载数据失败: " + (error.response?.data?.detail || error.message),
    );
  } finally {
    stackedLoading.value = false;
  }
};

// 站点筛选变化处理
const handleSiteFilterChange = (siteId) => {
  if (displayMode.value === "site") {
    loadData(); // 重新加载数据
  } else if (displayMode.value === "stacked") {
    loadStackedData(); // 堆叠模式也支持站点筛选
  }
};

// 将时间字符串（如 "2分43秒"）转换为秒数
const parseTimeToSeconds = (timeStr) => {
  if (!timeStr || timeStr === "-") return 0;
  let seconds = 0;
  const minutesMatch = timeStr.match(/(\d+)分/);
  const secondsMatch = timeStr.match(/(\d+)秒/);
  if (minutesMatch) seconds += parseInt(minutesMatch[1]) * 60;
  if (secondsMatch) seconds += parseInt(secondsMatch[1]);
  return seconds;
};

// ==================== 事件处理函数 ====================
const handleSearch = () => {
  if (displayMode.value === "site") {
    loadData();
  } else {
    loadStackedData();
  }
};

const handleDateRangeChange = (value) => {
  if (displayMode.value === "site") {
    if (value && value.length === 2) siteSummaryDate.value = value[0];
  } else {
    if (value && value.length === 2) stackedDateRange.value = [...value];
    else stackedDateRange.value = [];
  }
  if (displayMode.value === "site") loadData();
  else loadStackedData();
};

const handlePageSizeChange = (size) => {
  pageSize.value = size;
  currentPage.value = 1;
};

const handleCurrentPageChange = (page) => {
  currentPage.value = page;
};

const handleDatePageSizeChange = (size) => {
  datePageSize.value = size;
  dateCurrentPage.value = 1;
};

const handleDatePageChange = (page) => {
  dateCurrentPage.value = page;
};

const handleSortChange = ({ prop, order }) => {
  if (prop) {
    sortField.value = prop;
    sortOrder.value = order === "ascending" ? "ascending" : "descending";
  }
};

const switchMode = (mode) => {
  displayMode.value = mode;
  setDateRangeByMode();
  if (mode === "site") loadData();
  else loadStackedData();
};

const resetFilters = () => {
  filters.value = {
    siteId: "",
    employeeName: "",
    dateRange: [],
    employeeId: "",
  };
  if (displayMode.value === "site") {
    siteSummaryDate.value = getTodayDate();
    filters.value.dateRange = [siteSummaryDate.value, siteSummaryDate.value];
    loadData();
  } else {
    initStackedDateRange();
    filters.value.dateRange = [...stackedDateRange.value];
    loadStackedData();
    dateCurrentPage.value = 1;
  }
};

const showClearDialog = () => {
  clearForm.value = {
    siteId: filters.value.siteId || null,
    shift: activeShift.value,
    date: selectedDate.value || getTodayDate(),
  };
  clearDialogVisible.value = true;
};

const resetClearForm = () => {
  clearForm.value = { siteId: null, shift: "day", date: "" };
};

const confirmClearData = async () => {
  if (!clearForm.value.siteId) {
    ElMessage.warning("请选择站点");
    return;
  }
  if (!clearForm.value.date) {
    ElMessage.warning("请选择日期");
    return;
  }
  clearingData.value = true;
  try {
    await adminApi.clearSiteStatsByDate(
      clearForm.value.siteId,
      clearForm.value.shift,
      clearForm.value.date,
    );
    ElMessage.success("数据清除成功");
    clearDialogVisible.value = false;
    if (displayMode.value === "site") await loadData();
    else await loadStackedData();
  } catch (error) {
    console.error("清除数据失败:", error);
    ElMessage.error(error.message || "清除数据失败");
  } finally {
    clearingData.value = false;
  }
};

// 显示删除指定日期数据对话框
const showDeleteDateDialog = () => {
  deleteDateForm.value = {
    date: getTodayDate(),
  };
  deleteDateDialogVisible.value = true;
};

// 重置删除日期表单
const resetDeleteDateForm = () => {
  deleteDateForm.value = {
    date: "",
  };
};

// 确认删除指定日期数据
const confirmDeleteDateData = async () => {
  if (!deleteDateForm.value.date) {
    ElMessage.warning("请选择日期");
    return;
  }

  deletingDateData.value = true;
  try {
    // ✅ 调用新接口
    await adminApi.clearSiteStatsByDateOnly(deleteDateForm.value.date);
    ElMessage.success(`已删除 ${deleteDateForm.value.date} 的所有数据`);
    deleteDateDialogVisible.value = false;

    // 刷新数据
    if (displayMode.value === "site") {
      await loadData();
    } else {
      await loadStackedData();
    }
  } catch (error) {
    console.error("删除数据失败:", error);
    ElMessage.error(error.message || "删除数据失败");
  } finally {
    deletingDateData.value = false;
  }
};

const clearAllData = async () => {
  try {
    await ElMessageBox.confirm(
      "确定要清空所有站点数据吗？此操作不可恢复！\n\n注意：只会清空上传的数据记录，不会删除站点和员工配置。",
      "警告",
      {
        confirmButtonText: "确定清空",
        cancelButtonText: "取消",
        type: "warning",
      },
    );
    const loadingInstance = ElLoading.service({
      fullscreen: true,
      text: "正在清空数据...",
      background: "rgba(0, 0, 0, 0.7)",
    });
    try {
      const response = (await adminApi.clearAllSiteStats?.()) || {
        message: "数据已清空",
      };
      ElMessage.success(response.message || "数据已清空");
      if (displayMode.value === "site") await loadData();
      else await loadStackedData();
    } finally {
      loadingInstance.close();
    }
  } catch (error) {
    if (error !== "cancel") {
      console.error("清空数据失败:", error);
      ElMessage.error(error.response?.data?.detail || "清空数据失败");
    }
  }
};

// 鼠标进入卡片
const onCardMouseEnter = (dateGroup) => {
  hoverCardStats.value = {
    visible: true,
    date: dateGroup.date,
    employeeCount: dateGroup.employeeCount,
    totalValue: dateGroup.totalValue,
  };
};

// 鼠标离开卡片
const onCardMouseLeave = () => {
  hoverCardStats.value = {
    ...hoverCardStats.value,
    visible: false,
  };
};

const showUploadDialog = () => {
  uploadForm.value = {
    mode: "BX",
    siteId: filters.value.siteId || null,
    shift: activeShift.value,
    statDate: getYesterdayDate(),
    file: null,
    accountCol: null,
    startCol: null,
    endCol: null,
  };
  columnOptions.value = [];
  columnMappingHint.value = "";
  uploadPreview.value = null;
  if (uploadRef.value) uploadRef.value.clearFiles();
  uploadVisible.value = true;
};

const handleUploadModeChange = async () => {
  if (uploadForm.value.file) {
    await loadColumnPreview(uploadForm.value.file);
  }
};

const handleSiteChange = (siteId) => {
  if (uploadPreview.value) {
    const site = sites.value.find((s) => s.id === siteId);
    if (site) {
      uploadPreview.value.site_code = site.code;
      uploadPreview.value.site_name = site.name;
    }
  }
};

const handleFileChange = async (file) => {
  uploadForm.value.file = file.raw;
  const site = sites.value.find((s) => s.id === uploadForm.value.siteId);
  uploadPreview.value = {
    filename: file.name,
    site_id: uploadForm.value.siteId,
    site_code: site?.code,
    site_name: site?.name,
    shift: uploadForm.value.shift,
    statDate: uploadForm.value.statDate,
    total_records: 0,
    account_count: 0,
    details: [],
    unmatched_accounts: [],
  };
  await loadColumnPreview(file.raw);
};

const uploadData = async () => {
  if (!uploadForm.value.siteId) {
    ElMessage.warning("请选择站点");
    return;
  }
  if (!uploadForm.value.file) {
    ElMessage.warning("请选择文件");
    return;
  }
  if (!uploadForm.value.statDate) {
    ElMessage.warning("请选择统计日期");
    return;
  }
  if (!uploadForm.value.accountCol || !uploadForm.value.startCol || !uploadForm.value.endCol) {
    ElMessage.warning("请选择开始时间、完成时间和操作人对应的列");
    return;
  }
  if (
    uploadForm.value.accountCol === uploadForm.value.startCol ||
    uploadForm.value.accountCol === uploadForm.value.endCol ||
    uploadForm.value.startCol === uploadForm.value.endCol
  ) {
    ElMessage.warning("开始时间、完成时间、操作人必须选择不同的列");
    return;
  }

  uploading.value = true;
  try {
    const formData = new FormData();
    formData.append("file", uploadForm.value.file);
    formData.append("site_id", uploadForm.value.siteId);
    formData.append("shift", uploadForm.value.shift);
    formData.append("date", uploadForm.value.statDate);
    formData.append("mode", uploadForm.value.mode);
    formData.append("account_col", String(uploadForm.value.accountCol));
    formData.append("start_col", String(uploadForm.value.startCol));
    formData.append("end_col", String(uploadForm.value.endCol));

    const response = await adminApi.uploadSiteStats(formData);

    // ✅ response.data 直接就是 { matchedCount, unmatchedCount, stats }
    const data = response.data;

    // ✅ 使用驼峰命名（与后端返回一致）
    const matchedCount = data.matchedCount || 0;
    const unmatchedCount = data.unmatchedCount || 0;
    const totalRecords = data.stats?.total_records || 0;

    console.log("上传统计:", { matchedCount, unmatchedCount, totalRecords });

    let message = `上传成功！共处理 ${totalRecords} 条记录`;
    if (matchedCount > 0) {
      message += `，匹配 ${matchedCount} 个账号`;
    }
    if (unmatchedCount > 0) {
      message += `，未匹配 ${unmatchedCount} 个账号`;
    }
    ElMessage.success(message);

    if (unmatchedCount > 0) {
      const samples = (data.unmatchedAccounts || data.unmatched_accounts || [])
        .slice(0, 5)
        .join("、");
      const mapping = data.columnMapping || data.column_mapping;
      let warn = `发现 ${unmatchedCount} 个未匹配账号`;
      if (samples) warn += `（如: ${samples}）`;
      if (mapping?.accountCol) {
        warn += `；本次从第 ${mapping.accountCol} 列读取账号`;
      }
      ElMessage.warning(warn);
    }

    uploadVisible.value = false;

    // 刷新数据
    await loadData();
    await loadEmployees();
    if (displayMode.value === "stacked") {
      await loadStackedData();
    }
  } catch (error) {
    console.error("上传失败:", error);
    const errorMsg =
      error.response?.data?.detail ||
      error.response?.data?.message ||
      error.message ||
      "上传失败";
    ElMessage.error(errorMsg);
  } finally {
    uploading.value = false;
  }
};

const exportData = () => {
  if (summaryData.value.length === 0) {
    ElMessage.warning("暂无数据可导出");
    return;
  }
  const exportRows = summaryData.value.map((row) => {
    const rowData = {
      统计日期: selectedDate.value,
      员工姓名: row.employee_name,
    };
    siteColumns.value.forEach((site) => {
      const siteData = row.sites[site];
      rowData[`${site}_笔数`] = siteData?.value || "-";
      rowData[`${site}_平均时间`] = siteData?.avg_time_str || "-";
    });
    rowData["总计_笔数"] = row.total_value;
    rowData["总计_平均时间"] = row.total_avg_time;
    return rowData;
  });
  const headers = Object.keys(exportRows[0]);
  const csvRows = [
    headers.join(","),
    ...exportRows.map((row) =>
      headers
        .map((h) => `"${String(row[h] || "").replace(/"/g, '""')}"`)
        .join(","),
    ),
  ];
  const blob = new Blob(["\uFEFF" + csvRows.join("\n")], {
    type: "text/csv;charset=utf-8;",
  });
  const link = document.createElement("a");
  link.href = URL.createObjectURL(blob);
  const siteInfo = selectedSiteName.value || "全部站点";
  const shiftText = activeShift.value === "day" ? "A班" : "B班";
  link.setAttribute(
    "download",
    `${siteInfo}_${shiftText}_${selectedDate.value}_出款统计.csv`,
  );
  link.click();
  URL.revokeObjectURL(link.href);
};

const exportStackedData = () => {
  if (stackedDisplayData.value.length === 0) {
    ElMessage.warning("暂无数据可导出");
    return;
  }

  const exportRows = [];
  let globalIndex = 1;

  for (const dateGroup of paginatedDailyCards.value) {
    // 添加日期汇总行
    exportRows.push({
      序号: "",
      统计日期: `${dateGroup.date}\n${dateGroup.employeeCount}人出款\n当日总:${dateGroup.totalValue}笔`,
      姓名: "",
      ...Object.fromEntries(siteColumns.value.map((site) => [site, ""])),
      总计: "",
    });

    // 添加员工数据行
    for (const employee of dateGroup.employees) {
      const siteData = {};
      siteColumns.value.forEach((site) => {
        siteData[site] = employee.sites?.[site]?.value || "-";
      });
      exportRows.push({
        序号: globalIndex++,
        统计日期: "",
        姓名: employee.employee_name,
        ...siteData,
        总计: `${employee.total_value || 0}笔${employee.total_avg_time ? `/${employee.total_avg_time}` : ""}`,
      });
    }

    // 每个卡片后加空行分隔
    exportRows.push({});
  }

  const headers = ["序号", "统计日期", "姓名", ...siteColumns.value, "总计"];
  const csvRows = [
    headers.join(","),
    ...exportRows.map((row) =>
      headers
        .map((h) => {
          const val = row[h] || "";
          // 处理换行符，CSV 中需要用引号包裹
          const cleanedVal = String(val).replace(/"/g, '""');
          return `"${cleanedVal}"`;
        })
        .join(","),
    ),
  ];

  const blob = new Blob(["\uFEFF" + csvRows.join("\n")], {
    type: "text/csv;charset=utf-8;",
  });
  const link = document.createElement("a");
  link.href = URL.createObjectURL(blob);
  const shiftText = activeShift.value === "day" ? "A班" : "B班";
  link.setAttribute(
    "download",
    `${shiftText}_出款统计_堆叠视图_${new Date().toISOString().slice(0, 10)}.csv`,
  );
  link.click();
  URL.revokeObjectURL(link.href);
};

// ==================== 监听器与生命周期 ====================
watch(
  [
    () => filters.value.siteId,
    () => filters.value.employeeId,
    () => filters.value.dateRange,
    activeShift,
  ],
  () => {
    if (displayMode.value === "site") loadData();
    else if (displayMode.value === "stacked") loadStackedData();
  },
  { deep: true },
);

onMounted(() => {
  loadSites();
  loadEmployees();
  initStackedDateRange();
  displayMode.value = "stacked";
  filters.value.dateRange = [...stackedDateRange.value];
  loadStackedData();
});
</script>

<style scoped>
/* ==================== 背景装饰样式 ==================== */
.site-stats {
  padding: 0px;
  min-height: 100vh;
  background: #f0f2f6;
  position: relative;
}

/* 背景装饰容器 */
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

/* 彩色气泡 */
.blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.4;
  animation: float 20s ease-in-out infinite;
}

/* 气泡1 - 紫蓝色 */
.blob-1 {
  width: 500px;
  height: 500px;
  background: linear-gradient(135deg, #667eea, #764ba2);
  top: -200px;
  right: -100px;
  animation-delay: 0s;
}

/* 气泡2 - 粉红色 */
.blob-2 {
  width: 600px;
  height: 600px;
  background: linear-gradient(135deg, #f093fb, #f5576c);
  bottom: -250px;
  left: -150px;
  animation-delay: -5s;
}

/* 气泡3 - 青蓝色 */
.blob-3 {
  width: 400px;
  height: 400px;
  background: linear-gradient(135deg, #4facfe, #00f2fe);
  top: 40%;
  left: 30%;
  animation-delay: -10s;
}

/* 气泡浮动动画 */
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
.filter-bar,
.stats-row,
.summary-card {
  position: relative;
  z-index: 1;
}

/* 卡片添加半透明背景效果（可选） */
.filter-bar :deep(.el-card__body),
.summary-card :deep(.el-card__body) {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(0px);
}

/* 筛选栏样式 */
.filter-bar {
  margin-bottom: 20px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.3);
}

.filter-bar :deep(.el-card__body) {
  padding: 16px 20px;
}

/* 统计卡片样式 */
.stats-row {
  margin-bottom: 20px;
}

.stat-card {
  cursor: default;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.3);
  transition: all 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.stat-card :deep(.el-card__body) {
  padding: 16px;
}

.stat-content {
  display: flex;
  align-items: center;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  margin-right: 16px;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  line-height: 1.2;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 4px;
}

.summary-card {
  min-height: 400px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.3);
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 8px;
}

.header-tip {
  font-size: 12px;
  color: #909399;
}

/* 站点汇总模式样式 */
.site-cell {
  padding: 4px 0;
}

.site-cell .value {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.site-cell .time {
  font-size: 14px;
  color: #909399;
  margin-top: 2px;
}

.site-cell.empty .value {
  color: #c0c4cc;
}

.total-cell {
  padding: 4px 0;
}

.total-cell .value {
  font-size: 14px;
  font-weight: 600;
  color: #409eff;
}

.total-cell .time {
  font-size: 14px;
  color: #909399;
  margin-top: 2px;
}

/* 堆叠模式样式 */
.global-table-header {
  margin-bottom: 16px;
  border-radius: 8px;
  overflow: hidden;
}

.daily-cards-container {
  max-height: calc(100vh - 340px);
  overflow-y: auto;
  padding: 4px;
}

.daily-card {
  margin-bottom: 20px;
  border-radius: 12px;
  background: #fff;
  border: 1px solid #e4e7ed;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  transition: all 0.3s ease;
}

.daily-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.grid-header {
  display: grid;
  background: linear-gradient(180deg, #a6cdf7 0%, #62adf0 50%, #0971e0 100%);
  color: #1a2a3a;
  font-weight: bold;
  font-size: 14px;
  border-top: 1px solid rgba(255, 255, 255, 0.6);
  border-bottom: 2px solid #7cb3e0;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.4),
    0 2px 6px rgba(0, 0, 0, 0.08);
  text-shadow: 0 1px 0 rgba(255, 255, 255, 0.3);
}

.grid-header > div {
  padding: 12px 8px;
  text-align: center;
  border-right: 1px solid rgba(255, 255, 255, 0.3);
}

.grid-header > div:last-child {
  border-right: none;
}

.grid-row {
  display: grid;
  border-bottom: 1px solid #e4e7ed;
  transition: background 0.2s ease;
}

.grid-row > div {
  padding: 8px;
  text-align: center;
  border-right: 1px solid #e4e7ed;
  vertical-align: middle;
}

.grid-row > div:last-child {
  border-right: none;
}

.row-striped {
  background-color: #fafafa;
}

.grid-row:hover {
  background-color: #e6f7ff !important;
}

/* 日期单元格 */
.date-cell-content {
  text-align: center;
  line-height: 1.6;
  padding: 4px 0;
}

.date-stats {
  font-size: 12px;
  color: #409eff;
}

/* 站点数据 */
.site-value,
.total-value {
  font-size: 14px;
  font-weight: 500;
  color: #409eff;
}

.site-time,
.total-time {
  font-size: 11px;
  color: #909399;
  margin-top: 2px;
}

.total-value {
  color: #67c23a;
  font-weight: 600;
}

/* 搜索高亮 */
.search-highlight {
  background-color: #fff3cd;
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: bold;
  color: #856404;
  display: inline-block;
}

.empty-data {
  padding: 40px 0;
}

.pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  line-height: 1.2;
  transition: all 0.2s ease;
}

.stat-value.hover-value {
  color: #1890ff;
}

.stat-sub {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.stat-label span {
  display: inline-block;
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 姓名列颜色 */
.grid-row .col-name {
  color: #303133;
  font-weight: 500;
}

.grid-row:hover .col-name {
  color: #409eff;
}

/* 日期单元格简化样式 */
.date-cell-simple {
  font-weight: 500;
  color: #409eff;
  cursor: pointer;
}

.date-cell-simple:hover {
  text-decoration: underline;
}

/* 表单提示样式 */
.form-tip {
  margin-top: 6px;
  font-size: 12px;
  color: #909399;
  line-height: 1.5;
}

.column-mapping-row {
  width: 100%;
}

.column-mapping-label {
  margin-bottom: 6px;
  font-size: 13px;
  color: #606266;
  font-weight: 500;
}

/* 工具提示卡片样式 */
.card-tooltip {
  position: fixed;
  background: rgba(0, 0, 0, 0.85);
  color: white;
  padding: 8px 12px;
  border-radius: 8px;
  font-size: 12px;
  z-index: 9999;
  pointer-events: none;
  backdrop-filter: blur(8px);
  white-space: nowrap;
}

.tooltip-date {
  font-weight: bold;
  margin-bottom: 4px;
}

.tooltip-stats {
  font-size: 11px;
  opacity: 0.9;
}

/* 响应式调整 */
@media (max-width: 1200px) {
  .site-stats {
    padding: 16px;
  }

  .blob-1,
  .blob-2,
  .blob-3 {
    opacity: 0.2;
  }
}

@media (max-width: 768px) {
  .site-stats {
    padding: 12px;
  }

  .blob-1,
  .blob-2,
  .blob-3 {
    display: none;
  }
}
</style>

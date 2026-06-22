// frontend/api/admin_api.ts
import request from "./request";

const adminApi = {
  // ========== 考勤管理 ==========
  getEmployees: (params?: { skip?: number; limit?: number; search?: string }) =>
    request.get<any>("/admin/employees", { params }),

  createEmployee: (data: any) => request.post("/admin/employees", data),

  updateEmployee: (id: string, data: any) =>
    request.put(`/admin/employees/${id}`, data),

  deleteEmployee: (id: string) => request.delete(`/admin/employees/${id}`),

  getAttendanceRecords: (yearMonth: string, employeeIds: string[]) =>
    request.get("/admin/attendance/records", {
      params: { year_month: yearMonth, employee_ids: employeeIds },
    }),

  saveAttendanceRecords: (data: { year_month: string; data: any }) =>
    request.post("/admin/attendance/records", data),

  getRecordsByEmployees: (yearMonth: string, employeeIds: string[]) =>
    request.get("/admin/attendance/records", {
      params: { year_month: yearMonth, employee_ids: employeeIds },
    }),

  saveRecords: (data: { year_month: string; data: any }) =>
    request.post("/admin/attendance/records", data),

  getTodayAbsentees: () => request.get("/admin/attendance/today-absentees"),

  // ========== 绩效管理 ==========
  getPerformance: (params: { month: string; employee_id?: string }) =>
    request.get("/admin/performance", { params }),

  batchSavePerformance: (data: { month: string; items: any[] }) =>
    request.post("/admin/performance/batch", data),

  updatePerformance: (
    employeeId: string,
    month: string,
    data: {
      total_score?: number;
      grade?: string;
      score_records?: any[];
    },
  ) => request.put(`/admin/performance/${employeeId}/${month}`, data),

  // ✅ 新增：单条绩效删除
  deletePerformance: (employeeId: string, month: string) =>
    request.delete(`/admin/performance/${employeeId}/${month}`),

  // ========== 罚款管理 ==========
  getPenaltyRecords: (params: {
    month?: string;
    employee_id?: string;
    page?: number;
    page_size?: number;
  }) => request.get("/admin/penalty/records", { params }),

  createPenaltyRecord: (data: any) =>
    request.post("/admin/penalty/record", data),

  deletePenaltyRecord: (id: string) =>
    request.delete(`/admin/penalty/records/${id}`),

  updatePenaltyRecord: (
    id: string,
    data: {
      amount: number;
      category: string;
      reason: string;
      penalty_date: string;
    },
  ) => request.put(`/admin/penalty/records/${id}`, data),

  // ========== 出款站点管理 ==========
  getSites: (params?: { is_active?: boolean }) =>
    request.get("/admin/site-stats/sites", { params }),

  createSite: (data: any) => request.post("/admin/site-stats/sites", data),

  updateSite: (id: string, data: any) =>
    request.put(`/admin/site-stats/sites/${id}`, data),

  deleteSite: (id: string) => request.delete(`/admin/site-stats/sites/${id}`),

  // ========== 员工账号管理 ==========
  getEmployeeAccounts: (params?: {
    site_id?: string;
    account_name?: string;
    skip?: number;
    limit?: number;
    shift?: string;
  }) => request.get("/admin/site-stats/employee-accounts", { params }),

  createEmployeeAccount: (data: any) =>
    request.post("/admin/site-stats/employee-accounts", data),

  updateEmployeeAccount: (id: string, data: any) =>
    request.put(`/admin/site-stats/employee-accounts/${id}`, data),

  deleteEmployeeAccount: (id: string) =>
    request.delete(`/admin/site-stats/employee-accounts/${id}`),

  // ========== 出款统计 ==========
  previewSiteStatsUpload: (formData: FormData) =>
    request.post("/admin/site-stats/upload/preview", formData, {
      headers: { "Content-Type": "multipart/form-data" },
    }),

  uploadSiteStats: (formData: FormData) =>
    request.post("/admin/site-stats/upload", formData, {
      headers: { "Content-Type": "multipart/form-data" },
    }),

  getSiteStatsSummary: (params: {
    site_id?: string;
    employee_account_id?: string;
    start_date?: string;
    end_date?: string;
    shift?: string;
  }) => request.get("/admin/site-stats/summary", { params }),

  getSiteStatsStacked: (params: {
    site_id?: string;
    employee_account_id?: string;
    start_date?: string;
    end_date?: string;
    shift?: string;
  }) => request.get("/admin/site-stats/stacked-summary", { params }),
  clearSiteStatsByDate: (siteId: string, shift: string, date: string) =>
    request.delete("/admin/site-stats/data/clear-by-date", {
      params: { site_id: siteId, shift, date },
    }),
  clearAllSiteStats: () => request.delete("/admin/site-stats/data/clear"),
  // 删除指定日期的所有数据（不限站点、不限班次）
  clearSiteStatsByDateOnly: (date: string) =>
    request.delete("/admin/site-stats/clear-by-date-only", {
      params: { date },
    }),

  // ========== 操作日志管理 ==========
  getOperationLogs: (params: {
    page?: number;
    pageSize?: number;
    operatorName?: string;
    operationModule?: string;
    operationType?: string;
    startDate?: string;
    endDate?: string;
  }) => request.get("/operation-logs", { params }),

  getOperationModules: () => request.get<string[]>("/operation-logs/modules"),

  getOperationTypes: () => request.get<string[]>("/operation-logs/types"),

  exportOperationLogs: (params: {
    operatorName?: string;
    operationModule?: string;
    startDate?: string;
    endDate?: string;
  }) => request.get("/operation-logs/export", { params, responseType: "blob" }),

  deleteOperationLogs: (data: {
    ids?: string[];
    daysOld?: number;
    deleteAll?: boolean;
  }) => request.delete("/operation-logs", { data }),
};

export default adminApi;

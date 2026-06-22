// frontend/api/types.ts - 完整的类型定义（与后端 Go 模型对齐）

// ========== 认证相关 ==========
export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  refreshToken: string;
  expiresAt: string;
}

// ========== 设备相关（camelCase） ==========
export interface Device {
  id: string;
  /** 设备ID（可能与 id 相同或不同，用于兼容旧版本） */
  deviceId?: string;
  /** 设备指纹，用于唯一标识设备 */
  deviceFingerprint: string;
  /** 员工ID（系统内部） */
  employeeId: string;
  /** 员工姓名 */
  employeeName: string;
  /** 员工工号 */
  employeeNumber: string;
  /** 职位 */
  position?: string;
  /** 设备名称/主机名 */
  hostname: string;
  /** 操作系统版本 */
  osVersion: string;
  /** Agent 版本号 */
  agentVersion: string;
  /** 审核状态: pending(待审核), approved(已审核), rejected(已拒绝) */
  status: "pending" | "approved" | "rejected";
  /** 在线状态: online(在线), offline(离线) - 前端动态计算，后端不存储 */
  onlineStatus: "online" | "offline";
  /** 最后在线时间 (UTC) */
  lastSeenAt: string | null;
  /** 绑定时间 (UTC) */
  boundAt: string | null;
  /** 创建时间 (UTC) */
  createdAt: string;
  /** 更新时间 (UTC) */
  updatedAt?: string;
  /** 截图统计（前端扩展字段，从 screenshotApi.getEmployeeStats 获取） */
  screenshotStats?: {
    /** 今日截图数量 */
    todayCount: number;
    /** 总截图数量 */
    totalCount: number;
  };
}

export interface DeviceListResponse {
  items: Device[];
  total: number;
  page: number;
  pageSize: number;
}

export interface OnlineStatsResponse {
  total: number;
  online: number;
  offline: number;
  timestamp: string;
}

// ========== 活动日志相关 ==========
export interface ActivityLog {
  id: string;
  deviceId: string;
  processName: string;
  windowTitle?: string;
  browserTitle?: string;
  startedAt: string;
  endedAt?: string;
  durationSeconds: number;
  isIdle: boolean;
  createdAt: string;
  reported?: boolean; // ✅ 添加：是否已报告
  reportedAt?: string; // ✅ 添加：报告时间
  retryCount?: number; // ✅ 添加：重试次数
  errorMessage?: string; // ✅ 添加：错误信息
}

export interface ActivityStats {
  totalDuration: number;
  idleDuration: number;
  activeDuration: number;
  productivityRate: number;
  topApps: Array<{ name: string; duration: number; count: number }>;
  dailyActivity: Array<{ date: string; duration: number; idle: number }>;
  hourlyActivity: Record<number, number>;
}

export interface DailyActivitySummary {
  date: string;
  totalDuration: number;
  idleDuration: number;
  activeDuration: number;
  productivityRate: number;
  topApps: Array<{ name: string; duration: number }>;
}

// ========== 录制相关 ==========
export interface Recording {
  id: string;
  deviceId: string;
  fileName?: string;
  fileSize: number;
  duration: number; // ✅ 添加：时长（秒）
  startTime: string;
  endTime: string;
  uploaded: boolean; // ✅ 改为 uploaded (布尔值)
  uploadedAt?: string; // ✅ 可选，上传完成时间（如果有）
  status?: string; // ✅ 可选，前端计算使用
  format?: string;
  filePath?: string; // ✅ 添加：文件路径
  sha256?: string; // ✅ 添加：文件哈希
  createdAt?: string; // ✅ 添加：创建时间
}

export interface RecordingListResponse {
  items: Recording[];
  total: number;
  page: number;
  pageSize: number;
}

// ========== 策略相关 ==========
export interface PolicyContent {
  version: number;
  recording?: {
    enabled: boolean;
    fps: number;
    quality: number;
    enableGpu: boolean;
    sliceMinutes: number;
  };
  screenshot?: {
    enabled: boolean;
    intervalSeconds: number;
    quality: number;
    format: string;
  };
  activity?: {
    enabled: boolean;
    trackIdle: boolean;
    idleThresholdSeconds: number;
  };
  security?: {
    enableEncryption: boolean;
    enableVPN: boolean;
  };
  cleanup?: {
    enabled: boolean;
    recordingDays: number;
    screenshotDays: number;
    logDays: number;
    diskThresholdPercent: number;
  };
  rules?: Array<{
    name: string;
    condition: string;
    action: string;
  }>;
  config?: Record<string, any>;
}

export interface Policy {
  id: string;
  name: string;
  version: number;
  content: string;
  signature?: string;
  status: number;
  createdAt: string;
  updatedAt: string;
  isCurrent?: boolean;
}

export interface CleanupPolicy {
  autoCleanupEnabled: boolean; // ✅ 添加
  recordingRetentionDays: number; // ✅ 改为 recordingRetentionDays
  screenshotRetentionDays: number; // ✅ 改为 screenshotRetentionDays
  logRetentionDays: number; // ✅ 改为 logRetentionDays
  diskUsageThresholdPercent: number; // ✅ 改为 diskUsageThresholdPercent
  cleanupIntervalMinutes: number; // ✅ 添加
  lastRunAt: string | null;
  updatedAt: string;
}

// ========== 清理相关 ==========
export interface StorageStats {
  totalBytes: number;
  usedBytes: number; // ✅ 添加
  freeBytes: number; // ✅ 添加
  diskUsagePercent: number;
  recordingCount: number; // ✅ 添加
  screenshotCount: number; // ✅ 添加
  logCount: number;
  lastCleanupAt: string | null;
}

export interface CleanupLog {
  id: string;
  triggerType: "auto" | "manual" | "auto_disk_threshold" | "auto_disk_critical";
  status: "success" | "partial" | "failed";
  deletedRecordings: number;
  deletedScreenshots: number;
  deletedLogs: number;
  deletedBytes: number;
  message: string;
  startedAt: string;
  completedAt: string;
}

// 添加分页响应类型
export interface PageResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
}

/** 设备端任务（录屏按需上传等） */
export interface DeviceTask {
  id: string;
  deviceId: string;
  type: string;
  startTime?: string;
  endTime?: string;
  status: string;
  progress: number;
  result?: string;
  completedAt?: string;
  createdAt?: string;
}

// ========== 用户相关 ==========
export interface User {
  id: string;
  username: string;
  realName?: string;
  email?: string;
  status: number;
  roleIds?: string[];
  roles?: Role[];
  createdAt: string;
  updatedAt: string;
  lastLogin?: string | null;
}

export interface Role {
  id: string;
  name: string;
  description?: string;
  permissions: string[];
  createdAt: string;
  updatedAt: string;
}

// ========== 截图相关 ==========
export interface Screenshot {
  id: string;
  deviceId: string;
  fileName: string;
  fileSize: number;
  format: string;
  width: number; // ✅ 添加：图片宽度
  height: number; // ✅ 添加：图片高度
  capturedAt: string;
  uploadedAt: string;
  thumbnailUrl?: string;
  imageUrl?: string;
  filePath: string;
  uploaded: boolean;
  createdAt: string;
  encrypted?: boolean;
}

export interface ScreenshotListResponse {
  items: Screenshot[];
  total: number;
  page: number;
  pageSize: number;
}

// ========== 远程查看相关 ==========
export interface RemoteViewStartSession {
  sessionId: string;
  deviceId: string;
  deviceName: string;
  roomName: string;
  livekitUrl: string;
  viewerToken: string;
  expiresAt: string;
}

export interface RemoteViewStatus {
  isActive: boolean;
  sessionId?: string;
  status?: string;
  viewerCount: number;
  startedAt?: string;
  stats?: {
    fps: number;
    bitrateKbps: number;
    encoder: string;
  };
}

// ========== 响应类型 ==========
export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export interface ListResponse<T> {
  items: T[];
  total: number;
  page?: number;
  pageSize?: number;
}

// ========== WebSocket 消息类型 ==========
export interface WSMessage {
  type:
    | "online_devices"
    | "device_status"
    | "policy_update"
    | "device_info_update"
    | "task_notification"
    | "ping"
    | "pong";
  devices?: string[];
  deviceId?: string;
  status?: "online" | "offline";
  timestamp?: number;
  data?: any;
  [key: string]: any;
}

// 在文件末尾添加

// ========== 今日考勤异常人员 ==========
export interface TodayAbsenteeItem {
  employeeId: string;
  employeeName: string;
  position: string;
  status: string;
  statusLabel: string;
  date: string;
}

export interface TodayAbsenteesResponse {
  vacation: TodayAbsenteeItem[]; // 休假（全天+半天）
  leave: TodayAbsenteeItem[]; // 请假
  absent: TodayAbsenteeItem[]; // 旷工
  total: number;
}

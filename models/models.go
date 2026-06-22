// backend/models/models.go
// 完整替换文件内容

package models

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ========== 通用响应结构 ==========

// APIResponse API 统一响应结构
type APIResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp string      `json:"timestamp"`
	RequestID string      `json:"requestId"`
}

// ========== 用户模型 ==========

// User 用户模型
type User struct {
	ID           uuid.UUID  `json:"id"`
	Username     string     `json:"username"`
	PasswordHash string     `json:"-"`
	Email        string     `json:"email"`
	RealName     string     `json:"realName"`
	DepartmentID *uuid.UUID `json:"departmentId,omitempty"`
	Status       int        `json:"status"` // 1:启用, 0:禁用
	CreatedAt    time.Time  `json:"createdAt"`
	LastLogin    *time.Time `json:"lastLogin,omitempty"`
	Roles        []Role     `json:"roles,omitempty"`
}

// UserResponse 用户响应（隐藏敏感信息）
type UserResponse struct {
	ID        uuid.UUID  `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	RealName  string     `json:"realName"`
	Status    int        `json:"status"`
	Roles     []Role     `json:"roles,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	LastLogin *time.Time `json:"lastLogin,omitempty"`
}

// ========== 角色模型 ==========

// Role 角色模型
type Role struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// UserRole 用户角色关联
type UserRole struct {
	UserID    uuid.UUID `json:"userId"`
	RoleID    uuid.UUID `json:"roleId"`
	CreatedAt time.Time `json:"createdAt"`
}

// ========== 设备模型 ==========

// Device 设备模型
type Device struct {
	ID                uuid.UUID  `json:"id"`
	EmployeeID        string     `json:"employeeId"`   // 员工ID（系统内部）
	EmployeeName      string     `json:"employeeName"` // 员工姓名
	EmployeeNumber    string     `json:"employeeNumber"`
	Position          string     `json:"position"`
	DeviceFingerprint string     `json:"deviceFingerprint"`
	Hostname          string     `json:"hostname"`
	OSVersion         string     `json:"osVersion"`
	AgentVersion      string     `json:"agentVersion"`
	Status            string     `json:"status"` // online, offline, pending
	OnlineStatus      string     `json:"onlineStatus" gorm:"-"`
	LastSeenAt        *time.Time `json:"lastSeenAt"`
	BoundAt           *time.Time `json:"boundAt"`           // ✅ 添加：绑定时间
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         *time.Time `json:"updatedAt,omitempty"` // ✅ 添加：更新时间
	DeviceSecret      string     `json:"-"`
}

// ========== 策略模型 ==========

// Policy 策略模型
type Policy struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Version   int       `json:"version"`
	Content   string    `json:"content"`
	Signature string    `json:"signature"`
	Status    int       `json:"status"` // 1:已发布, 0:草稿
	IsCurrent bool      `json:"isCurrent"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// PolicyContent 策略内容
type PolicyContent struct {
	Version int                    `json:"version,omitempty"`
	Rules   []PolicyRule           `json:"rules,omitempty"`
	Config  map[string]interface{} `json:"config,omitempty"`
}

// PolicyRule 策略规则
type PolicyRule struct {
	ID     string `json:"id"`
	Action string `json:"action"`
	Match  string `json:"match"`
}

// ========== 录屏模型 ==========

// Recording 录屏记录模型
type Recording struct {
	ID        uuid.UUID `json:"id"`
	DeviceID  uuid.UUID `json:"deviceId"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	FilePath  string    `json:"filePath"`
	FileSize  int64     `json:"fileSize"`
	SHA256    string    `json:"sha256"`
	Duration  int       `json:"duration"` // ✅ 添加：录制时长（秒）
	Uploaded  bool      `json:"uploaded"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"` // ✅ 添加：更新时间
}

// ========== 截图模型 ==========

// Screenshot 截图模型
type Screenshot struct {
	ID         string     `json:"id"`
	DeviceID   string     `json:"deviceId"`
	FileName   string     `json:"fileName"`     // ✅ 添加：文件名
	FileSize   int64      `json:"fileSize"`
	Format     string     `json:"format"`       // ✅ 添加：格式（jpg, png等）
	Width      int        `json:"width"`        // ✅ 添加：宽度
	Height     int        `json:"height"`       // ✅ 添加：高度
	CapturedAt time.Time  `json:"capturedAt"`
	FilePath   string     `json:"filePath"`
	Uploaded   bool       `json:"uploaded"`
	UploadedAt *time.Time `json:"uploadedAt,omitempty"` // ✅ 添加：上传时间
	Encrypted  bool       `json:"encrypted,omitempty"`  // ✅ 添加：是否加密
	CreatedAt  time.Time  `json:"createdAt"`
}

// ScreenshotListResponse 截图列表响应
type ScreenshotListResponse struct {
	Items    []Screenshot `json:"items"`
	Total    int          `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"pageSize"`
}

// GetScreenshotsRequest 获取截图请求参数
type GetScreenshotsRequest struct {
	DeviceID  string `form:"deviceId"`
	StartDate string `form:"startDate"`
	EndDate   string `form:"endDate"`
	Page      int    `form:"page" default:"1"`
	PageSize  int    `form:"pageSize" default:"20"`
}

// ========== 活动日志模型 ==========

// ActivityLog 活动日志模型
type ActivityLog struct {
	ID              string     `json:"id"`
	DeviceID        string     `json:"deviceId"`
	ProcessName     string     `json:"processName"`
	WindowTitle     string     `json:"windowTitle"`
	BrowserTitle    string     `json:"browserTitle"`
	StartedAt       time.Time  `json:"startedAt"`
	EndedAt         *time.Time `json:"endedAt,omitempty"`
	IsIdle          bool       `json:"isIdle"`
	DurationSeconds int        `json:"durationSeconds"`
	CreatedAt       time.Time  `json:"createdAt"`
	Reported        bool       `json:"reported"`                    // ✅ 添加：是否已报告到后端
	ReportedAt      *time.Time `json:"reportedAt,omitempty"`       // ✅ 添加：报告时间
	RetryCount      int        `json:"retryCount,omitempty"`       // ✅ 添加：重试次数
	ErrorMessage    string     `json:"errorMessage,omitempty"`     // ✅ 添加：错误信息
}

// ActivityStats 活动统计
type ActivityStats struct {
	TotalDuration    int64            `json:"totalDuration"`
	IdleDuration     int64            `json:"idleDuration"`
	ActiveDuration   int64            `json:"activeDuration"`
	TopApps          []AppUsage       `json:"topApps"`
	DailyActivity    []DailyActivity  `json:"dailyActivity"`
	HourlyActivity   map[string]int64 `json:"hourlyActivity,omitempty"`
	ProductivityRate float64          `json:"productivityRate"`
}

// AppUsage 应用使用统计
type AppUsage struct {
	ProcessName string  `json:"processName"`
	Duration    int64   `json:"duration"`
	Percentage  float64 `json:"percentage"`
}

// DailyActivity 每日活动统计
type DailyActivity struct {
	Date   string  `json:"date"`
	Hours  float64 `json:"hours"`
	Idle   float64 `json:"idle"`
	Active float64 `json:"active"`
}

type DailyActivitySummary struct {
	Date          string  `json:"date"`
	ActiveMinutes int64   `json:"activeMinutes"`
	IdleMinutes   int64   `json:"idleMinutes"`
	AppCount      int64   `json:"appCount"`
	Productivity  float64 `json:"productivity"`
}

// ========== 上传任务模型 ==========

// UploadTask 上传任务模型
type UploadTask struct {
	ID          uuid.UUID  `json:"id"`
	DeviceID    uuid.UUID  `json:"deviceId"`
	Type        string     `json:"type"`
	StartTime   time.Time  `json:"startTime"`
	EndTime     time.Time  `json:"endTime"`
	Status      string     `json:"status"` // pending, running, completed, failed
	Progress    int        `json:"progress"`
	Result      string     `json:"result,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
}

// ========== 清理策略模型 ==========

// CleanupPolicyConfig 清理策略配置
type CleanupPolicyConfig struct {
	AutoCleanupEnabled        bool       `json:"autoCleanupEnabled"`
	RecordingRetentionDays    int        `json:"recordingRetentionDays"`
	ScreenshotRetentionDays   int        `json:"screenshotRetentionDays"`
	LogRetentionDays          int        `json:"logRetentionDays"`
	DiskUsageThresholdPercent int        `json:"diskUsageThresholdPercent"`
	CleanupIntervalMinutes    int        `json:"cleanupIntervalMinutes"`
	LastRunAt                 *time.Time `json:"lastRunAt,omitempty"`
	UpdatedAt                 time.Time  `json:"updatedAt"`
}

type CleanupPolicy = CleanupPolicyConfig

// StorageStatsInfo 存储统计信息
type StorageStatsInfo struct {
	TotalBytes       int64      `json:"totalBytes"`
	UsedBytes        int64      `json:"usedBytes"`
	FreeBytes        int64      `json:"freeBytes"`
	DiskUsagePercent int        `json:"diskUsagePercent"`
	RecordingCount   int        `json:"recordingCount"`
	ScreenshotCount  int        `json:"screenshotCount"`
	LogCount         int        `json:"logCount"`
	LastCleanupAt    *time.Time `json:"lastCleanupAt,omitempty"`
}

type StorageStats = StorageStatsInfo

// CleanupLogEntry 清理日志条目
type CleanupLogEntry struct {
	ID                 string    `json:"id"`
	TriggerType        string    `json:"triggerType"` // auto, manual
	Status             string    `json:"status"`      // success, failed
	DeletedRecordings  int       `json:"deletedRecordings"`
	DeletedScreenshots int       `json:"deletedScreenshots"`
	DeletedLogs        int       `json:"deletedLogs"`
	DeletedBytes       int64     `json:"deletedBytes"`
	Message            string    `json:"message"`
	StartedAt          time.Time `json:"startedAt"`
	CompletedAt        time.Time `json:"completedAt"`
}

type CleanupLog = CleanupLogEntry

// ========== 认证模型 ==========

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    string `json:"expiresAt"`
}

// ========== 上传签名模型 ==========

// UploadSignaturePayload 上传签名负载
type UploadSignaturePayload struct {
	UploadID    string `json:"uploadId"`
	DeviceID    string `json:"deviceId,omitempty"`
	FileName    string `json:"fileName,omitempty"`
	FileType    string `json:"fileType,omitempty"`
	FileID      string `json:"fileId,omitempty"`
	ChunkIndex  int    `json:"chunkIndex,omitempty"`
	TotalChunks int    `json:"totalChunks,omitempty"`
	Data        string `json:"data,omitempty"`
}

// UploadSignatureInput 上传签名输入
type UploadSignatureInput struct {
	Timestamp string `json:"timestamp"`
	Nonce     string `json:"nonce"`
	Signature string `json:"signature"`
}

// ========== 用户管理请求/响应模型 ==========

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string   `json:"username" binding:"required"`
	Password string   `json:"password" binding:"required,min=6"`
	Email    string   `json:"email"`
	RealName string   `json:"realName"`
	RoleIDs  []string `json:"roleIds"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Email    string   `json:"email"`
	RealName string   `json:"realName"`
	Status   int      `json:"status"`
	RoleIDs  []string `json:"roleIds"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Password string `json:"password" binding:"required,min=6"`
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions" binding:"required"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// ========== 权限常量定义 ==========

const (
	// 用户管理权限
	PermUserView   = "user:view"
	PermUserCreate = "user:create"
	PermUserUpdate = "user:update"
	PermUserDelete = "user:delete"
	PermUserManage = "user:manage"

	// 角色管理权限
	PermRoleView   = "role:view"
	PermRoleCreate = "role:create"
	PermRoleUpdate = "role:update"
	PermRoleDelete = "role:delete"

	// 设备管理权限
	PermDeviceView   = "device:view"
	PermDeviceManage = "device:manage"
	PermDeviceBind   = "device:bind"
	PermDeviceUnbind = "device:unbind"

	// 录屏管理权限
	PermRecordingView   = "recording:view"
	PermRecordingPlay   = "recording:play"
	PermRecordingExport = "recording:export"
	PermRecordingDelete = "recording:delete"
	PermRecordingUpload = "recording:upload"

	// 截图管理权限
	PermScreenshotView   = "screenshot:view"
	PermScreenshotDelete = "screenshot:delete"

	// 活动日志权限
	PermActivityView   = "activity:view"
	PermActivityExport = "activity:export"
	PermActivityDelete = "activity:delete"

	// 策略管理权限
	PermPolicyView   = "policy:view"
	PermPolicyCreate = "policy:create"
	PermPolicyUpdate = "policy:update"
	PermPolicyDelete = "policy:delete"

	// 审计权限
	PermAuditView = "audit:view"

	// ========== 出款管理权限（新增） ==========
	PermAttendanceView = "attendance:view"  // 查看考勤绩效
	PermAttendanceEdit = "attendance:edit"  // 编辑考勤绩效
	PermSiteView       = "site:view"        // 查看出款站点
	PermSiteManage     = "site:manage"      // 管理出款站点
	PermStatsView      = "stats:view"       // 查看出款统计
	PermStatsEdit      = "stats:edit"       // 编辑出款统计（上传数据）

)

// ========== 操作日志模型 ==========

// OperationLog 操作日志模型（只记录修改操作）
type OperationLog struct {
    ID              string    `json:"id"`
    OperatorID      string    `json:"operatorId"`
    OperatorName    string    `json:"operatorName"`
    OperatorRole    string    `json:"operatorRole"`
    OperationType   string    `json:"operationType"`   // CREATE, UPDATE, DELETE, UPLOAD
    OperationModule string    `json:"operationModule"` // USER, ROLE, DEVICE, SALARY, ATTENDANCE, PERFORMANCE, PENALTY, SITE_STATS
    OperationDesc   string    `json:"operationDesc"`   // 详细描述，如：修改员工阿富的工资: 底薪18000→18500
    TargetID        string    `json:"targetId"`
    TargetName      string    `json:"targetName"`
    IPAddress       string    `json:"ipAddress"`
    UserAgent       string    `json:"userAgent"`
    OldValue        string    `json:"oldValue"`        // 修改前的值（JSON格式）
    NewValue        string    `json:"newValue"` 
	ExecutionTimeMs int       `json:"executionTimeMs"`       // 修改后的值（JSON格式）
    CreatedAt       time.Time `json:"createdAt"`
}

// 操作类型常量
const (
    OperationTypeCreate = "CREATE"  // 创建
    OperationTypeUpdate = "UPDATE"  // 修改
    OperationTypeDelete = "DELETE"  // 删除
    OperationTypeUpload = "UPLOAD"  // 上传
)

// 操作模块常量
const (
    OperationModuleUser        = "USER"
    OperationModuleRole        = "ROLE"
    OperationModuleDevice      = "DEVICE"
    OperationModuleAttendance  = "ATTENDANCE"
    OperationModulePerformance = "PERFORMANCE"
    OperationModulePenalty     = "PENALTY"
    OperationModuleSiteStats   = "SITE_STATS"
    OperationModulePolicy      = "POLICY"
)

// 添加权限常量
const (
    // 操作日志权限
    PermOperationLogView   = "operation_log:view"
    PermOperationLogExport = "operation_log:export"
    PermOperationLogDelete = "operation_log:delete"
)

// DefaultRoles 默认角色定义
var DefaultRoles = []Role{
	{
		Name:        "管理员",
		Description: "拥有所有权限",
		Permissions: []string{
			PermUserView, PermUserCreate, PermUserUpdate, PermUserDelete, PermUserManage,
			PermRoleView, PermRoleCreate, PermRoleUpdate, PermRoleDelete,
			PermAttendanceView, PermAttendanceEdit,
			PermSiteView, PermSiteManage,
			PermStatsView, PermStatsEdit,
			PermOperationLogView, PermOperationLogExport, PermOperationLogDelete,
		},
	},
	{
		Name:        "审计员",
		Description: "只读权限，可查看业务数据",
		Permissions: []string{
			PermAttendanceView,
			PermSiteView,
			PermStatsView,
			PermOperationLogView,
		},
	},
	{
		Name:        "操作员",
		Description: "可编辑考勤与出款数据",
		Permissions: []string{
			PermAttendanceView, PermAttendanceEdit,
			PermSiteView, PermSiteManage,
			PermStatsView, PermStatsEdit,
		},
	},
}

// ========== 辅助函数 ==========

// UTCNow 获取当前 UTC 时间
func UTCNow() time.Time {
	return time.Now().UTC()
}

// ToUTC 转换为 UTC 时间
func ToUTC(value time.Time) time.Time {
	if value.IsZero() {
		return time.Time{}
	}
	return value.UTC()
}

// FormatUTC 格式化 UTC 时间为 RFC3339
func FormatUTC(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

// ParseUTC 解析 RFC3339 格式的时间字符串为 UTC 时间
func ParseUTC(value string) (time.Time, error) {
	return ParseFlexibleUTC(value)
}

// ParseFlexibleUTC 解析多种 ISO 8601 / RFC3339 时间格式为 UTC
func ParseFlexibleUTC(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("empty time value")
	}

	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	}
	for _, layout := range formats {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported time format: %s", value)
}

// ========== 签名函数（优化版 - 使用 sync.Once 延迟初始化） ==========

var (
	jwtSecretValue    []byte
	policySecretValue []byte
	secretsOnce       sync.Once
)

// initSecrets 初始化内存中的密钥值（延迟加载）
func initSecrets() {
    jwtEnv := os.Getenv("JWT_SECRET")
    if strings.TrimSpace(jwtEnv) == "" {
        // ✅ 生产环境必须设置，开发环境给警告但使用随机值
        log.Println("⚠️ WARNING: JWT_SECRET not set, using random value (insecure for production)")
        jwtEnv = generateRandomSecret(32)
    }
    jwtSecretValue = []byte(jwtEnv)

    policyEnv := os.Getenv("POLICY_SIGNATURE_SECRET")
    if strings.TrimSpace(policyEnv) == "" {
        log.Println("⚠️ WARNING: POLICY_SIGNATURE_SECRET not set, using random value (insecure for production)")
        policyEnv = generateRandomSecret(32)
    }
    policySecretValue = []byte(policyEnv)
}

// 添加辅助函数
func generateRandomSecret(length int) string {
    bytes := make([]byte, length)
    if _, err := rand.Read(bytes); err != nil {
        return "fallback-secret-do-not-use"
    }
    return hex.EncodeToString(bytes)[:length]
}

// JWTSecret 获取 JWT 密钥
func JWTSecret() []byte {
	secretsOnce.Do(initSecrets)
	return jwtSecretValue
}

// PolicySignatureSecret 获取策略签名密钥
func PolicySignatureSecret() []byte {
	secretsOnce.Do(initSecrets)
	return policySecretValue
}

// SignPolicyContent 使用当前 POLICY_SIGNATURE_SECRET 对策略 content 原文签名
func SignPolicyContent(content string) string {
	secret := PolicySignatureSecret()
	if len(secret) == 0 || strings.TrimSpace(content) == "" {
		return ""
	}
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(content))
	return hex.EncodeToString(mac.Sum(nil))
}

// marshalSignaturePayload 保留用于兼容结构体定义（目前签名流程改为基于 DeviceSecret）
func marshalSignaturePayload(payload UploadSignaturePayload) string {
	// 创建一个不包含 Data 字段的结构体用于序列化
	type sigPayload struct {
		UploadID    string `json:"uploadId"`
		DeviceID    string `json:"deviceId,omitempty"`
		FileName    string `json:"fileName,omitempty"`
		FileType    string `json:"fileType,omitempty"`
		FileID      string `json:"fileId,omitempty"`
		ChunkIndex  int    `json:"chunkIndex,omitempty"`
		TotalChunks int    `json:"totalChunks,omitempty"`
	}

	p := sigPayload{
		UploadID:    payload.UploadID,
		DeviceID:    payload.DeviceID,
		FileName:    payload.FileName,
		FileType:    payload.FileType,
		FileID:      payload.FileID,
		ChunkIndex:  payload.ChunkIndex,
		TotalChunks: payload.TotalChunks,
	}

	bytes, err := json.Marshal(p)
	if err != nil {
		log.Printf("Failed to marshal payload: %v", err)
		return "{}"
	}
	return string(bytes)
}
// escapeJSONString 转义 JSON 字符串中的特殊字符
func escapeJSONString(s string) string {
    // 简单的 JSON 字符串转义
    s = strings.ReplaceAll(s, "\\", "\\\\")
    s = strings.ReplaceAll(s, "\"", "\\\"")
    s = strings.ReplaceAll(s, "\n", "\\n")
    s = strings.ReplaceAll(s, "\r", "\\r")
    s = strings.ReplaceAll(s, "\t", "\\t")
    return s
}

// ========== 响应函数 ==========

// Send 发送统一格式的响应
func Send(c *gin.Context, statusCode int, code int, message string, data interface{}) {
	requestID := c.GetString("requestID")
	c.JSON(statusCode, APIResponse{
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: FormatUTC(UTCNow()),
		RequestID: requestID,
	})
}

// SendError 发送错误响应
// SendError 发送错误响应
func SendError(c *gin.Context, statusCode int, code int, message string) {
	// 添加错误日志记录
	log.Printf("[ERROR] SendError: status=%d, code=%d, message=%s, path=%s", 
		statusCode, code, message, c.Request.URL.Path)
	
	Send(c, statusCode, code, message, nil)
}

// SendSuccess 发送成功响应
func SendSuccess(c *gin.Context, data interface{}) {
	Send(c, 200, 0, "success", data)
}

// ========== 辅助方法 ==========

// HasPermission 检查角色是否拥有指定权限
func HasPermission(role Role, permission string) bool {
	for _, p := range role.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasAnyPermission 检查角色是否拥有任意一个权限
func HasAnyPermission(role Role, permissions ...string) bool {
	for _, permission := range permissions {
		if HasPermission(role, permission) {
			return true
		}
	}
	return false
}

// ToUserResponse 将 User 转换为 UserResponse
func (u *User) ToUserResponse(roles []Role) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		RealName:  u.RealName,
		Status:    u.Status,
		Roles:     roles,
		CreatedAt: u.CreatedAt,
		LastLogin: u.LastLogin,
	}
}

// ========== JWT Claims ==========

// Claims JWT 声明结构
type Claims struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// DeviceClaims 设备 JWT Claims
type DeviceClaims struct {
	DeviceID string `json:"deviceId"`
	jwt.RegisteredClaims
}

// ========== 配置验证 ==========

// ValidateConfig 验证所有必需的配置（开发环境会警告）
// backend/models/models.go - 修复 ValidateConfig 函数

// ValidateConfig 验证所有必需的配置
func ValidateConfig() error {
	env := strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))
	isProduction := env == "production" || env == "prod"

	jwtSecret := JWTSecret()
	jwtSecretStr := string(jwtSecret)

	// 检查是否使用默认密钥
	isDefaultJWT := jwtSecretStr == "development-secret-key-do-not-use-in-production"

	if isProduction {
		if isDefaultJWT {
			return fmt.Errorf("JWT_SECRET environment variable is required and must not be the default value")
		}
		if len(jwtSecret) < 32 {
			return fmt.Errorf("JWT_SECRET must be at least 32 characters in production")
		}
	} else if isDefaultJWT {
		fmt.Println("⚠️  WARNING: Using default JWT_SECRET. Set a strong secret before production deploy.")
	}

	return nil
}

// IsOnline 判断设备是否在线（5分钟内有活动）
func (d *Device) IsOnline() bool {
    if d.LastSeenAt == nil {
        return false
    }
    return time.Since(*d.LastSeenAt) < 3*time.Minute
}

// GetOnlineStatus 获取在线状态文本
func (d *Device) GetOnlineStatus() string {
    if d.IsOnline() {
        return "online"
    }
    return "offline"
}
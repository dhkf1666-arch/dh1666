package handlers

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"enterprise-agent/backend/models"
)

// ========== 操作日志处理器 ==========

type OperationLogHandler struct {
	db *sql.DB
}

func NewOperationLogHandler(db *sql.DB) *OperationLogHandler {
	return &OperationLogHandler{db: db}
}

// GetOperationLogs 获取操作日志列表
func (h *OperationLogHandler) GetOperationLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	operatorName := c.Query("operatorName")
	operationModule := c.Query("operationModule")
	operationType := c.Query("operationType")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	where := "1=1"
	args := []interface{}{}
	argIdx := 1

	if operatorName != "" {
		where += fmt.Sprintf(` AND (
			%s ILIKE $%d
			OR ol.operator_name ILIKE $%d
			OR u.real_name ILIKE $%d
			OR u.username ILIKE $%d
		)`, operationLogOperatorDisplay, argIdx, argIdx, argIdx, argIdx)
		args = append(args, "%"+operatorName+"%")
		argIdx++
	}
	if operationModule != "" {
		where += fmt.Sprintf(" AND ol.operation_module = $%d", argIdx)
		args = append(args, operationModule)
		argIdx++
	}
	if operationType != "" {
		where += fmt.Sprintf(" AND ol.operation_type = $%d", argIdx)
		args = append(args, operationType)
		argIdx++
	}
	if startDate != "" {
		where += fmt.Sprintf(" AND ol.created_at >= $%d", argIdx)
		args = append(args, startDate)
		argIdx++
	}
	if endDate != "" {
		where += fmt.Sprintf(" AND ol.created_at <= $%d", argIdx)
		args = append(args, endDate+" 23:59:59")
		argIdx++
	}

	// 查询总数
	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT ol.id) FROM operation_logs ol
		%s
		WHERE %s`, operationLogUserJoin, where)
	var total int
	if err := h.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}

	// 查询数据（添加 execution_time_ms 字段）
	query := fmt.Sprintf(`
		SELECT ol.id, ol.operator_id, %s AS operator_name, ol.operator_role, ol.operation_type,
			   ol.operation_module, ol.operation_desc, ol.target_id, ol.target_name,
			   ol.ip_address, ol.user_agent, ol.execution_time_ms, ol.created_at
		FROM operation_logs ol
		%s
		WHERE %s
		ORDER BY ol.created_at DESC
		LIMIT $%d OFFSET $%d
	`, operationLogOperatorDisplay, operationLogUserJoin, where, argIdx, argIdx+1)

	args = append(args, pageSize, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var log models.OperationLog
		var operatorID sql.NullString
		var operatorRole, targetID, targetName, userAgent sql.NullString
		var executionTimeMs sql.NullInt32

		err := rows.Scan(
			&log.ID, &operatorID, &log.OperatorName, &operatorRole,
			&log.OperationType, &log.OperationModule, &log.OperationDesc,
			&targetID, &targetName, &log.IPAddress, &userAgent,
			&executionTimeMs, &log.CreatedAt,
		)
		if err != nil {
			continue
		}

		if operatorID.Valid {
			log.OperatorID = operatorID.String
		}
		if operatorRole.Valid {
			log.OperatorRole = operatorRole.String
		}
		if targetID.Valid {
			log.TargetID = targetID.String
		}
		if targetName.Valid {
			log.TargetName = targetName.String
		}
		if userAgent.Valid {
			log.UserAgent = userAgent.String
		}
		if executionTimeMs.Valid {
			log.ExecutionTimeMs = int(executionTimeMs.Int32)
		}

		logs = append(logs, map[string]interface{}{
			"id":              log.ID,
			"operatorId":      log.OperatorID,
			"operatorName":    log.OperatorName,
			"operatorRole":    log.OperatorRole,
			"operationType":   log.OperationType,
			"operationModule": log.OperationModule,
			"operationDesc":   log.OperationDesc,
			"targetId":        log.TargetID,
			"targetName":      log.TargetName,
			"ipAddress":       log.IPAddress,
			"userAgent":       log.UserAgent,
			"executionTimeMs": log.ExecutionTimeMs,
			"createdAt":       log.CreatedAt,
		})
	}

	models.Send(c, 200, 0, "success", gin.H{
		"items":    logs,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetOperationModules 获取操作模块列表
func (h *OperationLogHandler) GetOperationModules(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT DISTINCT operation_module FROM operation_logs 
		ORDER BY operation_module
	`)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}
	defer rows.Close()

	var modules []string
	for rows.Next() {
		var module string
		rows.Scan(&module)
		modules = append(modules, module)
	}

	models.Send(c, 200, 0, "success", modules)
}

// GetOperationTypes 获取操作类型列表
func (h *OperationLogHandler) GetOperationTypes(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT DISTINCT operation_type FROM operation_logs 
		ORDER BY operation_type
	`)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}
	defer rows.Close()

	var types []string
	for rows.Next() {
		var t string
		rows.Scan(&t)
		types = append(types, t)
	}

	models.Send(c, 200, 0, "success", types)
}

// ExportOperationLogs 导出操作日志
func (h *OperationLogHandler) ExportOperationLogs(c *gin.Context) {
	operatorName := c.Query("operatorName")
	operationModule := c.Query("operationModule")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	where := "1=1"
	args := []interface{}{}
	argIdx := 1

	if operatorName != "" {
		where += fmt.Sprintf(` AND (
			%s ILIKE $%d
			OR ol.operator_name ILIKE $%d
			OR u.real_name ILIKE $%d
			OR u.username ILIKE $%d
		)`, operationLogOperatorDisplay, argIdx, argIdx, argIdx, argIdx)
		args = append(args, "%"+operatorName+"%")
		argIdx++
	}
	if operationModule != "" {
		where += fmt.Sprintf(" AND ol.operation_module = $%d", argIdx)
		args = append(args, operationModule)
		argIdx++
	}
	if startDate != "" {
		where += fmt.Sprintf(" AND ol.created_at >= $%d", argIdx)
		args = append(args, startDate)
		argIdx++
	}
	if endDate != "" {
		where += fmt.Sprintf(" AND ol.created_at <= $%d", argIdx)
		args = append(args, endDate+" 23:59:59")
		argIdx++
	}

	query := fmt.Sprintf(`
		SELECT %s AS operator_name, ol.operator_role, ol.operation_type, ol.operation_module,
			   ol.operation_desc, ol.target_name, ol.ip_address, ol.execution_time_ms, ol.created_at
		FROM operation_logs ol
		%s
		WHERE %s
		ORDER BY ol.created_at DESC
		LIMIT 10000
	`, operationLogOperatorDisplay, operationLogUserJoin, where)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}
	defer rows.Close()

	var builder strings.Builder
	builder.WriteString("操作人,角色,操作类型,操作模块,操作描述,目标名称,IP地址,耗时(ms),操作时间\n")

	for rows.Next() {
		var operatorName, operatorRole, operationType, operationModule, operationDesc, targetName, ipAddress string
		var executionTimeMs int
		var createdAt time.Time
		rows.Scan(&operatorName, &operatorRole, &operationType, &operationModule,
			&operationDesc, &targetName, &ipAddress, &executionTimeMs, &createdAt)

		builder.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%d,%s\n",
			escapeCSV(operatorName), escapeCSV(operatorRole),
			escapeCSV(operationType), escapeCSV(operationModule),
			escapeCSV(operationDesc), escapeCSV(targetName),
			escapeCSV(ipAddress), executionTimeMs, createdAt.Format("2006-01-02 15:04:05")))
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=operation_logs_"+time.Now().Format("20060102")+".csv")
	c.String(200, builder.String())
}

// DeleteOperationLogs 批量删除操作日志
func (h *OperationLogHandler) DeleteOperationLogs(c *gin.Context) {
	var req struct {
		IDs       []string `json:"ids"`
		DaysOld   int      `json:"daysOld"`
		DeleteAll bool     `json:"deleteAll"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		models.SendError(c, 400, 1001, err.Error())
		return
	}

	if req.DeleteAll {
		_, err := h.db.Exec("DELETE FROM operation_logs")
		if err != nil {
			models.SendError(c, 500, 5000, err.Error())
			return
		}
		models.Send(c, 200, 0, "所有操作日志已删除", gin.H{"deleted": "all"})
		return
	}

	if req.DaysOld > 0 {
		cutoff := time.Now().AddDate(0, 0, -req.DaysOld)
		result, err := h.db.Exec("DELETE FROM operation_logs WHERE created_at < $1", cutoff)
		if err != nil {
			models.SendError(c, 500, 5000, err.Error())
			return
		}
		rowsAffected, _ := result.RowsAffected()
		models.Send(c, 200, 0, fmt.Sprintf("已删除 %d 条记录", rowsAffected), gin.H{"deleted": rowsAffected})
		return
	}

	if len(req.IDs) == 0 {
		models.SendError(c, 400, 1001, "请提供要删除的ID")
		return
	}

	placeholders := make([]string, len(req.IDs))
	args := make([]interface{}, len(req.IDs))
	for i, id := range req.IDs {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = id
	}

	query := fmt.Sprintf("DELETE FROM operation_logs WHERE id IN (%s)", strings.Join(placeholders, ","))
	result, err := h.db.Exec(query, args...)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}
	rowsAffected, _ := result.RowsAffected()
	models.Send(c, 200, 0, fmt.Sprintf("已删除 %d 条记录", rowsAffected), gin.H{"deleted": rowsAffected})
}

func escapeCSV(s string) string {
	if s == "" {
		return ""
	}
	if strings.ContainsAny(s, ",\"\n\r") {
		return fmt.Sprintf("\"%s\"", strings.ReplaceAll(s, "\"", "\"\""))
	}
	return s
}

// ========== 操作日志中间件（只记录修改操作）==========

// OperationLogMiddleware 操作日志中间件 - 只记录 POST/PUT/DELETE 请求
func OperationLogMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只记录修改操作（POST、PUT、DELETE）
		method := c.Request.Method
		if method != "POST" && method != "PUT" && method != "DELETE" {
			c.Next()
			return
		}

		// 跳过不需要记录日志的路径
		skipPaths := []string{
			"/api/v1/auth/login",
			"/api/v1/auth/refresh",
			"/api/v1/auth/logout",
			"/health",
			"/api/v1/ws/",
			"/api/v1/screenshots",
			"/api/v1/upload/chunk",
			"/api/v1/upload/complete",
			"/api/v1/operation-logs",
			"/api/v1/activities/report",
			"/api/v1/tasks",
			"/api/v1/upload-tasks",
		}

		requestPath := c.Request.URL.Path
		for _, skip := range skipPaths {
			if strings.HasPrefix(requestPath, skip) {
				c.Next()
				return
			}
		}

		// 记录开始时间 ⭐ 新增
		startTime := time.Now()

		// 读取请求体（仅用于读取后重新设置，不保存到日志）
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			// 重新设置body，让后续handler可以读取
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		c.Next()

		// ⭐ 计算执行耗时（毫秒）
		executionTimeMs := int(time.Since(startTime).Milliseconds())

		// 获取操作人信息
		operatorID := c.GetString("userID")
		username := c.GetString("username")
		operatorName := resolveOperatorDisplayName(db, operatorID, username)
		operatorRole := c.GetString("role")

		// 如果是设备认证，不记录操作日志
		if operatorID == "" && operatorName == "" {
			return
		}

		// 判断操作类型和模块
		operationType := getOperationType(method)
		operationModule := getOperationModule(requestPath)
		responseStatus := c.Writer.Status()

		// 只记录成功的操作（2xx状态码）
		if responseStatus < 200 || responseStatus >= 300 {
			return
		}

		// 业务 handler 已写入更详细日志的路径，跳过中间件记录，避免重复且描述不清
		if shouldSkipMiddlewareOperationLog(requestPath) {
			return
		}

		targetID, targetName := extractLogTarget(c)
		operationDesc := buildOperationDesc(requestPath, method)
		operationDesc = appendTargetHint(operationDesc, targetName)

		log := &models.OperationLog{
			ID:              uuid.New().String(),
			OperatorID:      operatorID,
			OperatorName:    operatorName,
			OperatorRole:    operatorRole,
			OperationType:   operationType,
			OperationModule: operationModule,
			OperationDesc:   operationDesc,
			TargetID:        targetID,
			TargetName:      targetName,
			IPAddress:       c.ClientIP(), // ⭐ 确保IP正确获取
			UserAgent:       c.Request.UserAgent(),
			ExecutionTimeMs: executionTimeMs, // ⭐ 新增：耗时
			CreatedAt:       time.Now(),
		}

		// 异步保存日志
		go saveOperationLog(db, log)
	}
}

func getOperationType(method string) string {
	switch method {
	case "POST":
		return "CREATE"
	case "PUT":
		return "UPDATE"
	case "DELETE":
		return "DELETE"
	default:
		return "OTHER"
	}
}

func getOperationModule(path string) string {
	switch {
	case strings.Contains(path, "/users"):
		return "USER"
	case strings.Contains(path, "/roles"):
		return "ROLE"
	case strings.Contains(path, "/devices"):
		return "DEVICE"
	case strings.Contains(path, "/recordings"):
		return "RECORDING"
	case strings.Contains(path, "/screenshots"):
		return "SCREENSHOT"
	case strings.Contains(path, "/activities"):
		return "ACTIVITY"
	case strings.Contains(path, "/policies"):
		return "POLICY"
	case strings.Contains(path, "/employees"):
		return "EMPLOYEE"
	case strings.Contains(path, "/attendance"):
		return "ATTENDANCE"
	case strings.Contains(path, "/performance"):
		return "PERFORMANCE"
	case strings.Contains(path, "/penalty"):
		return "PENALTY"
	case strings.Contains(path, "/site-stats"):
		return "SITE_STATS"
	case strings.Contains(path, "/salary"):
		return "SALARY"
	case strings.Contains(path, "/tasks"):
		return "TASK"
	case strings.Contains(path, "/cleanup"):
		return "POLICY"
	default:
		return "OTHER"
	}
}

type operationDescRule struct {
	pathKeyword string
	method      string
	description string
}

// 路径匹配顺序：越具体的规则越靠前
var operationDescRules = []operationDescRule{
	{"/site-stats/upload/preview", "POST", "预览出款统计上传表头"},
	{"/site-stats/upload", "POST", "上传出款统计数据"},
	{"/site-stats/data/clear-by-date", "DELETE", "清除指定站点出款统计"},
	{"/site-stats/clear-by-date-only", "DELETE", "清除指定日期全部出款统计"},
	{"/site-stats/data/clear", "DELETE", "清空全部出款统计数据"},
	{"/site-stats/sites", "POST", "新增出款站点"},
	{"/site-stats/sites", "PUT", "修改出款站点"},
	{"/site-stats/sites", "DELETE", "删除出款站点"},
	{"/site-stats/employee-accounts", "POST", "新增出款员工账号"},
	{"/site-stats/employee-accounts", "PUT", "修改出款员工账号"},
	{"/site-stats/employee-accounts", "DELETE", "删除出款员工账号"},
	{"/employees", "POST", "新增员工"},
	{"/employees", "PUT", "修改员工信息"},
	{"/employees", "DELETE", "删除员工"},
	{"/attendance/records", "POST", "保存考勤记录"},
	{"/performance/batch", "POST", "批量保存绩效考核"},
	{"/performance/", "PUT", "修改员工绩效考核"},
	{"/performance/", "DELETE", "删除员工绩效考核"},
	{"/penalty/record", "POST", "添加罚款记录"},
	{"/penalty/records", "PUT", "修改罚款记录"},
	{"/penalty/records", "DELETE", "删除罚款记录"},
	{"/salary/upload", "POST", "上传工资表"},
	{"/salary/records", "PUT", "修改工资记录"},
	{"/salary/records", "DELETE", "删除工资记录"},
	{"/users", "POST", "创建系统用户"},
	{"/users", "PUT", "修改系统用户"},
	{"/users", "DELETE", "删除系统用户"},
	{"/roles", "POST", "创建角色"},
	{"/roles", "PUT", "修改角色"},
	{"/roles", "DELETE", "删除角色"},
	{"/policies/", "publish", "发布监控策略"},
	{"/policies", "POST", "创建监控策略"},
	{"/policies", "PUT", "修改监控策略"},
	{"/policies", "DELETE", "删除监控策略"},
	{"/policies/", "rollback", "回滚监控策略"},
	{"/devices", "POST", "绑定监控设备"},
	{"/devices", "PUT", "修改设备信息"},
	{"/devices", "DELETE", "解绑监控设备"},
	{"/recordings", "DELETE", "删除录屏文件"},
	{"/screenshots/batch-delete", "POST", "批量删除截图"},
	{"/screenshots", "DELETE", "删除截图"},
	{"/cleanup/manual", "POST", "手动执行存储清理"},
	{"/cleanup/policy", "PUT", "修改存储清理策略"},
	{"/tasks/batch-status", "POST", "批量更新任务状态"},
	{"/tasks/", "cancel", "取消上传任务"},
	{"/tasks/", "retry", "重试上传任务"},
	{"/tasks", "POST", "创建上传任务"},
	{"/tasks", "DELETE", "删除上传任务"},
	{"/user/profile", "PUT", "修改个人资料"},
	{"/user/change-password", "POST", "修改登录密码"},
}

func shouldSkipMiddlewareOperationLog(path string) bool {
	prefixes := []string{
		"/api/v1/users",
		"/api/v1/user/profile",
		"/api/v1/user/change-password",
		"/api/v1/roles",
		"/api/v1/admin/attendance/records",
		"/api/v1/admin/penalty",
		"/api/v1/admin/performance",
		"/api/v1/admin/employees",
		"/api/v1/admin/site-stats",
		"/api/v1/salary",
		"/api/v1/tasks",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func buildOperationDesc(path, method string) string {
	if strings.Contains(path, "/users/") && strings.Contains(path, "/roles") {
		switch method {
		case "POST":
			return "分配用户角色"
		case "DELETE":
			return "移除用户角色"
		}
	}
	if strings.Contains(path, "reset-password") && method == "POST" {
		return "重置用户密码"
	}

	for _, rule := range operationDescRules {
		if !strings.Contains(path, rule.pathKeyword) {
			continue
		}
		if rule.method != "" && rule.method != method && !strings.Contains(path, rule.method) {
			continue
		}
		return rule.description
	}

	actionName := map[string]string{
		"POST":   "新增",
		"PUT":    "修改",
		"DELETE": "删除",
		"PATCH":  "修改",
	}[method]
	moduleName := getOperationModuleLabel(path)
	if actionName != "" && moduleName != "" {
		return actionName + moduleName
	}
	return method + " " + path
}

func getOperationModuleLabel(path string) string {
	switch {
	case strings.Contains(path, "/users"):
		return "系统用户"
	case strings.Contains(path, "/roles"):
		return "角色"
	case strings.Contains(path, "/devices"):
		return "设备"
	case strings.Contains(path, "/recordings"):
		return "录屏"
	case strings.Contains(path, "/screenshots"):
		return "截图"
	case strings.Contains(path, "/policies"):
		return "监控策略"
	case strings.Contains(path, "/cleanup"):
		return "清理策略"
	case strings.Contains(path, "/employees"):
		return "员工"
	case strings.Contains(path, "/attendance"):
		return "考勤"
	case strings.Contains(path, "/performance"):
		return "绩效考核"
	case strings.Contains(path, "/penalty"):
		return "罚款"
	case strings.Contains(path, "/site-stats"):
		return "出款统计"
	case strings.Contains(path, "/salary"):
		return "工资"
	case strings.Contains(path, "/tasks"):
		return "任务"
	default:
		return "数据"
	}
}

func extractLogTarget(c *gin.Context) (targetID, targetName string) {
	if id := c.Param("id"); id != "" {
		return id, ""
	}
	if employeeID := c.Param("employeeId"); employeeID != "" {
		if month := c.Param("month"); month != "" {
			return employeeID, month + " 月绩效"
		}
		return employeeID, ""
	}
	return "", ""
}

func appendTargetHint(desc, targetName string) string {
	targetName = strings.TrimSpace(targetName)
	if targetName == "" || strings.Contains(desc, targetName) {
		return desc
	}
	return desc + "：" + targetName
}

func getOperationDesc(path, method string) string {
	return buildOperationDesc(path, method)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// saveOperationLog 保存操作日志到数据库（包含 IP 和耗时）
func saveOperationLog(db *sql.DB, log *models.OperationLog) {
    println("=== saveOperationLog called ===")
    println("ID:", log.ID)
    println("OperatorName:", log.OperatorName)
    println("OperationType:", log.OperationType)
    println("OperationModule:", log.OperationModule)
    println("OperationDesc:", log.OperationDesc)
    
    query := `
        INSERT INTO operation_logs (
            id, operator_id, operator_name, operator_role, operation_type,
            operation_module, operation_desc, target_id, target_name,
            ip_address, user_agent, execution_time_ms, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
    `
    _, err := db.Exec(query,
        log.ID, log.OperatorID, log.OperatorName, log.OperatorRole,
        log.OperationType, log.OperationModule, log.OperationDesc,
        log.TargetID, log.TargetName, log.IPAddress, log.UserAgent,
        log.ExecutionTimeMs, log.CreatedAt,
    )
    if err != nil {
        println("Failed to save operation log:", err.Error())
    } else {
        println("Operation log saved successfully!")
    }
}

// SaveOperationLog 公共方法：供其他 handler 调用记录操作日志
func (h *OperationLogHandler) SaveOperationLog(c *gin.Context, operationType, operationModule, operationDesc, targetID, targetName string) {
	operatorID := c.GetString("userID")
	username := c.GetString("username")
	operatorName := resolveOperatorDisplayName(h.db, operatorID, username)
	operatorRole := c.GetString("role")

	if operatorID == "" && operatorName == "" {
		return
	}

	log := &models.OperationLog{
		ID:              uuid.New().String(),
		OperatorID:      operatorID,
		OperatorName:    operatorName,
		OperatorRole:    operatorRole,
		OperationType:   operationType,
		OperationModule: operationModule,
		OperationDesc:   operationDesc,
		TargetID:        targetID,
		TargetName:      targetName,
		IPAddress:       c.ClientIP(),
		UserAgent:       c.Request.UserAgent(),
		ExecutionTimeMs: 0, // 手动调用时无法计算耗时，设为0
		CreatedAt:       time.Now(),
	}

	go saveOperationLog(h.db, log)
}
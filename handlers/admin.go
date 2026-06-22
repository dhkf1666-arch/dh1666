// backend/handlers/admin.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"enterprise-agent/backend/models"
)

type AdminHandler struct {
	db *sql.DB
}

// AccountResult 账号统计结果
type AccountResult struct {
    AccountName string
    OrderCount  int
    AvgSeconds  int
}

func NewAdminHandler(db *sql.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

// ==================== 考勤管理 ====================

// GetEmployees 获取员工列表
func (h *AdminHandler) GetEmployees(c *gin.Context) {
	skip, _ := strconv.Atoi(c.DefaultQuery("skip", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	search := c.Query("search")

	where := "1=1"
	args := []interface{}{}
	argIdx := 1

	if search != "" {
		where += fmt.Sprintf(" AND name ILIKE $%d", argIdx)
		args = append(args, "%"+search+"%")
		argIdx++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM employees WHERE %s", where)
	var total int
	if err := h.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}

	query := fmt.Sprintf(`
		SELECT id, employee_id, name, position, hire_date, work_location, created_at, updated_at
		FROM employees WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	args = append(args, limit, skip)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}
	defer rows.Close()

	var employees []map[string]interface{}
	for rows.Next() {
		var id, employeeID, name, position, workLocation string
		var hireDate, createdAt, updatedAt sql.NullTime
		if err := rows.Scan(&id, &employeeID, &name, &position, &hireDate, &workLocation, &createdAt, &updatedAt); err != nil {
			continue
		}
		emp := map[string]interface{}{
			"id":            id,
			"employee_id":   employeeID,
			"name":          name,
			"position":      position,
			"hire_date":     nullTimeToString(hireDate),
			"work_location": workLocation,
			"created_at":    createdAt.Time,
			"updated_at":    updatedAt.Time,
		}
		employees = append(employees, emp)
	}

	models.Send(c, 200, 0, "success", gin.H{
		"items": employees,
		"total": total,
	})
}

// CreateEmployee 创建员工
func (h *AdminHandler) CreateEmployee(c *gin.Context) {
    var req struct {
        EmployeeID   string `json:"employee_id"`
        Name         string `json:"name"`
        Position     string `json:"position"`
        HireDate     string `json:"hire_date"`
        WorkLocation string `json:"work_location"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        models.SendError(c, 400, 1001, err.Error())
        return
    }

    id := uuid.New().String()
    var hireDate interface{}
    if req.HireDate != "" {
        hireDate = req.HireDate
    }

    tx, err := h.db.Begin()
    if err != nil {
        models.SendError(c, 500, 5000, err.Error())
        return
    }
    defer tx.Rollback()

    // 1. 插入员工
    _, err = tx.Exec(`
        INSERT INTO employees (id, employee_id, name, position, hire_date, work_location, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, NOW())
    `, id, req.EmployeeID, req.Name, req.Position, hireDate, req.WorkLocation)
    if err != nil {
        models.SendError(c, 500, 5000, err.Error())
        return
    }

    // 2. ✅ 自动初始化当前月份的绩效记录
    currentMonth := time.Now().Format("2006-01")
    _, err = tx.Exec(`
        INSERT INTO performance_records (employee_id, employee_name, position, score_records, total_score, grade, month, created_at, updated_at)
        VALUES ($1, $2, $3, '[]'::jsonb, 10, '合格', $4, NOW(), NOW())
        ON CONFLICT (employee_id, month) DO NOTHING
    `, id, req.Name, req.Position, currentMonth)
    if err != nil {
        // 绩效记录初始化失败不影响员工创建，只记录日志
        log.Printf("初始化绩效记录失败: %v", err)
    }

    if err := tx.Commit(); err != nil {
        models.SendError(c, 500, 5000, err.Error())
        return
    }

    Log(c, "CREATE", "EMPLOYEE",
        fmt.Sprintf("新增员工：%s（工号 %s，岗位 %s）", req.Name, req.EmployeeID, req.Position),
        id, req.Name)

    models.Send(c, 200, 0, "Employee created", gin.H{"id": id})
}

// UpdateEmployee 更新员工
func (h *AdminHandler) UpdateEmployee(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name         string `json:"name"`
		EmployeeID   string `json:"employee_id"`
		Position     string `json:"position"`
		HireDate     string `json:"hire_date"`
		WorkLocation string `json:"work_location"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		models.SendError(c, 400, 1001, err.Error())
		return
	}

	var hireDate interface{}
	if req.HireDate != "" {
		hireDate = req.HireDate
	}

	_, err := h.db.Exec(`
		UPDATE employees SET name=$1, employee_id=$2, position=$3, hire_date=$4, work_location=$5, updated_at=NOW()
		WHERE id=$6
	`, req.Name, req.EmployeeID, req.Position, hireDate, req.WorkLocation, id)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}

	Log(c, "UPDATE", "EMPLOYEE",
		fmt.Sprintf("修改员工：%s（工号 %s，岗位 %s）", req.Name, req.EmployeeID, req.Position),
		id, req.Name)

	models.Send(c, 200, 0, "Employee updated", nil)
}

// DeleteEmployee 删除员工
func (h *AdminHandler) DeleteEmployee(c *gin.Context) {
	id := c.Param("id")
	var employeeName string
	_ = h.db.QueryRow(`SELECT name FROM employees WHERE id=$1`, id).Scan(&employeeName)

	_, err := h.db.Exec("DELETE FROM employees WHERE id=$1", id)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}

	if employeeName != "" {
		Log(c, "DELETE", "EMPLOYEE",
			fmt.Sprintf("删除员工：%s", employeeName), id, employeeName)
	}

	models.Send(c, 200, 0, "Employee deleted", nil)
}

// GetAttendanceRecords 批量获取考勤记录
func (h *AdminHandler) GetAttendanceRecords(c *gin.Context) {
    yearMonth := c.Query("year_month")
    
    employeeIDs := c.QueryArray("employee_ids[]")
    if len(employeeIDs) == 0 {
        employeeIDs = c.QueryArray("employee_ids")
    }
    if len(employeeIDs) == 0 {
        if ids := c.Query("employee_ids"); ids != "" {
            employeeIDs = strings.Split(ids, ",")
        }
    }

    if yearMonth == "" || len(employeeIDs) == 0 {
        models.Send(c, 200, 0, "success", gin.H{})
        return
    }

    // ✅ 验证并过滤有效的 UUID
    validUUIDs := make([]string, 0)
    for _, id := range employeeIDs {
        if _, err := uuid.Parse(id); err == nil {
            validUUIDs = append(validUUIDs, id)
        }
    }

    if len(validUUIDs) == 0 {
        models.Send(c, 200, 0, "success", gin.H{})
        return
    }

    // ✅ 构建占位符，将字符串 UUID 转换为 UUID 类型
    placeholders := make([]string, len(validUUIDs))
    args := []interface{}{yearMonth}
    for i, id := range validUUIDs {
        placeholders[i] = fmt.Sprintf("$%d::uuid", i+2)  // ✅ 关键：::uuid 转换
        args = append(args, id)
    }

    query := fmt.Sprintf(`
        SELECT employee_id, date, status
        FROM attendance_records
        WHERE year_month = $1 AND employee_id IN (%s)
    `, strings.Join(placeholders, ","))

    rows, err := h.db.Query(query, args...)
    if err != nil {
        log.Printf("SQL错误: %v, query: %s", err, query)
        models.SendError(c, 500, 5000, err.Error())
        return
    }
    defer rows.Close()

    // 构建返回结果
    result := make(map[string]map[string]interface{})
    for _, id := range employeeIDs {
        result[id] = make(map[string]interface{})
    }

    for rows.Next() {
        var employeeUUID, date, status string
        if err := rows.Scan(&employeeUUID, &date, &status); err != nil {
            continue
        }
        // 匹配原始输入的 ID
        for _, inputID := range employeeIDs {
            if inputID == employeeUUID {
                if result[inputID] == nil {
                    result[inputID] = make(map[string]interface{})
                }
                result[inputID][date] = map[string]string{"status": status}
                break
            }
        }
    }

    models.Send(c, 200, 0, "success", result)
}

// SaveAttendanceRecords 保存考勤记录
func (h *AdminHandler) SaveAttendanceRecords(c *gin.Context) {
    var req struct {
        YearMonth string                 `json:"year_month"`
        Data      map[string]interface{} `json:"data"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        models.SendError(c, 400, 1001, err.Error())
        return
    }

    tx, err := h.db.Begin()
    if err != nil {
        models.SendError(c, 500, 5000, err.Error())
        return
    }
    defer tx.Rollback()

    for empID, records := range req.Data {
        // ✅ 添加验证：确保 empID 是有效的员工 UUID
        var exists bool
        err := tx.QueryRow(`
            SELECT EXISTS(SELECT 1 FROM employees WHERE id = $1)
        `, empID).Scan(&exists)
        if err != nil || !exists {
            // 如果不是有效的 UUID，尝试通过 employee_id 查找
            var realUUID string
            err = tx.QueryRow(`
                SELECT id FROM employees WHERE employee_id = $1
            `, empID).Scan(&realUUID)
            if err != nil {
                // 如果还是找不到，跳过这条记录
                continue
            }
            empID = realUUID // 使用找到的 UUID
        }

        recordsMap, ok := records.(map[string]interface{})
        if !ok {
            continue
        }
        for date, record := range recordsMap {
            recordMap, ok := record.(map[string]interface{})
            if !ok {
                continue
            }
            status, _ := recordMap["status"].(string)
            if status == "" {
                continue
            }
            _, err = tx.Exec(`
                INSERT INTO attendance_records (employee_id, year_month, date, status, updated_at)
                VALUES ($1, $2, $3, $4, NOW())
                ON CONFLICT (employee_id, year_month, date) DO UPDATE SET status=$4, updated_at=NOW()
            `, empID, req.YearMonth, date, status)
            if err != nil {
                models.SendError(c, 500, 5000, err.Error())
                return
            }
        }
    }

    if err := tx.Commit(); err != nil {
        models.SendError(c, 500, 5000, err.Error())
        return
    }

    // ✅ 记录操作日志
	Log(c, "UPDATE", "ATTENDANCE",
		fmt.Sprintf("保存 %s 月考勤记录", req.YearMonth), "", req.YearMonth)

    models.Send(c, 200, 0, "Attendance saved", nil)
}

// ==================== 绩效管理 ====================

// GetPerformance 获取绩效数据
func (h *AdminHandler) GetPerformance(c *gin.Context) {
	month := c.Query("month")
	employeeID := c.Query("employee_id")

	var where string
	var args []interface{}
	argIdx := 1

	if month != "" {
		where += fmt.Sprintf(" AND month = $%d", argIdx)
		args = append(args, month)
		argIdx++
	}
	if employeeID != "" {
		where += fmt.Sprintf(" AND employee_id = $%d", argIdx)
		args = append(args, employeeID)
		argIdx++
	}

	query := fmt.Sprintf(`
		SELECT employee_id, employee_name, position, score_records, total_score, grade, month
		FROM performance_records WHERE 1=1 %s
	`, where)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var employeeID, employeeName, position, month string
		var scoreRecordsJSON string
		var totalScore int
		var grade string
		if err := rows.Scan(&employeeID, &employeeName, &position, &scoreRecordsJSON, &totalScore, &grade, &month); err != nil {
			continue
		}
		var scoreRecords []interface{}
		json.Unmarshal([]byte(scoreRecordsJSON), &scoreRecords)
		items = append(items, map[string]interface{}{
			"employee_id":   employeeID,
			"employee_name": employeeName,
			"position":      position,
			"score_records": scoreRecords,
			"total_score":   totalScore,
			"grade":         grade,
			"month":         month,
		})
	}

	models.Send(c, 200, 0, "success", gin.H{"items": items})
}

// BatchSavePerformance 批量保存绩效
func (h *AdminHandler) BatchSavePerformance(c *gin.Context) {
	var req struct {
		Month string                   `json:"month"`
		Items []map[string]interface{} `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		models.SendError(c, 400, 1001, err.Error())
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}
	defer tx.Rollback()

	for _, item := range req.Items {
		employeeID, _ := item["employee_id"].(string)
		employeeName, _ := item["employee_name"].(string)
		position, _ := item["position"].(string)
		totalScore, _ := item["total_score"].(float64)
		grade, _ := item["grade"].(string)
		scoreRecords, _ := json.Marshal(item["score_records"])

		_, err = tx.Exec(`
			INSERT INTO performance_records (employee_id, employee_name, position, score_records, total_score, grade, month, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
			ON CONFLICT (employee_id, month) DO UPDATE SET
				employee_name=$2, position=$3, score_records=$4, total_score=$5, grade=$6, updated_at=NOW()
		`, employeeID, employeeName, position, string(scoreRecords), int(totalScore), grade, req.Month)
		if err != nil {
			models.SendError(c, 500, 5000, err.Error())
			return
		}
	}

	if err := tx.Commit(); err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}

	Log(c, "UPDATE", "PERFORMANCE",
		fmt.Sprintf("批量保存 %s 月绩效考核（共 %d 人）", req.Month, len(req.Items)),
		"", req.Month)

	models.Send(c, 200, 0, "Performance saved", nil)
}

// ==================== 罚款管理 ====================

// GetPenaltyRecords 获取罚款记录
func (h *AdminHandler) GetPenaltyRecords(c *gin.Context) {
	month := c.Query("month")
	employeeID := c.Query("employee_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	offset := (page - 1) * pageSize
	where := "1=1"
	args := []interface{}{}
	argIdx := 1

	if month != "" {
		where += fmt.Sprintf(" AND DATE_TRUNC('month', penalty_date) = DATE_TRUNC('month', $%d::date)", argIdx)
		args = append(args, month+"-01")
		argIdx++
	}
	if employeeID != "" {
		where += fmt.Sprintf(" AND employee_id = $%d", argIdx)
		args = append(args, employeeID)
		argIdx++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM penalty_records WHERE %s", where)
	var total int
	h.db.QueryRow(countQuery, args...).Scan(&total)

	query := fmt.Sprintf(`
		SELECT pr.id, pr.employee_id, pr.employee_name, pr.position, pr.amount, pr.category, pr.reason, pr.penalty_date,
			pr.created_by,
			COALESCE(NULLIF(TRIM(u.real_name), ''), pr.created_by) AS created_by_name,
			pr.created_at
		FROM penalty_records pr
		LEFT JOIN users u ON u.username = pr.created_by
		WHERE %s
		ORDER BY pr.penalty_date DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var id, employeeID, employeeName, position, category, reason, createdBy, createdByName string
		var amount float64
		var penaltyDate, createdAt time.Time
		rows.Scan(&id, &employeeID, &employeeName, &position, &amount, &category, &reason, &penaltyDate, &createdBy, &createdByName, &createdAt)
		items = append(items, map[string]interface{}{
			"id":              id,
			"employee_id":     employeeID,
			"employee_name":   employeeName,
			"position":        position,
			"amount":          amount,
			"category":        category,
			"reason":          reason,
			"penalty_date":    penaltyDate.Format("2006-01-02"),
			"created_by":      createdBy,
			"created_by_name": createdByName,
		})
	}

	// 统计汇总
	var totalAmount float64
	var employeeCount int
	h.db.QueryRow(fmt.Sprintf(`
		SELECT COALESCE(SUM(amount), 0), COUNT(DISTINCT employee_id)
		FROM penalty_records WHERE %s
	`, where), args[:len(args)-2]...).Scan(&totalAmount, &employeeCount)

	models.Send(c, 200, 0, "success", gin.H{
		"items": items,
		"total": total,
		"stats": gin.H{
			"total_amount":   totalAmount,
			"employee_count": employeeCount,
			"record_count":   total,
			"avg_amount": func() float64 {
				if employeeCount > 0 {
					return totalAmount / float64(employeeCount)
				}
				return 0
			}(),
		},
	})
}

// CreatePenaltyRecord 创建罚款记录
func (h *AdminHandler) CreatePenaltyRecord(c *gin.Context) {
	var req struct {
		EmployeeID  string  `json:"employee_id"`
		PenaltyDate string  `json:"penalty_date"`
		Amount      float64 `json:"amount"`
		Category    string  `json:"category"`
		Reason      string  `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		models.SendError(c, 400, 1001, err.Error())
		return
	}

	username := c.GetString("username")
	id := uuid.New().String()

	_, err := h.db.Exec(`
		INSERT INTO penalty_records (id, employee_id, employee_name, position, amount, category, reason, penalty_date, created_by, created_at)
		SELECT $1, $2, e.name, e.position, $3, $4, $5, $6, $7, NOW()
		FROM employees e WHERE e.id = $2
	`, id, req.EmployeeID, req.Amount, req.Category, req.Reason, req.PenaltyDate, username)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}

	var employeeName string
	h.db.QueryRow(`SELECT name FROM employees WHERE id = $1`, req.EmployeeID).Scan(&employeeName)
	Log(c, "CREATE", "PENALTY",
		fmt.Sprintf("添加罚款：员工 %s，金额 ¥%.2f，类型 %s", employeeName, req.Amount, req.Category),
		id, employeeName)

	models.Send(c, 200, 0, "Penalty record created", nil)
}

// DeletePenaltyRecord 删除罚款记录
func (h *AdminHandler) DeletePenaltyRecord(c *gin.Context) {
	id := c.Param("id")
    
    // ✅ 获取删除的记录信息用于日志
    var employeeName string
    var amount float64
    h.db.QueryRow(`SELECT employee_name, amount FROM penalty_records WHERE id = $1`, id).Scan(&employeeName, &amount)
    
	_, err := h.db.Exec("DELETE FROM penalty_records WHERE id=$1", id)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}
    
    // ✅ 修复: 使用全局 Log 函数
    Log(c, "DELETE", "PENALTY",
        fmt.Sprintf("删除罚款：员工 %s，金额 ¥%.2f", employeeName, amount),
        id, employeeName)
        
	models.Send(c, 200, 0, "Penalty record deleted", nil)
}

// ==================== 出款站点管理 ====================
// GetSites 获取站点列表
func (h *AdminHandler) GetSites(c *gin.Context) {
    isActive := c.Query("is_active")
    where := "1=1"
    args := []interface{}{}
    if isActive != "" {
        where += " AND is_active = $1"
        args = append(args, isActive == "true")
    }
    
    query := fmt.Sprintf(`
        SELECT id, code, name, sort_order, is_active, created_at, updated_at,
            COALESCE((SELECT COUNT(*) FROM employee_accounts WHERE site_id = sites.id), 0) as account_count,
            COALESCE((SELECT COUNT(*) FROM site_stats WHERE site_id = sites.id), 0) as data_count
        FROM sites WHERE %s ORDER BY sort_order ASC
    `, where)
    
    rows, err := h.db.Query(query, args...)
    if err != nil {
        models.SendError(c, 500, 5000, err.Error())
        return
    }
    defer rows.Close()

    var items []map[string]interface{}
    for rows.Next() {
        var id, code, name string
        var sortOrder int
        var isActive bool
        var createdAt, updatedAt sql.NullTime  // ✅ 使用 sql.NullTime
        var accountCount, dataCount int
        
        err := rows.Scan(&id, &code, &name, &sortOrder, &isActive, &createdAt, &updatedAt, &accountCount, &dataCount)
        if err != nil {
            log.Printf("Scan error in GetSites: %v", err)
            continue
        }
        
        log.Printf("Site: id=%s, code=%s, name=%s, accountCount=%d, isActive=%v", 
            id, code, name, accountCount, isActive)
        
        items = append(items, map[string]interface{}{
            "id":            id,
            "code":          code,
            "name":          name,
            "sort_order":    sortOrder,
            "is_active":     isActive,
            "account_count": accountCount,
            "data_count":    dataCount,
        })
    }
    
    models.Send(c, 200, 0, "success", gin.H{"items": items})
}

// CreateSite 创建站点
func (h *AdminHandler) CreateSite(c *gin.Context) {
	var req struct {
		Code      string `json:"code"`
		Name      string `json:"name"`
		SortOrder int    `json:"sort_order"`
		IsActive  bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		models.SendError(c, 400, 1001, err.Error())
		return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO sites (id, code, name, sort_order, is_active, created_at) VALUES ($1, $2, $3, $4, $5, NOW())`,
		id, req.Code, req.Name, req.SortOrder, req.IsActive)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}
	Log(c, "CREATE", "SITE_STATS",
		fmt.Sprintf("新增出款站点：%s（编码 %s）", req.Name, req.Code), id, req.Name)
	models.Send(c, 200, 0, "Site created", nil)
}

// UpdateSite 更新站点
func (h *AdminHandler) UpdateSite(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name      string `json:"name"`
		SortOrder int    `json:"sort_order"`
		IsActive  bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		models.SendError(c, 400, 1001, err.Error())
		return
	}
	var siteCode string
	_ = h.db.QueryRow(`SELECT code FROM sites WHERE id=$1`, id).Scan(&siteCode)

	_, err := h.db.Exec(`UPDATE sites SET name=$1, sort_order=$2, is_active=$3, updated_at=NOW() WHERE id=$4`,
		req.Name, req.SortOrder, req.IsActive, id)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}
	Log(c, "UPDATE", "SITE_STATS",
		fmt.Sprintf("修改出款站点：%s（编码 %s）", req.Name, siteCode), id, req.Name)
	models.Send(c, 200, 0, "Site updated", nil)
}

// DeleteSite 删除站点
func (h *AdminHandler) DeleteSite(c *gin.Context) {
	id := c.Param("id")
	var siteName, siteCode string
	_ = h.db.QueryRow(`SELECT name, code FROM sites WHERE id=$1`, id).Scan(&siteName, &siteCode)

	_, err := h.db.Exec("DELETE FROM sites WHERE id=$1", id)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}
	if siteName != "" {
		Log(c, "DELETE", "SITE_STATS",
			fmt.Sprintf("删除出款站点：%s（编码 %s）", siteName, siteCode), id, siteName)
	}
	models.Send(c, 200, 0, "Site deleted", nil)
}

// GetEmployeeAccounts 获取员工账号列表
func (h *AdminHandler) GetEmployeeAccounts(c *gin.Context) {
	siteID := c.Query("site_id")
	accountName := c.Query("account_name")
	shift := c.Query("shift")
	skip, _ := strconv.Atoi(c.DefaultQuery("skip", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	where := "1=1"
	args := []interface{}{}
	argIdx := 1
	if siteID != "" {
		where += fmt.Sprintf(" AND site_id = $%d", argIdx)
		args = append(args, siteID)
		argIdx++
	}
	if accountName != "" {
		where += fmt.Sprintf(" AND account_name ILIKE $%d", argIdx)
		args = append(args, "%"+accountName+"%")
		argIdx++
	}
	if shift != "" {
		where += fmt.Sprintf(" AND shift = $%d", argIdx)
		args = append(args, shift)
		argIdx++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM employee_accounts WHERE %s", where)
	var total int
	h.db.QueryRow(countQuery, args...).Scan(&total)

	query := fmt.Sprintf(`
		SELECT ea.id, ea.site_id, s.code as site_code, ea.name, ea.account_name, ea.shift, ea.is_active, ea.created_at,
			(SELECT COUNT(*) FROM site_stats WHERE employee_account_id = ea.id) as data_count
		FROM employee_accounts ea
		LEFT JOIN sites s ON ea.site_id = s.id
		WHERE %s
		ORDER BY ea.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	args = append(args, limit, skip)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var id, siteID, siteCode, name, accountName, shift string
		var isActive bool
		var createdAt time.Time
		var dataCount int
		rows.Scan(&id, &siteID, &siteCode, &name, &accountName, &shift, &isActive, &createdAt, &dataCount)
		items = append(items, map[string]interface{}{
			"id":           id,
			"site_id":      siteID,
			"site_code":    siteCode,
			"name":         name,
			"account_name": accountName,
			"shift":        shift,
			"is_active":    isActive,
			"data_count":   dataCount,
		})
	}
	models.Send(c, 200, 0, "success", gin.H{"items": items, "total": total})
}

// CreateEmployeeAccount 创建员工账号
func (h *AdminHandler) CreateEmployeeAccount(c *gin.Context) {
	var req struct {
		SiteID      string `json:"site_id"`
		Name        string `json:"name"`
		AccountName string `json:"account_name"`
		Shift       string `json:"shift"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		models.SendError(c, 400, 1001, err.Error())
		return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`
		INSERT INTO employee_accounts (id, site_id, name, account_name, shift, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, true, NOW())
	`, id, req.SiteID, req.Name, req.AccountName, req.Shift)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}
	Log(c, "CREATE", "SITE_STATS",
		fmt.Sprintf("新增出款员工账号：%s（账号 %s，%s班）", req.Name, req.AccountName, req.Shift),
		id, req.Name)
	models.Send(c, 200, 0, "Employee account created", nil)
}

// UpdateEmployeeAccount 更新员工账号
func (h *AdminHandler) UpdateEmployeeAccount(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name        string `json:"name"`
		AccountName string `json:"account_name"`
		Shift       string `json:"shift"`
		IsActive    bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		models.SendError(c, 400, 1001, err.Error())
		return
	}
	_, err := h.db.Exec(`
		UPDATE employee_accounts SET name=$1, account_name=$2, shift=$3, is_active=$4, updated_at=NOW()
		WHERE id=$5
	`, req.Name, req.AccountName, req.Shift, req.IsActive, id)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}
	Log(c, "UPDATE", "SITE_STATS",
		fmt.Sprintf("修改出款员工账号：%s（账号 %s，%s班）", req.Name, req.AccountName, req.Shift),
		id, req.Name)
	models.Send(c, 200, 0, "Employee account updated", nil)
}

// DeleteEmployeeAccount 删除员工账号
func (h *AdminHandler) DeleteEmployeeAccount(c *gin.Context) {
	id := c.Param("id")
	var accountName, displayName string
	_ = h.db.QueryRow(`SELECT account_name, COALESCE(NULLIF(name, ''), account_name) FROM employee_accounts WHERE id=$1`, id).
		Scan(&accountName, &displayName)

	_, err := h.db.Exec("DELETE FROM employee_accounts WHERE id=$1", id)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}
	if displayName != "" {
		Log(c, "DELETE", "SITE_STATS",
			fmt.Sprintf("删除出款员工账号：%s（账号 %s）", displayName, accountName), id, displayName)
	}
	models.Send(c, 200, 0, "Employee account deleted", nil)
}

// ==================== 出款统计 ====================

// PreviewSiteStatsUpload 预览上传文件表头，供前端手动选择列映射
func (h *AdminHandler) PreviewSiteStatsUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		models.SendError(c, 400, 1001, "请选择文件")
		return
	}

	mode := c.PostForm("mode")
	if mode == "" {
		mode = "BX"
	}

	rows, err := readSiteStatsRowsFromFileHeader(file)
	if err != nil {
		models.SendError(c, 400, 1001, err.Error())
		return
	}
	if len(rows) < 1 {
		models.SendError(c, 400, 1001, "文件无表头数据")
		return
	}

	rows = padSiteStatsRows(rows)
	header := rows[0]
	cols := resolveSiteStatsColumns(header, mode)

	models.Send(c, 200, 0, "success", gin.H{
		"columns":        buildSiteStatsColumnOptions(header),
		"total_columns":  len(header),
		"suggested": gin.H{
			"account_col": cols.AccountCol + 1,
			"start_col":   cols.StartCol + 1,
			"end_col":     cols.EndCol + 1,
			"source":      cols.Source,
		},
	})
}

func (h *AdminHandler) UploadSiteStats(c *gin.Context) {
    file, err := c.FormFile("file")
    if err != nil {
        models.SendError(c, 400, 1001, "请选择文件")
        return
    }
    siteID := c.PostForm("site_id")
    shift := c.PostForm("shift")
    statDate := c.PostForm("date")
    mode := c.PostForm("mode")
    if mode == "" {
        mode = "BX"
    }

    columnOverride, err := parseSiteStatsColumnOverrideFromForm(c)
    if err != nil {
        models.SendError(c, 400, 1001, err.Error())
        return
    }

    rows, err := readSiteStatsRowsFromFileHeader(file)
    if err != nil {
        log.Printf("文件解析失败: %v", err)
        models.SendError(c, 400, 1001, err.Error())
        return
    }
    log.Printf("文件解析成功，共 %d 行", len(rows))

    if len(rows) < 2 {
        models.SendError(c, 400, 1001, "文件无数据，请确保有表头和数据行")
        return
    }

    rows = padSiteStatsRows(rows)
    cols := resolveSiteStatsColumnMapping(rows[0], mode, columnOverride)
    accountNameCol := cols.AccountCol
    startTimeCol := cols.StartCol
    endTimeCol := cols.EndCol

    // 检查列数是否足够
    maxColNeeded := accountNameCol + 1
    if len(rows[0]) < maxColNeeded {
        models.SendError(c, 400, 1001, fmt.Sprintf("文件列数不足，需要至少 %d 列，实际 %d 列", maxColNeeded, len(rows[0])))
        return
    }

    log.Printf("[Upload] Mode: %s, MappingSource: %s, AccountCol: %d, StartCol: %d, EndCol: %d, TotalRows: %d",
        mode, cols.Source, accountNameCol, startTimeCol, endTimeCol, len(rows))
    
    // 打印表头信息（前几列）
    if len(rows[0]) > 0 {
        headerPreview := rows[0]
        if len(headerPreview) > maxColNeeded {
            headerPreview = headerPreview[:maxColNeeded]
        }
        log.Printf("表头前%d列: %v", len(headerPreview), headerPreview)
    }

    accountStats := make(map[string]struct {
        Count      int
        StartTimes []time.Time
        EndTimes   []time.Time
    })

    // 解析数据（跳过表头）
    for i, row := range rows[1:] {
        rowIndex := i + 2
        
        // 检查列数是否足够
        if len(row) <= accountNameCol {
            log.Printf("行 %d: 列数不足，需要至少 %d 列，实际 %d 列，跳过", rowIndex, accountNameCol+1, len(row))
            continue
        }
        
        // 获取账号
        accountName := strings.TrimSpace(getColumn(row, accountNameCol))
        accountName = strings.Trim(accountName, "\r\n\t ")
        if accountName == "" {
            log.Printf("行 %d: 账号为空，跳过", rowIndex)
            continue
        }
        if isLikelyHeaderAccountValue(accountName) {
            log.Printf("行 %d: 疑似表头行，跳过 [%s]", rowIndex, accountName)
            continue
        }

        cleanAccountName := normalizeAccountNameForMatch(accountName)
        if cleanAccountName == "" {
            log.Printf("行 %d: 账号规范化后为空，跳过", rowIndex)
            continue
        }

        if i < 5 {
            log.Printf("行 %d: 账号原始=[%s] 规范化=[%s]", rowIndex, accountName, cleanAccountName)
        }

        // 获取时间
        startTimeStr := ""
        if len(row) > startTimeCol {
            startTimeStr = strings.TrimSpace(getColumn(row, startTimeCol))
        }
        endTimeStr := ""
        if len(row) > endTimeCol {
            endTimeStr = strings.TrimSpace(getColumn(row, endTimeCol))
        }

        stats := accountStats[cleanAccountName]
        stats.Count++

        // 解析开始时间
        if startTimeStr != "" {
            if t, err := parseChineseTime(startTimeStr); err == nil {
                stats.StartTimes = append(stats.StartTimes, t)
            } else {
                log.Printf("行 %d: 开始时间解析失败: %s, error: %v", rowIndex, startTimeStr, err)
            }
        }

        // 解析结束时间
        if endTimeStr != "" {
            if t, err := parseChineseTime(endTimeStr); err == nil {
                stats.EndTimes = append(stats.EndTimes, t)
            } else {
                log.Printf("行 %d: 结束时间解析失败: %s, error: %v", rowIndex, endTimeStr, err)
            }
        }

        accountStats[cleanAccountName] = stats
    }

    if len(accountStats) == 0 {
        models.SendError(c, 400, 1001, "未找到有效的账号数据，请检查文件格式")
        return
    }

    // 计算并保存结果
    type AccountResult struct {
        AccountName string
        OrderCount  int
        AvgSeconds  int
    }
    
    var results []AccountResult
    for accountName, stats := range accountStats {
        var totalDuration float64
        var validCount int
        
        minLen := len(stats.StartTimes)
        if len(stats.EndTimes) < minLen {
            minLen = len(stats.EndTimes)
        }
        
        for j := 0; j < minLen; j++ {
            duration := stats.EndTimes[j].Sub(stats.StartTimes[j]).Seconds()
            if duration > 0 && duration < 3600 { // 小于1小时
                totalDuration += duration
                validCount++
            }
        }
        
        avgSeconds := 0
        if validCount > 0 {
            avgSeconds = int(totalDuration / float64(validCount))
        }
        
        results = append(results, AccountResult{
            AccountName: accountName,
            OrderCount:  stats.Count,
            AvgSeconds:  avgSeconds,
        })
    }

    // 保存到数据库
    tx, err := h.db.BeginTx(c.Request.Context(), nil)
    if err != nil {
        models.SendError(c, 500, 5000, err.Error())
        return
    }
    defer tx.Rollback()

    matchedCount := 0
    var unmatchedAccounts []string
    for _, res := range results {
        var employeeAccountID string

        cleanAccountName := normalizeAccountNameForMatch(res.AccountName)
        if cleanAccountName == "" {
            unmatchedAccounts = append(unmatchedAccounts, res.AccountName)
            continue
        }

        err := tx.QueryRow(`
            SELECT id FROM employee_accounts 
            WHERE site_id = $1 
              AND LOWER(REGEXP_REPLACE(TRIM(account_name), '[[:space:]]+', '', 'g')) = $2
              AND shift = $3
        `, siteID, cleanAccountName, shift).Scan(&employeeAccountID)

        if err != nil {
            log.Printf("账号 [%s] 未匹配到员工（规范化后: [%s]）", res.AccountName, cleanAccountName)
            unmatchedAccounts = append(unmatchedAccounts, res.AccountName)
            continue
        }

        _, err = tx.Exec(`
            INSERT INTO site_stats (employee_account_id, site_id, stat_date, shift, order_count, avg_time_seconds, created_at)
            VALUES ($1, $2, $3, $4, $5, $6, NOW())
            ON CONFLICT (employee_account_id, stat_date, shift) DO UPDATE SET
                order_count = EXCLUDED.order_count,
                avg_time_seconds = EXCLUDED.avg_time_seconds,
                updated_at = NOW()
        `, employeeAccountID, siteID, statDate, shift, res.OrderCount, res.AvgSeconds)
        if err != nil {
            models.SendError(c, 500, 5000, err.Error())
            return
        }
        matchedCount++
    }
    if err := tx.Commit(); err != nil {
        models.SendError(c, 500, 5000, err.Error())
        return
    }

    var siteName string
    _ = h.db.QueryRow(`SELECT name FROM sites WHERE id = $1`, siteID).Scan(&siteName)
    shiftLabel := map[string]string{"day": "A班", "night": "B班"}[shift]
    if shiftLabel == "" {
        shiftLabel = shift
    }
    Log(c, "UPLOAD", "SITE_STATS",
        fmt.Sprintf("上传出款统计：站点 %s，日期 %s，%s，成功匹配 %d 人", siteName, statDate, shiftLabel, matchedCount),
        siteID, siteName)

    models.Send(c, 200, 0, "Upload successful", gin.H{
        "matched_count":      matchedCount,
        "unmatched_count":    len(accountStats) - matchedCount,
        "unmatched_accounts": unmatchedAccounts,
        "column_mapping":     cols.toResponse(),
        "stats": gin.H{
            "total_records":  len(rows) - 1,
            "total_accounts": len(accountStats),
            "matched":        matchedCount,
            "unmatched":      len(accountStats) - matchedCount,
        },
    })
}

// ClearSiteStatsByDate 清除指定日期的站点统计数据
func (h *AdminHandler) ClearSiteStatsByDate(c *gin.Context) {
	siteID := c.Query("site_id")
	shift := c.Query("shift")
	date := c.Query("date")

	if siteID == "" || date == "" {
		models.SendError(c, 400, 1001, "site_id 和 date 不能为空")
		return
	}

	query := "DELETE FROM site_stats WHERE site_id = $1 AND stat_date = $2"
	args := []interface{}{siteID, date}

	if shift != "" {
		query += " AND shift = $3"
		args = append(args, shift)
	}

	result, err := h.db.Exec(query, args...)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}

	rowsAffected, _ := result.RowsAffected()
	var siteName string
	_ = h.db.QueryRow(`SELECT name FROM sites WHERE id=$1`, siteID).Scan(&siteName)
	Log(c, "DELETE", "SITE_STATS",
		fmt.Sprintf("清除出款统计：站点 %s，日期 %s，删除 %d 条", siteName, date, rowsAffected),
		siteID, siteName)
	models.Send(c, 200, 0, "Data cleared", gin.H{"deleted": rowsAffected})
}

// GetSiteStatsSummary 获取站点统计汇总
func (h *AdminHandler) GetSiteStatsSummary(c *gin.Context) {
	siteID := c.Query("site_id")
	employeeAccountID := c.Query("employee_account_id")
	employeeName := c.Query("employee_name")  // ✅ 新增：员工姓名搜索
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	shift := c.Query("shift")

	where := "1=1"
	args := []interface{}{}
	argIdx := 1
	
	if siteID != "" {
		where += fmt.Sprintf(" AND ss.site_id = $%d", argIdx)
		args = append(args, siteID)
		argIdx++
	}
	if employeeAccountID != "" {
		where += fmt.Sprintf(" AND ss.employee_account_id = $%d", argIdx)
		args = append(args, employeeAccountID)
		argIdx++
	}
	// ✅ 新增：员工姓名模糊搜索
	if employeeName != "" {
		where += fmt.Sprintf(" AND ea.name ILIKE $%d", argIdx)
		args = append(args, "%"+employeeName+"%")
		argIdx++
	}
	if startDate != "" {
		where += fmt.Sprintf(" AND ss.stat_date >= $%d", argIdx)
		args = append(args, startDate)
		argIdx++
	}
	if endDate != "" {
		where += fmt.Sprintf(" AND ss.stat_date <= $%d", argIdx)
		args = append(args, endDate)
		argIdx++
	}
	if shift != "" {
		where += fmt.Sprintf(" AND ss.shift = $%d", argIdx)
		args = append(args, shift)
		argIdx++
	}

	query := fmt.Sprintf(`
		SELECT 
			ea.name as employee_name,
			ea.account_name,
			ss.site_id,
			s.code as site_code,
            ss.shift, 
			SUM(ss.order_count) as total_value,
			SUM(ss.order_count * ss.avg_time_seconds) as weighted_time,
			SUM(ss.order_count) as total_weight
		FROM site_stats ss
		JOIN employee_accounts ea ON ss.employee_account_id = ea.id
		JOIN sites s ON ss.site_id = s.id
		WHERE %s
        GROUP BY ea.name, ea.account_name, ss.site_id, s.code, ss.shift 
		ORDER BY total_value DESC
	`, where)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}
	defer rows.Close()

	var items []map[string]interface{}
	siteColumns := make(map[string]bool)

	for rows.Next() {
		var employeeName, accountName, siteID, siteCode string
		var totalValue int
		var weightedTime, totalWeight int
		var shiftValue string

		err := rows.Scan(&employeeName, &accountName, &siteID, &siteCode, &shiftValue, &totalValue, &weightedTime, &totalWeight)
		if err != nil {
			continue
		}

		siteColumns[siteCode] = true

		items = append(items, map[string]interface{}{
			"employee_name": employeeName,
			"account_name":  accountName,
			"site_id":       siteID,
			"site_code":     siteCode,
            "shift":         shiftValue,
			"total_value":   totalValue,
			"total_avg_seconds": func() int {
				if totalWeight > 0 {
					return weightedTime / totalWeight
				}
				return 0
			}(),
		})
	}

	// 转换为按员工聚合的格式
	employeeMap := make(map[string]map[string]interface{})
	for _, item := range items {
		empName := item["employee_name"].(string)
		if employeeMap[empName] == nil {
			employeeMap[empName] = map[string]interface{}{
				"employee_name":       empName,
				"account_name":        item["account_name"],
				"sites":               make(map[string]interface{}),
				"total_value":         0,
				"total_weighted_time": 0,
				"total_weight":        0,
			}
		}
		emp := employeeMap[empName]
		siteCode := item["site_code"].(string)
		emp["sites"].(map[string]interface{})[siteCode] = map[string]interface{}{
			"value":            item["total_value"],
			"avg_time_seconds": item["total_avg_seconds"],
			"avg_time_str":     formatSecondsToTime(item["total_avg_seconds"].(int)),
		}
		emp["total_value"] = emp["total_value"].(int) + item["total_value"].(int)
		
		// 累加加权时间和权重用于计算总体平均时间
		if val, ok := item["total_avg_seconds"].(int); ok {
			emp["total_weighted_time"] = emp["total_weighted_time"].(int) + (val * item["total_value"].(int))
			emp["total_weight"] = emp["total_weight"].(int) + item["total_value"].(int)
		}
	}

	var result []map[string]interface{}
	for _, emp := range employeeMap {
		// 计算总体平均时间
		totalWeight := emp["total_weight"].(int)
		if totalWeight > 0 {
			avgSeconds := emp["total_weighted_time"].(int) / totalWeight
			emp["total_avg_seconds"] = avgSeconds
			emp["total_avg_time"] = formatSecondsToTime(avgSeconds)
		} else {
			emp["total_avg_seconds"] = 0
			emp["total_avg_time"] = "-"
		}
		result = append(result, emp)
	}

	// 获取站点列
	var siteCols []string
	for site := range siteColumns {
		siteCols = append(siteCols, site)
	}

	models.Send(c, 200, 0, "success", gin.H{
		"items":        result,
		"site_columns": siteCols,
	})
}

// 辅助函数
func getColumn(row []string, idx int) string {
	if idx < len(row) {
		return row[idx]
	}
	return ""
}

func parseChineseTime(timeStr string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006/1/2 15:04:05",
		"2006-01-02 15:04",
		"2006/1/2 15:04",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, timeStr, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("parse error: %s", timeStr)
}

func formatSecondsToTime(seconds int) string {
	if seconds <= 0 {
		return "0秒"
	}
	minutes := seconds / 60
	secs := seconds % 60
	if minutes > 0 && secs > 0 {
		return fmt.Sprintf("%d分%d秒", minutes, secs)
	}
	if minutes > 0 {
		return fmt.Sprintf("%d分", minutes)
	}
	return fmt.Sprintf("%d秒", secs)
}

func nullTimeToString(t sql.NullTime) string {
	if t.Valid {
		return t.Time.Format("2006-01-02")
	}
	return ""
}

// GetSiteStatsStacked 获取站点统计堆叠数据（按日期+员工+班次分组）
func (h *AdminHandler) GetSiteStatsStacked(c *gin.Context) {
	siteID := c.Query("site_id")
	employeeAccountID := c.Query("employee_account_id")
	employeeName := c.Query("employee_name")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	shift := c.Query("shift")

	where := "1=1"
	args := []interface{}{}
	argIdx := 1

	if siteID != "" {
		where += fmt.Sprintf(" AND ss.site_id = $%d", argIdx)
		args = append(args, siteID)
		argIdx++
	}
	if employeeAccountID != "" {
		where += fmt.Sprintf(" AND ss.employee_account_id = $%d", argIdx)
		args = append(args, employeeAccountID)
		argIdx++
	}
	if employeeName != "" {
		where += fmt.Sprintf(" AND ea.name ILIKE $%d", argIdx)
		args = append(args, "%"+employeeName+"%")
		argIdx++
	}
	if startDate != "" {
		where += fmt.Sprintf(" AND ss.stat_date >= $%d", argIdx)
		args = append(args, startDate)
		argIdx++
	}
	if endDate != "" {
		where += fmt.Sprintf(" AND ss.stat_date <= $%d", argIdx)
		args = append(args, endDate)
		argIdx++
	}
	if shift != "" {
		where += fmt.Sprintf(" AND ss.shift = $%d", argIdx)
		args = append(args, shift)
		argIdx++
	}

	// ✅ 修改：在 SELECT 中添加 ss.shift 字段
	query := fmt.Sprintf(`
        SELECT 
            ea.name as employee_name,
            ea.account_name,
            ss.stat_date as date,
            s.code as site_code,
            ss.shift,                       -- ✅ 新增：班次字段
            ss.order_count as value,
            ss.avg_time_seconds,
            ss.employee_account_id
        FROM site_stats ss
        JOIN employee_accounts ea ON ss.employee_account_id = ea.id
        JOIN sites s ON ss.site_id = s.id
        WHERE %s
        ORDER BY ss.stat_date DESC, ea.name ASC
    `, where)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}
	defer rows.Close()

	var items []map[string]interface{}
	siteColumns := make(map[string]bool)

	for rows.Next() {
		var employeeName, accountName, date, siteCode, shiftValue string // ✅ 新增 shiftValue
		var value, avgTimeSeconds int
		var employeeAccountID string

		// ✅ 修改：扫描 shift 字段
		err := rows.Scan(&employeeName, &accountName, &date, &siteCode, &shiftValue,
			&value, &avgTimeSeconds, &employeeAccountID)
		if err != nil {
			log.Printf("Scan error in GetSiteStatsStacked: %v", err)
			continue
		}

		siteColumns[siteCode] = true

		items = append(items, map[string]interface{}{
			"employee_name":    employeeName,
			"account_name":     accountName,
			"employee_id":      employeeAccountID,
			"date":             date,
			"site_code":        siteCode,
			"shift":            shiftValue, // ✅ 新增：返回班次
			"value":            value,
			"avg_time_seconds": avgTimeSeconds,
			"avg_time_str":     formatSecondsToTime(avgTimeSeconds),
		})
	}

	var siteCols []string
	for site := range siteColumns {
		siteCols = append(siteCols, site)
	}

	models.Send(c, 200, 0, "success", gin.H{
		"items":        items,
		"site_columns": siteCols,
	})
}

// ClearAllSiteStats 清空所有站点统计数据（保留站点与员工配置）
func (h *AdminHandler) ClearAllSiteStats(c *gin.Context) {
	result, err := h.db.Exec("DELETE FROM site_stats")
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}

	rowsAffected, _ := result.RowsAffected()
	Log(c, "DELETE", "SITE_STATS",
		fmt.Sprintf("清空全部出款统计数据，删除 %d 条", rowsAffected), "", "全部站点")
	models.Send(c, 200, 0, "All site stats data cleared", gin.H{"deleted": rowsAffected})
}

// ClearSiteStatsByDateOnly 删除指定日期的所有站点统计数据
func (h *AdminHandler) ClearSiteStatsByDateOnly(c *gin.Context) {
    date := c.Query("date")
    
    if date == "" {
        models.SendError(c, 400, 1001, "date 不能为空")
        return
    }
    
    query := "DELETE FROM site_stats WHERE stat_date = $1"
    
    result, err := h.db.Exec(query, date)
    if err != nil {
        models.SendError(c, 500, 5000, err.Error())
        return
    }
    
    rowsAffected, _ := result.RowsAffected()
    Log(c, "DELETE", "SITE_STATS",
        fmt.Sprintf("清除 %s 全部站点出款统计，删除 %d 条", date, rowsAffected), "", date)
    models.Send(c, 200, 0, "Data cleared", gin.H{"deleted": rowsAffected})
}



// ==================== 绩效管理（增强版：添加修改删除日志） ====================
// UpdatePerformance 更新单条绩效记录
func (h *AdminHandler) UpdatePerformance(c *gin.Context) {
    var req struct {
        EmployeeID  string      `json:"employee_id"`
        Month       string      `json:"month"`
        TotalScore  int         `json:"total_score"`
        Grade       string      `json:"grade"`
        ScoreRecords interface{} `json:"score_records"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        models.SendError(c, 400, 1001, err.Error())
        return
    }

    // 获取旧值用于日志（同时获取员工ID用于后续查询）
    var oldTotalScore int
    var oldGrade string
    var employeeID, employeeName string
    err := h.db.QueryRow(`
        SELECT employee_id, employee_name, total_score, grade 
        FROM performance_records 
        WHERE employee_id = $1 AND month = $2
    `, req.EmployeeID, req.Month).Scan(&employeeID, &employeeName, &oldTotalScore, &oldGrade)
    if err != nil && err != sql.ErrNoRows {
        models.SendError(c, 500, 5000, err.Error())
        return
    }

    // 序列化 score_records
    scoreRecordsJSON, _ := json.Marshal(req.ScoreRecords)

    // 执行更新
    _, err = h.db.Exec(`
        UPDATE performance_records 
        SET total_score = $1, grade = $2, score_records = $3, updated_at = NOW()
        WHERE employee_id = $4 AND month = $5
    `, req.TotalScore, req.Grade, string(scoreRecordsJSON), req.EmployeeID, req.Month)
    if err != nil {
        models.SendError(c, 500, 5000, err.Error())
        return
    }

    // ✅ 修复：更新后重新查询最新的员工姓名
    var latestEmployeeName string
    err = h.db.QueryRow(`
        SELECT e.name 
        FROM performance_records p
        JOIN employees e ON p.employee_id = e.id
        WHERE p.employee_id = $1 AND p.month = $2
    `, req.EmployeeID, req.Month).Scan(&latestEmployeeName)
    if err == nil && latestEmployeeName != "" {
        employeeName = latestEmployeeName
    } else {
        // 降级方案：使用绩效表中存储的姓名
        _ = h.db.QueryRow(`
            SELECT employee_name FROM performance_records 
            WHERE employee_id = $1 AND month = $2
        `, req.EmployeeID, req.Month).Scan(&employeeName)
    }

    month := c.Param("month")
    if month == "" {
        month = req.Month
    }

    // 记录操作日志（现在使用最新数据）
    Log(c, "UPDATE", "PERFORMANCE",
        fmt.Sprintf("修改绩效考核：员工 %s，%s 月，分数 %d→%d，等级 %s→%s",
            employeeName, month, oldTotalScore, req.TotalScore, oldGrade, req.Grade),
        req.EmployeeID, employeeName)

    models.Send(c, 200, 0, "Performance updated", nil)
}

// DeletePerformance 删除绩效记录
func (h *AdminHandler) DeletePerformance(c *gin.Context) {
    employeeID := c.Param("employeeId")
    month := c.Param("month")
    
    if employeeID == "" || month == "" {
        models.SendError(c, 400, 1001, "employeeId and month are required")
        return
    }
    
    // 获取员工姓名用于日志
    var employeeName string
    var oldTotalScore int
    h.db.QueryRow(`
        SELECT employee_name, total_score 
        FROM performance_records 
        WHERE employee_id = $1 AND month = $2
    `, employeeID, month).Scan(&employeeName, &oldTotalScore)
    
    _, err := h.db.Exec(`
        DELETE FROM performance_records 
        WHERE employee_id = $1 AND month = $2
    `, employeeID, month)
    if err != nil {
        models.SendError(c, 500, 5000, err.Error())
        return
    }
    
    // ✅ 记录操作日志
    Log(c, "DELETE", "PERFORMANCE",
        fmt.Sprintf("删除绩效考核：员工 %s，%s 月（原分数 %d）", employeeName, month, oldTotalScore),
        employeeID, employeeName)
    
    models.Send(c, 200, 0, "Performance deleted", nil)
}

// UpdatePenaltyRecord 更新罚款记录
func (h *AdminHandler) UpdatePenaltyRecord(c *gin.Context) {
    id := c.Param("id")
    
    var req struct {
        Amount      float64 `json:"amount"`
        Category    string  `json:"category"`
        Reason      string  `json:"reason"`
        PenaltyDate string  `json:"penalty_date"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        models.SendError(c, 400, 1001, err.Error())
        return
    }
    
    // 获取旧值用于日志
    var oldAmount float64
    var oldCategory, oldReason, employeeName string
    err := h.db.QueryRow(`
        SELECT employee_name, amount, category, reason 
        FROM penalty_records WHERE id = $1
    `, id).Scan(&employeeName, &oldAmount, &oldCategory, &oldReason)
    if err != nil {
        models.SendError(c, 500, 5000, err.Error())
        return
    }
    
    // 执行更新
    _, err = h.db.Exec(`
        UPDATE penalty_records 
        SET amount = $1, category = $2, reason = $3, penalty_date = $4
        WHERE id = $5
    `, req.Amount, req.Category, req.Reason, req.PenaltyDate, id)
    if err != nil {
        models.SendError(c, 500, 5000, err.Error())
        return
    }
    
    // ✅ 修复：更新后重新查询最新的员工姓名
    var latestEmployeeName string
    err = h.db.QueryRow(`
        SELECT employee_name FROM penalty_records WHERE id = $1
    `, id).Scan(&latestEmployeeName)
    if err == nil && latestEmployeeName != "" {
        employeeName = latestEmployeeName  // 使用最新的员工姓名
    }
    
    // 记录操作日志（现在使用最新数据）
    Log(c, "UPDATE", "PENALTY",
        fmt.Sprintf("修改罚款：员工 %s，金额 ¥%.2f→¥%.2f，类型 %s",
            employeeName, oldAmount, req.Amount, req.Category),
        id, employeeName)
    
    models.Send(c, 200, 0, "Penalty record updated", nil)
}

// ==================== 仪表盘：今日休假/请假/旷工人员 ====================

// TodayAbsenteeItem 今日异常考勤人员项
type TodayAbsenteeItem struct {
	EmployeeID   string `json:"employeeId"`
	EmployeeName string `json:"employeeName"`
	Position     string `json:"position"`
	Status       string `json:"status"`
	StatusLabel  string `json:"statusLabel"`
	Date         string `json:"date"`
}

// TodayAbsenteesResponse 今日异常考勤响应
type TodayAbsenteesResponse struct {
	Vacation []TodayAbsenteeItem `json:"vacation"` // 休假（全天+半天）
	Leave    []TodayAbsenteeItem `json:"leave"`    // 请假
	Absent   []TodayAbsenteeItem `json:"absent"`   // 旷工
	Total    int                 `json:"total"`
}

// GetTodayAbsentees 获取今日休假/请假/旷工人员列表
func (h *AdminHandler) GetTodayAbsentees(c *gin.Context) {
	today := time.Now().Format("2006-01-02")
	yearMonth := time.Now().Format("2006-01")

	// 查询今日有考勤记录且状态为休假/请假/旷工的员工
	query := `
		SELECT 
			e.id,
			e.employee_id,
			e.name,
			COALESCE(e.position, '') as position,
			ar.status,
			ar.date
		FROM employees e
		INNER JOIN attendance_records ar ON e.id = ar.employee_id
		WHERE ar.year_month = $1 
			AND ar.date = $2
			AND ar.status IN ('rest_full', 'rest_half', 'leave', 'absent')
		ORDER BY 
			CASE ar.status
				WHEN 'absent' THEN 1
				WHEN 'leave' THEN 2
				WHEN 'rest_full' THEN 3
				WHEN 'rest_half' THEN 4
				ELSE 5
			END,
			e.name
	`

	rows, err := h.db.QueryContext(c.Request.Context(), query, yearMonth, today)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}
	defer rows.Close()

	var vacationList []TodayAbsenteeItem
	var leaveList []TodayAbsenteeItem
	var absentList []TodayAbsenteeItem

	for rows.Next() {
		var id, employeeID, name, position, status, date string
		if err := rows.Scan(&id, &employeeID, &name, &position, &status, &date); err != nil {
			continue
		}

		item := TodayAbsenteeItem{
			EmployeeID:   employeeID,
			EmployeeName: name,
			Position:     position,
			Status:       status,
			StatusLabel:  getStatusLabel(status),
			Date:         date,
		}

		switch status {
		case "rest_full", "rest_half":
			vacationList = append(vacationList, item)
		case "leave":
			leaveList = append(leaveList, item)
		case "absent":
			absentList = append(absentList, item)
		}
	}

	response := TodayAbsenteesResponse{
		Vacation: vacationList,
		Leave:    leaveList,
		Absent:   absentList,
		Total:    len(vacationList) + len(leaveList) + len(absentList),
	}

	models.Send(c, 200, 0, "success", response)
}

// getStatusLabel 获取状态的中文标签
func getStatusLabel(status string) string {
	switch status {
	case "rest_full":
		return "全天休假"
	case "rest_half":
		return "半天休假"
	case "leave":
		return "请假"
	case "absent":
		return "旷工"
	default:
		return status
	}
}
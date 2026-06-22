// backend/handlers/operation_log_helper.go
package handlers

import (
	"database/sql"
	"strings"

	"github.com/gin-gonic/gin"
)

// GlobalLogger 全局操作日志处理器，在 main.go 中初始化
var GlobalLogger *OperationLogHandler

// Log 全局日志记录函数，方便在各个 handler 中调用
func Log(c *gin.Context, operationType, operationModule, operationDesc, targetID, targetName string) {
	if GlobalLogger != nil {
		GlobalLogger.SaveOperationLog(c, operationType, operationModule, operationDesc, targetID, targetName)
	}
}

func resolveOperatorDisplayName(db *sql.DB, operatorID, username string) string {
	var realName sql.NullString

	if operatorID != "" {
		if err := db.QueryRow(`SELECT real_name FROM users WHERE id::text = $1`, operatorID).Scan(&realName); err == nil {
			if realName.Valid && strings.TrimSpace(realName.String) != "" {
				return strings.TrimSpace(realName.String)
			}
		}
	}

	if username != "" {
		if err := db.QueryRow(`SELECT real_name FROM users WHERE username = $1`, username).Scan(&realName); err == nil {
			if realName.Valid && strings.TrimSpace(realName.String) != "" {
				return strings.TrimSpace(realName.String)
			}
		}
		return username
	}

	return ""
}

const operationLogUserJoin = `
LEFT JOIN users u ON (
	(ol.operator_id IS NOT NULL AND ol.operator_id <> '' AND u.id::text = ol.operator_id)
	OR u.username = ol.operator_name
)`

const operationLogOperatorDisplay = `COALESCE(NULLIF(TRIM(u.real_name), ''), ol.operator_name)`
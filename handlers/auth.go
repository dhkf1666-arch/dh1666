package handlers

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"enterprise-agent/backend/models"
)

type AuthHandler struct {
	db    *sql.DB
	redis *redis.Client
}

func NewAuthHandler(db *sql.DB, redis *redis.Client) *AuthHandler {
	return &AuthHandler{db: db, redis: redis}
}

func (h *AuthHandler) Login(c *gin.Context) {
    if h == nil || h.db == nil {
        log.Printf("Login failed: AuthHandler or db is nil")
        models.SendError(c, 500, 5000, "Database not initialized")
        return
    }
	
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Login bind error: %v", err)
		models.SendError(c, 400, 1001, "Invalid request")
		return
	}
	if h.db == nil {
		log.Printf("Login failed: database not initialized")
		models.SendError(c, 500, 5000, "Database not initialized")
		return
	}

	username := strings.TrimSpace(req.Username)
	if username == "" {
		models.SendError(c, 400, 1001, "Invalid request")
		return
	}

	// 查询用户基本信息及角色
	var user models.User
	var passwordHash string
	var roleName string

	err := h.db.QueryRowContext(c.Request.Context(),
		`SELECT u.id, u.username, u.password_hash, u.status,
			COALESCE(
				(SELECT r.name FROM user_roles ur JOIN roles r ON ur.role_id = r.id WHERE ur.user_id = u.id LIMIT 1),
				'auditor'
			) as role_name
		FROM users u WHERE u.username = $1`,
		username,
	).Scan(&user.ID, &user.Username, &passwordHash, &user.Status, &roleName)

	if err != nil {
		log.Printf("Login failed: invalid credentials (username=%s)", username)
		models.SendError(c, 401, 1002, "Invalid credentials")
		return
	}

	if user.Status != 1 {
		log.Printf("Login failed: account disabled (username=%s)", user.Username)
		models.SendError(c, 401, 1002, "Account disabled")
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		log.Printf("Login failed: invalid credentials (username=%s)", user.Username)
		models.SendError(c, 401, 1002, "Invalid credentials")
		return
	}

	_, err = h.db.ExecContext(c.Request.Context(),
		"UPDATE users SET last_login_at = $1 WHERE id = $2",
		models.UTCNow(), user.ID)
	if err != nil {
		log.Printf("Failed to update last_login_at for user %s: %v", user.Username, err)
	}

	userPermissions, err := h.getUserPermissions(c.Request.Context(), user.ID.String())
	if err != nil {
		log.Printf("Failed to get user permissions for %s: %v", user.Username, err)
		userPermissions = []string{}
	}

	expiresAt := models.UTCNow().Add(24 * time.Hour)
	normalizedRole := normalizeRoleName(roleName)

	claims := &models.Claims{
		UserID:      user.ID.String(),
		Username:    user.Username,
		Role:        normalizedRole,
		Permissions: userPermissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret := models.JWTSecret()
	if len(secret) == 0 {
		log.Printf("Login failed: JWT secret not configured")
		models.SendError(c, 500, 5000, "JWT secret not configured")
		return
	}

	tokenString, err := token.SignedString(secret)
	if err != nil {
		log.Printf("Login failed: token generation error for user %s: %v", user.Username, err)
		models.SendError(c, 500, 5000, "Token generation failed")
		return
	}

	refreshToken := uuid.New().String()
	if err := h.redis.Set(c.Request.Context(), "refresh:"+refreshToken, user.ID.String(), 7*24*time.Hour).Err(); err != nil {
		log.Printf("Login failed: refresh token store error for user %s: %v", user.Username, err)
		models.SendError(c, 500, 5000, err.Error())
		return
	}

	log.Printf("Login success: username=%s userId=%s", user.Username, user.ID)

	models.Send(c, 200, 0, "success", models.LoginResponse{
		Token:        tokenString,
		RefreshToken: refreshToken,
		ExpiresAt:    models.FormatUTC(expiresAt),
	})
}

// getUserRole 获取用户的角色，返回角色名称（默认返回 "auditor"）
func normalizeRoleName(roleName string) string {
	switch strings.TrimSpace(strings.ToLower(roleName)) {
	case "管理员", "admin", "administrator", "超级管理员", "superadmin":
		return "admin"
	case "审计员", "auditor":
		return "auditor"
	case "操作员", "operator", "ops":
		return "operator"
	case "查看员", "viewer":  // ✅ 添加查看员映射
		return "viewer"
	default:
		return "auditor"
	}
}

func (h *AuthHandler) getUserRole(ctx context.Context, userID string) string {
	if h.db == nil {
		return "auditor"
	}

	var roleName string
	var permissionsJSON string

	// 查询用户关联的角色
	err := h.db.QueryRowContext(ctx,
		`SELECT r.name, COALESCE(r.permissions, '[]')
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		LIMIT 1`,
		userID,
	).Scan(&roleName, &permissionsJSON)

	if err != nil {
		// 如果没有找到角色，尝试从默认角色表中获取
		defaultRole := h.getDefaultRole(ctx)
		if defaultRole != "" {
			return defaultRole
		}
		// 最后返回默认的 auditor
		return "auditor"
	}

	return normalizeRoleName(roleName)
}

func (h *AuthHandler) getUsernameByID(ctx context.Context, userID string) string {
	if h.db == nil {
		return ""
	}

	var username string
	if err := h.db.QueryRowContext(ctx, `SELECT username FROM users WHERE id = $1`, userID).Scan(&username); err != nil {
		return ""
	}
	return username
}

// getDefaultRole 获取默认角色（用于兼容旧数据）
func (h *AuthHandler) getDefaultRole(ctx context.Context) string {
	if h.db == nil {
		return "auditor"
	}

	var roleName string
	err := h.db.QueryRowContext(ctx,
		`SELECT name FROM roles WHERE name = '管理员' OR name = 'admin' LIMIT 1`,
	).Scan(&roleName)

	if err != nil {
		return "auditor"
	}
	return normalizeRoleName(roleName)
}

// getUserPermissions 获取用户的所有权限（去重）
func (h *AuthHandler) getUserPermissions(ctx context.Context, userID string) ([]string, error) {
    if h.db == nil {
        return []string{}, nil
    }

    // 使用 PostgreSQL 的 jsonb_array_elements_text 函数
    query := `
        SELECT DISTINCT jsonb_array_elements_text(r.permissions)
        FROM user_roles ur
        JOIN roles r ON ur.role_id = r.id
        WHERE ur.user_id = $1
    `
    rows, err := h.db.QueryContext(ctx, query, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    permissionSet := make(map[string]struct{})
    for rows.Next() {
        var perm string
        if err := rows.Scan(&perm); err != nil {
            continue
        }
        permissionSet[perm] = struct{}{}
    }

    permissions := make([]string, 0, len(permissionSet))
    for perm := range permissionSet {
        permissions = append(permissions, perm)
    }
    return permissions, nil
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		models.SendError(c, 400, 1001, "Invalid request")
		return
	}

	userID, err := h.redis.Get(c.Request.Context(), "refresh:"+req.RefreshToken).Result()
	if err != nil {
		models.SendError(c, 401, 1002, "Invalid refresh token")
		return
	}

	expiresAt := models.UTCNow().Add(24 * time.Hour)
	roleName := h.getUserRole(c.Request.Context(), userID)
	username := h.getUsernameByID(c.Request.Context(), userID)
	normalizedRole := normalizeRoleName(roleName)

	// ✅ 新增：获取用户权限
	userPermissions, err := h.getUserPermissions(c.Request.Context(), userID)
	if err != nil {
		log.Printf("Failed to get user permissions for refresh: %v", err)
		userPermissions = []string{}
	}

	// ✅ 修改：使用 Claims 结构体
	claims := &models.Claims{
		UserID:      userID,
		Username:    username,
		Role:        normalizedRole,
		Permissions: userPermissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret := models.JWTSecret()
	if len(secret) == 0 {
		models.SendError(c, 500, 5000, "JWT secret not configured")
		return
	}
	tokenString, err := token.SignedString(secret)
	if err != nil {
		models.SendError(c, 500, 5000, "Token generation failed")
		return
	}

	models.Send(c, 200, 0, "success", gin.H{
		"token":     tokenString,
		"expiresAt": models.FormatUTC(expiresAt),
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if len(authHeader) < len("Bearer ") || authHeader[:7] != "Bearer " {
		models.SendError(c, 400, 1001, "Invalid authorization header")
		return
	}

	if err := h.redis.Set(c.Request.Context(), "blacklist:"+authHeader[7:], "1", 24*time.Hour).Err(); err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}

	models.Send(c, 200, 0, "Logged out", nil)
}

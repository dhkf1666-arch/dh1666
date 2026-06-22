// backend/middleware/auth.go
// 完整修正版

package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"

	"enterprise-agent/backend/models"
)

// Auth 用户认证中间件
func Auth(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c)
		if tokenString == "" {
			models.SendError(c, 401, 1002, "Missing authorization token")
			c.Abort()
			return
		}

		claims := &models.Claims{}
		err := validateUserToken(tokenString, claims)
		if err != nil {
			if err == jwt.ErrTokenExpired {
				models.SendError(c, 401, 1002, "Token expired")
			} else {
				models.SendError(c, 401, 1002, "Invalid token")
			}
			c.Abort()
			return
		}

		// 检查黑名单（Redis 故障时拒绝鉴权）
		if revoked, unavailable := checkTokenBlacklist(c.Request.Context(), redisClient, tokenString); unavailable {
			abortAuthServiceUnavailable(c)
			return
		} else if revoked {
			models.SendError(c, 401, 1002, "Token has been revoked")
			c.Abort()
			return
		}

		// 设置用户信息
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", normalizeRoleName(claims.Role))
		// ✅ 关键修复：设置权限列表到context
		c.Set("permissions", claims.Permissions)
		c.Next()
	}
}

// DeviceAuth 设备认证中间件
func DeviceAuth(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Deprecated: Use OptionalAuth for upload endpoints where token may be absent.
		tokenString := extractToken(c)
		if tokenString == "" {
			models.SendError(c, 401, 1002, "Missing device token")
			c.Abort()
			return
		}

		claims := &models.DeviceClaims{}
		err := validateDeviceToken(tokenString, claims)

		if err != nil {
			if err == jwt.ErrTokenExpired {
				models.SendError(c, 401, 1002, "Device token expired")
			} else {
				models.SendError(c, 401, 1002, "Invalid device token")
			}
			c.Abort()
			return
		}

		// 检查黑名单（Redis 故障时拒绝鉴权）
		if revoked, unavailable := checkTokenBlacklist(c.Request.Context(), redisClient, tokenString); unavailable {
			abortAuthServiceUnavailable(c)
			return
		} else if revoked {
			models.SendError(c, 401, 1002, "Device token has been revoked")
			c.Abort()
			return
		}

	c.Set("deviceID", claims.DeviceID)
	c.Set("deviceId", claims.DeviceID)
	c.Set("device_id", claims.DeviceID)
	c.Next()
	}
}

// validateToken 验证 JWT token
func validateUserToken(tokenString string, claims *models.Claims) error {
	if err := parseTokenWithClaims(tokenString, claims); err != nil {
		return err
	}
	if claims.UserID == "" {
		return jwt.ErrTokenMalformed
	}
	return nil
}

func validateDeviceToken(tokenString string, claims *models.DeviceClaims) error {
	if err := parseTokenWithClaims(tokenString, claims); err != nil {
		return err
	}
	if claims.DeviceID == "" {
		return jwt.ErrTokenMalformed
	}
	return nil
}

func parseTokenWithClaims(tokenString string, claims jwt.Claims) error {
	secret := models.JWTSecret()
	if len(secret) == 0 {
		return jwt.ErrInvalidKey
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))

	if err != nil {
		return err
	}

	if !token.Valid {
		return jwt.ErrInvalidKey
	}

	return nil
}

// extractToken 从请求中提取 token
func extractToken(c *gin.Context) string {
	// 从 Authorization header 获取
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token != authHeader {
			return strings.TrimSpace(token)
		}
	}

	// 从 query 参数获取（供 <video>/<img> 等媒体流鉴权）
	if token := strings.TrimSpace(c.Query("token")); token != "" {
		return token
	}

	// 从 cookie 获取（可选）
	if token, err := c.Cookie("token"); err == nil && token != "" {
		return strings.TrimSpace(token)
	}

	return ""
}

func normalizeRoleName(roleName string) string {
	switch strings.TrimSpace(strings.ToLower(roleName)) {
	case "管理员", "admin", "administrator", "超级管理员", "superadmin":
		return "admin"
	case "审计员", "auditor":
		return "auditor"
	case "操作员", "operator", "ops":
		return "operator"
	default:
		return "auditor"
	}
}

// OptionalAuth 允许没有 token 的请求（用于 upload 分片兼容场景）
func OptionalAuth(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c)
		if tokenString == "" {
			// 允许匿名访问，但不设置用户信息
			c.Next()
			return
		}

		// 如果有 token，则优先尝试解析为用户 token
		claims := &models.Claims{}
		if err := validateUserToken(tokenString, claims); err == nil {
			if revoked, unavailable := checkTokenBlacklist(c.Request.Context(), redisClient, tokenString); unavailable {
				abortAuthServiceUnavailable(c)
				return
			} else if revoked {
				models.SendError(c, 401, 1002, "Token has been revoked")
				c.Abort()
				return
			}
			c.Set("userID", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("role", normalizeRoleName(claims.Role))
			c.Set("permissions", claims.Permissions)
			c.Next()
			return
		}

		// 再尝试解析为设备 token
		deviceClaims := &models.DeviceClaims{}
		if err := validateDeviceToken(tokenString, deviceClaims); err == nil {
			if revoked, unavailable := checkTokenBlacklist(c.Request.Context(), redisClient, tokenString); unavailable {
				abortAuthServiceUnavailable(c)
				return
			} else if revoked {
				models.SendError(c, 401, 1002, "Device token has been revoked")
				c.Abort()
				return
			}
			c.Set("deviceID", deviceClaims.DeviceID)
			c.Set("deviceId", deviceClaims.DeviceID)
			c.Set("device_id", deviceClaims.DeviceID)
			c.Next()
			return
		}

		// ✅ 修复：token 存在但解析失败 -> 拒绝访问（而不是放行）
		// 这里不再引用未定义的 err 变量
		models.SendError(c, 401, 1002, "Invalid authorization token")
		c.Abort()
	}
}

// UserOrDeviceAuth 接受用户 JWT 或设备 JWT（二者其一即可）
func UserOrDeviceAuth(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c)
		if tokenString == "" {
			models.SendError(c, 401, 1002, "Missing authorization token")
			c.Abort()
			return
		}

		claims := &models.Claims{}
		if err := validateUserToken(tokenString, claims); err == nil {
			if revoked, unavailable := checkTokenBlacklist(c.Request.Context(), redisClient, tokenString); unavailable {
				abortAuthServiceUnavailable(c)
				return
			} else if revoked {
				models.SendError(c, 401, 1002, "Token has been revoked")
				c.Abort()
				return
			}
			c.Set("userID", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("role", normalizeRoleName(claims.Role))
			c.Set("permissions", claims.Permissions)
			c.Next()
			return
		}

		deviceClaims := &models.DeviceClaims{}
		if err := validateDeviceToken(tokenString, deviceClaims); err == nil {
			if revoked, unavailable := checkTokenBlacklist(c.Request.Context(), redisClient, tokenString); unavailable {
				abortAuthServiceUnavailable(c)
				return
			} else if revoked {
				models.SendError(c, 401, 1002, "Device token has been revoked")
				c.Abort()
				return
			}
			c.Set("deviceID", deviceClaims.DeviceID)
			c.Set("deviceId", deviceClaims.DeviceID)
			c.Set("device_id", deviceClaims.DeviceID)
			c.Next()
			return
		}

		models.SendError(c, 401, 1002, "Invalid authorization token")
		c.Abort()
	}
}

// RequirePermission 权限检查中间件
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		if userID == "" {
			models.SendError(c, 401, 1002, "User not authenticated")
			c.Abort()
			return
		}

		// ✅ 优先检查 admin 角色（管理员拥有所有权限）
		role := c.GetString("role")
		if role == "admin" {
			c.Next()
			return
		}

		// 从 context 中获取权限列表
		perms, exists := c.Get("permissions")
		if !exists {
			models.SendError(c, 403, 1003, "Permission denied: "+permission)
			c.Abort()
			return
		}

		userPermissions, ok := perms.([]string)
		if !ok {
			models.SendError(c, 403, 1003, "Permission denied: "+permission)
			c.Abort()
			return
		}

		// 检查用户是否拥有该权限
		for _, p := range userPermissions {
			if p == permission {
				c.Next()
				return
			}
		}

		models.SendError(c, 403, 1003, "Permission denied: "+permission)
		c.Abort()
	}
}

// hasPermission 检查角色权限 - ✅ 已废弃，由 RequirePermission 替代
// 保留此函数以兼容可能的其他引用
func hasPermission(role string, permission string) bool {
	// 硬编码映射仅作为后退方案
	rolePermissions := map[string][]string{
		"admin": {
			"user:view", "user:create", "user:update", "user:delete", "user:manage",
			"role:view", "role:create", "role:update", "role:delete",
			"device:view", "device:manage", "device:bind", "device:unbind",
			"recording:view", "recording:play", "recording:export", "recording:delete", "recording:upload",
			"screenshot:view", "screenshot:delete",
			"activity:view", "activity:export", "activity:delete",
			"policy:view", "policy:create", "policy:update", "policy:delete",
			"audit:view",
		},
		"operator": {
			"user:view",
			"device:view", "device:manage", "device:bind", "device:unbind",
			"recording:view", "recording:play", "recording:export",
			"screenshot:view",
			"activity:view",
			"policy:view", "policy:create", "policy:update",
		},
		"auditor": {
			"device:view",
			"recording:view", "recording:play",
			"screenshot:view",
			"activity:view",
			"audit:view",
		},
	}

	perms, ok := rolePermissions[role]
	if !ok {
		return false
	}

	for _, p := range perms {
		if p == permission {
			return true
		}
	}
	return false
}

// RequireAnyPermission 检查是否有任意一个权限
func RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role == "" {
			models.SendError(c, 401, 1002, "User not authenticated")
			c.Abort()
			return
		}
		for _, perm := range permissions {
			if hasPermission(role, perm) {
				c.Next()
				return
			}
		}
		models.SendError(c, 403, 1003, "Permission denied")
		c.Abort()
	}
}


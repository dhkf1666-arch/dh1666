package middleware

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"enterprise-agent/backend/models"
)

// checkTokenBlacklist 检查 token 是否在黑名单中。
// Redis 不可用且客户端已配置时 fail-close，返回 serviceUnavailable=true。
func checkTokenBlacklist(ctx context.Context, redisClient *redis.Client, tokenString string) (revoked bool, serviceUnavailable bool) {
	if redisClient == nil {
		return false, false
	}

	exists, err := redisClient.Exists(ctx, "blacklist:"+tokenString).Result()
	if err != nil {
		log.Printf("[Auth] Redis blacklist check failed: %v", err)
		return false, true
	}
	return exists > 0, false
}

func abortAuthServiceUnavailable(c *gin.Context) {
	models.SendError(c, 503, 5003, "Authentication service temporarily unavailable")
	c.Abort()
}

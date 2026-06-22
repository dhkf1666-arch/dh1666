package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"enterprise-agent/backend/handlers"
	"enterprise-agent/backend/middleware"
)

// Handlers 仅保留管理后台所需处理器
type Handlers struct {
	Auth         *handlers.AuthHandler
	User         *handlers.UserHandler
	Admin        *handlers.AdminHandler
	OperationLog *handlers.OperationLogHandler
}

// SetupRoutes 配置管理后台路由
func SetupRoutes(router *gin.Engine, h *Handlers, redisClient *redis.Client) {
	api := router.Group("/api/v1")

	public := api.Group("")
	{
		public.POST("/auth/login", middleware.AuthRateLimit(), h.Auth.Login)
		public.POST("/auth/refresh", middleware.AuthRateLimit(), h.Auth.Refresh)
		public.POST("/auth/logout", middleware.AuthRateLimit(), h.Auth.Logout)
	}

	protected := api.Group("")
	protected.Use(middleware.Auth(redisClient))
	{
		users := protected.Group("/users")
		{
			users.GET("", middleware.RequirePermission("user:view"), h.User.ListUsers)
			users.POST("", middleware.RequirePermission("user:create"), h.User.CreateUser)
			users.GET("/:id", middleware.RequirePermission("user:view"), h.User.GetUser)
			users.PUT("/:id", middleware.RequirePermission("user:update"), h.User.UpdateUser)
			users.DELETE("/:id", middleware.RequirePermission("user:delete"), h.User.DeleteUser)
			users.POST("/:id/reset-password", middleware.RequirePermission("user:manage"), h.User.ResetPassword)
			users.POST("/:id/roles", middleware.RequirePermission("user:manage"), h.User.AssignRole)
			users.DELETE("/:id/roles/:roleId", middleware.RequirePermission("user:manage"), h.User.RemoveRole)
		}

		profile := protected.Group("/user")
		{
			profile.GET("/profile", h.User.GetProfile)
			profile.PUT("/profile", h.User.UpdateProfile)
			profile.POST("/change-password", h.User.ChangeOwnPassword)
		}

		roles := protected.Group("/roles")
		{
			roles.GET("", middleware.RequirePermission("role:view"), h.User.GetAllRoles)
			roles.GET("/:id", middleware.RequirePermission("role:view"), h.User.GetRole)
			roles.POST("", middleware.RequirePermission("role:create"), h.User.CreateRole)
			roles.PUT("/:id", middleware.RequirePermission("role:update"), h.User.UpdateRole)
			roles.DELETE("/:id", middleware.RequirePermission("role:delete"), h.User.DeleteRole)
		}

		admin := protected.Group("/admin")
		{
			admin.GET("/employees", middleware.RequirePermission("attendance:view"), h.Admin.GetEmployees)
			admin.POST("/employees", middleware.RequirePermission("attendance:edit"), h.Admin.CreateEmployee)
			admin.PUT("/employees/:id", middleware.RequirePermission("attendance:edit"), h.Admin.UpdateEmployee)
			admin.DELETE("/employees/:id", middleware.RequirePermission("attendance:edit"), h.Admin.DeleteEmployee)

			admin.GET("/attendance/records", middleware.RequirePermission("attendance:view"), h.Admin.GetAttendanceRecords)
			admin.POST("/attendance/records", middleware.RequirePermission("attendance:edit"), h.Admin.SaveAttendanceRecords)
			admin.GET("/attendance/today-absentees", middleware.RequirePermission("attendance:view"), h.Admin.GetTodayAbsentees)

			admin.GET("/performance", middleware.RequirePermission("attendance:view"), h.Admin.GetPerformance)
			admin.POST("/performance/batch", middleware.RequirePermission("attendance:edit"), h.Admin.BatchSavePerformance)
			admin.PUT("/performance/:employeeId/:month", middleware.RequirePermission("attendance:edit"), h.Admin.UpdatePerformance)
			admin.DELETE("/performance/:employeeId/:month", middleware.RequirePermission("attendance:edit"), h.Admin.DeletePerformance)

			admin.GET("/penalty/records", middleware.RequirePermission("attendance:view"), h.Admin.GetPenaltyRecords)
			admin.POST("/penalty/record", middleware.RequirePermission("attendance:edit"), h.Admin.CreatePenaltyRecord)
			admin.PUT("/penalty/records/:id", middleware.RequirePermission("attendance:edit"), h.Admin.UpdatePenaltyRecord)
			admin.DELETE("/penalty/records/:id", middleware.RequirePermission("attendance:edit"), h.Admin.DeletePenaltyRecord)

			admin.GET("/site-stats/sites", middleware.RequirePermission("site:view"), h.Admin.GetSites)
			admin.POST("/site-stats/sites", middleware.RequirePermission("site:manage"), h.Admin.CreateSite)
			admin.PUT("/site-stats/sites/:id", middleware.RequirePermission("site:manage"), h.Admin.UpdateSite)
			admin.DELETE("/site-stats/sites/:id", middleware.RequirePermission("site:manage"), h.Admin.DeleteSite)

			admin.GET("/site-stats/employee-accounts", middleware.RequirePermission("site:view"), h.Admin.GetEmployeeAccounts)
			admin.POST("/site-stats/employee-accounts", middleware.RequirePermission("site:manage"), h.Admin.CreateEmployeeAccount)
			admin.PUT("/site-stats/employee-accounts/:id", middleware.RequirePermission("site:manage"), h.Admin.UpdateEmployeeAccount)
			admin.DELETE("/site-stats/employee-accounts/:id", middleware.RequirePermission("site:manage"), h.Admin.DeleteEmployeeAccount)

			admin.POST("/site-stats/upload/preview", middleware.RequirePermission("stats:edit"), h.Admin.PreviewSiteStatsUpload)
			admin.POST("/site-stats/upload", middleware.RequirePermission("stats:edit"), h.Admin.UploadSiteStats)
			admin.DELETE("/site-stats/data/clear-by-date", middleware.RequirePermission("stats:edit"), h.Admin.ClearSiteStatsByDate)
			admin.DELETE("/site-stats/data/clear", middleware.RequirePermission("stats:edit"), h.Admin.ClearAllSiteStats)
			admin.GET("/site-stats/summary", middleware.RequirePermission("stats:view"), h.Admin.GetSiteStatsSummary)
			admin.GET("/site-stats/stacked-summary", middleware.RequirePermission("stats:view"), h.Admin.GetSiteStatsStacked)
			admin.DELETE("/site-stats/clear-by-date-only", middleware.RequirePermission("stats:edit"), h.Admin.ClearSiteStatsByDateOnly)
		}

		operationLogs := protected.Group("/operation-logs")
		{
			operationLogs.GET("", middleware.RequirePermission("operation_log:view"), h.OperationLog.GetOperationLogs)
			operationLogs.GET("/modules", middleware.RequirePermission("operation_log:view"), h.OperationLog.GetOperationModules)
			operationLogs.GET("/types", middleware.RequirePermission("operation_log:view"), h.OperationLog.GetOperationTypes)
			operationLogs.GET("/export", middleware.RequirePermission("operation_log:export"), h.OperationLog.ExportOperationLogs)
			operationLogs.DELETE("", middleware.RequirePermission("operation_log:delete"), h.OperationLog.DeleteOperationLogs)
		}
	}
}

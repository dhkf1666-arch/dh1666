// backend/handlers/user.go
// 完整替换文件内容 - 修复循环导入问题

package handlers

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"enterprise-agent/backend/models"
	"enterprise-agent/backend/services"
)

type UserHandler struct {
	repo *services.UserRepository
}

func NewUserHandler(repo *services.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

func userLogDisplayName(username, realName string) string {
	realName = strings.TrimSpace(realName)
	if realName != "" {
		return realName
	}
	return username
}

// ListUsers 获取用户列表
func (h *UserHandler) ListUsers(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "20")
	keyword := c.Query("keyword")

	// 转换页码和每页大小
	pageNum, _ := strconv.Atoi(page)
	pageSizeNum, _ := strconv.Atoi(pageSize)
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSizeNum < 1 {
		pageSizeNum = 20
	}
	if pageSizeNum > 100 {
		pageSizeNum = 100
	}

	users, total, err := h.repo.List(c.Request.Context(), page, pageSize, keyword)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}

	models.Send(c, 200, 0, "success", gin.H{
		"items":    users,
		"total":    total,
		"page":     pageNum,
		"pageSize": pageSizeNum,
	})
}

// CreateUser 创建用户
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.SendError(c, 400, 1001, "Invalid request: "+err.Error())
		return
	}

	user, err := h.repo.Create(c.Request.Context(), &req)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}

	// ✅ 记录操作日志
	Log(c, "CREATE", "USER",
		fmt.Sprintf("创建系统用户：%s（登录名 %s）", userLogDisplayName(user.Username, user.RealName), user.Username),
		user.ID.String(), userLogDisplayName(user.Username, user.RealName))

	models.Send(c, 200, 0, "User created", user)
}

// GetUser 获取用户详情
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		models.SendError(c, 400, 1001, "Invalid user ID")
		return
	}

	user, err := h.repo.Get(c.Request.Context(), id)
	if err == sql.ErrNoRows {
		models.SendError(c, 404, 2001, "User not found")
		return
	}
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}

	models.Send(c, 200, 0, "success", user)
}

// UpdateUser 更新用户
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		models.SendError(c, 400, 1001, "Invalid user ID")
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.SendError(c, 400, 1001, "Invalid request")
		return
	}

	user, err := h.repo.Update(c.Request.Context(), id, &req)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}

	// ✅ 记录操作日志
	Log(c, "UPDATE", "USER",
		fmt.Sprintf("修改系统用户：%s（登录名 %s）", userLogDisplayName(user.Username, user.RealName), user.Username),
		id.String(), userLogDisplayName(user.Username, user.RealName))

	models.Send(c, 200, 0, "User updated", user)
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		models.SendError(c, 400, 1001, "Invalid user ID")
		return
	}

	// 不能删除自己
	currentUserID := c.GetString("userID")
	if currentUserID == id.String() {
		models.SendError(c, 400, 1001, "Cannot delete yourself")
		return
	}

	// 获取用户信息用于日志
	user, err := h.repo.Get(c.Request.Context(), id)
	if err == nil && user.Username == "admin" {
		models.SendError(c, 400, 1001, "Cannot delete admin user")
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}

	// ✅ 记录操作日志
	if user != nil {
		Log(c, "DELETE", "USER",
			fmt.Sprintf("删除系统用户：%s（登录名 %s）", userLogDisplayName(user.Username, user.RealName), user.Username),
			id.String(), userLogDisplayName(user.Username, user.RealName))
	}

	models.Send(c, 200, 0, "User deleted", nil)
}

// ResetPassword 重置用户密码（管理员）
func (h *UserHandler) ResetPassword(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		models.SendError(c, 400, 1001, "Invalid user ID")
		return
	}

	var req struct {
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		models.SendError(c, 400, 1001, "Invalid request: password is required and must be at least 6 characters")
		return
	}

	// 先获取用户信息用于日志
	user, err := h.repo.Get(c.Request.Context(), id)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}

	if err := h.repo.ChangePassword(c.Request.Context(), id, req.Password); err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}

	// ✅ 记录操作日志
	Log(c, "UPDATE", "USER",
		fmt.Sprintf("重置密码：用户 %s（登录名 %s）", userLogDisplayName(user.Username, user.RealName), user.Username),
		id.String(), userLogDisplayName(user.Username, user.RealName))

	models.Send(c, 200, 0, "Password reset successfully", nil)
}

// GetProfile 获取当前用户信息
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("userID")
	id, err := uuid.Parse(userID)
	if err != nil {
		models.SendError(c, 401, 1002, "Invalid user")
		return
	}

	user, err := h.repo.Get(c.Request.Context(), id)
	if err != nil {
		models.SendError(c, 404, 2001, "User not found")
		return
	}

	models.Send(c, 200, 0, "success", user)
}

// UpdateProfile 更新个人信息
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("userID")
	id, err := uuid.Parse(userID)
	if err != nil {
		models.SendError(c, 401, 1002, "Invalid user")
		return
	}

	var req struct {
		Email    string `json:"email"`
		RealName string `json:"realName"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		models.SendError(c, 400, 1001, "Invalid request")
		return
	}

	user, err := h.repo.UpdateProfile(c.Request.Context(), id, req.Email, req.RealName)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}

	// ✅ 记录操作日志
	Log(c, "UPDATE", "USER",
		fmt.Sprintf("修改个人资料：%s", userLogDisplayName(user.Username, user.RealName)),
		id.String(), userLogDisplayName(user.Username, user.RealName))

	models.Send(c, 200, 0, "Profile updated", user)
}

// ChangeOwnPassword 修改自己的密码
func (h *UserHandler) ChangeOwnPassword(c *gin.Context) {
	userID := c.GetString("userID")
	id, err := uuid.Parse(userID)
	if err != nil {
		models.SendError(c, 401, 1002, "Invalid user")
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.SendError(c, 400, 1001, "Invalid request")
		return
	}

	if err := h.repo.ChangeOwnPassword(c.Request.Context(), id, req.OldPassword, req.NewPassword); err != nil {
		models.SendError(c, 400, 1001, err.Error())
		return
	}

	// ✅ 记录操作日志
	username := c.GetString("username")
	Log(c, "UPDATE", "USER",
		fmt.Sprintf("修改登录密码：%s", username), id.String(), username)

	models.Send(c, 200, 0, "Password changed successfully", nil)
}

// ========== 角色管理 ==========

// GetAllRoles 获取所有角色
func (h *UserHandler) GetAllRoles(c *gin.Context) {
	roles, err := h.repo.GetAllRoles(c.Request.Context())
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}
	models.Send(c, 200, 0, "success", roles)
}

// GetRole 获取单个角色
func (h *UserHandler) GetRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		models.SendError(c, 400, 1001, "Invalid role ID")
		return
	}

	role, err := h.repo.GetRoleByID(c.Request.Context(), id)
	if err != nil {
		models.SendError(c, 404, 2001, err.Error())
		return
	}
	models.Send(c, 200, 0, "success", role)
}

// CreateRole 创建角色
func (h *UserHandler) CreateRole(c *gin.Context) {
	var req models.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.SendError(c, 400, 1001, "Invalid request: "+err.Error())
		return
	}

	role, err := h.repo.CreateRole(c.Request.Context(), &req)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}

	// ✅ 记录操作日志
	Log(c, "CREATE", "ROLE",
		fmt.Sprintf("创建角色：%s", role.Name), role.ID.String(), role.Name)

	models.Send(c, 200, 0, "Role created", role)
}

// UpdateRole 更新角色
func (h *UserHandler) UpdateRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		models.SendError(c, 400, 1001, "Invalid role ID")
		return
	}

	var req models.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.SendError(c, 400, 1001, "Invalid request")
		return
	}

	role, err := h.repo.UpdateRole(c.Request.Context(), id, &req)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}

	// ✅ 记录操作日志
	Log(c, "UPDATE", "ROLE",
		fmt.Sprintf("修改角色：%s", role.Name), id.String(), role.Name)

	models.Send(c, 200, 0, "Role updated", role)
}

// DeleteRole 删除角色
func (h *UserHandler) DeleteRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		models.SendError(c, 400, 1001, "Invalid role ID")
		return
	}

	// 获取角色信息用于日志
	role, err := h.repo.GetRoleByID(c.Request.Context(), id)
	if err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}

	if err := h.repo.DeleteRole(c.Request.Context(), id); err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}

	// ✅ 记录操作日志
	Log(c, "DELETE", "ROLE",
		fmt.Sprintf("删除角色：%s", role.Name), id.String(), role.Name)

	models.Send(c, 200, 0, "Role deleted", nil)
}

// AssignRole 为用户分配角色
func (h *UserHandler) AssignRole(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		models.SendError(c, 400, 1001, "Invalid user ID")
		return
	}

	var req struct {
		RoleID string `json:"roleId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		models.SendError(c, 400, 1001, "Invalid request")
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		models.SendError(c, 400, 1001, "Invalid role ID")
		return
	}

	// 获取用户和角色信息
	user, _ := h.repo.Get(c.Request.Context(), userID)
	role, _ := h.repo.GetRoleByID(c.Request.Context(), roleID)

	if err := h.repo.AssignRoleToUser(c.Request.Context(), userID, roleID); err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}

	// ✅ 记录操作日志
	roleName := ""
	if role != nil {
		roleName = role.Name
	}
	username := ""
	if user != nil {
		username = user.Username
	}
	Log(c, "UPDATE", "ROLE",
		fmt.Sprintf("分配角色：用户 %s → 角色 %s", username, roleName), userID.String(), username)

	models.Send(c, 200, 0, "Role assigned", nil)
}

// RemoveRole 从用户移除角色
func (h *UserHandler) RemoveRole(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		models.SendError(c, 400, 1001, "Invalid user ID")
		return
	}

	var req struct {
		RoleID string `json:"roleId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		models.SendError(c, 400, 1001, "Invalid request")
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		models.SendError(c, 400, 1001, "Invalid role ID")
		return
	}

	// 获取用户和角色信息
	user, _ := h.repo.Get(c.Request.Context(), userID)
	role, _ := h.repo.GetRoleByID(c.Request.Context(), roleID)

	if err := h.repo.RemoveRoleFromUser(c.Request.Context(), userID, roleID); err != nil {
		models.SendError(c, 500, 5000, err.Error())
		return
	}

	// ✅ 记录操作日志
	roleName := ""
	if role != nil {
		roleName = role.Name
	}
	username := ""
	if user != nil {
		username = user.Username
	}
	Log(c, "UPDATE", "ROLE",
		fmt.Sprintf("移除角色：用户 %s ← 角色 %s", username, roleName), userID.String(), username)

	models.Send(c, 200, 0, "Role removed", nil)
}
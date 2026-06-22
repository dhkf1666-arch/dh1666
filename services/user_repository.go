// backend/services/user_repository.go
// 完整替换文件内容

package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"enterprise-agent/backend/models"
)

// UserRepository 用户仓库
// 主存储：PostgreSQL，内存作为一级缓存
type UserRepository struct {
	db    *sql.DB
	mu    sync.RWMutex
	users map[uuid.UUID]*models.User // 缓存
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *sql.DB) *UserRepository {
	repo := &UserRepository{
		db:    db,
		users: make(map[uuid.UUID]*models.User),
	}
	if db != nil {
		repo.loadUsers() // 加载缓存
	}
	return repo
}

// loadUsers 加载所有用户到缓存
func (r *UserRepository) loadUsers() {
	// PostgreSQL 兼容的用户+角色查询
	query := `
		SELECT 
			u.id, 
			u.username, 
			u.password_hash, 
			u.email, 
			u.real_name, 
			u.status, 
			u.created_at,
			u.last_login_at,
			COALESCE(
				(
					SELECT json_agg(
						json_build_object(
							'id', r.id,
							'name', r.name,
							'description', r.description,
							'permissions', r.permissions,
							'created_at', r.created_at,
							'updated_at', r.updated_at
						)
					)
					FROM user_roles ur
					JOIN roles r ON ur.role_id = r.id
					WHERE ur.user_id = u.id
				),
				'[]'::json
			) as roles
		FROM users u
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		var passwordHash string
		var createdAt time.Time
		var lastLogin sql.NullTime
		var rolesJSON string

		if err := rows.Scan(&user.ID, &user.Username, &passwordHash, &user.Email, &user.RealName, &user.Status, &createdAt, &lastLogin, &rolesJSON); err != nil {
			continue
		}
		user.PasswordHash = passwordHash
		user.CreatedAt = createdAt
		if lastLogin.Valid {
			user.LastLogin = &lastLogin.Time
		}

		// 解析用户关联的角色
		var roles []models.Role
		if err := json.Unmarshal([]byte(rolesJSON), &roles); err == nil {
			user.Roles = roles
		}

		r.users[user.ID] = &user
	}
}

// List 获取用户列表（带分页）
func (r *UserRepository) List(ctx context.Context, page, pageSize, keyword string) ([]models.User, int, error) {
	pageNum := 1
	pageSizeNum := 20

	if page != "" {
		if v, err := parseInt(page); err == nil && v > 0 {
			pageNum = v
		}
	}
	if pageSize != "" {
		if v, err := parseInt(pageSize); err == nil && v > 0 && v <= 100 {
			pageSizeNum = v
		}
	}

	offset := (pageNum - 1) * pageSizeNum
	args := []interface{}{}
	where := "1=1"
	argIdx := 1

	if keyword != "" {
		where = fmt.Sprintf("(username ILIKE $%d OR real_name ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+keyword+"%")
		argIdx++
	}

	// 查询总数
	countQuery := `SELECT COUNT(*) FROM users WHERE ` + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 查询用户列表
	dataQuery := fmt.Sprintf(`
		SELECT id, username, email, real_name, status, created_at, last_login_at
		FROM users 
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)

	args = append(args, pageSizeNum, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var lastLogin sql.NullTime
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.RealName, &u.Status, &u.CreatedAt, &lastLogin); err != nil {
			return nil, 0, err
		}
		if lastLogin.Valid {
			u.LastLogin = &lastLogin.Time
		}
		// 获取用户角色
		roles, _ := r.GetUserRoles(ctx, u.ID)
		u.Roles = roles
		users = append(users, u)
	}

	return users, total, nil
}

// Get 根据ID获取用户
func (r *UserRepository) Get(ctx context.Context, id uuid.UUID) (*models.User, error) {
	// 先查缓存
	r.mu.RLock()
	if user, ok := r.users[id]; ok {
		userCopy := *user
		userCopy.PasswordHash = ""
		r.mu.RUnlock()
		return &userCopy, nil
	}
	r.mu.RUnlock()

	// 缓存未命中，查询数据库
	query := `
		SELECT id, username, password_hash, email, real_name, status, created_at, last_login_at
		FROM users WHERE id = $1
	`
	var user models.User
	var passwordHash string
	var createdAt time.Time
	var lastLogin sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &passwordHash, &user.Email, &user.RealName, &user.Status, &createdAt, &lastLogin)
	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	user.PasswordHash = passwordHash
	user.CreatedAt = createdAt

	// 获取角色
	roles, _ := r.GetUserRoles(ctx, user.ID)
	user.Roles = roles

	// 更新缓存
	r.mu.Lock()
	r.users[user.ID] = &user
	r.mu.Unlock()

	userCopy := user
	userCopy.PasswordHash = ""
	return &userCopy, nil
}

// GetByUsername 根据用户名获取用户（用于登录验证）
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	// 先查缓存
	r.mu.RLock()
	for _, user := range r.users {
		if user.Username == username {
			userCopy := *user
			r.mu.RUnlock()
			return &userCopy, nil
		}
	}
	r.mu.RUnlock()

	// 缓存未命中，查询数据库
	query := `
		SELECT id, username, password_hash, email, real_name, status, created_at
		FROM users WHERE username = $1
	`
	var user models.User
	var passwordHash string
	var createdAt time.Time

	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &passwordHash, &user.Email, &user.RealName, &user.Status, &createdAt)
	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	user.PasswordHash = passwordHash
	user.CreatedAt = createdAt

	// 获取角色
	roles, _ := r.GetUserRoles(ctx, user.ID)
	user.Roles = roles

	// 更新缓存
	r.mu.Lock()
	r.users[user.ID] = &user
	r.mu.Unlock()

	userCopy := user
	return &userCopy, nil
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	// 检查用户名是否存在
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`, req.Username).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("username already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           uuid.New(),
		Username:     req.Username,
		PasswordHash: string(hash),
		Email:        req.Email,
		RealName:     req.RealName,
		Status:       1,
		CreatedAt:    models.UTCNow(),
		Roles:        []models.Role{},
	}

	// 使用事务
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 插入用户
	_, err = tx.Exec(`
		INSERT INTO users (id, username, password_hash, email, real_name, status, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		user.ID, user.Username, user.PasswordHash, user.Email, user.RealName, user.Status, user.CreatedAt)
	if err != nil {
		return nil, err
	}

	// 分配角色
	if len(req.RoleIDs) > 0 {
		for _, roleIDStr := range req.RoleIDs {
			roleID, err := uuid.Parse(roleIDStr)
			if err != nil {
				continue
			}
			_, err = tx.Exec(`
				INSERT INTO user_roles (user_id, role_id, created_at) 
				VALUES ($1, $2, $3)`,
				user.ID, roleID, models.UTCNow())
			if err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 获取角色信息
	roles, _ := r.GetUserRoles(ctx, user.ID)
	user.Roles = roles

	// 更新缓存
	r.mu.Lock()
	r.users[user.ID] = user
	r.mu.Unlock()

	userCopy := *user
	userCopy.PasswordHash = ""
	return &userCopy, nil
}

// Update 更新用户
func (r *UserRepository) Update(ctx context.Context, id uuid.UUID, req *models.UpdateUserRequest) (*models.User, error) {
	// 使用事务
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 构建更新语句
	updates := []string{}
	args := []interface{}{}
	argIdx := 1

	if req.Email != "" {
		updates = append(updates, fmt.Sprintf("email = $%d", argIdx))
		args = append(args, req.Email)
		argIdx++
	}
	if req.RealName != "" {
		updates = append(updates, fmt.Sprintf("real_name = $%d", argIdx))
		args = append(args, req.RealName)
		argIdx++
	}
	if req.Status != 0 {
		updates = append(updates, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, req.Status)
		argIdx++
	}

	if len(updates) > 0 {
		updates = append(updates, "updated_at = NOW()")
		query := fmt.Sprintf(`UPDATE users SET %s WHERE id = $%d`, strings.Join(updates, ", "), argIdx)
		args = append(args, id)
		_, err = tx.Exec(query, args...)
		if err != nil {
			return nil, err
		}
	}

	// 更新角色关联
	if req.RoleIDs != nil {
		// 删除原有角色关联
		_, err = tx.Exec(`DELETE FROM user_roles WHERE user_id = $1`, id)
		if err != nil {
			return nil, err
		}

		// 添加新角色关联
		for _, roleIDStr := range req.RoleIDs {
			roleID, err := uuid.Parse(roleIDStr)
			if err != nil {
				continue
			}
			_, err = tx.Exec(`
				INSERT INTO user_roles (user_id, role_id, created_at) 
				VALUES ($1, $2, $3)`,
				id, roleID, models.UTCNow())
			if err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 重新加载用户信息到缓存
	return r.Get(ctx, id)
}

// Delete 删除用户
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// 不能删除管理员
	var username string
	err := r.db.QueryRowContext(ctx, `SELECT username FROM users WHERE id = $1`, id).Scan(&username)
	if err != nil {
		return err
	}
	if username == "admin" {
		return fmt.Errorf("cannot delete admin user")
	}

	// 使用事务
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 删除用户角色关联
	_, err = tx.Exec(`DELETE FROM user_roles WHERE user_id = $1`, id)
	if err != nil {
		return err
	}

	// 删除用户
	_, err = tx.Exec(`DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	// 从缓存中删除
	r.mu.Lock()
	delete(r.users, id)
	r.mu.Unlock()

	return nil
}

// ChangePassword 修改用户密码（管理员操作）
func (r *UserRepository) ChangePassword(ctx context.Context, id uuid.UUID, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `UPDATE users SET password_hash = $1 WHERE id = $2`, string(hash), id)
	if err != nil {
		return err
	}

	// 更新缓存
	r.mu.Lock()
	if user, ok := r.users[id]; ok {
		user.PasswordHash = string(hash)
	}
	r.mu.Unlock()

	return nil
}

// ChangeOwnPassword 修改自己的密码
func (r *UserRepository) ChangeOwnPassword(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error {
	// 获取当前密码哈希
	var currentHash string
	err := r.db.QueryRowContext(ctx, `SELECT password_hash FROM users WHERE id = $1`, id).Scan(&currentHash)
	if err != nil {
		return err
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(oldPassword)); err != nil {
		return fmt.Errorf("old password is incorrect")
	}

	// 更新密码
	return r.ChangePassword(ctx, id, newPassword)
}

// UpdateProfile 更新个人信息
func (r *UserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, email, realName string) (*models.User, error) {
	updates := []string{}
	args := []interface{}{}
	argIdx := 1

	if email != "" {
		updates = append(updates, fmt.Sprintf("email = $%d", argIdx))
		args = append(args, email)
		argIdx++
	}
	if realName != "" {
		updates = append(updates, fmt.Sprintf("real_name = $%d", argIdx))
		args = append(args, realName)
		argIdx++
	}

	if len(updates) > 0 {
		updates = append(updates, "updated_at = NOW()")
		query := fmt.Sprintf(`UPDATE users SET %s WHERE id = $%d`, strings.Join(updates, ", "), argIdx)
		args = append(args, id)
		_, err := r.db.ExecContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
	}

	// 重新加载用户信息
	return r.Get(ctx, id)
}

// ========== 角色管理方法 ==========

// GetAllRoles 获取所有角色
func (r *UserRepository) GetAllRoles(ctx context.Context) ([]models.Role, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, description, permissions, created_at, updated_at 
		FROM roles ORDER BY created_at
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		var permissionsJSON string
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &permissionsJSON, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(permissionsJSON), &role.Permissions); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

// CreateRole 创建角色
func (r *UserRepository) CreateRole(ctx context.Context, req *models.CreateRoleRequest) (*models.Role, error) {
	// 检查角色名是否已存在
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM roles WHERE name = $1)`, req.Name).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("role name already exists")
	}

	permissionsJSON, err := json.Marshal(req.Permissions)
	if err != nil {
		return nil, err
	}

	role := &models.Role{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
		CreatedAt:   models.UTCNow(),
		UpdatedAt:   models.UTCNow(),
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO roles (id, name, description, permissions, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6)`,
		role.ID, role.Name, role.Description, string(permissionsJSON), role.CreatedAt, role.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return role, nil
}

// UpdateRole 更新角色
func (r *UserRepository) UpdateRole(ctx context.Context, id uuid.UUID, req *models.UpdateRoleRequest) (*models.Role, error) {
	updates := []string{}
	args := []interface{}{}
	argIdx := 1

	if req.Name != "" {
		updates = append(updates, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, req.Name)
		argIdx++
	}
	if req.Description != "" {
		updates = append(updates, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, req.Description)
		argIdx++
	}
	if req.Permissions != nil {
		permissionsJSON, err := json.Marshal(req.Permissions)
		if err != nil {
			return nil, err
		}
		updates = append(updates, fmt.Sprintf("permissions = $%d", argIdx))
		args = append(args, string(permissionsJSON))
		argIdx++
	}

	if len(updates) == 0 {
		return r.GetRoleByID(ctx, id)
	}

	updates = append(updates, fmt.Sprintf("updated_at = $%d", argIdx))
	args = append(args, models.UTCNow())
	argIdx++

	args = append(args, id)
	query := fmt.Sprintf(`UPDATE roles SET %s WHERE id = $%d`, strings.Join(updates, ", "), argIdx)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return r.GetRoleByID(ctx, id)
}

// DeleteRole 删除角色
func (r *UserRepository) DeleteRole(ctx context.Context, id uuid.UUID) error {
	// 检查是否有用户关联此角色
	var userCount int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_roles WHERE role_id = $1`, id).Scan(&userCount)
	if err != nil {
		return err
	}
	if userCount > 0 {
		return fmt.Errorf("cannot delete role: it is assigned to %d user(s)", userCount)
	}

	// 先删除角色关联
	_, err = r.db.ExecContext(ctx, `DELETE FROM user_roles WHERE role_id = $1`, id)
	if err != nil {
		return err
	}

	// 删除角色
	_, err = r.db.ExecContext(ctx, `DELETE FROM roles WHERE id = $1`, id)
	return err
}

// GetRoleByID 根据ID获取角色
func (r *UserRepository) GetRoleByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	var role models.Role
	var permissionsJSON string

	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, description, permissions, created_at, updated_at 
		FROM roles WHERE id = $1`, id).
		Scan(&role.ID, &role.Name, &role.Description, &permissionsJSON, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("role not found")
		}
		return nil, err
	}

	if err := json.Unmarshal([]byte(permissionsJSON), &role.Permissions); err != nil {
		return nil, err
	}

	return &role, nil
}

// GetUserRoles 获取用户的所有角色
func (r *UserRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]models.Role, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT r.id, r.name, r.description, r.permissions, r.created_at, r.updated_at
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		var permissionsJSON string
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &permissionsJSON, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(permissionsJSON), &role.Permissions); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

// AssignRoleToUser 为用户分配角色
func (r *UserRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	// 检查是否已存在
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM user_roles WHERE user_id = $1 AND role_id = $2)`,
		userID, roleID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return nil // 已存在，无需重复添加
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO user_roles (user_id, role_id, created_at) 
		VALUES ($1, $2, $3)`,
		userID, roleID, models.UTCNow())
	if err != nil {
		return err
	}

	// 更新缓存中的用户角色信息
	r.mu.Lock()
	defer r.mu.Unlock()
	if user, ok := r.users[userID]; ok {
		roles, _ := r.GetUserRoles(ctx, userID)
		user.Roles = roles
	}

	return nil
}

// RemoveRoleFromUser 从用户移除角色
func (r *UserRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`,
		userID, roleID)
	if err != nil {
		return err
	}

	// 更新缓存中的用户角色信息
	r.mu.Lock()
	defer r.mu.Unlock()
	if user, ok := r.users[userID]; ok {
		roles, _ := r.GetUserRoles(ctx, userID)
		user.Roles = roles
	}

	return nil
}

// InvalidateCache 清除指定用户的缓存
func (r *UserRepository) InvalidateCache(userID uuid.UUID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.users, userID)
}

// ClearCache 清除所有缓存
func (r *UserRepository) ClearCache() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users = make(map[uuid.UUID]*models.User)
}

// RefreshCache 刷新缓存
func (r *UserRepository) RefreshCache(ctx context.Context) error {
	r.mu.Lock()
	r.users = make(map[uuid.UUID]*models.User)
	r.mu.Unlock()
	r.loadUsers()
	return nil
}

// parseInt 辅助函数：字符串转整数
func parseInt(s string) (int, error) {
	var result int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid integer: %s", s)
		}
		result = result*10 + int(c-'0')
	}
	return result, nil
}

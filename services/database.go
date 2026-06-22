package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	// "strconv"
	"strings"
	"time"

	// "github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/lib/pq"

	"enterprise-agent/backend/models"
)

func NewDatabase() *sql.DB {
	// ✅ 优先使用 DATABASE_URL（Render/Aiven 推荐方式）
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		log.Println("Using DATABASE_URL for PostgreSQL connection")
		db, err := sql.Open("postgres", dbURL)
		if err != nil {
			panic(fmt.Sprintf("failed to open database: %v", err))
		}

		db.SetMaxOpenConns(30)
		db.SetMaxIdleConns(10)
		db.SetConnMaxLifetime(5 * time.Minute)

		if err := db.Ping(); err != nil {
			panic(fmt.Sprintf("failed to ping database: %v", err))
		}

		if err := initSchema(db); err != nil {
			panic(fmt.Sprintf("failed to initialize schema: %v", err))
		}
		if err := seedDefaultAdmin(db); err != nil {
			panic(fmt.Sprintf("failed to seed default admin: %v", err))
		}

		log.Println("PostgreSQL connected successfully via DATABASE_URL")
		return db
	}

	// 原有的单独配置方式（支持 SSL）
	host := envOr("DB_HOST", "localhost")
	port := envOr("DB_PORT", "5432")
	user := envOr("DB_USER", "admin")
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "admin123"
	}
	dbname := envOr("DB_NAME", "dhpg_admin")
	sslmode := envOr("DB_SSLMODE", "require") // ✅ 默认改为 require（Aiven 需要）

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(fmt.Sprintf("failed to open database: %v", err))
	}

	db.SetMaxOpenConns(30)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		panic(fmt.Sprintf("failed to ping database: %v", err))
	}

	if err := initSchema(db); err != nil {
		panic(fmt.Sprintf("failed to initialize schema: %v", err))
	}
	if err := seedDefaultAdmin(db); err != nil {
		panic(fmt.Sprintf("failed to seed default admin: %v", err))
	}

	log.Println("PostgreSQL connected successfully")
	return db
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func initSchema(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			username VARCHAR(100) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			email VARCHAR(255),
			real_name VARCHAR(100),
			status SMALLINT NOT NULL DEFAULT 1,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ,
			last_login_at TIMESTAMPTZ
		)`,
		`CREATE TABLE IF NOT EXISTS roles (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(100) NOT NULL UNIQUE,
			description TEXT,
			permissions JSONB NOT NULL DEFAULT '[]',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS user_roles (
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (user_id, role_id)
		)`,
		`CREATE TABLE IF NOT EXISTS employees (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			employee_id VARCHAR(100),
			name VARCHAR(255),
			position VARCHAR(100),
			hire_date DATE,
			work_location VARCHAR(255),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ
		)`,
		`CREATE INDEX IF NOT EXISTS idx_employees_employee_id ON employees(employee_id)`,
		`CREATE TABLE IF NOT EXISTS attendance_records (
			employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
			year_month VARCHAR(7) NOT NULL,
			date DATE NOT NULL,
			status VARCHAR(50) NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (employee_id, year_month, date)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_attendance_records_employee_id ON attendance_records(employee_id)`,
		`CREATE TABLE IF NOT EXISTS performance_records (
			employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
			employee_name VARCHAR(255),
			position VARCHAR(100),
			score_records JSONB,
			total_score INTEGER DEFAULT 0,
			grade VARCHAR(50),
			month VARCHAR(7) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (employee_id, month)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_performance_records_month ON performance_records(month)`,
		`CREATE TABLE IF NOT EXISTS penalty_records (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			employee_id UUID REFERENCES employees(id) ON DELETE SET NULL,
			employee_name VARCHAR(255),
			position VARCHAR(100),
			amount NUMERIC(12,2) NOT NULL DEFAULT 0,
			category VARCHAR(100),
			reason TEXT,
			penalty_date DATE NOT NULL,
			created_by VARCHAR(100),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_penalty_records_employee_id ON penalty_records(employee_id)`,
		`CREATE INDEX IF NOT EXISTS idx_penalty_records_penalty_date ON penalty_records(penalty_date)`,
		`CREATE TABLE IF NOT EXISTS sites (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			code VARCHAR(100) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			sort_order INTEGER NOT NULL DEFAULT 0,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ
		)`,
		`CREATE TABLE IF NOT EXISTS employee_accounts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			site_id UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
			name VARCHAR(255),
			account_name VARCHAR(255) NOT NULL,
			shift VARCHAR(50),
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ,
			UNIQUE (site_id, account_name)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_employee_accounts_site_id ON employee_accounts(site_id)`,
		`CREATE TABLE IF NOT EXISTS site_stats (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			employee_account_id UUID NOT NULL REFERENCES employee_accounts(id) ON DELETE CASCADE,
			site_id UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
			stat_date DATE NOT NULL,
			shift VARCHAR(50),
			order_count INTEGER NOT NULL DEFAULT 0,
			avg_time_seconds INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ,
			UNIQUE (employee_account_id, stat_date, shift)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_site_stats_site_id ON site_stats(site_id)`,
		`CREATE INDEX IF NOT EXISTS idx_site_stats_employee_account_id ON site_stats(employee_account_id)`,
		`CREATE INDEX IF NOT EXISTS idx_site_stats_stat_date ON site_stats(stat_date)`,
		`CREATE TABLE IF NOT EXISTS operation_logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			operator_id VARCHAR(36),
			operator_name VARCHAR(100) NOT NULL,
			operator_role VARCHAR(50),
			operation_type VARCHAR(50) NOT NULL,
			operation_module VARCHAR(50) NOT NULL,
			operation_desc TEXT NOT NULL,
			target_id VARCHAR(100),
			target_name VARCHAR(255),
			ip_address VARCHAR(45),
			user_agent TEXT,
			request_method VARCHAR(10),
			request_path VARCHAR(255),
			request_params TEXT,
			response_status INTEGER,
			execution_time_ms INTEGER DEFAULT 0,
			old_value TEXT,
			new_value TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_operation_logs_operator_name ON operation_logs(operator_name)`,
		`CREATE INDEX IF NOT EXISTS idx_operation_logs_created_at ON operation_logs(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_operation_logs_operation_module ON operation_logs(operation_module)`,
		`CREATE INDEX IF NOT EXISTS idx_operation_logs_operation_type ON operation_logs(operation_type)`,
		`CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql`,
		`DROP TRIGGER IF EXISTS update_users_updated_at ON users`,
		`CREATE TRIGGER update_users_updated_at
			BEFORE UPDATE ON users
			FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()`,
		`DROP TRIGGER IF EXISTS update_roles_updated_at ON roles`,
		`CREATE TRIGGER update_roles_updated_at
			BEFORE UPDATE ON roles
			FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			if strings.Contains(err.Error(), "already exists") {
				continue
			}
			return fmt.Errorf("schema init failed: %w\nSQL: %s", err, stmt)
		}
	}
	return nil
}

func seedDefaultAdmin(db *sql.DB) error {
	var adminRoleID string
	for _, role := range models.DefaultRoles {
		permissionsJSON, err := json.Marshal(role.Permissions)
		if err != nil {
			return err
		}

		var roleID string
		err = db.QueryRow(`
			INSERT INTO roles (id, name, description, permissions, created_at, updated_at)
			VALUES (gen_random_uuid(), $1, $2, $3, NOW(), NOW())
			ON CONFLICT (name) DO UPDATE SET
				description = EXCLUDED.description,
				permissions = EXCLUDED.permissions,
				updated_at = NOW()
			RETURNING id`,
			role.Name, role.Description, string(permissionsJSON),
		).Scan(&roleID)
		if err != nil {
			return err
		}
		if role.Name == "管理员" {
			adminRoleID = roleID
		}
	}

	if adminRoleID == "" {
		return fmt.Errorf("admin role not found")
	}

	username := envOr("ADMIN_USERNAME", "admin")
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		password = "admin123"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	var userID string
	var userExists bool
	if err := db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`, username,
	).Scan(&userExists); err != nil {
		return err
	}

	if userExists {
		if err := db.QueryRow(
			`SELECT id FROM users WHERE username = $1`, username,
		).Scan(&userID); err != nil {
			return err
		}
	} else {
		err = db.QueryRow(`
			INSERT INTO users (id, username, password_hash, email, real_name, status, created_at)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, 1, NOW())
			RETURNING id`,
			username, string(hash), "admin@example.com", "系统管理员",
		).Scan(&userID)
		if err != nil {
			return err
		}
	}

	_, err = db.Exec(`
		INSERT INTO user_roles (user_id, role_id, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id, role_id) DO NOTHING`,
		userID, adminRoleID,
	)
	return err
}

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole enum
type UserRole string

const (
	RoleAdmin   UserRole = "admin"
	RoleManager UserRole = "manager"
	RoleCashier UserRole = "cashier"
	RoleStaff   UserRole = "staff"
)

// UserStatus enum
type UserStatus string

const (
	StatusActive    UserStatus = "active"
	StatusInactive  UserStatus = "inactive"
	StatusSuspended UserStatus = "suspended"
)

// User model
type User struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Username  string     `gorm:"type:varchar(50);uniqueIndex;notNull" json:"username"`
	Email     string     `gorm:"type:varchar(100);uniqueIndex;notNull" json:"email"`
	Password  string     `gorm:"type:text;notNull" json:"-"`
	FullName  string     `gorm:"type:varchar(100);notNull" json:"full_name"`
	Role      UserRole   `gorm:"type:user_role;notNull;default:'staff'" json:"role"`
	Status    UserStatus `gorm:"type:user_status;notNull;default:'active'" json:"status"`
	CompanyID uuid.UUID  `gorm:"type:uuid;notNull" json:"company_id"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	LastLogin *time.Time `json:"last_login,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// UserSession model
type UserSession struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;notNull;index" json:"user_id"`
	Token     string    `gorm:"type:text;uniqueIndex;notNull" json:"token"`
	IPAddress string    `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	UserAgent string    `gorm:"type:text" json:"user_agent,omitempty"`
	ExpiresAt time.Time `gorm:"notNull" json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (s *UserSession) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// EmailVerification model
type EmailVerification struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID  `gorm:"type:uuid;notNull;index" json:"user_id"`
	Token      string     `gorm:"type:varchar(255);uniqueIndex;notNull" json:"token"`
	ExpiresAt  time.Time  `gorm:"notNull" json:"expires_at"`
	IsVerified bool       `gorm:"default:false;notNull" json:"is_verified"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
}

func (e *EmailVerification) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

// PasswordReset model
type PasswordReset struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;notNull;index" json:"user_id"`
	Token     string     `gorm:"type:varchar(255);uniqueIndex;notNull" json:"token"`
	ExpiresAt time.Time  `gorm:"notNull" json:"expires_at"`
	IsUsed    bool       `gorm:"default:false;notNull" json:"is_used"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
}

func (p *PasswordReset) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// TableName overrides
func (User) TableName() string {
	return "users"
}

func (UserSession) TableName() string {
	return "user_sessions"
}

func (EmailVerification) TableName() string {
	return "email_verifications"
}

func (PasswordReset) TableName() string {
	return "password_resets"
}

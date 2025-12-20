// Package auth provides user repository for authentication
package auth

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository handles user data operations
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// CreateUser creates a new user in the database
func (r *UserRepository) CreateUser(user *User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	if len(user.Roles) == 0 {
		user.Roles = GetDefaultRoles()
	}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(id string) (*User, error) {
	var user User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(email string) (*User, error) {
	var user User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (r *UserRepository) GetUserByUsername(username string) (*User, error) {
	var user User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

// UpdateUser updates an existing user
func (r *UserRepository) UpdateUser(user *User) error {
	user.UpdatedAt = time.Now()

	if err := r.db.Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// DeleteUser deletes a user by ID
func (r *UserRepository) DeleteUser(id string) error {
	result := r.db.Where("id = ?", id).Delete(&User{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UserExists checks if a user exists with the given email or username
func (r *UserRepository) UserExists(email, username string) (bool, error) {
	var count int64
	if err := r.db.Model(&User{}).Where("email = ? OR username = ?", email, username).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}

	return count > 0, nil
}

// ListUsers retrieves all users with pagination
func (r *UserRepository) ListUsers(limit, offset int) ([]User, error) {
	var users []User
	if err := r.db.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

// CountUsers returns the total number of users
func (r *UserRepository) CountUsers() (int64, error) {
	var count int64
	if err := r.db.Model(&User{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// UpdateUserRoles updates a user's roles
func (r *UserRepository) UpdateUserRoles(userID string, roles []string) error {
	user, err := r.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.Roles = roles
	user.UpdatedAt = time.Now()

	if err := r.db.Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user roles: %w", err)
	}

	return nil
}

// GetUsersByRole retrieves all users with a specific role
func (r *UserRepository) GetUsersByRole(role string) ([]User, error) {
	var users []User
	if err := r.db.Where("roles @> ?", fmt.Sprintf("[\"%s\"]", role)).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get users by role: %w", err)
	}

	return users, nil
}

// CreateInitialAdminUser creates the initial admin user if no users exist
func (r *UserRepository) CreateInitialAdminUser(email, username, password string, authService *AuthService) error {
	count, err := r.CountUsers()
	if err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}

	if count > 0 {
		return nil // Users already exist, no need to create admin
	}

	hashedPassword, err := authService.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	adminUser := &User{
		ID:           uuid.New().String(),
		Email:        email,
		Username:     username,
		PasswordHash: hashedPassword,
		Roles:        []string{RoleAdmin, RoleUser},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := r.CreateUser(adminUser); err != nil {
		return fmt.Errorf("failed to create initial admin user: %w", err)
	}

	return nil
}

// MigrateUserTable creates the users table if it doesn't exist
func (r *UserRepository) MigrateUserTable() error {
	if err := r.db.AutoMigrate(&User{}); err != nil {
		return fmt.Errorf("failed to migrate user table: %w", err)
	}

	return nil
}

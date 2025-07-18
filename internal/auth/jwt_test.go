// Package auth provides unit tests for JWT authentication
package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthService_HashPassword(t *testing.T) {
	authService := NewAuthService("test-secret-key", 24*time.Hour)
	
	password := "testpassword123"
	hash, err := authService.HashPassword(password)
	
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
}

func TestAuthService_CheckPassword(t *testing.T) {
	authService := NewAuthService("test-secret-key", 24*time.Hour)
	
	password := "testpassword123"
	hash, err := authService.HashPassword(password)
	require.NoError(t, err)
	
	// Test correct password
	assert.True(t, authService.CheckPassword(password, hash))
	
	// Test incorrect password
	assert.False(t, authService.CheckPassword("wrongpassword", hash))
}

func TestAuthService_GenerateToken(t *testing.T) {
	authService := NewAuthService("test-secret-key", 24*time.Hour)
	
	user := User{
		ID:       "test-user-id",
		Email:    "test@example.com",
		Username: "testuser",
		Roles:    []string{RoleUser},
	}
	
	token, err := authService.GenerateToken(user)
	
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestAuthService_ValidateToken(t *testing.T) {
	authService := NewAuthService("test-secret-key", 24*time.Hour)
	
	user := User{
		ID:       "test-user-id",
		Email:    "test@example.com",
		Username: "testuser",
		Roles:    []string{RoleUser},
	}
	
	token, err := authService.GenerateToken(user)
	require.NoError(t, err)
	
	claims, err := authService.ValidateToken(token)
	require.NoError(t, err)
	
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.Username, claims.Username)
	assert.Equal(t, user.Roles, claims.Roles)
}

func TestAuthService_ValidateExpiredToken(t *testing.T) {
	// Create service with very short token duration
	authService := NewAuthService("test-secret-key", 1*time.Nanosecond)
	
	user := User{
		ID:       "test-user-id",
		Email:    "test@example.com",
		Username: "testuser",
		Roles:    []string{RoleUser},
	}
	
	token, err := authService.GenerateToken(user)
	require.NoError(t, err)
	
	// Wait for token to expire
	time.Sleep(2 * time.Nanosecond)
	
	_, err = authService.ValidateToken(token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token is expired")
}

func TestAuthService_HasRole(t *testing.T) {
	authService := NewAuthService("test-secret-key", 24*time.Hour)
	
	claims := &UserClaims{
		Roles: []string{RoleUser, RoleDeveloper},
	}
	
	assert.True(t, authService.HasRole(claims, RoleUser))
	assert.True(t, authService.HasRole(claims, RoleDeveloper))
	assert.False(t, authService.HasRole(claims, RoleAdmin))
}

func TestAuthService_HasAnyRole(t *testing.T) {
	authService := NewAuthService("test-secret-key", 24*time.Hour)
	
	claims := &UserClaims{
		Roles: []string{RoleUser},
	}
	
	assert.True(t, authService.HasAnyRole(claims, []string{RoleUser, RoleAdmin}))
	assert.True(t, authService.HasAnyRole(claims, []string{RoleUser}))
	assert.False(t, authService.HasAnyRole(claims, []string{RoleAdmin, RoleDeveloper}))
}

func TestGetDefaultRoles(t *testing.T) {
	roles := GetDefaultRoles()
	
	assert.NotEmpty(t, roles)
	assert.Contains(t, roles, RoleUser)
}

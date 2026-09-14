package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUserRole(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    UserRole
		wantErr bool
	}{
		{
			name:    "valid user role",
			input:   "user",
			want:    UserRoleUser,
			wantErr: false,
		},
		{
			name:    "valid admin role",
			input:   "admin",
			want:    UserRoleAdmin,
			wantErr: false,
		},
		{
			name:    "invalid role",
			input:   "moderator",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty role",
			input:   "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "uppercase role",
			input:   "Admin",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewUserRole(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, UserRole(""), got)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUserRole_IsAdmin(t *testing.T) {
	tests := []struct {
		name string
		role UserRole
		want bool
	}{
		{
			name: "admin is admin",
			role: UserRoleAdmin,
			want: true,
		},
		{
			name: "user is not admin",
			role: UserRoleUser,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.role.IsAdmin())
		})
	}
}

func TestUserRole_String(t *testing.T) {
	tests := []struct {
		name string
		role UserRole
		want string
	}{
		{
			name: "user string",
			role: UserRoleUser,
			want: "user",
		},
		{
			name: "admin string",
			role: UserRoleAdmin,
			want: "admin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.role.String())
		})
	}
}

package repository

import (
	"context"
	"gorm.io/gorm"
	"service-account/internal/domain"
)

type User struct {
	db *gorm.DB
}

func NewUsersRepo(db *gorm.DB) *User {
	return &User{db}
}

// Create user.
func (r *User) Create(ctx context.Context, user *domain.User) error {
	db := r.db.WithContext(ctx).Table("tb_users").Create(user)
	if db.Error != nil {
		return db.Error
	}

	// user.ID - returns inserted data's primary key

	return nil
}

func (r *User) GetByCredentials(ctx context.Context, email string) (*domain.User, error) {
	user := new(domain.User)
	db := r.db.WithContext(ctx).Table("tb_users").Where("email = ?", email).Take(user)
	if db.Error != nil {
		return nil, db.Error
	}

	return user, nil
}

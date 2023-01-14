package repository

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"service-account/internal/domain"
)

var (
	ErrRecordNotFound     = errors.New("Record not found")
	ErrRecordAlreadyExist = errors.New("Record already exist")
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserById(ctx context.Context, id uint32) (*domain.User, error)
}

type UserRepositoryGorm struct {
	db *gorm.DB
}

var _ UserRepository = &UserRepositoryGorm{}

func NewUsersRepo(db *gorm.DB) *UserRepositoryGorm {
	return &UserRepositoryGorm{db}
}

// Create user.
func (r *UserRepositoryGorm) Create(ctx context.Context, user *domain.User) error {
	db := r.db.WithContext(ctx).Table("tb_users").Create(user)
	if db.Error != nil {
		return ErrRecordAlreadyExist
	}

	// user.ID - returns inserted data's primary key

	return nil
}

func (r *UserRepositoryGorm) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user := new(domain.User)
	db := r.db.WithContext(ctx).Table("tb_users").Where("email = ?", email).Take(user)
	if db.Error != nil {
		if errors.Is(db.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}

		return nil, db.Error
	}

	return user, nil
}

func (r *UserRepositoryGorm) GetUserById(ctx context.Context, id uint32) (*domain.User, error) {
	user := new(domain.User)
	db := r.db.WithContext(ctx).Table("tb_users").Where("id = ?", id).Take(user)
	if db.Error != nil {
		if errors.Is(db.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}

		return nil, db.Error
	}

	return user, nil
}

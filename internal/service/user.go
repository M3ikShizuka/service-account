package service

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	"service-account/internal/domain"
	"service-account/internal/repository"
	"time"
)

var (
	ErrUserNotFound      = errors.New("User not found")
	ErrPasswordIncorrect = errors.New("Password is incorrect")
)

type UserSignUpInput struct {
	Username string
	Email    string
	Password string
}

type UserSignInInput struct {
	Email    string
	Password string
}

type UserService struct {
	repo   *repository.User
	hasher Hasher
}

func NewUserSerices(userRepo *repository.User, hasher Hasher) *UserService {
	return &UserService{
		repo:   userRepo,
		hasher: hasher,
	}
}

func (s *UserService) SignUp(ctx context.Context, inputUserData *UserSignUpInput) error {
	// Hashing password.
	// TODO: get password salt from config file and .env.
	SALT := []byte("1234567890qwerty")
	passwordHash := s.hasher.Hash(inputUserData.Password, SALT)

	// Prepare user data.
	user := &domain.User{
		Username:         inputUserData.Username,
		Email:            inputUserData.Email,
		PasswordHash:     passwordHash,
		DateRegistration: time.Now(),
		DateLastOnline:   time.Now(),
	}

	// Create record in database.
	if err := s.repo.Create(ctx, user); err != nil {
		return err
	}

	return nil
}

func (s *UserService) SignIn(ctx context.Context, inputUserData *UserSignInInput) error {
	// Hashing password.
	// TODO: get password salt from config file and .env.
	SALT := []byte("1234567890qwerty")
	passwordHash := s.hasher.Hash(inputUserData.Password, SALT)

	// Get user record from database.
	user := &domain.User{}
	user, err := s.repo.GetByCredentials(ctx, inputUserData.Email)
	if err != nil {
		return ErrUserNotFound
	}

	// Check password hash.
	if bytes.Compare(user.PasswordHash, passwordHash) != 0 {
		// Not equal!
		return ErrPasswordIncorrect
	}

	// User sign in.
	return nil
}

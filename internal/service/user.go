package service

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	"service-account/internal/config"
	"service-account/internal/domain"
	"service-account/internal/repository"
	"time"
)

var (
	ErrUserNotFound      = errors.New("UserRepositoryGorm not found")
	ErrPasswordIncorrect = errors.New("Password is incorrect")
	ErrUserAlreadyExist  = errors.New("UserRepositoryGorm already exist")
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
	repo   UserRepository
	hasher Hasher
	config *config.Config
}

func NewUserSerices(userRepo UserRepository, hasher Hasher, config *config.Config) *UserService {
	return &UserService{
		repo:   userRepo,
		hasher: hasher,
		config: config,
	}
}

func (s *UserService) SignUp(ctx context.Context, inputUserData *UserSignUpInput) error {
	// Hashing password.
	// TODO: get password salt from config file and .env.
	passwordHash := s.hasher.Hash(inputUserData.Password, []byte(s.config.DB.Salt))

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
		if errors.Is(err, repository.ErrRecordAlreadyExist) {
			return ErrUserAlreadyExist
		}

		return err
	}

	return nil
}

func (s *UserService) SignIn(ctx context.Context, inputUserData *UserSignInInput) (*domain.User, error) {
	// Hashing password.
	// TODO: get password salt from config file and .env.
	passwordHash := s.hasher.Hash(inputUserData.Password, []byte(s.config.DB.Salt))

	// Get user record from database.
	user := &domain.User{}
	user, err := s.repo.GetUserByEmail(ctx, inputUserData.Email)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	// Check password hash.
	if bytes.Compare(user.PasswordHash, passwordHash) != 0 {
		// Not equal!
		return nil, ErrPasswordIncorrect
	}

	// UserRepositoryGorm sign in.
	return user, nil
}

func (s *UserService) GetUserById(ctx context.Context, id uint32) (*domain.User, error) {
	// Get user info.
	// Get user record from database.
	user := &domain.User{}
	user, err := s.repo.GetUserById(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return user, nil
}

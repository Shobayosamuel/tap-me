package auth

import (
	"errors"

	"github.com/Shobayosamuel/tap-me/internal/models"
	"github.com/Shobayosamuel/tap-me/internal/repository"
	"github.com/Shobayosamuel/tap-me/internal/utils"
)


type Service interface {
	Register(req RegisterRequest) (*TokenResponse, error)
	Login(req LoginRequest) (*TokenResponse, error)
	RefreshToken(refreshToken string) (*TokenResponse, error)
	GetUserFromToken(tokenString string) (*models.User, error)
}

type service struct {
	userRepo repository.UserRepository
}

func NewService(userRepo repository.UserRepository) Service {
	return &service{
		userRepo: userRepo,
	}
}

func (s *service) Register(req RegisterRequest) (*TokenResponse, error) {
	// check if user exists
	if _, err := s.userRepo.GetByEmail(req.Email); err == nil {
		return nil, errors.New("email already exists")
	}
	if _, err := s.userRepo.GetByUsername(req.Username); err == nil {
		return nil, errors.New("username already exists")
	}

	// hash password
	hashed_password, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// create the user
	user := &models.User{
		Username: req.Username,
		Email: req.Email,
		Password: hashed_password,
		IsActive: true,
	}
	if err := s.userRepo.Create(user);err == nil {
		return nil, errors.New("failed to create new user")
	}

	// return generated token
	return s.generateTokens(*user)

}

func (s *service) Login(req LoginRequest) (*TokenResponse, error) {
	// get user by username
	user, err := s.userRepo.GetByUsername(req.Username)

	if err != nil {
		return nil, errors.New("failed to hash password")
	}
	if !utils.CheckPassword(req.Password, user.Password) {
		return nil, errors.New("Incorrect credentials")
	}

	if !user.IsActive {
		return nil, errors.New("User is not activated")
	}
	// return generated token
	return s.generateTokens(*user)

}

func (s *service) generateTokens(user models.User) (*TokenResponse, error) {
	accessToken, err := utils.GenerateAccessToken(user)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	refreshToken, err := utils.GenerateRefreshToken(user)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(utils.GetAccessTokenTTL().Seconds()),
	}, nil
}

func (s *service) RefreshToken(refreshToken string) (*TokenResponse, error) {
	// Validate refresh token
	claims, err := utils.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Get user
	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Generate new tokens
	return s.generateTokens(*user)
}

func (s *service) GetUserFromToken(tokenString string) (*models.User, error) {
	claims, err := utils.ValidateAccessToken(tokenString)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}
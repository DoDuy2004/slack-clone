package service

import (
	"errors"

	"github.com/DoDuy2004/slack-clone/backend/internal/models"
	"github.com/DoDuy2004/slack-clone/backend/internal/models/dto"
	"github.com/DoDuy2004/slack-clone/backend/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserService interface {
	UpdateProfile(userID uuid.UUID, req *dto.UpdateProfileRequest) (*models.User, error)
	GetProfile(userID uuid.UUID) (*models.User, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) UpdateProfile(userID uuid.UUID, req *dto.UpdateProfileRequest) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if req.FullName != nil {
		user.FullName = req.FullName
	}
	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}
	if req.StatusMessage != nil {
		user.StatusMessage = req.StatusMessage
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) GetProfile(userID uuid.UUID) (*models.User, error) {
	return s.userRepo.FindByID(userID)
}

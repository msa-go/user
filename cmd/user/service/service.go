package service

import (
	// golang package
	"context"
	"user/cmd/user/repository"
	"user/models"
)

type UserService struct {
	UserRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{
		UserRepo: userRepo,
	}
}

/* 이메일을 통한 유저 조회 */
func (svc *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := svc.UserRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

/* 유저 ID를 통한 유저 조회 */
func (svc *UserService) GetUserByID(ctx context.Context, userID int64) (*models.User, error) {
	user, err := svc.UserRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

/* 새로운 유저 등록 */
func (svc *UserService) CreateNewUser(ctx context.Context, user *models.User) (int64, error) {
	userID, err := svc.UserRepo.InsertNewUser(ctx, user)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

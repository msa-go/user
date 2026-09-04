package repository

import (
	// golang package
	"context"
	"errors"
	"user/models"

	// external package
	"gorm.io/gorm"
)

/* Email을 통한 유저 조회 */
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.Database.WithContext(ctx).Where("email = ?", email).Last(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, models.ErrUserNotFound
		}

		return nil, err
	}

	return &user, nil
}

/* 유저 ID를 통한 유저 조회 */
func (r *UserRepository) FindByUserID(ctx context.Context, userID int64) (*models.User, error) {
	var user models.User
	err := r.Database.WithContext(ctx).Where("id = ?", userID).Last(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, models.ErrUserNotFound
		}

		return nil, err
	}

	return &user, nil
}

/* 새 유저 등록 */
func (r *UserRepository) InsertNewUser(ctx context.Context, user *models.User) (int64, error) {
	err := r.Database.WithContext(ctx).Create(user).Error
	if err != nil {
		return 0, err
	}

	return user.ID, nil
}

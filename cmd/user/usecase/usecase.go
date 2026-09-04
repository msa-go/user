package usecase

import (
	// golang package
	"context"
	"errors"
	"time"
	"user/cmd/user/service"
	"user/infrastructure/log"
	"user/models"
	"user/utils"

	// external package
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

// UserUsecase orchestrates user-related business scenarios (login, registration)
// on top of UserService. Both the REST handler and the gRPC handler depend on
// this interface so they can share the same business logic.
type UserUsecase interface {
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, userID int64) (*models.User, error)
	Register(ctx context.Context, user *models.User) error
	Login(ctx context.Context, param models.LoginParameter, userID int64, storedPassword string) (string, error)
}

type userUsecase struct {
	UserService service.UserService
	JWTSecret   string
}

func NewUserUsecase(userService service.UserService, jwtSecret string) UserUsecase {
	return &userUsecase{
		UserService: userService,
		JWTSecret:   jwtSecret,
	}
}

func (uc *userUsecase) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	tracer := otel.Tracer("userfc-Usecase")
	ctx, span := tracer.Start(ctx, "GetUserByEmail")
	defer span.End()

	user, err := uc.UserService.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (uc *userUsecase) GetUserByID(ctx context.Context, userID int64) (*models.User, error) {
	trace := otel.Tracer("userfc-Usecase")
	ctx, span := trace.Start(ctx, "GetUserByID")
	defer span.End()

	user, err := uc.UserService.GetUserByID(ctx, userID)
	if err != nil {
		log.LogWithTrace(ctx)
		return nil, err
	}

	return user, nil
}

// Register creates a new user from the given user pointer of models.User.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (uc *userUsecase) Register(ctx context.Context, user *models.User) error {
	// hash password
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"email": user.Email,
		}).Errorf("utils.HashPassword() got error %v", err)
		return err
	}

	// insert db
	user.Password = hashedPassword
	_, err = uc.UserService.CreateNewUser(ctx, user)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"email": user.Email,
			"name":  user.Name,
		}).Errorf("uc.UserService.CreateNewUser() got error %v", err)
		return err
	}

	return nil
}

func (uc *userUsecase) Login(ctx context.Context, param models.LoginParameter, userID int64, storedPassword string) (string, error) {
	tracer := otel.Tracer("userfc-Usecase")
	ctx, span := tracer.Start(ctx, "Login")
	defer span.End()

	isMatch, err := utils.CheckPasswordHash(storedPassword, param.Password)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"email": param.Email,
		}).Errorf("utils.CheckPasswordHash got error: %v", err)
	}

	if !isMatch {
		return "", errors.New("Email atau password salah")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(uc.JWTSecret))
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"email": param.Email,
		}).Errorf("token.SignedString got error: %v", err)
		return "", err
	}

	return tokenString, nil
}

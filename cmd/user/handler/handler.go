package handler

import (
	// golang package
	"errors"
	"net/http"
	"user/cmd/user/usecase"
	"user/infrastructure/log"
	"user/models"

	// external package
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
)

type UserHandler struct {
	UserUsecase usecase.UserUsecase
}

func NewUserHandler(useUsecase usecase.UserUsecase) *UserHandler {
	return &UserHandler{
		UserUsecase: useUsecase,
	}
}

/* 사용자 로그인 */
func (h *UserHandler) Login(c *gin.Context) {
	trace := otel.Tracer("User-handler")
	ctx, span := trace.Start(c.Request.Context(), "HandleLogin")
	defer span.End()

	// 유효성 검사 1 - 로그인 요청 파라미터(LoginParameter, DTO) 만족
	var param models.LoginParameter
	if err := c.ShouldBindJSON(&param); err != nil {
		log.Logger.Info(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error_message": "잘못된 요청 파라미터 입니다.",
		})
		return
	}

	// 유효성 검사 2 - 비밀번호 길이 8자 이상
	if len(param.Password) < 8 {
		log.Logger.Info("Invalid Input")
		c.JSON(http.StatusBadRequest, gin.H{
			"error_message": "비밀번호는 8자 이상이어야 합니다.",
		})
		return
	}

	// 유효성 검사 3 - 존재하지 않는 유저 확인
	user, err := h.UserUsecase.GetUserByEmail(ctx, param.Email)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error_message": "존재하지 않는 사용자입니다.",
			})
			return
		}

		log.Logger.Error("Error login", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error_message": err.Error(),
		})
		return
	}

	token, err := h.UserUsecase.Login(ctx, param, user.ID, user.Password)
	if err != nil {
		log.Logger.Error(err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "잘못된 이메일 또는 비밀번호 입니다.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

/* 회원가입 */
func (h *UserHandler) Register(c *gin.Context) {
	// trace := otel.Tracer("User-handler")
	// ctx, span := trace.Start(c.Request.Context(), "HandleRegister")
	// defer span.End()

	var param models.RegisterParameter

	if err := c.ShouldBindJSON(&param); err != nil {
		log.Logger.Info(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error_message": "잘못된 요청 파라미터 입니다.",
		})
		return
	}

	if len(param.Password) < 8 ||
		len(param.ConfirmPassword) < 8 {
		log.Logger.Info("Invalid Input")
		c.JSON(http.StatusBadRequest, gin.H{
			"error_message": "비밀번호는 8자 이상이어야 합니다.",
		})
		return
	}

	if param.Password != param.ConfirmPassword {
		log.Logger.Info("Invalid Credentials")
		c.JSON(http.StatusBadRequest, gin.H{
			"error_message": "비밀번호와 비밀번호 확인이 일치하지 않습니다.",
		})
		return
	}

	err := h.UserUsecase.Register(c.Request.Context(), &models.User{
		Name:     param.Name,
		Email:    param.Email,
		Password: param.Password, // plain text --> hashing password (bcrypt)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error_message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "회원가입 성공",
	})
}

/* 유저 정보 조회 */
func (h *UserHandler) GetUserInfo(c *gin.Context) {

	// 인증 정보에서 유저 ID 추출
	userIDStr, isExist := c.Get("user_id")
	if !isExist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error_message": "인증되지 않은 사용자",
		})
		return
	}

	// float64 -> int64 형변환
	userID, ok := userIDStr.(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error_message": "유효하지 않은 유저 ID",
		})
		return
	}

	// 유효성 검사 1 - 유저 ID를 통한 유저 조회
	user, err := h.UserUsecase.GetUserByID(c.Request.Context(), int64(userID))
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error_message": "존재하지 않는 사용자입니다.",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error_message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":  user.Name,
		"email": user.Email,
		// profile picture
	})
}

/* 핑 테스트 */
func (h *UserHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "OK",
	})
}

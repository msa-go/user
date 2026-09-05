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

// genericLoginError 는 이메일 미존재/비밀번호 불일치를 구분하지 않는 메시지다.
// 둘을 구분해서 응답하면 공격자가 응답 문구만으로 가입 여부를 알아낼 수 있다(계정 목록화).
const genericLoginError = "이메일 또는 비밀번호가 올바르지 않습니다."

// internalErrorMessage 는 클라이언트에 내려주는 고정 메시지다.
// DB/드라이버 에러 문자열을 그대로 노출하면 내부 구현이 드러날 수 있어,
// 상세는 서버 로그로만 남기고 클라이언트에는 이 메시지만 반환한다.
const internalErrorMessage = "요청을 처리하는 중 오류가 발생했습니다."

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

	// 유효성 검사 3 - 유저 조회
	//
	// 이메일 미존재와 비밀번호 불일치를 같은 응답(genericLoginError, 401)으로
	// 처리한다. 구분해서 응답하면 계정 존재 여부가 노출된다(user enumeration).
	user, err := h.UserUsecase.GetUserByEmail(ctx, param.Email)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error_message": genericLoginError,
			})
			return
		}

		log.LogWithTrace(ctx).Errorf("h.UserUsecase.GetUserByEmail() got error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error_message": internalErrorMessage,
		})
		return
	}

	token, err := h.UserUsecase.Login(ctx, param, user.ID, user.Password)
	if err != nil {
		log.LogWithTrace(ctx).Infof("login failed for %s: %v", param.Email, err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error_message": genericLoginError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

/* 회원가입 */
func (h *UserHandler) Register(c *gin.Context) {
	trace := otel.Tracer("User-handler")
	ctx, span := trace.Start(c.Request.Context(), "HandleRegister")
	defer span.End()

	var param models.RegisterParameter

	if err := c.ShouldBindJSON(&param); err != nil {
		log.Logger.Info(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error_message": "잘못된 요청 파라미터 입니다.",
		})
		return
	}

	if len(param.Password) < 8 || len(param.ConfirmPassword) < 8 {
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

	err := h.UserUsecase.Register(ctx, &models.User{
		Name:     param.Name,
		Email:    param.Email,
		Password: param.Password, // plain text --> hashing password (bcrypt)
	})

	if err != nil {
		log.LogWithTrace(ctx).Errorf("h.UserUsecase.Register() got error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error_message": internalErrorMessage,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "회원가입 성공",
	})
}

/* 유저 정보 조회 */
func (h *UserHandler) GetUserInfo(c *gin.Context) {
	ctx := c.Request.Context()

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
	user, err := h.UserUsecase.GetUserByID(ctx, int64(userID))
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error_message": "존재하지 않는 사용자입니다.",
			})
			return
		}

		log.LogWithTrace(ctx).Errorf("h.UserUsecase.GetUserByID() got error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error_message": internalErrorMessage,
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

package middleware

import (
	// golang package
	"context"
	"time"
	"user/infrastructure/log"

	// external package
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// contextKey 는 컨텍스트 키 충돌을 막기 위한 전용 타입이다.
// 문자열 리터럴을 키로 쓰면 다른 패키지의 키와 충돌할 수 있다.
type contextKey string

const RequestIDKey contextKey = "request_id"

// RequestLogger 는 요청마다 request_id 를 발급해 컨텍스트에 싣고,
// 처리 결과를 구조화 로그로 남긴다. timeout 은 설정에서 주입받는다.
func RequestLogger(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()

		// 클라이언트 연결 종료 시 함께 취소되도록 요청 컨텍스트에서 파생시킨다.
		timeoutCtx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		ctx := context.WithValue(timeoutCtx, RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		startTime := time.Now()
		c.Next()
		latency := time.Since(startTime)

		requestLog := logrus.Fields{
			"request_id": requestID,
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"status":     c.Writer.Status(),
			"latency":    latency,
		}

		if c.Writer.Status() < 400 {
			log.Logger.WithFields(requestLog).Info("Request Success")
		} else {
			log.Logger.WithFields(requestLog).Error("Request Error")
		}
	}
}

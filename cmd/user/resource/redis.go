package resource

import (
	// golang package

	"context"
	"fmt"
	"log"
	"user/config"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis(cfg *config.Config) *redis.Client {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
	})

	/*
		앱이 켜지는 시점(main()에서 서버 기동 전) 코드
			-> HTTP 요청도 없음 (→ c.Request.Context() 같은 게 존재하지 않음),  상위 함수가 넘겨준 ctx도 없음
			즉, 파생시킬 부모 컨텍스트가 없음 따라서 Background를 통해서 시작점 생성
	*/
	ctx := context.Background()              // 새 컨텍스트 값 생성 — 그냥 값 하나 만든 것뿐
	_, err := RedisClient.Ping(ctx).Result() // blocking, 끝날 때까지 기다림
	if err != nil {
		log.Fatalf("failed connect to redis: %v", err)
	}

	log.Println("Connected to Redis")
	return RedisClient
}

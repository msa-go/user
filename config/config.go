package config

import (
	// golang package
	"log"
	"strings"

	// external package
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// envPrefix 는 설정을 덮어쓸 환경변수의 접두사다.
// 예) secret.jwtsecret -> USER_SECRET_JWTSECRET
const envPrefix = "USER"

// secretKeys 는 yaml 에 값을 두지 않고 환경변수로만 주입받는 키 목록이다.
//
// AutomaticEnv 는 viper 가 이미 아는 키(yaml 에 존재하거나 기본값이 설정된 키)에만
// 적용된다. yaml 에서 제거한 키는 AllKeys() 에 없어 Unmarshal 이 건너뛰므로,
// 환경변수를 설정해도 조용히 무시된다. 따라서 명시적으로 BindEnv 해야 한다.
var secretKeys = []string{
	"database.password",
	"redis.password",
	"secret.jwtsecret",
}

func LoadConfig() Config {
	var cfg Config

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./files/config")

	// 환경변수가 yaml 값보다 우선한다.
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	for _, key := range secretKeys {
		if err := viper.BindEnv(key); err != nil {
			log.Fatalf("error bind env %q: %v", key, err)
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("error read config file: %v", err)
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("error unmarshal config: %v", err)
	}

	// validate 태그는 검증기를 직접 호출해야 동작한다.
	// 설정 누락을 런타임이 아닌 기동 시점에 잡는다.
	if err := validator.New().Struct(cfg); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	return cfg
}

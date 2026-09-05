package config

// viper 는 mapstructure 를 사용하므로 yaml 태그가 아닌 mapstructure 태그를 읽는다.
type Config struct {
	App      AppConfig      `mapstructure:"app" validate:"required"`
	Database DatabaseConfig `mapstructure:"database" validate:"required"`
	Redis    RedisConfig    `mapstructure:"redis" validate:"required"`
	Secret   SecretConfig   `mapstructure:"secret" validate:"required"`
}

type AppConfig struct {
	Port string `mapstructure:"port" validate:"required"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host" validate:"required"`
	User     string `mapstructure:"user" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
	Port     string `mapstructure:"port" validate:"required"`
	Name     string `mapstructure:"name" validate:"required"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host" validate:"required"`
	Port     string `mapstructure:"port" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
}

type SecretConfig struct {
	JWTSecret string `mapstructure:"jwtsecret" validate:"required,min=32"`
}

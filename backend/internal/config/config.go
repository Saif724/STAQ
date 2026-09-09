package config

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Email    EmailConfig
}

type AppConfig struct {
	Env         string
	Port        string
	FrontendURL string
	BackendURL  string
}

type DatabaseConfig struct {
	URL string
}

type RedisConfig struct {
	Address  string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret string
}

type EmailConfig struct {
	APIKey string
	From   string
}

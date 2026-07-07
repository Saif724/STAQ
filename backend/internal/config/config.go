package config

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	SMTP     SMTPConfig
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

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

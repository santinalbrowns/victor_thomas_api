package config

import "syscall"

type Config struct {
	PORT                 string
	DB_USER              string
	DB_PSWD              string
	DB_HOST              string
	DB_PORT              string
	DB_NAME              string
	JWT_CERT_PATH        string
	JWT_PUB_CERT_PATH    string
	REDIS                string
	REDIS_PSWD           string
	TWILIO_SID           string
	TWILIO_TOKEN         string
	TWILIO_SERVICE_SID   string
	IMAGES_PATH          string
	THUMBNAILS_PATH      string
	PAYCHANGU_SECRET_KEY string
	PAYCHANGU_PUBLIC_KEY string
}

func Load() *Config {
	return &Config{
		PORT:    Get("PORT"),
		DB_USER: Get("DB_USER"),
		DB_PSWD: Get("DB_PSWD"),
		DB_HOST: Get("DB_HOST"),
		DB_PORT: Get("DB_PORT"),
		DB_NAME: Get("DB_NAME"),
	}
}

func Get(key string) string {
	if value, oky := syscall.Getenv(key); oky {
		return value
	}

	panic("failed to get env: " + key)
}

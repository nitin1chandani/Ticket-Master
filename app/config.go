package app

import (
	"os"
	"strconv"
)

func loadConfig() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	jwtExpiryHours := 24
	if raw := os.Getenv("JWT_EXPIRY_HOURS"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err != nil && parsed > 0 {
			jwtExpiryHours = parsed
		}
	}

	return Config{
		Port:           port,
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		RedisAddr:      redisAddr,
		RedisPassword:  os.Getenv("REDIS_PASSWORD"),
		RedisDB:        0,
		JWTSecret:      os.Getenv("JWT_SECRET"),
		JWTExpiryHours: jwtExpiryHours,
	}
}

package app

import "os"

func loadConfig() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	return Config{
		Port:          port,
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		RedisAddr:     redisAddr,
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       0,
	}
}

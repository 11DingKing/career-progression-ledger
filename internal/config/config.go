package config

import (
	"os"
	"time"
)

type Config struct {
	HTTPAddr       string
	DBPath         string
	SessionTTL     time.Duration
	WorkerInterval time.Duration
}

func Load() Config {
	return Config{HTTPAddr: env("HTTP_ADDR", ":8080"), DBPath: env("DB_PATH", "career-progression.db"), SessionTTL: duration("SESSION_TTL", 12*time.Hour), WorkerInterval: duration("WORKER_INTERVAL", 5*time.Second)}
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func duration(k string, d time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if parsed, e := time.ParseDuration(v); e == nil {
			return parsed
		}
	}
	return d
}

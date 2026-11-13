package config

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Port           string
	DatabaseURL    string
	AllowedOrigins []string
	Environment    string
}

func Load() *Config {
	loadEnv()

	port := getEnv("PORT", "")
	dbURL := getEnv("DATABASE_URL", "")
	env := getEnv("APP_ENV", "")

	if port == "" {
		port = "8080"
	}

	if dbURL == "" {
		log.Println("[config] WARNING: DATABASE_URL is not set in .env or environment")
	}

	rawOrigins := getEnv("ALLOWED_ORIGINS", "")
	var origins []string
	if rawOrigins != "" {
		for _, o := range strings.Split(rawOrigins, ",") {
			trimmed := strings.TrimSpace(o)
			if trimmed != "" {
				origins = append(origins, trimmed)
			}
		}
	}

	return &Config{
		Port:           port,
		DatabaseURL:    dbURL,
		AllowedOrigins: origins,
		Environment:    env,
	}
}

func loadEnv() {
	// Cari file .env mulai dari direktori saat ini ke atas (parent directories)
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}

	current := dir
	for i := 0; i < 5; i++ {
		// Cek file .env di folder saat ini
		envPath := filepath.Join(current, ".env")
		if parseEnvFile(envPath) {
			return
		}

		// Cek file .env di subfolder backend/.env
		backendEnv := filepath.Join(current, "backend", ".env")
		if parseEnvFile(backendEnv) {
			return
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
}

func parseEnvFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	loadedCount := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)
			if key != "" {
				if _, exists := os.LookupEnv(key); !exists {
					_ = os.Setenv(key, val)
				}
				loadedCount++
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[config] error reading %s: %v", path, err)
	}

	if loadedCount > 0 {
		log.Printf("[config] Successfully loaded %d variables from %s", loadedCount, path)
		return true
	}
	return false
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}

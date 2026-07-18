package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// База данных
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Сервер
	ServerHost string
	ServerPort int

	// Источники данных
	ProductsURLPattern string
	ProductsRange      []int
	ClientsURL         string

	// Пагинация
	PageStart int
	PageLimit int
	MaxPages  int

	// Конкурентность
	MaxConcurrentRequests int
	RateLimit             int
	BurstSize             int
	RequestDelay          time.Duration

	// Настройки загрузки
	DownloadTimeout time.Duration
	MaxRetries      int
	RetryDelay      time.Duration

	// Логирование
	LogLevel string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env file not found, using system environment variables")
	}

	return &Config{
		// База данных
		DBHost:     getEnv("DB_HOST", "db"),
		DBPort:     getEnvAsInt("DB_PORT", 5432),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "testdb"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		// Сервер
		ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),
		ServerPort: getEnvAsInt("SERVER_PORT", 8080),

		// Источники данных (НОВЫЙ ФОРМАТ!)
		ProductsURLPattern: getEnv("PRODUCTS_URL_PATTERN", "https://api.dynamica.space/sources/source%d.php"),
		ProductsRange:      parseRange(getEnv("PRODUCTS_SOURCES_RANGE", "1-3")),
		ClientsURL:         getEnv("CLIENTS_URL", "https://api.dynamica.space/sources/clients.php"),

		// Пагинация
		PageStart: getEnvAsInt("PAGE_START", 1),
		PageLimit: getEnvAsInt("PAGE_LIMIT", 50),
		MaxPages:  getEnvAsInt("MAX_PAGES", 100),

		// Конкурентность
		MaxConcurrentRequests: getEnvAsInt("MAX_CONCURRENT_REQUESTS", 5),
		RateLimit:             getEnvAsInt("RATE_LIMIT", 10),
		BurstSize:             getEnvAsInt("BURST_SIZE", 5),
		RequestDelay:          getEnvAsDuration("REQUEST_DELAY", 100*time.Millisecond),

		// Загрузка
		DownloadTimeout: getEnvAsDuration("DOWNLOAD_TIMEOUT", 30*time.Second),
		MaxRetries:      getEnvAsInt("MAX_RETRIES", 3),
		RetryDelay:      getEnvAsDuration("RETRY_DELAY", 1*time.Second),

		// Логирование
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
}

// GetProductURLs генерирует список URL из шаблона и диапазона
func (c *Config) GetProductURLs() []string {
	urls := make([]string, 0, len(c.ProductsRange))
	for _, num := range c.ProductsRange {
		urls = append(urls, fmt.Sprintf(c.ProductsURLPattern, num))
	}
	return urls
}

// parseRange парсит строку вида "1-3" или "1,3,5,7" или "5"
func parseRange(rangeStr string) []int {
	if rangeStr == "" {
		return []int{1, 2, 3} // fallback
	}

	// Убираем пробелы
	rangeStr = strings.TrimSpace(rangeStr)

	// Если через запятую: "1,3,5,7"
	if strings.Contains(rangeStr, ",") {
		parts := strings.Split(rangeStr, ",")
		nums := make([]int, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if n, err := strconv.Atoi(p); err == nil && n > 0 {
				nums = append(nums, n)
			}
		}
		if len(nums) > 0 {
			return nums
		}
	}

	// Если через дефис: "1-3" или "1-10"
	if strings.Contains(rangeStr, "-") {
		parts := strings.Split(rangeStr, "-")
		if len(parts) == 2 {
			start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err1 == nil && err2 == nil && start > 0 && end >= start {
				nums := make([]int, 0, end-start+1)
				for i := start; i <= end; i++ {
					nums = append(nums, i)
				}
				return nums
			}
		}
	}

	// Если одно число: "5"
	if n, err := strconv.Atoi(rangeStr); err == nil && n > 0 {
		return []int{n}
	}

	// Fallback
	return []int{1, 2, 3}
}

// Вспомогательные функции
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return fallback
}

func (c *Config) GetDSN() string {
	return "host=" + c.DBHost +
		" port=" + strconv.Itoa(c.DBPort) +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" sslmode=" + c.DBSSLMode
}

func (c *Config) GetServerAddress() string {
	return c.ServerHost + ":" + strconv.Itoa(c.ServerPort)
}

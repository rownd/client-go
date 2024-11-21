package testing

import (
    "os"
    
    "github.com/joho/godotenv"
)

type TestConfig struct {
    AppKey    string
    AppSecret string
    AppID     string
    BaseURL   string
}

func init() {
    // Load .env file if it exists
    godotenv.Load()
}

// GetTestConfig returns test configuration from environment variables
func GetTestConfig() TestConfig {
    return TestConfig{
        AppKey:    getEnvOrDefault("ROWND_TEST_APP_KEY", ""),
        AppSecret: getEnvOrDefault("ROWND_TEST_APP_SECRET", ""),
        AppID:     getEnvOrDefault("ROWND_TEST_APP_ID", ""),
        BaseURL:   getEnvOrDefault("ROWND_TEST_BASE_URL", "https://api.rownd.io"),
    }
}

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
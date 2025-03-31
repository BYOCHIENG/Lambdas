package config

import (
    "os"
    "strings"
)

type Config struct {
    Region      string
    ToEmail     string
    FromEmail   string
    AlertEmails []string
}

func LoadConfig() Config {
    alertEmails := strings.Split(getEnv("ALERT_EMAILS", "admin@example.com"), ",")

    return Config {
        Region:      getEnv("AWS_REGION", "us-east-1"),
        ToEmail:     getEnv("TO_EMAIL", "admin@example.com"),
        FromEmail:   getEnv("FROM_EMAIL", "contact@example.com"),
        AlertEmails: alertEmails,
    }
}

func getEnv(key, defaultValue string) string {
    value := os.Getenv(key)
    
    if value = "" {
        return defaultValue
    }

    return value
}

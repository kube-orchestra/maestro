// Copyright Contributors to the Open Cluster Management project

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
)

var Cfg = new()

// Define a Config type to hold our config properties.
type Config struct {
	MQTTClientId           string // mqtt client id for the maeastro service
	MQTTBrokerURL          string // mqtt broker url
	MQTTBrokerUserName     string // mqtt broker username
	MQTTBrokerUserPassword string // mqtt broker password

	DBHost string // database host name
	DBName string // dataabse name
	DBPass string // database password
	DBPort int    // database port
	DBUser string // database user
	DBSSL  string // database ssl mode
	DBTmz  string // database timezone (need to remve this need)
}

func new() *Config {
	// If environment variables are set, use default values
	// Simply put, the order of preference is env -> default values (from left to right)
	conf := &Config{
		MQTTClientId:           getEnv("MQTT_CLIENT_ID", "maestro-server"),
		MQTTBrokerURL:          getEnv("MQTT_BROKER_URL", "tcp://localhost:1883"),
		MQTTBrokerUserName:     getEnv("MQTT_BROKER_USERNAME", "admin"),
		MQTTBrokerUserPassword: getEnv("MQTT_BROKER_PASSWORD", "password"),

		DBHost: getEnv("DB_HOST", "localhost"),
		DBName: getEnv("DB_NAME", ""),
		DBUser: getEnv("DB_USER", ""),
		DBPass: getEnv("DB_PASS", ""),
		DBPort: getEnvAsInt("DB_PORT", 5432),
		DBSSL:  getEnv("DB_SSL", "disable"),
		DBTmz:  getEnv("DB_TMZ", "America/Los_Angeles"),
	}

	conf.DBPass = url.QueryEscape(conf.DBPass)
	return conf
}

// Format and print environment to logger.
func (cfg *Config) PrintConfig() {
	// Make a copy to redact secrets and sensitive information.
	tmp := *cfg
	tmp.DBPass = "[REDACTED]"
	tmp.MQTTBrokerUserPassword = "[REDACTED]"

	// for user friendly indented printing
	prettyJSON, err := json.MarshalIndent(tmp, "", "    ")
	if err != nil {
		fmt.Println("Encountered a problem formatting configuration:", err)
		return
	}
	fmt.Println(string(prettyJSON))

}

// Validate required configuration.
func (cfg *Config) Validate() error {
	if cfg.DBName == "" {
		return errors.New("required environment DB_NAME is not set")
	}
	if cfg.DBUser == "" {
		return errors.New("required environment DB_USER is not set")
	}
	if cfg.DBPass == "" {
		return errors.New("required environment DB_PASS is not set")
	}
	return nil
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// Simple helper function to read an environment variable into integer or return a default value
func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

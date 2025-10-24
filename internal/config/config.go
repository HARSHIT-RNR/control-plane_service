package config

import (
	"time"

	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variable.
type Config struct {
	// Server configuration
	GRPCPort int `mapstructure:"GRPC_PORT"`

	// Database configuration
	DatabaseURL string `mapstructure:"DATABASE_URL"`

	// Kafka configuration
	KafkaBrokers                         string `mapstructure:"KAFKA_BROKERS"`
	KafkaConsumerGroup                   string `mapstructure:"KAFKA_CONSUMER_GROUP"`
	KafkaTopicIAMCreateInitialAdmin      string `mapstructure:"KAFKA_TOPIC_IAM_CREATE_INITIAL_ADMIN"`
	KafkaTopicUserLifecycle              string `mapstructure:"KAFKA_TOPIC_USER_LIFECYCLE"`
	KafkaTopicNotificationPasswordSetup  string `mapstructure:"KAFKA_TOPIC_NOTIFICATION_PASSWORD_SETUP"`

	// Token configuration
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`

	// SAML configuration (COMMENTED OUT FOR NOW)
	// SAMLEnabled           bool   `mapstructure:"SAML_ENABLED"`
	// SAMLEntityID          string `mapstructure:"SAML_ENTITY_ID"`
	// SAMLSSOURL            string `mapstructure:"SAML_SSO_URL"`
	// SAMLIDPMetadataURL    string `mapstructure:"SAML_IDP_METADATA_URL"`
	// SAMLCertificateFile   string `mapstructure:"SAML_CERTIFICATE_FILE"`
	// SAMLPrivateKeyFile    string `mapstructure:"SAML_PRIVATE_KEY_FILE"`
	// SAMLRootURL           string `mapstructure:"SAML_ROOT_URL"`
	// SAMLAllowIDPInitiated bool   `mapstructure:"SAML_ALLOW_IDP_INITIATED"`

	// OPA configuration
	OPAEnabled bool   `mapstructure:"OPA_ENABLED"`
	OPAURL     string `mapstructure:"OPA_URL"`
}

// Load reads configuration from environment variables and config files
func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("GRPC_PORT", 50051)
	viper.SetDefault("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/cp_db?sslmode=disable")
	viper.SetDefault("KAFKA_BROKERS", "localhost:9092")
	viper.SetDefault("KAFKA_TOPIC_IAM_CREATE_INITIAL_ADMIN", "iam.create-initial-admin")
	viper.SetDefault("KAFKA_TOPIC_USER_LIFECYCLE", "user.lifecycle")
	viper.SetDefault("KAFKA_TOPIC_NOTIFICATION_PASSWORD_SETUP", "notification.send-password-setup")
	viper.SetDefault("ACCESS_TOKEN_DURATION", 15*time.Minute)
	viper.SetDefault("REFRESH_TOKEN_DURATION", 24*time.Hour)
	viper.SetDefault("SAML_ENABLED", false)
	viper.SetDefault("OPA_ENABLED", false)

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

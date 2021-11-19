// Copyright Yahoo 2021
// Licensed under the terms of the Apache License 2.0.
// See LICENSE file in project root for terms.
package config

import (
	"fmt"
	"net/url"

	"github.com/grafeas/grafeas/go/config"
)

const emptyFieldErrTemplate = `invalid field: "%s" must not be empty`

// default values for Config
const (
	defaultPort    = 5432
	defaultDBName  = "grafeas"
	defaultSSLMode = "verify-full"
)

// Config is the configuration for PostgreSQL store.
// json tags are required because
// config.ConvertGenericConfigToSpecificType internally uses json package.
type Config struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	// For rds_prostgres, DBName has to alrady exist and can be accessed by User.
	DBName   string `json:"db_name"`
	User     string `json:"user"`
	Password string `json:"password"`
	// Valid sslmodes: disable, allow, prefer, require, verify-ca, verify-full.
	// See https://www.postgresql.org/docs/current/static/libpq-connect.html for details
	SSLMode     string `json:"ssl_mode"`
	SSLRootCert string `json:"ssl_root_cert"`
	// PaginationKey is a 32-bit URL-safe base64 key used to encrypt pagination tokens.
	// Check the underlying DB implementation to see it's supported.
	// Regarding PostgreSQL, if one is not provided, it will be generated [1].
	// Multiple grafeas instances in the same cluster must share the same value.
	//
	// [1] https://github.com/grafeas/grafeas-pgsql
	PaginationKey string `json:"pagination_key"`

	ConnPool ConnPoolConfig `json:"conn_pool"`

	// IAMAuth is only used when Password is empty.
	IAMAuth IAMAuthConfig `json:"iam_auth"`
}

func New(ci *config.StorageConfiguration) (*Config, error) {
	var c Config

	err := config.ConvertGenericConfigToSpecificType(ci, &c)
	if err != nil {
		return nil, fmt.Errorf("failed to convert the generic storage config to a rds config, err: %v", err)
	}
	c.populateDefaultValues()

	if err := c.validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Config) populateDefaultValues() {
	if c.Port == 0 {
		c.Port = defaultPort
	}
	if c.DBName == "" {
		c.DBName = defaultDBName
	}
	if c.SSLMode == "" {
		c.SSLMode = defaultSSLMode
	}
	c.IAMAuth.populateDefaultValues()
}

func (c *Config) validate() error {
	if c.Host == "" {
		return fmt.Errorf(emptyFieldErrTemplate, "Config.Host")
	}
	if c.Port <= 0 {
		return fmt.Errorf(`invalid field: "Config.Port" must be larger than zero, got %v`, c.Port)
	}
	if c.User == "" {
		return fmt.Errorf(emptyFieldErrTemplate, "Config.User")
	}
	if c.SSLRootCert == "" && (c.SSLMode == "verify-ca" || c.SSLMode == "verify-full") {
		return fmt.Errorf(`invalid field: Config.SSLRootCert must not be empty because SSLMode is %s`, c.SSLMode)
	}
	if c.Password != "" {
		return nil
	}
	// The password is empty, so IAMAuth must be valid.
	return c.IAMAuth.validate()
}

// ConnPoolConfig contains the configuration related to connection pool management.
// The explanation of each field can be found in the following functions in sql package:
//
// - SetMaxOpenConns
// - SetMaxIdleConns
// - SetConnMaxLifetime
// - SetConnMaxIdleTime
//
// Default values are not provided for these fields because
// 0 is the zero value of `int`, but it's also a valid value for these fields.
type ConnPoolConfig struct {
	MaxOpenConns             int `json:"max_open_conns"`
	MaxIdleConns             int `json:"max_idle_conns"`
	ConnMaxLifetimeInSeconds int `json:"conn_max_lifetime_in_seconds"`
	ConnMaxIdleTimeInSeconds int `json:"conn_max_idle_time_in_seconds"`
}

// IAMAuthConfig contains configuration required to
// get a temporary DB password (i.e. token) from AWS API.
type IAMAuthConfig struct {
	// Region refers to the AWS region in which the DB resides.
	Region string `json:"region"`
	// CredentialsProvider specifies how to configure the AWS credentials provider.
	CredentialsProvider ZTSCredentialProviderConfig `json:"credentials_provider"`
}

func (c *IAMAuthConfig) populateDefaultValues() {
	c.CredentialsProvider.populateDefaultValues()
}

func (c *IAMAuthConfig) validate() error {
	if c.Region == "" {
		return fmt.Errorf(emptyFieldErrTemplate, "IAMAuthConfig.Region")
	}
	return c.CredentialsProvider.validate()
}

// default values for ZTSCredentialProviderConfig
const (
	defaultRenewThresholdInSeconds = 600
)

// ZTSCredentialProviderConfig stores the configurations for configuring ZTSCredentialsProvider.
type ZTSCredentialProviderConfig struct {
	// APIEndpoint is the endpoint for requesting AWS temporary credentials.
	APIEndpoint string `json:"api_endpoint"`
	// AthenzDomain is the Athenz domain associated with the AWS account.
	AthenzDomain string `json:"athenz_domain"`
	// IAMRole is the AWS IAM role who has access to the DB.
	IAMRole string `json:"iam_role"`
	// ExternalID refers to the one defined in AWS documentation.
	// More info: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user_externalid.html
	ExternalID string `json:"external_id"`
	// RenewThresholdInSeconds defines the time period to refresh the credentials before it is expired.
	RenewThresholdInSeconds int `json:"renew_threshold_in_seconds"`
}

func (c *ZTSCredentialProviderConfig) populateDefaultValues() {
	if c.RenewThresholdInSeconds == 0 {
		c.RenewThresholdInSeconds = defaultRenewThresholdInSeconds
	}
}

func (c *ZTSCredentialProviderConfig) validate() error {
	if _, err := url.Parse(c.APIEndpoint); err != nil {
		return fmt.Errorf(`invalid field: "ZTSCredentialProviderConfig.APIEndpoint" should be a valid url, err: %v`, err)
	}
	if c.AthenzDomain == "" {
		return fmt.Errorf(emptyFieldErrTemplate, "ZTSCredentialProviderConfig.AthenzDomain")
	}
	if c.IAMRole == "" {
		return fmt.Errorf(emptyFieldErrTemplate, "ZTSCredentialProviderConfig.IAMRole")
	}
	if c.RenewThresholdInSeconds <= 0 {
		return fmt.Errorf(`invalid field: "ZTSCredentialProviderConfig.RenewThreshold" must be greater than 0, got %v`,
			c.RenewThresholdInSeconds)
	}
	return nil
}

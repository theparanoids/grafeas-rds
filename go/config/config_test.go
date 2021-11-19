// Copyright Yahoo 2021
// Licensed under the terms of the Apache License 2.0.
// See LICENSE file in project root for terms.
package config

import (
	"fmt"
	"log"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/grafeas/grafeas/go/config"
)

func TestNewConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		file       string
		wantErrMsg string
		wantConfig Config
	}{
		{
			file: "valid.yaml",
			wantConfig: Config{
				Host:          "some-host.rds.amazonaws.com",
				Port:          defaultPort,
				DBName:        defaultDBName,
				User:          "grafeas_rw",
				SSLMode:       defaultSSLMode,
				SSLRootCert:   "/opt/rds-ca-2019-root.pem",
				PaginationKey: "some_random_key",
				ConnPool: ConnPoolConfig{
					MaxOpenConns:             50,
					MaxIdleConns:             25,
					ConnMaxLifetimeInSeconds: 1800,
					ConnMaxIdleTimeInSeconds: 900,
				},
				IAMAuth: IAMAuthConfig{
					Region: "us-west-2",
					CredentialsProvider: ZTSCredentialProviderConfig{
						APIEndpoint:             "https://zts.athenz.company.com:4443/zts/v1",
						AthenzDomain:            "grafeas",
						IAMRole:                 "some-role.grafeas",
						RenewThresholdInSeconds: defaultRenewThresholdInSeconds,
					},
				},
			},
		},
		{
			file:       "invalid_api_endpoint.yaml",
			wantErrMsg: `invalid field: "ZTSCredentialProviderConfig.APIEndpoint" should be a valid url`,
		},
		{
			file:       "invalid_missing_athenz_domain.yaml",
			wantErrMsg: fmt.Sprintf(emptyFieldErrTemplate, "ZTSCredentialProviderConfig.AthenzDomain"),
		},
		{
			file:       "invalid_missing_host.yaml",
			wantErrMsg: fmt.Sprintf(emptyFieldErrTemplate, "Config.Host"),
		},
		{
			file:       "invalid_missing_user.yaml",
			wantErrMsg: fmt.Sprintf(emptyFieldErrTemplate, "Config.User"),
		},
		{
			file:       "invalid_missing_iam_role.yaml",
			wantErrMsg: fmt.Sprintf(emptyFieldErrTemplate, "ZTSCredentialProviderConfig.IAMRole"),
		},
		{
			file:       "invalid_missing_region.yaml",
			wantErrMsg: fmt.Sprintf(emptyFieldErrTemplate, "IAMAuthConfig.Region"),
		},
		{
			file:       "invalid_missing_ssl_root_cert.yaml",
			wantErrMsg: `invalid field: Config.SSLRootCert must not be empty because SSLMode is`,
		},
		{
			file:       "invalid_port.yaml",
			wantErrMsg: `invalid field: "Config.Port" must be larger than zero, got`,
		},
		{
			file:       "invalid_renew_threshold.yaml",
			wantErrMsg: `invalid field: "ZTSCredentialProviderConfig.RenewThreshold" must be greater than 0, got`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.file, func(t *testing.T) {
			t.Parallel()

			// We simulate how Grafeas loads the config instead of coining our own loading logic because:
			// 1. We want to use YAML for testing config files because Grafeas uses YAML for config files,
			//    but we don't want to add YAML tags by duplicating the existing JSON tags.
			// 2. We can make sure that a tested config file can be directly plugged into a Grafeas config,
			//    and it will work transparently.
			gc, err := config.LoadConfig(path.Join("testdata", tt.file))
			if err != nil {
				log.Fatalf("failed to load config: %v", err)
			}
			c, err := New(gc.StorageConfig)
			if (err != nil) != (tt.wantErrMsg != "") {
				if err != nil {
					t.Errorf("don't want error, but got %q", err)
				} else {
					t.Errorf("got nil error, but want error to include %q", tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("want %q to include %q", err.Error(), tt.wantErrMsg)
				}
				return
			}
			if !reflect.DeepEqual(tt.wantConfig, *c) {
				t.Errorf("config mismatch, want %v, got %v", tt.wantConfig, *c)
			}
		})
	}
	t.Run("failed to convert config", func(t *testing.T) {
		t.Parallel()

		wantErrMsg := "failed to convert the generic storage config to a rds config"
		invalidConf := config.StorageConfiguration(make(chan struct{}))
		_, err := New(&invalidConf)
		if err == nil {
			t.Fatalf("got nil error, but want error to be %q", wantErrMsg)
		}
		gotErrMsg := err.Error()
		if !strings.Contains(gotErrMsg, wantErrMsg) {
			t.Errorf("want %q to include %q", gotErrMsg, wantErrMsg)
		}
	})
}

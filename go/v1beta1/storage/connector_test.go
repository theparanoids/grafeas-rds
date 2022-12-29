// Copyright Yahoo 2021
// Licensed under the terms of the Apache License 2.0.
// See LICENSE file in project root for terms.
package storage

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/golang/mock/gomock"
	"github.com/theparanoids/grafeas-rds/go/config"
	"github.com/theparanoids/grafeas-rds/go/v1beta1/mocks"
)

func TestNewConnector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		creds      *credentials.Credentials
		wantLogs   string
		wantErrMsg string
	}{
		{
			name: "happy path - no IAM auth",
		},
		{
			name:     "happy path - IAM auth",
			creds:    credentials.NewStaticCredentials("a", "b", "c"),
			wantLogs: logsOptInIAMAuth,
		},
		{
			name:       "failed to set up IAM auth - invalid credentials",
			creds:      credentials.AnonymousCredentials,
			wantLogs:   logsOptInIAMAuth,
			wantErrMsg: errMsgSetupIAMAuth,
		},
	}

	mockCtrl := gomock.NewController(t)
	mockDriver := mocks.NewMockDriver(mockCtrl)
	mockCredentialsCreator := mocks.NewMockCredentialsCreator(mockCtrl)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			conf := config.Config{}
			var cc CredentialsCreator
			if tt.creds != nil {
				// Region is initialized here to make sure that
				// the value of IAMAuthConfig is different in each test case.
				conf.IAMAuth.Region = tt.name
				mockCredentialsCreator.EXPECT().Create(conf.IAMAuth).Return(tt.creds, nil)
				cc = mockCredentialsCreator
			}
			var buf bytes.Buffer
			logger := log.New(&buf, "", 0)
			c, err := newConnector(ctx, &conf, mockDriver, cc, logger, "")
			if (err == nil) != (tt.wantErrMsg == "") {
				if err == nil {
					t.Error("want error, but no error is returned")
				} else {
					t.Errorf("want no error, but an error is returned: %v", err)
				}
				return
			}
			logs := buf.String()
			if !strings.Contains(logs, tt.wantLogs) {
				t.Errorf("got %q, but want it to include %q", logs, tt.wantLogs)
			}
			if err != nil {
				return
			}
			dsn := c.readDSN()
			if dsn == "" {
				t.Errorf("the dsn %q should not be empty", dsn)
			}
		})
	}
}

func TestConnectorConnect(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	mockDriver := mocks.NewMockDriver(mockCtrl)
	c := &connector{
		dsn:    "some dsn",
		driver: mockDriver,
	}
	mockDriver.EXPECT().Open(c.dsn).Times(1)
	_, err := c.Connect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestConnectorDriver(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	want := mocks.NewMockDriver(mockCtrl)
	c := &connector{driver: want}
	got := c.Driver()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSetupIAMAuth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		creds      *credentials.Credentials
		wantErrMsg string
	}{
		{
			name:  "happy path",
			creds: credentials.NewStaticCredentials("a", "b", "c"),
		},
		{
			name:       "failed to create credentials",
			wantErrMsg: errMsgCreateCredentials,
		},
		{
			name:       "invalid credentials",
			creds:      credentials.AnonymousCredentials,
			wantErrMsg: errMsgRefreshAuthToken,
		},
	}

	mockCtrl := gomock.NewController(t)
	mockCredentialsCreator := mocks.NewMockCredentialsCreator(mockCtrl)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Region is initialized here to make sure that
			// the value of IAMAuthConfig is different in each test case.
			conf := config.IAMAuthConfig{Region: tt.name}
			var err error
			if tt.creds == nil {
				err = errors.New("some error")
			}
			mockCredentialsCreator.EXPECT().Create(conf).Return(tt.creds, err)
			c := &connector{}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			err = c.setupIAMAuth(ctx, conf, mockCredentialsCreator, log.Default())
			if (err == nil) != (tt.wantErrMsg == "") {
				if err == nil {
					t.Error("want error, but no error is returned")
				} else {
					t.Errorf("want no error, but an error is returned: %v", err)
				}
				return
			}
			if err != nil {
				return
			}
			dsn := c.readDSN()
			if dsn == "" {
				t.Errorf("the dsn %q should not be empty", dsn)
			}
		})
	}
}

func TestRefreshAuthToken(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		region  string
		creds   *credentials.Credentials
		wantErr bool
	}{
		{
			name:    "happy path",
			region:  "some-region",
			creds:   credentials.NewStaticCredentials("a", "b", "c"),
			wantErr: false,
		},
		{
			name:    "empty secret key in credentials should fail",
			creds:   credentials.AnonymousCredentials,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &connector{}
			err := c.refreshAuthToken(tt.creds, tt.region)
			hasErr := err != nil
			if tt.wantErr != hasErr {
				if err == nil {
					t.Error("want error, but no error is returned")
				} else {
					t.Errorf("want no error, but an error is returned: %v", err)
				}
				return
			}
			if hasErr {
				return
			}
			if !strings.Contains(c.password, tt.region) {
				t.Errorf("the password %q should contain the specified region %q", c.password, tt.region)
			}
			dsn := c.readDSN()
			if !strings.Contains(dsn, tt.region) {
				t.Errorf("the dsn %q should contain the specified region %q", dsn, tt.region)
			}
		})
	}
}

func TestRefreshAuthTokenPeriodically(t *testing.T) {
	t.Parallel()

	c := &connector{}
	creds := credentials.NewStaticCredentials("a", "b", "c")

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		refreshInterval := 1000 * time.Millisecond
		checkInterval := 1300 * time.Millisecond
		go c.refreshAuthTokenPeriodically(ctx, creds, "", refreshInterval, log.Default())
		time.Sleep(checkInterval)
		oldDsn := c.readDSN()
		if oldDsn == "" {
			t.Error("the dsn should have been updated with the refreshed token, but it's not")
		}
		time.Sleep(checkInterval)
		dsn := c.readDSN()
		if oldDsn == dsn {
			t.Error("the dsn is not updated on the correct interval, hence the token")
		}
	})
	t.Run("context is done", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		startTime := time.Now()
		c.refreshAuthTokenPeriodically(ctx, creds, "", refreshAuthTokenInterval, log.Default())
		if time.Since(startTime) >= refreshAuthTokenInterval {
			t.Error("context is done, but the function does not return immediately")
		}
	})
	t.Run("invalid credentials", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		var buf bytes.Buffer
		const interval = 10 * time.Millisecond
		go func() {
			time.Sleep(interval * 5)
			cancel()
		}()
		// A blocking call is used here to avoid race condition on buf.
		// The writer (i.e. refreshAuthTokenPeriodically) should stop writing to buf
		// before the reader (i.e. buf.String()) attemps to read it.
		c.refreshAuthTokenPeriodically(ctx, credentials.AnonymousCredentials, "", interval, log.New(&buf, "", 0))
		logs := buf.String()
		if !strings.Contains(logs, errMsgRefreshAuthToken) {
			t.Errorf("got %q, but want it to include %q", logs, errMsgRefreshAuthToken)
		}
	})
}

func TestUpdatePassword(t *testing.T) {
	t.Parallel()

	c := &connector{}
	pwd := "abc"
	c.updatePassword(pwd)
	if c.password != pwd {
		t.Errorf("got %q, want %q", c.password, pwd)
	}
	dsn := c.readDSN()
	if !strings.Contains(dsn, pwd) {
		t.Errorf("the dsn %q should contain the password %q", dsn, pwd)
	}
}

func TestAssembleDSN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		sslRootCert string
		want        string
	}{
		{
			name: "ssl root cert is not given",
			want: "host=localhost port=5432 dbname=grafeas user=grafeas_rw password=dummy-password-for-unit-tests-only sslmode=verify-full",
		},
		{
			name:        "ssl root cert is given",
			sslRootCert: "ca.pem",
			want:        "host=localhost port=5432 dbname=grafeas user=grafeas_rw password=dummy-password-for-unit-tests-only sslmode=verify-full sslrootcert=ca.pem",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := connector{
				host:        "localhost",
				port:        5432,
				dbName:      "grafeas",
				user:        "grafeas_rw",
				password:    "dummy-password-for-unit-tests-only",
				sslMode:     "verify-full",
				sslRootCert: tt.sslRootCert,
			}
			got := c.assembleDSN()
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// Copyright Yahoo 2021
// Licensed under the terms of the Apache License 2.0.
// See LICENSE file in project root for terms.
package rds

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	gomock "github.com/golang/mock/gomock"
	"github.com/grafeas/grafeas/go/config"

	rdsconfig "github.com/theparanoids/grafeas-rds/rds/config"
)

func TestStorageProviderProvide(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	validConf := config.StorageConfiguration(rdsconfig.Config{
		Host:        "some-host.rds.amazonaws.com",
		User:        "grafeas_rw",
		Password:    "dummy-password-for-unit-tests-only",
		SSLRootCert: "/opt/rds-ca-2019-root.pem",
		// ConnPool is populated in order to test if setConnPoolParams is invoked in Provide.
		ConnPool: rdsconfig.ConnPoolConfig{
			MaxOpenConns:             1,
			MaxIdleConns:             2,
			ConnMaxLifetimeInSeconds: 3,
			ConnMaxIdleTimeInSeconds: 4,
		},
	})

	type testCase struct {
		name         string
		expect       func(*testCase)
		conf         config.StorageConfiguration
		store        *MockStorage
		storeCreator *MockStorageCreator
		credsCreator *MockCredentialsCreator
		wantErrMsg   string
	}
	tests := []testCase{
		{
			name: "happy path",
			expect: func(tt *testCase) {
				tt.credsCreator.EXPECT().Create(gomock.Any()).Times(1).Return(credentials.NewStaticCredentials("a", "b", "c"), nil)
				tt.storeCreator.EXPECT().Create(gomock.Any(), gomock.Any()).Times(1).Return(tt.store, nil)

				conf := tt.conf.(rdsconfig.Config).ConnPool
				tt.store.EXPECT().SetMaxOpenConns(conf.MaxOpenConns).Times(1)
				tt.store.EXPECT().SetMaxIdleConns(conf.MaxIdleConns).Times(1)
				tt.store.EXPECT().SetConnMaxLifetime(time.Duration(conf.ConnMaxLifetimeInSeconds) * time.Second).Times(1)
				tt.store.EXPECT().SetConnMaxIdleTime(time.Duration(conf.ConnMaxIdleTimeInSeconds) * time.Second).Times(1)
			},
			conf:         validConf,
			store:        NewMockStorage(mockCtrl),
			storeCreator: NewMockStorageCreator(mockCtrl),
			credsCreator: NewMockCredentialsCreator(mockCtrl),
		},
		{
			name: "invalid config",
			// An empty Config is invalid because the Host field does not have a default value.
			conf:       config.StorageConfiguration(rdsconfig.Config{}),
			wantErrMsg: errMsgInitConfig,
		},
		{
			name: "invalid connector",
			expect: func(tt *testCase) {
				tt.credsCreator.EXPECT().Create(gomock.Any()).Times(1).Return(credentials.AnonymousCredentials, nil)
			},
			conf:         validConf,
			credsCreator: NewMockCredentialsCreator(mockCtrl),
			wantErrMsg:   errMsgInitConnector,
		},
		{
			name: "invalid store",
			expect: func(tt *testCase) {
				tt.credsCreator.EXPECT().Create(gomock.Any()).Times(1).Return(credentials.NewStaticCredentials("a", "b", "c"), nil)
				tt.storeCreator.EXPECT().Create(gomock.Any(), gomock.Any()).Times(1).Return(nil, errors.New("random error"))
			},
			conf:         validConf,
			storeCreator: NewMockStorageCreator(mockCtrl),
			credsCreator: NewMockCredentialsCreator(mockCtrl),
			wantErrMsg:   errMsgInitStorage,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.expect != nil {
				tt.expect(&tt)
			}
			storageProvider := NewGrafeasStorageProvider(NewMockDriver(mockCtrl), tt.credsCreator, tt.storeCreator)
			storage, err := storageProvider.Provide("", &tt.conf)
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

			if storage.Gs != tt.store || storage.Ps != tt.store {
				t.Errorf("unexpected fields: %v", storage)
			}
		})
	}
}

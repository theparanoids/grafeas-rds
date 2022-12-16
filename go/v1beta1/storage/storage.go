// Copyright Yahoo 2021
// Licensed under the terms of the Apache License 2.0.
// See LICENSE file in project root for terms.
package storage

import (
	"context"
	"database/sql/driver"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/grafeas/grafeas/go/config"
	"github.com/grafeas/grafeas/go/v1beta1/storage"
	rdsconfig "github.com/theparanoids/grafeas-rds/go/config"
)

const (
	errMsgInitConfig    = "failed to initialize config"
	errMsgInitConnector = "failed to initialize connector"
	errMsgInitStorage   = "failed to initialize store"
)

// ConnPoolMgr manages a RDBMS connection pool.
// The methods are defined on sql.DB.
// Ref. https://golang.org/pkg/database/sql/
type ConnPoolMgr interface {
	SetMaxOpenConns(n int)
	SetMaxIdleConns(n int)
	SetConnMaxLifetime(d time.Duration)
	SetConnMaxIdleTime(d time.Duration)
}

// Storage contains all the methods to
// 1. be used as a backend for a Grafeas server AND
// 2. manage a RDBMS connection pool.
type Storage interface {
	storage.Gs
	storage.Ps
	ConnPoolMgr
}

// StorageCreator can be implemented based on the backend storage types (e.g. PostgreSQL, MySQL, etc.).
type StorageCreator interface {
	Create(connector driver.Connector, paginationKey string) (Storage, error)
	CreateRW(readerConnector driver.Connector, writerConnector driver.Connector, paginationKey string) (Storage, error)
}

// CredentialsCreator can be implemented to create Credentials based on different types of providers.
// Fields of IAMAuthConfig can be updated to incorporate such changes in the future.
type CredentialsCreator interface {
	Create(rdsconfig.IAMAuthConfig) (*credentials.Credentials, error)
}

// GrafeasStorageProvider contains the fields to flexibily create a storage.Storage.
type GrafeasStorageProvider struct {
	drv                driver.Driver
	credentialsCreator CredentialsCreator
	storageCreator     StorageCreator
}

// NewGrafeasStorageProvider returns a StorageProvider whose fields are populated with the arguments.
func NewGrafeasStorageProvider(drv driver.Driver, credentialsCreator CredentialsCreator, storageCreator StorageCreator) *GrafeasStorageProvider {
	return &GrafeasStorageProvider{
		drv:                drv,
		credentialsCreator: credentialsCreator,
		storageCreator:     storageCreator,
	}
}

// Provide returns a storage which is configured based on the receiver's fields.
func (p GrafeasStorageProvider) Provide(_ string, confi *config.StorageConfiguration) (*storage.Storage, error) {
	conf, err := rdsconfig.New(confi)
	if err != nil {
		return nil, fmt.Errorf("%s, err: %v", errMsgInitConfig, err)
	}

	// TODO: Use the context passed from main after
	// the signature of RegisterStorageTypeProvider is updated to include it.
	connector, err := newConnector(context.Background(), conf, p.drv, p.credentialsCreator, log.Default(), "")
	if err != nil {
		return nil, fmt.Errorf("%s, err: %v", errMsgInitConnector, err)
	}

	rdsStorage, err := p.storageCreator.Create(connector, conf.PaginationKey)
	if err != nil {
		return nil, fmt.Errorf("%s, err: %v", errMsgInitStorage, err)
	}
	setConnPoolParams(rdsStorage, conf.ConnPool)

	grafeasStorage := &storage.Storage{
		Ps: rdsStorage,
		Gs: rdsStorage,
	}
	return grafeasStorage, nil
}

// ProvideRW returns a storage which is configured based on the receiver's fields. The storage connects to different reader/writer. If no reader is provided, then it will only connect to the writer.
func (p GrafeasStorageProvider) ProvideRW(_ string, confi *config.StorageConfiguration) (*storage.Storage, error) {
	conf, err := rdsconfig.New(confi)
	if err != nil {
		return nil, fmt.Errorf("%s, err: %v", errMsgInitConfig, err)
	}

	// TODO: Use the context passed from main after
	// the signature of RegisterStorageTypeProvider is updated to include it.
	writerConnector, err := newConnector(context.Background(), conf, p.drv, p.credentialsCreator, log.Default(), "")
	if err != nil {
		return nil, fmt.Errorf("%s, err: %v", errMsgInitConnector, err)
	}
	readerConnector, err := newConnector(context.Background(), conf, p.drv, p.credentialsCreator, log.Default(), conf.Reader)
	if err != nil {
		return nil, fmt.Errorf("%s, err: %v", errMsgInitConnector, err)
	}

	rdsStorage, err := p.storageCreator.CreateRW(readerConnector, writerConnector, conf.PaginationKey)
	if err != nil {
		return nil, fmt.Errorf("%s, err: %v", errMsgInitStorage, err)
	}
	setConnPoolParams(rdsStorage, conf.ConnPool)

	grafeasStorage := &storage.Storage{
		Ps: rdsStorage,
		Gs: rdsStorage,
	}
	return grafeasStorage, nil
}

func setConnPoolParams(mgr ConnPoolMgr, conf rdsconfig.ConnPoolConfig) {
	mgr.SetMaxOpenConns(conf.MaxOpenConns)
	mgr.SetMaxIdleConns(conf.MaxIdleConns)
	mgr.SetConnMaxLifetime(time.Duration(conf.ConnMaxLifetimeInSeconds) * time.Second)
	mgr.SetConnMaxIdleTime(time.Duration(conf.ConnMaxIdleTimeInSeconds) * time.Second)
}

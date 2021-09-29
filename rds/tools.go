// Copyright Yahoo 2021
// Licensed under the terms of the Apache License 2.0.
// See LICENSE file in project root for terms.
//
// +build tools

package rds

// Although gomock is already in go.mod because
// it is also used as a library by the testing code,
// we still put it here to make sure that
// all dependant tool binaries are listed here for the sake of consistency.
import _ "github.com/golang/mock/mockgen"

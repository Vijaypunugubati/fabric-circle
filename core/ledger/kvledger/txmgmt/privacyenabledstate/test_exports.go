/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package privacyenabledstate

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hyperledger/fabric/common/metrics/disabled"
	"github.com/hyperledger/fabric/core/ledger"
	"github.com/hyperledger/fabric/core/ledger/kvledger/bookkeeping"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/statedb/statecouchdb"
	"github.com/hyperledger/fabric/core/ledger/ledgerconfig"
	"github.com/hyperledger/fabric/core/ledger/mock"
	"github.com/hyperledger/fabric/core/ledger/util/couchdb"
	"github.com/hyperledger/fabric/integration/runner"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// TestEnv - an interface that a test environment implements
type TestEnv interface {
	Init(t testing.TB)
	GetDBHandle(id string) DB
	GetName() string
	Cleanup()
}

// Tests will be run against each environment in this array
// For example, to skip CouchDB tests, remove &couchDBLockBasedEnv{}
//var testEnvs = []testEnv{&levelDBCommonStorageTestEnv{}, &couchDBCommonStorageTestEnv{}}
var testEnvs = []TestEnv{&LevelDBCommonStorageTestEnv{}, &CouchDBCommonStorageTestEnv{}}

///////////// LevelDB Environment //////////////

// LevelDBCommonStorageTestEnv implements TestEnv interface for leveldb based storage
type LevelDBCommonStorageTestEnv struct {
	t                 testing.TB
	provider          DBProvider
	bookkeeperTestEnv *bookkeeping.TestEnv
}

// Init implements corresponding function from interface TestEnv
func (env *LevelDBCommonStorageTestEnv) Init(t testing.TB) {
	viper.Set("ledger.state.stateDatabase", "")
	removeDBPath(t)
	env.bookkeeperTestEnv = bookkeeping.NewTestEnv(t)
	dbProvider, err := NewCommonStorageDBProvider(
		env.bookkeeperTestEnv.TestProvider,
		&disabled.Provider{},
		&mock.HealthCheckRegistry{},
		nil,
	)
	assert.NoError(t, err)
	env.t = t
	env.provider = dbProvider
}

// GetDBHandle implements corresponding function from interface TestEnv
func (env *LevelDBCommonStorageTestEnv) GetDBHandle(id string) DB {
	db, err := env.provider.GetDBHandle(id)
	assert.NoError(env.t, err)
	return db
}

// GetName implements corresponding function from interface TestEnv
func (env *LevelDBCommonStorageTestEnv) GetName() string {
	return "levelDBCommonStorageTestEnv"
}

// Cleanup implements corresponding function from interface TestEnv
func (env *LevelDBCommonStorageTestEnv) Cleanup() {
	env.provider.Close()
	env.bookkeeperTestEnv.Cleanup()
	removeDBPath(env.t)
}

///////////// CouchDB Environment //////////////

// CouchDBCommonStorageTestEnv implements TestEnv interface for couchdb based storage
type CouchDBCommonStorageTestEnv struct {
	t                 testing.TB
	provider          DBProvider
	bookkeeperTestEnv *bookkeeping.TestEnv
	couchCleanup      func()
}

func (env *CouchDBCommonStorageTestEnv) setupCouch() string {
	externalCouch, set := os.LookupEnv("COUCHDB_ADDR")
	if set {
		env.couchCleanup = func() {}
		return externalCouch
	}

	couchDB := &runner.CouchDB{}
	if err := couchDB.Start(); err != nil {
		err := fmt.Errorf("failed to start couchDB: %s", err)
		panic(err)
	}
	env.couchCleanup = func() { couchDB.Stop() }
	return couchDB.Address()
}

// Init implements corresponding function from interface TestEnv
func (env *CouchDBCommonStorageTestEnv) Init(t testing.TB) {
	redologsPath := ledgerconfig.GetCouchdbRedologsPath()
	assert.NoError(t, os.RemoveAll(redologsPath))
	viper.Set("ledger.state.stateDatabase", "CouchDB")
	couchAddress := env.setupCouch()

	stateDBConfig := &ledger.StateDB{
		StateDatabase: "CouchDB",
		CouchDB: &couchdb.Config{
			Address:             couchAddress,
			Username:            "",
			Password:            "",
			MaxRetries:          3,
			MaxRetriesOnStartup: 20,
			RequestTimeout:      35 * time.Second,
			InternalQueryLimit:  1000,
			MaxBatchUpdateSize:  1000,
		},
	}

	env.bookkeeperTestEnv = bookkeeping.NewTestEnv(t)
	dbProvider, err := NewCommonStorageDBProvider(
		env.bookkeeperTestEnv.TestProvider,
		&disabled.Provider{},
		&mock.HealthCheckRegistry{},
		stateDBConfig,
	)
	assert.NoError(t, err)
	env.t = t
	env.provider = dbProvider
}

// GetDBHandle implements corresponding function from interface TestEnv
func (env *CouchDBCommonStorageTestEnv) GetDBHandle(id string) DB {
	db, err := env.provider.GetDBHandle(id)
	assert.NoError(env.t, err)
	return db
}

// GetName implements corresponding function from interface TestEnv
func (env *CouchDBCommonStorageTestEnv) GetName() string {
	return "couchDBCommonStorageTestEnv"
}

// Cleanup implements corresponding function from interface TestEnv
func (env *CouchDBCommonStorageTestEnv) Cleanup() {
	csdbProvider, _ := env.provider.(*CommonStorageDBProvider)
	statecouchdb.CleanupDB(env.t, csdbProvider.VersionedDBProvider)
	redologsPath := ledgerconfig.GetCouchdbRedologsPath()
	assert.NoError(env.t, os.RemoveAll(redologsPath))
	env.bookkeeperTestEnv.Cleanup()
	env.provider.Close()
	env.couchCleanup()
}

func removeDBPath(t testing.TB) {
	dbPath := ledgerconfig.GetStateLevelDBPath()
	if err := os.RemoveAll(dbPath); err != nil {
		t.Fatalf("Err: %s", err)
		t.FailNow()
	}
}

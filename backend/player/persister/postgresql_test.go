package persister_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/player/persister"
)

type mockStatementExecutor struct {
	ReturnForExec         sql.Result
	ReturnedErrorForExec  error
	ReturnForQuery        *sql.Rows
	ReturnedErrorForQuery error
}

func (mockExecutor *mockStatementExecutor) Exec(
	query string,
	args ...interface{}) (sql.Result, error) {
	return mockExecutor.ReturnForExec, mockExecutor.ReturnedErrorForExec
}

func (mockExecutor *mockStatementExecutor) Query(
	query string,
	args ...interface{}) (*sql.Rows, error) {
	return mockExecutor.ReturnForQuery, mockExecutor.ReturnedErrorForQuery
}

func TestReturnErrorFromInvalidConnectionString(unitTest *testing.T) {
	invalidConnection := "user=INVALID password=WRONG dbname=DOES_NOT_EXIST"
	postgresqlPersister, errorFromPostgresql :=
		persister.NewInPostgresql(invalidConnection)

	if errorFromPostgresql == nil {
		unitTest.Fatalf(
			"Successfully created PostgreSQL persister %+v from connection string %v instead producing error",
			postgresqlPersister,
			invalidConnection)
	}
}

func TestReturnErrorFromInvalidInitialStatement(unitTest *testing.T) {
	mockExecutor := &mockStatementExecutor{}
	mockExecutor.ReturnForExec = nil
	mockExecutor.ReturnedErrorForExec = fmt.Errorf("expected error")
	invalidStatement := "does not matter as the mock will return an error anyway"
	postgresqlPersister, errorFromPostgresql :=
		persister.NewInPostgresqlWithInitialStatements(mockExecutor, invalidStatement)

	if errorFromPostgresql == nil {
		unitTest.Fatalf(
			"Successfully created PostgreSQL persister %+v from invalid statement %v"+
				" instead of producing error",
			postgresqlPersister,
			invalidStatement)
	}
}

func TestReturnErrorFromQueryDuringGet(unitTest *testing.T) {
	playerName := "Does Not Matter"
	mockExecutor := &mockStatementExecutor{}
	mockExecutor.ReturnForQuery = nil
	mockExecutor.ReturnedErrorForQuery = fmt.Errorf("expected error")

	postgresqlPersister, errorFromPostgresql :=
		persister.NewInPostgresqlWithInitialStatements(mockExecutor)
	if errorFromPostgresql != nil {
		unitTest.Fatalf(
			"Produced error %v when trying to create persister using mock",
			errorFromPostgresql)
	}

	playerState, errorFromGet := postgresqlPersister.Get(playerName)
	if errorFromGet == nil {
		unitTest.Fatalf(
			"Get(%v) produced %+v, nil",
			playerName,
			playerState)
	}
}

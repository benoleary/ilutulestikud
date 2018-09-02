package persister_test

import (
	"fmt"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/player/persister"
)

type mockRowsAsResult struct {
	ErrorToReturn               error
	NumberOfRowsAtOrAfterCursor int
}

func (mockResult *mockRowsAsResult) Next() bool {
	hasRowAtCursor := mockResult.NumberOfRowsAtOrAfterCursor > 0
	mockResult.NumberOfRowsAtOrAfterCursor--
	return hasRowAtCursor
}

func (mockResult *mockRowsAsResult) Close() error {
	return mockResult.ErrorToReturn
}

func (mockResult *mockRowsAsResult) Err() error {
	return mockResult.ErrorToReturn
}

func (mockResult *mockRowsAsResult) Scan(rowDestination ...interface{}) error {
	return mockResult.ErrorToReturn
}

type mockMetadataAsResult struct {
	NumberOfRowsAffected int64
	ErrorToReturn        error
}

func (mockResult *mockMetadataAsResult) RowsAffected() (int64, error) {
	return mockResult.NumberOfRowsAffected, mockResult.ErrorToReturn
}

type mockStatementExecutor struct {
	ReturnForExecuteQuery            persister.RowsAsResult
	ReturnedErrorForExecuteQuery     error
	ReturnForExecuteStatement        persister.MetadataAsResult
	ReturnedErrorForExecuteStatement error
}

func (mockExecutor *mockStatementExecutor) ExecuteQuery(
	statementWithoutRowsAsResult string,
	argumentsForStatement ...interface{}) (persister.RowsAsResult, error) {
	return mockExecutor.ReturnForExecuteQuery, mockExecutor.ReturnedErrorForExecuteQuery
}

func (mockExecutor *mockStatementExecutor) ExecuteStatement(
	statementWithoutRowsAsResult string,
	argumentsForStatement ...interface{}) (persister.MetadataAsResult, error) {
	return mockExecutor.ReturnForExecuteStatement, mockExecutor.ReturnedErrorForExecuteStatement
}

func TestReturnErrorFromInvalidConnectionString(unitTest *testing.T) {
	invalidConnection := "user=INVALID password=WRONG dbname=DOES_NOT_EXIST"
	postgresqlPersister, errorFromPostgresql :=
		persister.NewInPostgresql(context.Background(), invalidConnection)

	if errorFromPostgresql == nil {
		unitTest.Fatalf(
			"Successfully created PostgreSQL persister %+v from connection string %v instead producing error",
			postgresqlPersister,
			invalidConnection)
	}
}

func TestReturnErrorFromInvalidInitialStatement(unitTest *testing.T) {
	mockExecutor := &mockStatementExecutor{}
	mockExecutor.ReturnForExecuteStatement = nil
	mockExecutor.ReturnedErrorForExecuteStatement = fmt.Errorf("expected error")
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
	mockExecutor.ReturnForExecuteQuery = nil
	mockExecutor.ReturnedErrorForExecuteQuery = fmt.Errorf("expected error")

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

func TestReturnErrorFromScanDuringGet(unitTest *testing.T) {
	playerName := "Does Not Matter"

	mockResult := &mockRowsAsResult{}
	mockResult.NumberOfRowsAtOrAfterCursor = 1
	mockResult.ErrorToReturn = fmt.Errorf("expected error")

	mockExecutor := &mockStatementExecutor{}
	mockExecutor.ReturnForExecuteQuery = mockResult
	mockExecutor.ReturnedErrorForExecuteQuery = nil

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

func TestReturnErrorWhenMultipleRowsFromGet(unitTest *testing.T) {
	playerName := "Does Not Matter"

	mockResult := &mockRowsAsResult{}
	mockResult.NumberOfRowsAtOrAfterCursor = 2
	mockResult.ErrorToReturn = nil

	mockExecutor := &mockStatementExecutor{}
	mockExecutor.ReturnForExecuteQuery = mockResult
	mockExecutor.ReturnedErrorForExecuteQuery = nil

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

func TestReturnErrorFromQueryDuringAll(unitTest *testing.T) {
	mockExecutor := &mockStatementExecutor{}
	mockExecutor.ReturnForExecuteQuery = nil
	mockExecutor.ReturnedErrorForExecuteQuery = fmt.Errorf("expected error")

	postgresqlPersister, errorFromPostgresql :=
		persister.NewInPostgresqlWithInitialStatements(mockExecutor)
	if errorFromPostgresql != nil {
		unitTest.Fatalf(
			"Produced error %v when trying to create persister using mock",
			errorFromPostgresql)
	}

	playerStates, errorFromAll := postgresqlPersister.All()
	if errorFromAll == nil {
		unitTest.Fatalf(
			"All() produced %+v, nil",
			playerStates)
	}
}

func TestReturnErrorFromScanDuringAll(unitTest *testing.T) {
	mockResult := &mockRowsAsResult{}
	mockResult.NumberOfRowsAtOrAfterCursor = 1
	mockResult.ErrorToReturn = fmt.Errorf("expected error")

	mockExecutor := &mockStatementExecutor{}
	mockExecutor.ReturnForExecuteQuery = mockResult
	mockExecutor.ReturnedErrorForExecuteQuery = nil

	postgresqlPersister, errorFromPostgresql :=
		persister.NewInPostgresqlWithInitialStatements(mockExecutor)
	if errorFromPostgresql != nil {
		unitTest.Fatalf(
			"Produced error %v when trying to create persister using mock",
			errorFromPostgresql)
	}

	playerStates, errorFromAll := postgresqlPersister.All()
	if errorFromAll == nil {
		unitTest.Fatalf(
			"All() produced %+v, nil",
			playerStates)
	}
}

func TestReturnErrorFromQueryDuringDelete(unitTest *testing.T) {
	playerName := "Does Not Matter"

	mockExecutor := &mockStatementExecutor{}
	mockExecutor.ReturnForExecuteStatement = nil
	mockExecutor.ReturnedErrorForExecuteStatement = fmt.Errorf("expected error")

	postgresqlPersister, errorFromPostgresql :=
		persister.NewInPostgresqlWithInitialStatements(mockExecutor)
	if errorFromPostgresql != nil {
		unitTest.Fatalf(
			"Produced error %v when trying to create persister using mock",
			errorFromPostgresql)
	}

	errorFromDelete := postgresqlPersister.Delete(playerName)
	if errorFromDelete == nil {
		unitTest.Fatalf(
			"Delete(%v) produced nil error",
			playerName)
	}
}

func TestReturnErrorWhenErrorParsingDeletionResult(unitTest *testing.T) {
	playerName := "Does Not Matter"

	mockResult := &mockMetadataAsResult{}
	mockResult.NumberOfRowsAffected = 1
	mockResult.ErrorToReturn = fmt.Errorf("expected error")

	mockExecutor := &mockStatementExecutor{}
	mockExecutor.ReturnForExecuteStatement = mockResult
	mockExecutor.ReturnedErrorForExecuteStatement = nil

	postgresqlPersister, errorFromPostgresql :=
		persister.NewInPostgresqlWithInitialStatements(mockExecutor)
	if errorFromPostgresql != nil {
		unitTest.Fatalf(
			"Produced error %v when trying to create persister using mock",
			errorFromPostgresql)
	}

	errorFromDelete := postgresqlPersister.Delete(playerName)
	if errorFromDelete == nil {
		unitTest.Fatalf(
			"Delete(%v) produced nil error",
			playerName)
	}
}

func TestReturnErrorWhenMultipleRowsDeleted(unitTest *testing.T) {
	playerName := "Does Not Matter"

	mockResult := &mockMetadataAsResult{}
	mockResult.NumberOfRowsAffected = 2
	mockResult.ErrorToReturn = nil

	mockExecutor := &mockStatementExecutor{}
	mockExecutor.ReturnForExecuteStatement = mockResult
	mockExecutor.ReturnedErrorForExecuteStatement = nil

	postgresqlPersister, errorFromPostgresql :=
		persister.NewInPostgresqlWithInitialStatements(mockExecutor)
	if errorFromPostgresql != nil {
		unitTest.Fatalf(
			"Produced error %v when trying to create persister using mock",
			errorFromPostgresql)
	}

	errorFromDelete := postgresqlPersister.Delete(playerName)
	if errorFromDelete == nil {
		unitTest.Fatalf(
			"Delete(%v) produced nil error",
			playerName)
	}
}

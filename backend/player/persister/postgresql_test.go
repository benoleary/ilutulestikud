package persister_test

import (
	"context"
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
	executionContext context.Context,
	statementWithoutRowsAsResult string,
	argumentsForStatement ...interface{}) (persister.RowsAsResult, error) {
	return mockExecutor.ReturnForExecuteQuery, mockExecutor.ReturnedErrorForExecuteQuery
}

func (mockExecutor *mockStatementExecutor) ExecuteStatement(
	executionContext context.Context,
	statementWithoutRowsAsResult string,
	argumentsForStatement ...interface{}) (persister.MetadataAsResult, error) {
	return mockExecutor.ReturnForExecuteStatement, mockExecutor.ReturnedErrorForExecuteStatement
}

func TestReturnErrorFromInvalidConnectionString(unitTest *testing.T) {
	invalidConnectionString := "user=INVALID password=WRONG dbname=DOES_NOT_EXIST"
	postgresqlPersister := persister.NewInPostgresql(invalidConnectionString)

	executionContext := context.Background()

	// We test that every kind of request generates an error.
	playerName := "Should Not Matter"
	playerColor := "should not matter"
	errorFromAddRequest :=
		postgresqlPersister.Add(executionContext, playerName, playerColor)

	if errorFromAddRequest == nil {
		unitTest.Fatalf(
			"Successfully created PostgreSQL persister %+v from connection"+
				" string %v, and got nil error from .Add(%v, %v, %v)",
			postgresqlPersister,
			invalidConnectionString,
			executionContext,
			playerName,
			playerColor)
	}

	errorFromUpdateColorRequest :=
		postgresqlPersister.UpdateColor(executionContext, playerName, playerColor)

	if errorFromUpdateColorRequest == nil {
		unitTest.Fatalf(
			"Successfully created PostgreSQL persister %+v from connection"+
				" string %v, and got nil error from .UpdateColor(%v, %v, %v)",
			postgresqlPersister,
			invalidConnectionString,
			executionContext,
			playerName,
			playerColor)
	}

	unexpectedGetResult, errorFromGetRequest :=
		postgresqlPersister.Get(executionContext, playerName)

	if errorFromGetRequest == nil {
		unitTest.Fatalf(
			"Successfully created PostgreSQL persister %+v from connection"+
				" string %v, and got %v from .All(%v, %v) instead of producing error",
			postgresqlPersister,
			invalidConnectionString,
			unexpectedGetResult,
			executionContext,
			playerName)
	}

	unexpectedAllResult, errorFromAllRequest :=
		postgresqlPersister.All(executionContext)

	if errorFromAllRequest == nil {
		unitTest.Fatalf(
			"Successfully created PostgreSQL persister %+v from connection"+
				" string %v, and got %v from .All(%v) instead of producing error",
			postgresqlPersister,
			invalidConnectionString,
			unexpectedAllResult,
			executionContext)
	}

	errorFromDeleteRequest :=
		postgresqlPersister.Delete(executionContext, playerName)

	if errorFromDeleteRequest == nil {
		unitTest.Fatalf(
			"Successfully created PostgreSQL persister %+v from connection"+
				" string %v, and got nil error from .Delete(%v, %v)",
			postgresqlPersister,
			invalidConnectionString,
			executionContext,
			playerName)
	}
}

func TestReturnErrorFromQueryDuringGet(unitTest *testing.T) {
	playerName := "Does Not Matter"

	mockExecutor := &mockStatementExecutor{}
	mockExecutor.ReturnForExecuteQuery = nil
	mockExecutor.ReturnedErrorForExecuteQuery = fmt.Errorf("expected error")

	ignoredConnectionString := "does not matter as the mock executor is provided"
	postgresqlPersister :=
		persister.NewInPostgresqlWithGivenLimitedExecutor(
			ignoredConnectionString,
			mockExecutor)

	playerState, errorFromGet :=
		postgresqlPersister.Get(context.Background(), playerName)
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

	ignoredConnectionString := "does not matter as the mock executor is provided"
	postgresqlPersister :=
		persister.NewInPostgresqlWithGivenLimitedExecutor(
			ignoredConnectionString,
			mockExecutor)

	playerState, errorFromGet :=
		postgresqlPersister.Get(context.Background(), playerName)
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

	ignoredConnectionString := "does not matter as the mock executor is provided"
	postgresqlPersister :=
		persister.NewInPostgresqlWithGivenLimitedExecutor(
			ignoredConnectionString,
			mockExecutor)

	playerState, errorFromGet :=
		postgresqlPersister.Get(context.Background(), playerName)
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

	ignoredConnectionString := "does not matter as the mock executor is provided"
	postgresqlPersister :=
		persister.NewInPostgresqlWithGivenLimitedExecutor(
			ignoredConnectionString,
			mockExecutor)

	playerStates, errorFromAll :=
		postgresqlPersister.All(context.Background())
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

	ignoredConnectionString := "does not matter as the mock executor is provided"
	postgresqlPersister :=
		persister.NewInPostgresqlWithGivenLimitedExecutor(
			ignoredConnectionString,
			mockExecutor)

	playerStates, errorFromAll :=
		postgresqlPersister.All(context.Background())
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

	ignoredConnectionString := "does not matter as the mock executor is provided"
	postgresqlPersister :=
		persister.NewInPostgresqlWithGivenLimitedExecutor(
			ignoredConnectionString,
			mockExecutor)

	errorFromDelete :=
		postgresqlPersister.Delete(context.Background(), playerName)
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

	ignoredConnectionString := "does not matter as the mock executor is provided"
	postgresqlPersister :=
		persister.NewInPostgresqlWithGivenLimitedExecutor(
			ignoredConnectionString,
			mockExecutor)

	errorFromDelete :=
		postgresqlPersister.Delete(context.Background(), playerName)
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

	ignoredConnectionString := "does not matter as the mock executor is provided"
	postgresqlPersister :=
		persister.NewInPostgresqlWithGivenLimitedExecutor(
			ignoredConnectionString,
			mockExecutor)

	errorFromDelete :=
		postgresqlPersister.Delete(context.Background(), playerName)
	if errorFromDelete == nil {
		unitTest.Fatalf(
			"Delete(%v) produced nil error",
			playerName)
	}
}

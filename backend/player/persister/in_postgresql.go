package persister

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/benoleary/ilutulestikud/backend/player"

	// We need the side-effects of importing this library, but do not use it directly.
	_ "github.com/lib/pq"
)

// RowsAsResult defines the subset of the functions of the sql.Rows struct used by the
// inPostgresqlPersister struct.
type RowsAsResult interface {
	Next() bool
	Close() error
	Err() error
	Scan(rowDestination ...interface{}) error
}

// MetadataAsResult defines the subset of the functions of the sql.Result interface used
// by the inPostgresqlPersister struct.
type MetadataAsResult interface {
	RowsAffected() (int64, error)
}

// LimitedExecutor defines the subset of the functions of the sql.DB interface used
// by the inPostgresqlPersister struct.
type LimitedExecutor interface {
	// ExecuteQuery should execute a query and return the resulting rows.
	ExecuteQuery(
		executionContext context.Context,
		queryStatement string,
		argumentsForStatement ...interface{}) (RowsAsResult, error)

	// ExecuteStatement should execute a statement and return metadata about the operation.
	ExecuteStatement(
		executionContext context.Context,
		statementWithoutRowsAsResult string,
		argumentsForStatement ...interface{}) (MetadataAsResult, error)
}

type wrappingLimitedExecutor struct {
	wrappedInterface *sql.DB
}

func (wrappingExecutor *wrappingLimitedExecutor) ExecuteQuery(
	executionContext context.Context,
	statementWithoutRowsAsResult string,
	argumentsForStatement ...interface{}) (RowsAsResult, error) {
	return wrappingExecutor.wrappedInterface.QueryContext(
		executionContext,
		statementWithoutRowsAsResult,
		argumentsForStatement...)
}

func (wrappingExecutor *wrappingLimitedExecutor) ExecuteStatement(
	executionContext context.Context,
	statementWithoutRowsAsResult string,
	argumentsForStatement ...interface{}) (MetadataAsResult, error) {
	return wrappingExecutor.wrappedInterface.ExecContext(
		executionContext,
		statementWithoutRowsAsResult,
		argumentsForStatement...)
}

// inPostgresqlPersister stores players in a PostgreSQL database.
type inPostgresqlPersister struct {
	connectionArguments  string
	connectionToDatabase LimitedExecutor
}

// NewInPostgresql creates a player state persister which connects to a
// PostgreSQL database by the given connection string.
func NewInPostgresql(connectionArguments string) player.StatePersister {
	return NewInPostgresqlWithGivenLimitedExecutor(
		connectionArguments,
		nil)
}

// NewInPostgresqlWithGivenLimitedExecutor creates a player state persister
// which connects to a PostgreSQL database by the given connection string,
// initialized with the given LimitedExecutor.
func NewInPostgresqlWithGivenLimitedExecutor(
	connectionArguments string,
	connectionToDatabase LimitedExecutor) player.StatePersister {
	return &inPostgresqlPersister{
		connectionArguments:  connectionArguments,
		connectionToDatabase: connectionToDatabase,
	}
}

// Add inserts the given name and color as a row in the database.
func (playerPersister *inPostgresqlPersister) Add(
	executionContext context.Context,
	playerName string,
	chatColor string) error {
	initializedExecutor, errorFromAcquiral :=
		playerPersister.acquireExecutor(executionContext)

	if errorFromAcquiral != nil {
		return errorFromAcquiral
	}

	playerCreationStatement := "INSERT INTO player (name, color) VALUES ($1, $2)"
	_, errorFromExecution :=
		initializedExecutor.ExecuteStatement(
			executionContext,
			playerCreationStatement,
			playerName,
			chatColor)

	return errorFromExecution
}

// UpdateColor updates the given player to have the given chat color. It
// relies on the PostgreSQL driver to ensure thread safety.
func (playerPersister *inPostgresqlPersister) UpdateColor(
	executionContext context.Context,
	playerName string,
	chatColor string) error {
	initializedExecutor, errorFromAcquiral :=
		playerPersister.acquireExecutor(executionContext)

	if errorFromAcquiral != nil {
		return errorFromAcquiral
	}

	playerUpdateStatement := "UPDATE player SET color = $1 WHERE name = $2"
	resultFromExecution, errorFromExecution :=
		initializedExecutor.ExecuteStatement(
			executionContext,
			playerUpdateStatement,
			chatColor,
			playerName)

	return errorUnlessExactlyOneRowAffected(
		playerName,
		resultFromExecution,
		errorFromExecution)
}

// Get returns the ReadOnly corresponding to the given player name if it exists.
func (playerPersister *inPostgresqlPersister) Get(
	executionContext context.Context,
	playerName string) (player.ReadonlyState, error) {
	initializedExecutor, errorFromAcquiral :=
		playerPersister.acquireExecutor(executionContext)

	if errorFromAcquiral != nil {
		return nil, errorFromAcquiral
	}

	playerSelectStatement :=
		"SELECT color FROM player WHERE name = $1"
	playerRows, errorFromExecution :=
		initializedExecutor.ExecuteQuery(
			executionContext,
			playerSelectStatement,
			playerName)
	if errorFromExecution != nil {
		return nil, errorFromExecution
	}

	defer playerRows.Close()

	hasAtLeastOnePlayer := playerRows.Next()
	if !hasAtLeastOnePlayer {
		errorToReturn :=
			fmt.Errorf(
				"No player with name %v is registered",
				playerName)
		return nil, errorToReturn
	}

	playerState :=
		player.ReadAndWriteState{
			PlayerName: playerName,
			ChatColor:  "error: not read in from DB correctly",
		}

	errorFromScan := playerRows.Scan(&playerState.ChatColor)
	if errorFromScan != nil {
		return nil, errorFromScan
	}

	hasMoreThanOnePlayer := playerRows.Next()
	if hasMoreThanOnePlayer {
		errorToReturn :=
			fmt.Errorf(
				"Player with name %v is registered more than once",
				playerName)
		return nil, errorToReturn
	}

	return &playerState, playerRows.Err()
}

// All returns a slice of all the players in the collection as ReadonlyState
// instances, ordered as given by the database.
func (playerPersister *inPostgresqlPersister) All(
	executionContext context.Context) ([]player.ReadonlyState, error) {
	initializedExecutor, errorFromAcquiral :=
		playerPersister.acquireExecutor(executionContext)

	if errorFromAcquiral != nil {
		return nil, errorFromAcquiral
	}

	playerSelectStatement := "SELECT name, color FROM player"
	playerRows, errorFromExecution :=
		initializedExecutor.ExecuteQuery(
			executionContext,
			playerSelectStatement)
	if errorFromExecution != nil {
		return nil, errorFromExecution
	}

	defer playerRows.Close()

	allStates := []player.ReadonlyState{}

	for playerRows.Next() {
		playerState := player.ReadAndWriteState{}
		errorFromScan := playerRows.Scan(&playerState.PlayerName, &playerState.ChatColor)
		if errorFromScan != nil {
			return nil, errorFromScan
		}

		allStates = append(allStates, &playerState)
	}

	return allStates, playerRows.Err()
}

// Delete deletes the given player from the collection. It returns an error
// if the player does not exist before the deletion attempt.
func (playerPersister *inPostgresqlPersister) Delete(
	executionContext context.Context,
	playerName string) error {
	initializedExecutor, errorFromAcquiral :=
		playerPersister.acquireExecutor(executionContext)

	if errorFromAcquiral != nil {
		return errorFromAcquiral
	}

	playerDeletionStatement := "DELETE FROM player WHERE name = $1"
	resultFromExecution, errorFromExecution :=
		initializedExecutor.ExecuteStatement(
			executionContext,
			playerDeletionStatement,
			playerName)

	return errorUnlessExactlyOneRowAffected(
		playerName,
		resultFromExecution,
		errorFromExecution)
}

// acquireExecutor returns the connection to the PostgreSQL database,
// initializing it if it has not already been initialized.
func (playerPersister *inPostgresqlPersister) acquireExecutor(
	executionContext context.Context) (LimitedExecutor, error) {
	if playerPersister.connectionToDatabase == nil {
		errorFromInitialization :=
			playerPersister.initializeExecutor(executionContext)
		if errorFromInitialization != nil {
			return nil, errorFromInitialization
		}
	}

	return playerPersister.connectionToDatabase, nil
}

// initializeExecutor initializes the connection to the PostgreSQL
// database using the given context along with the stored connection
// string, and then creates the player table if it does not yet
// exist in the database.
func (playerPersister *inPostgresqlPersister) initializeExecutor(
	executionContext context.Context) error {
	postgresqlDatabase, errorFromConnection :=
		sql.Open("postgres", playerPersister.connectionArguments)

	// Even if the connection string is junk, sql.Open(...) might not return an
	// error because it might not yet have opened any connection. In this case,
	// the appropriate thing to do is to check the connection with a ping.
	if errorFromConnection == nil {
		errorFromConnection = postgresqlDatabase.Ping()
	}

	if errorFromConnection != nil {
		return errorFromConnection
	}

	// This is PostgreSQL-specific (IF NOT EXISTS), and would not work for another
	// dialect of SQL.
	tableCreationStatement :=
		`CREATE TABLE IF NOT EXISTS player (
			name VARCHAR(255) PRIMARY KEY NOT NULL UNIQUE,
			color VARCHAR(255)
		)`

	playerPersister.connectionToDatabase =
		&wrappingLimitedExecutor{wrappedInterface: postgresqlDatabase}

	_, errorFromExecution :=
		playerPersister.connectionToDatabase.ExecuteStatement(
			executionContext,
			tableCreationStatement)

	return errorFromExecution
}

func errorUnlessExactlyOneRowAffected(
	playerName string,
	resultFromExecution MetadataAsResult,
	errorFromExecution error) error {
	if errorFromExecution != nil {
		return errorFromExecution
	}

	numberOfRowsAffected, errorFromParsing := resultFromExecution.RowsAffected()
	if errorFromParsing != nil {
		return errorFromParsing
	}

	if numberOfRowsAffected != 1 {
		return fmt.Errorf(
			"Expected to affect 1 row (for player %v), instead affected %v rows",
			playerName,
			numberOfRowsAffected)
	}

	return nil
}

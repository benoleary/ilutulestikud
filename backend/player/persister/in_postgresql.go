package persister

import (
	"database/sql"
	"fmt"

	"github.com/benoleary/ilutulestikud/backend/player"

	// We need the side-effects of importing this library, but do not use it directly.
	_ "github.com/lib/pq"
)

// inPostgresqlPersister stores players in a PostgreSQL database.
type inPostgresqlPersister struct {
	connectionToDatabase *sql.DB
}

// NewInPostgresql creates a player state persister which connects to a
// PostgreSQL database by the given connection string.
func NewInPostgresql(connectionArguments string) (player.StatePersister, error) {
	postgresqlDatabase, errorFromConnection :=
		sql.Open("postgres", connectionArguments)
	if errorFromConnection != nil {
		return nil, errorFromConnection
	}

	// This is PostgreSQL-specific (IF NOT EXISTS), and would not work for another
	// dialect of SQL.
	tableCreationStatement :=
		`CREATE TABLE IF NOT EXISTS player (
			name VARCHAR(255) PRIMARY KEY NOT NULL UNIQUE,
			color VARCHAR(255)
		)`

	_, errorFromTableCreation := postgresqlDatabase.Exec(tableCreationStatement)
	if errorFromTableCreation != nil {
		return nil, errorFromTableCreation
	}

	playerPersister :=
		&inPostgresqlPersister{
			connectionToDatabase: postgresqlDatabase,
		}

	return playerPersister, errorFromTableCreation
}

// Add inserts the given name and color as a row in the database.
func (playerPersister *inPostgresqlPersister) Add(
	playerName string,
	chatColor string) error {
	playerCreationStatement := "INSERT INTO player (name, color) VALUES ($1, $2)"
	_, errorFromExecution :=
		playerPersister.connectionToDatabase.Exec(
			playerCreationStatement,
			playerName,
			chatColor)

	return errorFromExecution
}

// UpdateColor updates the given player to have the given chat color. It
// relies on the PostgreSQL driver to ensure thread safety.
func (playerPersister *inPostgresqlPersister) UpdateColor(
	playerName string,
	chatColor string) error {
	playerUpdateStatement := "UPDATE player SET color = $1 WHERE name = $2"
	resultFromExecution, errorFromExecution :=
		playerPersister.connectionToDatabase.Exec(
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
	playerName string) (player.ReadonlyState, error) {
	playerSelectStatement :=
		"SELECT color FROM player WHERE name = $1"
	playerRows, errorFromExecution :=
		playerPersister.connectionToDatabase.Query(playerSelectStatement, playerName)
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
func (playerPersister *inPostgresqlPersister) All() ([]player.ReadonlyState, error) {
	playerSelectStatement := "SELECT name, color FROM player"
	playerRows, errorFromExecution :=
		playerPersister.connectionToDatabase.Query(playerSelectStatement)
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
func (playerPersister *inPostgresqlPersister) Delete(playerName string) error {
	playerDeletionStatement := "DELETE FROM player WHERE name = $1"
	resultFromExecution, errorFromExecution :=
		playerPersister.connectionToDatabase.Exec(
			playerDeletionStatement,
			playerName)

	return errorUnlessExactlyOneRowAffected(
		playerName,
		resultFromExecution,
		errorFromExecution)
}

func errorUnlessExactlyOneRowAffected(
	playerName string,
	resultFromExecution sql.Result,
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

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

	// This is PostgreSQL-specific, and would not work for another dialect of SQL.
	tableCreationStatement :=
		"CREATE TABLE IF NOT EXISTS player (name VARCHAR(255), color VARCHAR(255))"
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
			playerName,
			chatColor)
	if errorFromExecution != nil {
		return errorFromExecution
	}

	numberOfRowsUpdated, errorFromParsing := resultFromExecution.RowsAffected()
	if errorFromParsing != nil {
		return errorFromParsing
	}

	if numberOfRowsUpdated != 1 {
		return fmt.Errorf(
			"Expected to update 1 row (for player %v), instead updated %v rows",
			playerName,
			numberOfRowsUpdated)
	}

	return nil
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
	playerSelectStatement := "SELECT name FROM player"
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
	_, errorFromExecution :=
		playerPersister.connectionToDatabase.Exec(
			playerDeletionStatement,
			playerName)

	return errorFromExecution
}

// Reset removes all players.
func (playerPersister *inPostgresqlPersister) Reset() error {
	playerDeletionStatement := "DELETE FROM player"
	_, errorFromExecution :=
		playerPersister.connectionToDatabase.Exec(
			playerDeletionStatement)

	return errorFromExecution
}

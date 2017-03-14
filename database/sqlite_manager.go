/*
 * Implements a database manager using sqlite3 as the data storage.
 */

package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	cmpb "github.com/nguyenmq/ytbox-go/proto/common"
)

const (
	createUsersTable = `
		CREATE TABLE users (
			user_id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT,
			logged_in BOOLEAN NOT NULL,
			last_access DATETIME NOT NULL); `

	createSongsTable = `
		CREATE TABLE songs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			service TEXT NOT NULL,
			service_id TEXT NOT NULL,
			date DATETIME NOT NULL,
			user_id INTEGER NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(user_id)); `

	insertSong = `
		INSERT INTO songs VALUES
		(NULL, ?, ?, ?, datetime('now'), ?); `

	insertUser = `
		INSERT INTO users VALUES
		(NULL, ?, 1, datetime('now')); `

	queryUserById = `
		SELECT * FROM users WHERE user_id = ?; `

	updateUsername = `
		UPDATE users SET username=?
		WHERE user_id=?; `
)

type SqliteManager struct {
	db *sql.DB
}

/*
 * Clean up resources used by the database manager
 */
func (mgr *SqliteManager) Close() {
	mgr.db.Close()
}

/*
 * Initialize the sqlite database
 */
func (mgr *SqliteManager) Init(dbPath string) {
	fil, err := os.Open(dbPath)
	shouldFound := err != nil

	mgr.db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database connect with error: %v", err)
	}

	if shouldFound {
		foundDatabase(mgr.db)
	} else {
		fil.Close()
	}
}

/*
 * Add a new song to the database
 */
func (mgr *SqliteManager) AddSong(song *cmpb.Song) error {
	stmt, err := mgr.db.Prepare(insertSong)
	if err != nil {
		log.Printf("Error preparing add song statement: %v", err)
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(song.Title, song.Service, song.ServiceId, song.UserId)
	if err != nil {
		log.Printf("Error adding new song: %v", err)
		return err
	}

	songId, err := res.LastInsertId()
	if err != nil {
		log.Printf("Error getting auto-increment id of new song: %v", err)
		return err
	}
	song.SongId = uint32(songId)
	log.Printf("Added new song to db: { %v}", song)

	return nil
}

/*
 * Add a new user to the database
 */
func (mgr *SqliteManager) AddUser(username string) (*UserData, error) {
	stmt, err := mgr.db.Prepare(insertUser)
	if err != nil {
		log.Printf("Error preparing add user statement: %v", err)
		return nil, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(username)
	if err != nil {
		log.Printf("Error adding new user: %v", err)
		return nil, err
	}

	userId, err := res.LastInsertId()
	if err != nil {
		log.Printf("Error getting auto-increment id of new user: %v", err)
		return nil, err
	}
	log.Printf("Added new user: {name: %s, id: %d}", username, userId)

	return mgr.GetUserById(uint32(userId))
}

/*
 * Query for the user data of the givern user id
 */
func (mgr *SqliteManager) GetUserById(userId uint32) (*UserData, error) {
	userData := new(UserData)

	err := mgr.db.QueryRow(queryUserById, userId).Scan(&userData.User.UserId,
		&userData.User.Username, &userData.LoggedIn, &userData.LastAccess)

	// if an error occurred or there was no result, then return nil
	if err != nil {
		return nil, err
	}

	return userData, nil
}

/*
 * Updates the username of an existing user.
 */
func (mgr *SqliteManager) UpdateUsername(username string, userId uint32) error {
	stmt, err := mgr.db.Prepare(updateUsername)
	if err != nil {
		log.Printf("Error preparing update username statement: %v", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, userId)
	if err != nil {
		log.Printf("Error updating username: %v", err)
		return err
	}

	log.Printf("Updated username: {name: %s, id: %d}", username, userId)

	return nil
}

/*
 * Creates a new database with the necessary tables
 */
func foundDatabase(db *sql.DB) {
	_, err := db.Exec(createUsersTable)
	if err != nil {
		log.Fatal("Error creating users table: %v", err)
	}

	_, err = db.Exec(createSongsTable)
	if err != nil {
		log.Fatal("Error creating users table: %v", err)
	}
}

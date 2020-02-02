/*
 * Implements a database manager using sqlite3 as the data storage.
 */

package database

import (
	"database/sql"
	"log"
	"os"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	bepb "github.com/nguyenmq/ytbox-go/proto/backend"
	cmpb "github.com/nguyenmq/ytbox-go/proto/common"
)

const (
	createUsersTable = `
		CREATE TABLE users (
			user_id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT,
			room_id INTEGER NOT NULL,
			logged_in BOOLEAN NOT NULL,
			last_access DATETIME NOT NULL,
			FOREIGN KEY (room_id) REFERENCES rooms(room_id));`

	createRoomsTable = `
		CREATE TABLE rooms (
			room_id INTEGER PRIMARY KEY AUTOINCREMENT,
			room_name TEXT,
			create_date DATETIME NOT NULL,
			last_access DATETIME NOT NULL);`

	createSongsTable = `
		CREATE TABLE songs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			service TEXT NOT NULL,
			service_id TEXT NOT NULL,
			date DATETIME NOT NULL,
			user_id INTEGER NOT NULL,
			room_id INTEGER NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(user_id),
			FOREIGN KEY (room_id) REFERENCES rooms(room_id));`

	enableForeignKeySupport = `
		PRAGMA foreign_keys = ON;`

	insertRoom = `
		INSERT INTO rooms VALUES
		(NULL, ?, datetime('now'), datetime('now'));`

	insertSong = `
		INSERT INTO songs VALUES
		(NULL, ?, ?, ?, datetime('now'), ?, ?);`

	insertUser = `
		INSERT INTO users VALUES
		(NULL, ?, ?, 1, datetime('now'));`

	queryUserById = `
		SELECT * FROM users WHERE user_id = ?;`

	queryRoomByName = `
		SELECT * FROM rooms where room_name = ?;`

	updateUsername = `
		UPDATE users SET username=?
		WHERE user_id=?;`
)

type SqliteManager struct {
	db   *sql.DB
	lock *sync.RWMutex
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
func (mgr *SqliteManager) Init(dbPath string) error {
	fil, err := os.Open(dbPath)
	shouldFound := err != nil

	mgr.db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database connect with error: %v", err)
		return err
	}

	_, err = mgr.db.Exec(enableForeignKeySupport)
	if err != nil {
		log.Fatalf("Error enabling foreign key support: %v", err)
		return err
	}

	if shouldFound {
		if err = foundDatabase(mgr.db); err != nil {
			return err
		}
	} else {
		fil.Close()
	}

	mgr.lock = new(sync.RWMutex)
	return nil
}

/*
 * Add a new song to the database
 */
func (mgr *SqliteManager) AddSong(song *cmpb.Song) error {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	stmt, err := mgr.db.Prepare(insertSong)
	if err != nil {
		log.Printf("Error preparing add song statement: %v", err)
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(song.Title, song.Service, song.ServiceId, song.UserId, song.RoomId)
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
func (mgr *SqliteManager) AddUser(username string, roomId uint32) (*UserData, error) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	stmt, err := mgr.db.Prepare(insertUser)
	if err != nil {
		log.Printf("Error preparing add user statement: %v", err)
		return nil, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(username, roomId)
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

	// todo: fix this. If there's an error in the retrieval, then the user gets
	// created, but we return an error.
	return mgr.unsyncGetUserById(uint32(userId))
}

/*
 * Query for the user data of the given user id
 */
func (mgr *SqliteManager) GetUserById(userId uint32) (*UserData, error) {
	mgr.lock.RLock()
	defer mgr.lock.RUnlock()
	return mgr.unsyncGetUserById(userId)
}

/*
 * Query for the user data of the given user id. This is a helper method that
 * relies on other callers to already have the mutex lock.
 */
func (mgr *SqliteManager) unsyncGetUserById(userId uint32) (*UserData, error) {
	userData := new(UserData)

	err := mgr.db.QueryRow(queryUserById, userId).Scan(&userData.User.UserId,
		&userData.User.Username, &userData.User.RoomId, &userData.LoggedIn, &userData.LastAccess)

	// if an error occurred or there was no result, then return nil
	if err != nil {
		userData = nil
		return nil, err
	}

	return userData, nil
}

/*
 * Updates the username of an existing user.
 */
func (mgr *SqliteManager) UpdateUsername(username string, userId uint32) error {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

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
 * Adds a new room with given name
 */
func (mgr *SqliteManager) AddRoom(roomName string) (*RoomData, error) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	stmt, err := mgr.db.Prepare(insertRoom)
	if err != nil {
		log.Printf("Error preparing add room statement: %v", err)
		return nil, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(roomName)
	if err != nil {
		log.Printf("Error adding new room: %v", err)
		return nil, err
	}

	roomId, err := res.LastInsertId()
	log.Printf("Added new room: {name: %s, id: %d}", roomName, roomId)

	return mgr.unsyncGetRoomByName(roomName)
}

func (mgr *SqliteManager) GetRoomByName(roomName string) (*RoomData, error) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	return mgr.unsyncGetRoomByName(roomName)
}

func (mgr *SqliteManager) unsyncGetRoomByName(roomName string) (*RoomData, error) {
	roomData := new(RoomData)

	err := mgr.db.QueryRow(queryRoomByName, roomName).Scan(&roomData.Room.Id,
		&roomData.Room.Name, &roomData.CreateDate, &roomData.LastAccess)

	if err != nil {
		roomData = nil
		return nil, err
	}

	roomData.Room.Err = &bepb.Error{Success: true}
	return roomData, nil
}

/*
 * Creates a new database with the necessary tables
 */
func foundDatabase(db *sql.DB) error {
	_, err := db.Exec(createRoomsTable)
	if err != nil {
		log.Fatalf("Error creating rooms table: %v", err)
		return err
	}

	_, err = db.Exec(createUsersTable)
	if err != nil {
		log.Fatalf("Error creating users table: %v", err)
		return err
	}

	_, err = db.Exec(createSongsTable)
	if err != nil {
		log.Fatalf("Error creating users table: %v", err)
		return err
	}

	return nil
}

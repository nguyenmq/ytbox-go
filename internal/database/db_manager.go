/*
 * Database manager interface
 */

package database

import (
	"time"

	bepb "github.com/nguyenmq/ytbox-go/internal/proto/backend"
	cmpb "github.com/nguyenmq/ytbox-go/internal/proto/common"
)

type UserData struct {
	User       bepb.User
	LoggedIn   bool
	LastAccess time.Time
}

type RoomData struct {
	Room       bepb.Room
	CreateDate time.Time
	LastAccess time.Time
}

/*
 * Interface for manager the backend database
 */
type DbManager interface {
	// Add a new song to the database
	AddSong(song *cmpb.Song) error

	// Add a new user to the users table and returns the user's id
	AddUser(username string, roomId uint32) (*UserData, error)

	// Add a new room to the database
	AddRoom(roomName string) (*RoomData, error)

	// Close the database connection
	Close()

	// Get user by id
	GetUserById(userId uint32) (*UserData, error)

	// Updates the given user's name
	UpdateUsername(username string, userId uint32) error

	// Queries for a room given its name
	GetRoomByName(roomName string) (*RoomData, error)

	// Initialize the database interface
	Init(dbPath string) error
}

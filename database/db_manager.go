/*
 * Database manager interface
 */

package database

import (
	"time"

	pb "github.com/nguyenmq/ytbox-go/proto/backend"
)

type UserData struct {
	User       pb.User
	LoggedIn   bool
	LastAccess time.Time
}

/*
 * Interface for manager the backend database
 */
type DbManager interface {
	// Add a new song to the database
	AddSong(song *pb.Song) error

	// Add a new user to the users table and returns the user's id
	AddUser(username string) (*UserData, error)

	// Close the database connection
	Close()

	// Get user by id
	GetUserById(userId uint32) (*UserData, error)

	// Updates the given user's name
	UpdateUsername(username string, userId uint32) error

	// Initialize the database interface
	Init(dbPath string)
}

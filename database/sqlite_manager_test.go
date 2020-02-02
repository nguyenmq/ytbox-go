/*
 * Tests for the sqlite database manager
 */

package database

import (
	"os"
	"testing"

	sqlite "github.com/mattn/go-sqlite3"
	cmpb "github.com/nguyenmq/ytbox-go/proto/common"
)

const (
	testDbLocation = "/tmp/test_db.db"
	testRoomId     = 1
	testRoomName   = "Wizard's Keep"
	testSongId     = 1
	testUserId     = 1
	testUserName   = "Zedd"
)

var testSong = cmpb.Song{"Bags!!", 0, testUserName, testUserId, cmpb.ServiceType_Youtube, "0xdeadbeef", testRoomId}

func initDatabase() (*SqliteManager, error) {
	dbManager := new(SqliteManager)
	err := dbManager.Init(testDbLocation)
	return dbManager, err
}

func cleanUp(dbManager *SqliteManager) {
	if dbManager != nil {
		dbManager.Close()
	}

	os.Remove(testDbLocation)
}

func TestInit_when_success(t *testing.T) {
	dbManager, err := initDatabase()

	if err != nil {
		t.Error("Error when initializing the database", err)
	}

	cleanUp(dbManager)
}

func TestAddRoom_when_success(t *testing.T) {
	dbManager, err := initDatabase()

	if err != nil {
		t.Error("Error when initializing the database", err)
	}

	err = dbManager.AddRoom(testRoomName)
	if err != nil {
		t.Error("Error when adding new room", err)
	}

	cleanUp(dbManager)
}

func TestAddUser_when_success(t *testing.T) {
	dbManager, err := initDatabase()

	if err != nil {
		t.Error("Error when initializing the database", err)
	}

	userData, err := dbManager.AddUser(testUserName)
	if err != nil {
		t.Error("Error when adding new user", err)
	}

	expectedUserName := testUserName
	if userData.User.Username != expectedUserName {
		t.Error("Username should be", expectedUserName, "but was", userData.User.Username)
	}

	var expectedUserId uint32 = testUserId
	if userData.User.UserId != expectedUserId {
		t.Error("User id should be", expectedUserId, "but was", userData.User.UserId)
	}

	cleanUp(dbManager)
}

func TestAddSong_when_success(t *testing.T) {
	dbManager, err := initDatabase()

	if err != nil {
		t.Error("Error when initializing the database", err)
	}

	err = dbManager.AddRoom(testRoomName)
	if err != nil {
		t.Error("Error when adding new room", err)
	}

	_, err = dbManager.AddUser(testUserName)
	if err != nil {
		t.Error("Error when adding new user", err)
	}

	actualSong := testSong
	err = dbManager.AddSong(&actualSong)
	if err != nil {
		t.Error("Error when adding new song", err)
	}

	if actualSong.SongId != testSongId {
		t.Error("DB manager should set the song id to", testSongId)
	}

	cleanUp(dbManager)
}

func TestAddSong_whenRoomDoesNotExist_failToAdd(t *testing.T) {
	dbManager, err := initDatabase()

	if err != nil {
		t.Error("Error when initializing the database", err)
	}

	err = dbManager.AddRoom(testRoomName)
	if err != nil {
		t.Error("Error when adding new room", err)
	}

	_, err = dbManager.AddUser(testUserName)
	if err != nil {
		t.Error("Error when adding new user", err)
	}

	actualSong := testSong
	actualSong.RoomId = testRoomId + 1
	err = dbManager.AddSong(&actualSong)
	if err == nil {
		t.Error("DB manager did not return an error when adding a song with a room id that doesn't exist")
	}

	if err.(sqlite.Error).Code != sqlite.ErrConstraint {
		t.Error("DB manager did not fail with a constraint mismatch error:", err.(sqlite.Error))
	}

	cleanUp(dbManager)
}

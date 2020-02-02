/*
 * Tests for the sqlite database manager
 */

package database

import (
	"database/sql"
	"errors"
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

	expectedRoomName := testRoomName
	expectedRoomId := uint32(testRoomId)
	actualRoomData, err := dbManager.AddRoom(testRoomName)
	if err != nil {
		t.Error("Error when adding new room", err)
	}

	if actualRoomData.Room.Name != expectedRoomName {
		t.Error("DB manager should return expected room name", expectedRoomName, "but was", actualRoomData.Room.Name)
	}

	if actualRoomData.Room.Id != expectedRoomId {
		t.Error("DB manager should return expected room id", expectedRoomId, "but was", actualRoomData.Room.Id)
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

	_, err = dbManager.AddRoom(testRoomName)
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

	_, err = dbManager.AddUser(testUserName)
	if err != nil {
		t.Error("Error when adding new user", err)
	}

	actualSong := testSong
	actualSong.RoomId = testRoomId
	err = dbManager.AddSong(&actualSong)
	if err == nil {
		t.Error("DB manager did not return an error when adding a song with a room id that doesn't exist")
	}

	if err.(sqlite.Error).Code != sqlite.ErrConstraint {
		t.Error("DB manager did not fail with a constraint mismatch error:", err.(sqlite.Error))
	}

	cleanUp(dbManager)
}

func TestGetRoomByName_when_success(t *testing.T) {
	dbManager, err := initDatabase()

	if err != nil {
		t.Error("Error when initializing the database", err)
	}

	_, err = dbManager.AddRoom(testRoomName)
	if err != nil {
		t.Error("Error when adding new room", err)
	}

	expectedRoomName := testRoomName
	expectedRoomId := uint32(testRoomId)

	roomData, err := dbManager.GetRoomByName(expectedRoomName)
	if err != nil {
		t.Error("Get room by name failed with error:", err)
	}

	if roomData.Room.Name != expectedRoomName {
		t.Error("DB manager did not fetch the correct room name:", roomData.Room.Name)
	}

	if roomData.Room.Id != expectedRoomId {
		t.Error("DB manager did not fetch the correct room id:", roomData.Room.Id)
	}

	cleanUp(dbManager)
}

func TestGetRoomByName_whenRoomDoesNotExist_returnsNil(t *testing.T) {
	dbManager, err := initDatabase()

	if err != nil {
		t.Error("Error when initializing the database", err)
	}

	expectedRoomName := testRoomName
	roomData, err := dbManager.GetRoomByName(expectedRoomName)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Error("DB manager should return no rows error when room doesn't exist")
	}

	if roomData != nil {
		t.Error("DB manager should return nil room data when the room doesn't exist")
	}

	cleanUp(dbManager)
}

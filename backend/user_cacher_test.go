package backend

import (
	"testing"
)

const (
	testUserId   = 1
	testRoomId   = 1
	testUserName = "Kahlan"
)

func setup() *UserCache {
	cache := new(UserCache)
	cache.Init()
	return cache
}

func TestLookupUsername_when_success(t *testing.T) {
	cache := setup()
	cache.AddUserToCache(testUserId, testUserName, testRoomId)

	name, exists := cache.LookupUsername(testUserId)

	if !exists {
		t.Fatalf("User should exist")
	}

	if name != testUserName {
		t.Fatalf("Cache should return %s, but was %s", testUserName, name)
	}
}

func TestLookupRoomId_when_success(t *testing.T) {
	cache := setup()
	cache.AddUserToCache(testUserId, testUserName, testRoomId)

	roomId, exists := cache.LookupRoomId(testRoomId)

	if !exists {
		t.Fatalf("User's room id should exist")
	}

	if roomId != uint32(testRoomId) {
		t.Fatalf("Cache should return %d, but was %d", testRoomId, uint32(testRoomId))
	}
}

/*
 * Implements an in-memory cache for usernames and user ids
 */

package backend

import (
	"log"
	"sync"
)

type UserEntry struct {
	name   string
	roomId uint32
}

type UserCache struct {
	lock  *sync.RWMutex         // RW mutex on the cache
	cache map[uint32]*UserEntry // in-memory cache of user id -> username
}

/*
 * Initialize the user cache
 */
func (c *UserCache) Init() {
	c.lock = new(sync.RWMutex)
	c.cache = make(map[uint32]*UserEntry)
}

/*
 * Find the username associated with the user id
 */
func (c *UserCache) LookupUsername(userId uint32) (string, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	entry, exists := c.cache[userId]

	if exists {
		return entry.name, exists
	} else {
		return "", exists
	}
}

func (c *UserCache) LookupRoomId(userId uint32) (uint32, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	entry, exists := c.cache[userId]

	if exists {
		return entry.roomId, exists
	} else {
		return 0, exists
	}
}

/*
 * Adds a username and id to the cache
 */
func (c *UserCache) AddUserToCache(userId uint32, username string, roomId uint32) {
	c.lock.Lock()
	defer c.lock.Unlock()

	entry, exists := c.cache[userId]
	if exists && entry.name != username {
		log.Printf("Warning: overwriting cached user: {id: %d, cur name: %s, new name: %s}",
			userId, entry.name, username)
	} else {
		log.Printf("Cached: {user id: %d, username: %s}", userId, username)
	}

	c.cache[userId] = &UserEntry{username, roomId}
}

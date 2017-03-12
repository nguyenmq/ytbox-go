/*
 * Implements an in-memory cache for usernames and user ids
 */

package backend

import (
	"log"
	"sync"
)

type UserCache struct {
	lock  *sync.RWMutex     // RW mutex on the cache
	cache map[uint32]string // in-memory cache of user id -> username
}

/*
 * Initialize the user cache
 */
func (c *UserCache) Init() {
	c.lock = new(sync.RWMutex)
	c.cache = make(map[uint32]string)
}

/*
 * Find the username associated with the user id
 */
func (c *UserCache) LookupUsername(userId uint32) (string, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	username, exists := c.cache[userId]
	return username, exists
}

/*
 * Adds a username and id to the cache
 */
func (c *UserCache) AddUserToCache(userId uint32, username string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	cachedName, exists := c.cache[userId]
	if exists && username == cachedName {
		log.Printf("Warning: overwriting cached user: {id: %d, cur name: %s, new name: %s", userId, cachedName, username)
	} else {
		log.Printf("Cached: {user id: %d, username: %s}", userId, username)
	}

	c.cache[userId] = username
}

/*
 * Manages the song queue and the state currently playing song.
 */

package song_queue

import (
	"io/ioutil"
	"log"
	"sync"

	"github.com/golang/protobuf/proto"

	bepb "github.com/nguyenmq/ytbox-go/proto/backend"
	cmpb "github.com/nguyenmq/ytbox-go/proto/common"
)

const (
	QueueSnapshot string = "/tmp/ytbox.queue" // location of the queue snapshot
)

/*
 * Manages the song queue
 */
type SongQueueManager struct {
	queue      songQueuer    // the playlist of songs
	lock       *sync.RWMutex // read/write lock on the playlist
	npLock     *sync.Mutex   // lock on the now playing value
	cLock      *sync.Mutex   // mutex for condition variable
	cond       *sync.Cond    // condition variable on the queue
	nowPlaying *cmpb.Song    // the currently playing song
}

/*
 * Initializes the queue
 */
func (manager *SongQueueManager) Init(queuer songQueuer) {
	manager.queue = queuer
	manager.lock = new(sync.RWMutex)
	manager.npLock = new(sync.Mutex)
	manager.cLock = new(sync.Mutex)
	manager.cond = sync.NewCond(manager.cLock)
}

/*
 * Adds a song to the queue
 */
func (manager *SongQueueManager) AddSong(song *cmpb.Song) {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	manager.queue.push(song)

	if manager.queue.length() == 1 {
		manager.cond.Broadcast()
	}
}

/*
 * Returns the length of the queue
 */
func (manager *SongQueueManager) Len() int {
	manager.lock.RLock()
	defer manager.lock.RUnlock()
	return manager.queue.length()
}

/*
 * Clear the now playing state
 */
func (manager *SongQueueManager) ClearNowPlaying() {
	manager.npLock.Lock()
	defer manager.npLock.Unlock()

	manager.nowPlaying = nil
}

/*
 * Returns the data for the currently playing song
 */
func (manager *SongQueueManager) NowPlaying() *cmpb.Song {
	manager.npLock.Lock()
	defer manager.npLock.Unlock()

	return manager.nowPlaying
}

/*
 * Returns a list of songs in the queue
 */
func (manager *SongQueueManager) GetPlaylist() *bepb.Playlist {
	songs := make([]*cmpb.Song, manager.queue.length())
	var idx int = 0

	manager.lock.RLock()
	defer manager.lock.RUnlock()

	for e := manager.queue.front(); e != nil; e = e.next() {
		songs[idx] = e.value()
		idx++
	}

	return &bepb.Playlist{Songs: songs}
}

/*
 * Blocks the current thread while the size of the playlist is zero. The playlist
 * will notify all blocked threads that the size is once again greater than one
 * when a new song is added.
 */
func (manager *SongQueueManager) WaitForMoreSongs() {
	manager.cond.L.Lock()
	for manager.Len() == 0 {
		manager.ClearNowPlaying()
		manager.cond.Wait()
	}
	manager.cond.L.Unlock()
}

/*
 * Pops the next song off the queue and returns it
 */
func (manager *SongQueueManager) PopQueue() *cmpb.Song {
	manager.npLock.Lock()
	defer manager.npLock.Unlock()
	manager.nowPlaying = nil

	manager.lock.Lock()
	defer manager.lock.Unlock()

	if manager.queue.length() > 0 {
		manager.nowPlaying = manager.queue.pop()
	}

	return manager.nowPlaying
}

/*
 * Removes the identified song from the queue. Both the song id and uesr id
 * must match in order for the song to be successfully removed.
 */
func (manager *SongQueueManager) RemoveSong(songId uint32, userId uint32) error {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	return manager.queue.remove(songId, userId)
}

/*
 * Saves the playlist to a file
 */
func (manager *SongQueueManager) SavePlaylist(path string) error {
	playlist := manager.GetPlaylist()

	out, err := proto.Marshal(playlist)
	if err != nil {
		log.Printf("Failed to encode Playlist with error: %v", err)
		return err
	}

	err = ioutil.WriteFile(path, out, 0644)
	if err != nil {
		log.Printf("Failed to write playlist to file \"%s\" with error: %v", path, err)
		return err
	}

	return nil
}

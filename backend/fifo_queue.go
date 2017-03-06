/*
 * Implements a first-in-first-out song queue
 */

package ytbbe

import (
	"container/list"
	"sync"
)

/*
 * Data for the currently playing song
 */
var nowPlaying *SongData = nil

/*
 * Contains the state data for the queue
 */
type FifoQueue struct {
	playQueue *list.List
	lock      *sync.Mutex
}

/*
 * Adds a song to the queue
 */
func (fifo *FifoQueue) AddSong(song *SongData) {
	fifo.lock.Lock()
	defer fifo.lock.Unlock()
	fifo.playQueue.PushBack(song)
}

/*
 * Initializes the queue
 */
func (fifo *FifoQueue) Init() {
	fifo.playQueue = list.New()
	fifo.lock = new(sync.Mutex)
}

/*
 * Returns the length of the queue
 */
func (fifo *FifoQueue) Len() int {
	fifo.lock.Lock()
	defer fifo.lock.Unlock()
	return fifo.playQueue.Len()
}

/*
 * Returns the data for the currently playing song
 */
func (fifo *FifoQueue) NowPlaying() *SongData {
	return nowPlaying
}

/*
 * Returns a list of songs in the queue
 */
func (fifo *FifoQueue) Playlist() PlaylistType {
	var playlist PlaylistType = make(PlaylistType, fifo.playQueue.Len())
	var idx int = 0

	fifo.lock.Lock()
	defer fifo.lock.Unlock()

	for e := fifo.playQueue.Front(); e != nil; e = e.Next() {
		playlist[idx] = e.Value.(*SongData)
		idx++
	}

	return playlist
}

/*
 * Pops the next song off the queue and returns it
 */
func (fifo *FifoQueue) PopSong() *SongData {
	var front *SongData = nil
	nowPlaying = nil

	fifo.lock.Lock()
	defer fifo.lock.Unlock()

	if fifo.playQueue.Len() > 0 {
		front = fifo.playQueue.Remove(fifo.playQueue.Front()).(*SongData)
		nowPlaying = front
	}

	return front
}

/*
 * Removes the identified song from the queue
 */
func (fifo *FifoQueue) RemoveSong(serviceID string, userID UserIDType) bool {
	var found bool = false

	fifo.lock.Lock()
	defer fifo.lock.Unlock()

	for e := fifo.playQueue.Front(); e != nil; e = e.Next() {
		var song *SongData = e.Value.(*SongData)

		if song.ServiceID == serviceID && song.UserID == userID {
			fifo.playQueue.Remove(e)
			found = true
			break
		}
	}

	return found
}

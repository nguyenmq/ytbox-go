/*
 * Implements a first-in-first-out song queue
 */

package scheduler

import (
	"container/list"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"sync"

	"github.com/golang/protobuf/proto"

	bepb "github.com/nguyenmq/ytbox-go/proto/backend"
	cmpb "github.com/nguyenmq/ytbox-go/proto/common"
)

/*
 * Data for the currently playing song
 */
var nowPlaying *cmpb.Song = nil

/*
 * Contains the state data for the queue
 */
type FifoQueue struct {
	playQueue *list.List    // the playlist of songs
	lock      *sync.RWMutex // read/write lock on the playlist
	npLock    *sync.Mutex   // lock on the now playing value
	cLock     *sync.Mutex   // mutex for condition variable
	cond      *sync.Cond    // condition variable on the queue
}

/*
 * Adds a song to the queue
 */
func (fifo *FifoQueue) AddSong(song *cmpb.Song) {
	fifo.lock.Lock()
	defer fifo.lock.Unlock()
	fifo.playQueue.PushBack(song)

	if fifo.playQueue.Len() == 1 {
		fifo.cond.Broadcast()
	}
}

/*
 * Initializes the queue
 */
func (fifo *FifoQueue) Init() {
	fifo.playQueue = list.New()
	fifo.lock = new(sync.RWMutex)
	fifo.npLock = new(sync.Mutex)
	fifo.cLock = new(sync.Mutex)
	fifo.cond = sync.NewCond(fifo.cLock)
}

/*
 * Returns the length of the queue
 */
func (fifo *FifoQueue) Len() int {
	fifo.lock.RLock()
	defer fifo.lock.RUnlock()
	return fifo.playQueue.Len()
}

/*
 * Returns the data for the currently playing song
 */
func (fifo *FifoQueue) NowPlaying() *cmpb.Song {
	fifo.npLock.Lock()
	defer fifo.npLock.Unlock()

	return nowPlaying
}

/*
 * Returns a list of songs in the queue
 */
func (fifo *FifoQueue) GetPlaylist() *bepb.Playlist {
	songs := make([]*cmpb.Song, fifo.playQueue.Len())
	var idx int = 0

	fifo.lock.RLock()
	defer fifo.lock.RUnlock()

	for e := fifo.playQueue.Front(); e != nil; e = e.Next() {
		songs[idx] = e.Value.(*cmpb.Song)
		idx++
	}

	return &bepb.Playlist{Songs: songs}
}

/*
 * Returns a condition variable on the queue
 */
func (fifo *FifoQueue) GetConditionVar() *sync.Cond {
	return fifo.cond
}

/*
 * Pops the next song off the queue and returns it
 */
func (fifo *FifoQueue) PopQueue() *cmpb.Song {
	var front *cmpb.Song = nil

	fifo.npLock.Lock()
	defer fifo.npLock.Unlock()
	nowPlaying = nil

	fifo.lock.Lock()
	defer fifo.lock.Unlock()

	if fifo.playQueue.Len() > 0 {
		front = fifo.playQueue.Remove(fifo.playQueue.Front()).(*cmpb.Song)

		nowPlaying = front
	}

	return front
}

/*
 * Removes the identified song from the queue. Both the song id and uesr id
 * must match in order for the song to be successfully removed.
 */
func (fifo *FifoQueue) RemoveSong(songId uint32, userId uint32) error {
	fifo.lock.Lock()
	defer fifo.lock.Unlock()

	for e := fifo.playQueue.Front(); e != nil; e = e.Next() {
		var song *cmpb.Song = e.Value.(*cmpb.Song)

		if song.GetSongId() == songId {
			if song.GetUserId() == userId {
				fifo.playQueue.Remove(e)
				return nil
			} else {
				return errors.New(fmt.Sprintf("The user id %d for song %d does not match the id of the submitter",
					userId, songId))
			}
		}
	}

	return errors.New(fmt.Sprintf("Song with id %d does not exist in the queue", songId))
}

/*
 * Saves the playlist to a file
 */
func (fifo *FifoQueue) SavePlaylist(path string) error {
	playlist := fifo.GetPlaylist()

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

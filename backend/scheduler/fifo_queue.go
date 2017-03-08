/*
 * Implements a first-in-first-out song queue
 */

package scheduler

import (
	"container/list"
	"sync"

	pb "github.com/nguyenmq/ytbox-go/proto/backend"
)

/*
 * Data for the currently playing song
 */
var nowPlaying *pb.Song = nil

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
func (fifo *FifoQueue) AddSong(song *pb.Song) {
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
func (fifo *FifoQueue) NowPlaying() *pb.Song {
	return nowPlaying
}

/*
 * Returns a list of songs in the queue
 */
func (fifo *FifoQueue) GetPlaylist() *pb.Playlist {
	songs := make([]*pb.Song, fifo.playQueue.Len())
	var idx int = 0

	fifo.lock.Lock()
	defer fifo.lock.Unlock()

	for e := fifo.playQueue.Front(); e != nil; e = e.Next() {
		songs[idx] = e.Value.(*pb.Song)
		idx++
	}

	return &pb.Playlist{Songs: songs}
}

/*
 * Pops the next song off the queue and returns it
 */
func (fifo *FifoQueue) PopQueue() *pb.Song {
	var front *pb.Song = nil
	nowPlaying = nil

	fifo.lock.Lock()
	defer fifo.lock.Unlock()

	if fifo.playQueue.Len() > 0 {
		front = fifo.playQueue.Remove(fifo.playQueue.Front()).(*pb.Song)
		nowPlaying = front
	}

	return front
}

/*
 * Removes the identified song from the queue
 */
func (fifo *FifoQueue) RemoveSong(serviceId string, userId uint32) bool {
	var found bool = false

	fifo.lock.Lock()
	defer fifo.lock.Unlock()

	for e := fifo.playQueue.Front(); e != nil; e = e.Next() {
		var song *pb.Song = e.Value.(*pb.Song)

		if song.ServiceId == serviceId && song.UserId == userId {
			fifo.playQueue.Remove(e)
			found = true
			break
		}
	}

	return found
}

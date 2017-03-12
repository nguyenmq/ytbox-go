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
	lock      *sync.RWMutex
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
	fifo.lock = new(sync.RWMutex)
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
func (fifo *FifoQueue) NowPlaying() *pb.Song {
	return nowPlaying
}

/*
 * Returns a list of songs in the queue
 */
func (fifo *FifoQueue) GetPlaylist() *pb.Playlist {
	songs := make([]*pb.Song, fifo.playQueue.Len())
	var idx int = 0

	fifo.lock.RLock()
	defer fifo.lock.RUnlock()

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
 * Removes the identified song from the queue. Both the song id and uesr id
 * must match in order for the song to be successfully removed.
 */
func (fifo *FifoQueue) RemoveSong(songId uint32, userId uint32) error {
	fifo.lock.Lock()
	defer fifo.lock.Unlock()

	for e := fifo.playQueue.Front(); e != nil; e = e.Next() {
		var song *pb.Song = e.Value.(*pb.Song)

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

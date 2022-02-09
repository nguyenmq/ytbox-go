/*
 * Implements a first-in-first-out songQueuer
 */

package song_queue

import (
	"container/list"
	"errors"
	"fmt"

	cmpb "github.com/nguyenmq/ytbox-go/internal/proto/common"
)

/*
 * Contains the state data for the queue
 */
type FifoQueuer struct {
	queue *list.List // the playlist of songs
}

func NewFifoQueuer() *FifoQueuer {
	fifo := new(FifoQueuer)
	fifo.queue = list.New()
	return fifo
}

func (fifo *FifoQueuer) push(song *cmpb.Song) {
	fifo.queue.PushBack(song)
}

func (fifo *FifoQueuer) length() int {
	return fifo.queue.Len()
}

func (fifo *FifoQueuer) pop() *cmpb.Song {
	if fifo.queue.Len() > 0 {
		return fifo.queue.Remove(fifo.queue.Front()).(*cmpb.Song)
	}

	return nil
}

func (fifo *FifoQueuer) RemoveSong(songId uint32, userId uint32) error {
	for e := fifo.queue.Front(); e != nil; e = e.Next() {
		var song *cmpb.Song = e.Value.(*cmpb.Song)

		if song.GetSongId() == songId && song.GetUserId() == userId {
			fifo.queue.Remove(e)
			return nil
		}
	}

	return errors.New(fmt.Sprintf("Song with id %d does not exist in the queue", songId))
}

func (fifo *FifoQueuer) front() fifoElement {
	return fifoElement{
		current: fifo.queue.Front(),
	}
}

type fifoElement struct {
	current *list.Element
}

func (e fifoElement) value() *cmpb.Song {
	return e.current.Value.(*cmpb.Song)
}

func (e fifoElement) next() fifoElement {
	return fifoElement{
		current: e.current.Next(),
	}
}

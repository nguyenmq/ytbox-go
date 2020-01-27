package song_queue

import (
	cmpb "github.com/nguyenmq/ytbox-go/proto/common"
)

/*
 * A songQueuer maintains a list of songs
 */
type songQueuer interface {
	// Get an element pointer to the front of the queue. User for interation
	front() queueElement

	// Get the length of the playlist
	length() int

	// Pop a song off the front of the queue. Returns nil if queue is empty.
	pop() *cmpb.Song

	// Push a new song onto the queue
	push(song *cmpb.Song)

	// Remove song from the queue
	remove(songId uint32, userId uint32) error
}

type queueElement interface {
	// Get the song at this element in the queue
	value() *cmpb.Song

	// Get the next song in the queue. The end of the queue is reached when nil
	// is returned.
	next() queueElement
}

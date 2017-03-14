/*
 * This file describes the types and interfaces used by yt_box backend
 */

package scheduler

import (
	"sync"

	bepb "github.com/nguyenmq/ytbox-go/proto/backend"
	cmpb "github.com/nguyenmq/ytbox-go/proto/common"
)

/*
 * Interface to a queue scheduler. A queue scheduler represents a queue of
 * songs. The queue has the ability to prioritize the song in the queue however
 * it sees fit.
 */
type QueueScheduler interface {
	// Add a song to the queue
	AddSong(song *cmpb.Song)

	// Initialize the queue scheduler
	Init()

	// Get the length of the playlist
	Len() int

	// Get the song that's now playing
	NowPlaying() *cmpb.Song

	// Get a list of the currents songs in the queue
	GetPlaylist() *bepb.Playlist

	// Get condition variable on the queue
	GetConditionVar() *sync.Cond

	// Pop a song off the queue
	PopQueue() *cmpb.Song

	// Remove song from the queue
	RemoveSong(songId uint32, userId uint32) error

	// Saves the playlist to a file
	SavePlaylist(path string) error
}

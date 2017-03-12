/*
 * This file describes the types and interfaces used by yt_box backend
 */

package scheduler

import (
	pb "github.com/nguyenmq/ytbox-go/proto/backend"
)

/*
 * Interface to a queue scheduler. A queue scheduler represents a queue of
 * songs. The queue has the ability to prioritize the song in the queue however
 * it sees fit.
 */
type QueueScheduler interface {
	// Add a song to the queue
	AddSong(song *pb.Song)

	// Initialize the queue scheduler
	Init()

	// Get the length of the playlist
	Len() int

	// Get the song that's now playing
	NowPlaying() *pb.Song

	// Get a list of the currents songs in the queue
	GetPlaylist() *pb.Playlist

	// Pop a song off the queue
	PopQueue() *pb.Song

	// Remove song from the queue
	RemoveSong(songId uint32, userId uint32) error

	// Saves the playlist to a file
	SavePlaylist(path string) error
}

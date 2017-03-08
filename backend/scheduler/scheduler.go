package scheduler

/*
 * This file describes the types and interfaces used by yt_box backend
 */

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
	PopSong() *pb.Song

	// Remove song from the queue
	RemoveSong(serviceId string, userId uint32) bool
}

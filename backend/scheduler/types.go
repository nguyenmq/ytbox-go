package scheduler

/*
 * This file describes the types and interfaces used by yt_box backend
 */

import (
	"fmt"

	pb "github.com/nguyenmq/ytbox-go/proto"
)

/*
 * Describes a song in the queue and the data that's necessary to track it
 * throughout the system
 */
type SongData struct {
	// Title of song
	Title string

	// Id of the song stores in our database
	SongId uint32

	// Identifier of the music service the song is on
	Service pb.ServiceType

	// Id used by the service the song resides on
	ServiceId string

	// Username of submitter
	Username string

	// User Id of submitter
	UserId uint32
}

/*
 * Stringer for a *SongData
 */
func (song *SongData) String() string {
	return fmt.Sprintf("{title: %s, serviceId: %s, songId: %d, username: %s, userId: %d}",
		song.Title,
		song.ServiceId,
		song.SongId,
		song.Username,
		song.UserId)
}

/*
 * A list of songs
 */
type PlaylistType []*SongData

/*
 * Interface to a queue scheduler. A queue scheduler represents a queue of
 * songs. The queue has the ability to prioritize the song in the queue however
 * it sees fit.
 */
type QueueScheduler interface {
	// Add a song to the queue
	AddSong(song *SongData)

	// Initialize the queue scheduler
	Init()

	// Get the length of the playlist
	Len() int

	// Get the song that's now playing
	NowPlaying() *SongData

	// Get a list of the currents songs in the queue
	Playlist() PlaylistType

	// Pop a song off the queue
	PopSong() *SongData

	// Remove song from the queue
	RemoveSong(serviceId string, userId uint32) bool
}

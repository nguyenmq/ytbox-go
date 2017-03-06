package ytbbe

/*
 * This file describes the types and interfaces used by yt_box backend
 */

import (
	"fmt"
)

/*
 * Enum type for the music service identifier
 */
type ServiceType uint32

/*
 * Enum for the music service identifier
 */
const (
	ServiceYoutube ServiceType = iota
	ServiceSpotify
)

/*
 * ID types
 */
type UserIDType uint32
type SongIDType uint32

/*
 * Describes a song in the queue and the data that's necessary to track it
 * throughout the system
 */
type SongData struct {
	// Title of song
	Title string

	// ID of the song
	SongID SongIDType

	// Identifier of the music service the song is on
	Service ServiceType

	ServiceID string

	// Username of submitter
	Username string

	// User ID of submitter
	UserID UserIDType
}

/*
 * Stringer for a *SongData
 */
func (song *SongData) String() string {
	return fmt.Sprintf("(%s, %s, %d, %s, %d)",
		song.Title,
		song.ServiceID,
		song.SongID,
		song.Username,
		song.UserID)
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
	RemoveSong(serviceID string, userID UserIDType) bool
}

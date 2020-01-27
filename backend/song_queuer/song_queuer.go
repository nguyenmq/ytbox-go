package song_queue

import (
	bepb "github.com/nguyenmq/ytbox-go/proto/backend"
	cmpb "github.com/nguyenmq/ytbox-go/proto/common"
)

/*
 * A SongQueuer maintains a list of songs in its queue and the state of which
 * is the currently playing song.
 */
type SongQueuer interface {
	// Add a song to the queue
	AddSong(song *cmpb.Song)

	// Initialize the queue
	Init()

	// Get the length of the playlist
	Len() int

	// Clear the now playing state
	ClearNowPlaying()

	// Get the song that's now playing
	NowPlaying() *cmpb.Song

	// Get a list of the currents songs in the queue
	GetPlaylist() *bepb.Playlist

	// Blocks the current thread while the size of the playlist is zero
	WaitForMoreSongs()

	// Pop a song off the front of the queue
	PopQueue() *cmpb.Song

	// Remove song from the queue
	RemoveSong(songId uint32, userId uint32) error

	// Saves the playlist to a file
	SavePlaylist(path string) error
}

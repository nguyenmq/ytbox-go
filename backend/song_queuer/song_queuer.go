/*
 * This file describes the types and interfaces used by yt_box backend
 */

package song_queue

import (
	bepb "github.com/nguyenmq/ytbox-go/proto/backend"
	cmpb "github.com/nguyenmq/ytbox-go/proto/common"
)

/*
 * A SongQueuer maintains a list of songs.
 */
type SongQueuer interface {
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

	// Blocks the current thread while the size of the playlist is zero
	WaitForMoreSongs()

	// Pop a song off the queue
	PopQueue() *cmpb.Song

	// Remove song from the queue
	RemoveSong(songId uint32, userId uint32) error

	// Saves the playlist to a file
	SavePlaylist(path string) error
}

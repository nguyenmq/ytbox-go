package song_queue

import (
	"testing"

	cmpb "github.com/nguyenmq/ytbox-go/proto/common"
)

/*
 * List of sample song data to test against
 */
var sampleSongs = []cmpb.Song{
	{"title 1", 1, "Kid A", 1, cmpb.ServiceType_Youtube, "0xdeadbeef"},
	{"title 2", 2, "Kid B", 2, cmpb.ServiceType_Youtube, "0xba5eba11"},
	{"title 3", 3, "Kid A", 1, cmpb.ServiceType_Youtube, "0xf01dab1e"},
	{"title 4", 4, "Kid B", 2, cmpb.ServiceType_Youtube, "0xb01dface"},
	{"title 5", 5, "Kid A", 1, cmpb.ServiceType_Youtube, "0xca55e77e"},
}

/*
 * Compares two songs and returns true if they are the same
 */
func compareSongs(first *cmpb.Song, second *cmpb.Song) bool {
	var same bool = false
	same = (first.Title == second.Title)
	same = (first.ServiceId == second.ServiceId) && same
	same = (first.SongId == second.SongId) && same
	same = (first.Username == second.Username) && same
	same = (first.UserId == second.UserId) && same

	return same
}

/*
 * Tests an empty queue
 */
func TestEmptyQueue(t *testing.T) {
	fifo := NewFifoQueuer()

	var nextSong *cmpb.Song = fifo.pop()
	if nextSong != nil {
		t.Error("Expected nil, but got", nextSong)
	}
}

/*
 * Tests a queue with a single item
 */
func TestOneQueue(t *testing.T) {
	fifo := NewFifoQueuer()
	fifo.push(&sampleSongs[0])

	var nextSong *cmpb.Song = fifo.pop()
	if nextSong == nil {
		t.Error("Expected a song but got nil")
	} else if compareSongs(nextSong, &sampleSongs[0]) == false {
		t.Error("Expected", &sampleSongs[0], "but got", nextSong)
	}

	nextSong = fifo.pop()
	if nextSong != nil {
		t.Error("Expected nil, but got", nextSong)
	}
}

/*
 * Tests a queue with many items
 */
func TestManyQueue(t *testing.T) {
	fifo := NewFifoQueuer()

	for i := 0; i < len(sampleSongs); i++ {
		fifo.push(&sampleSongs[i])
	}

	for i := 0; i < len(sampleSongs); i++ {
		nextSong := fifo.pop()

		if nextSong == nil {
			t.Error("Expected a song but got nil")
		} else if compareSongs(nextSong, &sampleSongs[i]) == false {
			t.Error("Expected", &sampleSongs[i], "but got", nextSong)
		}
	}

	nextSong := fifo.pop()
	if nextSong != nil {
		t.Error("Expected nil, but got", nextSong)
	}
}

package ytbbe

import (
	"testing"
)

/*
 * List of sample song data to test against
 */
var sampleSongs = []SongData{
	{"title 1", 1, ServiceYoutube, "0xdeadbeef", "Kid A", 1},
	{"title 2", 2, ServiceYoutube, "0xba5eba11", "Kid B", 2},
	{"title 3", 3, ServiceYoutube, "0xf01dab1e", "Kid A", 1},
	{"title 4", 4, ServiceYoutube, "0xb01dface", "Kid B", 2},
	{"title 5", 5, ServiceYoutube, "0xca55e77e", "Kid A", 1},
}

/*
 * Compares two songs and returns true if they are the same
 */
func compareSongs(first *SongData, second *SongData) bool {
	var same bool = false
	same = (first.Title == second.Title)
	same = (first.ServiceID == second.ServiceID) && same
	same = (first.SongID == second.SongID) && same
	same = (first.Username == second.Username) && same
	same = (first.UserID == second.UserID) && same

	return same
}

/*
 * Tests an empty queue
 */
func TestEmptyQueue(t *testing.T) {
	var fifo FifoQueue
	fifo.Init()

	var nextSong *SongData = fifo.PopSong()
	if nextSong != nil {
		t.Error("Expected nil, but got", nextSong)
	}
}

/*
 * Tests a queue with a single item
 */
func TestOneQueue(t *testing.T) {
	var fifo FifoQueue
	fifo.Init()

	fifo.AddSong(&sampleSongs[0])

	var nextSong *SongData = fifo.PopSong()
	if nextSong == nil {
		t.Error("Expected a song but got nil")
	} else if compareSongs(nextSong, &sampleSongs[0]) == false {
		t.Error("Expected", &sampleSongs[0], "but got", nextSong)
	}

	nextSong = fifo.PopSong()
	if nextSong != nil {
		t.Error("Expected nil, but got", nextSong)
	}
}

/*
 * Tests a queue with many items
 */
func TestManyQueue(t *testing.T) {
	var fifo FifoQueue
	var nextSong *SongData
	fifo.Init()

	for i := 0; i < len(sampleSongs); i++ {
		fifo.AddSong(&sampleSongs[i])
	}

	for i := 0; i < len(sampleSongs); i++ {
		nextSong = fifo.PopSong()

		if nextSong == nil {
			t.Error("Expected a song but got nil")
		} else if compareSongs(nextSong, &sampleSongs[i]) == false {
			t.Error("Expected", &sampleSongs[i], "but got", nextSong)
		}
	}

	nextSong = fifo.PopSong()
	if nextSong != nil {
		t.Error("Expected nil, but got", nextSong)
	}
}

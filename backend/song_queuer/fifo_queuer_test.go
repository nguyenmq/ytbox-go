package song_queue

import (
	"testing"

	pb "github.com/nguyenmq/ytbox-go/proto/backend"
)

/*
 * List of sample song data to test against
 */
var sampleSongs = []pb.Song{
	{"title 1", 1, "Kid A", 1, pb.ServiceType_ServiceYoutube, "0xdeadbeef"},
	{"title 2", 2, "Kid B", 2, pb.ServiceType_ServiceYoutube, "0xba5eba11"},
	{"title 3", 3, "Kid A", 1, pb.ServiceType_ServiceYoutube, "0xf01dab1e"},
	{"title 4", 4, "Kid B", 2, pb.ServiceType_ServiceYoutube, "0xb01dface"},
	{"title 5", 5, "Kid A", 1, pb.ServiceType_ServiceYoutube, "0xca55e77e"},
}

/*
 * Compares two songs and returns true if they are the same
 */
func compareSongs(first *pb.Song, second *pb.Song) bool {
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
	var fifo FifoQueuer
	fifo.Init()

	var nextSong *pb.Song = fifo.PopQueue()
	if nextSong != nil {
		t.Error("Expected nil, but got", nextSong)
	}
}

/*
 * Tests a queue with a single item
 */
func TestOneQueue(t *testing.T) {
	var fifo FifoQueuer
	fifo.Init()

	fifo.AddSong(&sampleSongs[0])

	var nextSong *pb.Song = fifo.PopQueue()
	if nextSong == nil {
		t.Error("Expected a song but got nil")
	} else if compareSongs(nextSong, &sampleSongs[0]) == false {
		t.Error("Expected", &sampleSongs[0], "but got", nextSong)
	}

	nextSong = fifo.PopQueue()
	if nextSong != nil {
		t.Error("Expected nil, but got", nextSong)
	}
}

/*
 * Tests a queue with many items
 */
func TestManyQueue(t *testing.T) {
	var fifo FifoQueuer
	var nextSong *pb.Song
	fifo.Init()

	for i := 0; i < len(sampleSongs); i++ {
		fifo.AddSong(&sampleSongs[i])
	}

	for i := 0; i < len(sampleSongs); i++ {
		nextSong = fifo.PopQueue()

		if nextSong == nil {
			t.Error("Expected a song but got nil")
		} else if compareSongs(nextSong, &sampleSongs[i]) == false {
			t.Error("Expected", &sampleSongs[i], "but got", nextSong)
		}
	}

	nextSong = fifo.PopQueue()
	if nextSong != nil {
		t.Error("Expected nil, but got", nextSong)
	}
}

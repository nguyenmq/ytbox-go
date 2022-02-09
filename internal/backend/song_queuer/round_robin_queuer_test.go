package song_queue

import (
	"sort"
	"testing"
	"time"
)

/*
 * List of sample submissions to the round robin queuer
 */
var sampleSubmissions = []submission{
	{&sampleSongs[0], 3, time.Date(2020, time.January, 1, 0, 0, 4, 0, time.UTC)},
	{&sampleSongs[1], 2, time.Date(2020, time.January, 1, 0, 0, 3, 0, time.UTC)},
	{&sampleSongs[2], 2, time.Date(2020, time.January, 1, 0, 0, 2, 0, time.UTC)},
	{&sampleSongs[3], 1, time.Date(2020, time.January, 1, 0, 0, 1, 0, time.UTC)},
	{&sampleSongs[4], 1, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)},
}

func TestPushPop(t *testing.T) {
	queuer := NewRoundRobinQueuer()

	for i := 0; i < len(sampleSongs); i++ {
		queuer.push(&sampleSongs[i])
	}

	index := 0
	for queuer.length() > 0 {
		actualSong := queuer.pop()

		expectedSong := &sampleSongs[index]
		if compareSongs(expectedSong, actualSong) == false {
			t.Error("Expected", expectedSong, "but got", actualSong)
		}

		index++
	}

	if index != len(sampleSongs) {
		t.Error("Did not pop off all songs from queue. Index:", index)
	}
}

func TestRemove(t *testing.T) {
	queuer := NewRoundRobinQueuer()

	for i := 0; i < len(sampleSongs); i++ {
		queuer.push(&sampleSongs[i])
	}

	expectedSong := &sampleSongs[len(sampleSongs)-1]
	queuer.remove(expectedSong.SongId, expectedSong.UserId)

	expectedLength := len(sampleSongs) - 1
	actualLength := queuer.length()
	if expectedLength != actualLength {
		t.Error("Expected length", expectedLength, "but got", actualLength)
	}

	for queuer.length() > 0 {
		actualSong := queuer.pop()

		if compareSongs(expectedSong, actualSong) == true {
			t.Error("Expected song was not deleted")
		}
	}
}

func TestRoundRobinSort(t *testing.T) {
	actualList := make([]*submission, len(sampleSubmissions))

	for i := range sampleSubmissions {
		actualList[i] = &sampleSubmissions[i]
	}

	sort.Sort(byRoundRobin(actualList))

	for i := range sampleSubmissions {
		expectedSong := sampleSubmissions[len(sampleSubmissions)-i-1].song
		actualSong := actualList[i].song

		if compareSongs(expectedSong, actualSong) == false {
			t.Error("Expected", expectedSong, "but got", actualSong)
		}
	}
}

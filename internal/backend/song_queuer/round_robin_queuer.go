/*
 * A RoundRobinQueuer sorts the songs in the queue based on round-robin priority.
 */

package song_queue

import (
	"errors"
	"fmt"
	"sort"
	"time"

	cmpb "github.com/nguyenmq/ytbox-go/internal/proto/common"
)

const markedForRemoval = -1

// A user submission managed by the round robin queuer
type submission struct {
	song  *cmpb.Song // A song in the queue
	round int        // The round in which the song was submitted
	time  time.Time  // Time at which t he song was submitted
}

// Implements sort.Interface for a slice of submissions
type byRoundRobin []*submission

func (list byRoundRobin) Len() int {
	return len(list)
}

// Sort by round and then by most recently queue song in the round
func (list byRoundRobin) Less(i, j int) bool {
	if list[i].round == list[j].round {
		return list[i].time.Before(list[j].time)
	} else {
		return list[i].round < list[j].round
	}
}

func (list byRoundRobin) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

type RoundRobinQueuer struct {
	queue []*submission  // the queue of songs
	users map[uint32]int // keeps track of round robin count
	round int            // the current round
}

func NewRoundRobinQueuer() *RoundRobinQueuer {
	roundRobin := new(RoundRobinQueuer)
	roundRobin.queue = make([]*submission, 0)
	roundRobin.users = make(map[uint32]int)
	roundRobin.round = 0
	return roundRobin
}

func (roundRobin *RoundRobinQueuer) push(song *cmpb.Song) {
	var round int = 0

	// get the current round belonging to the user
	if user_round, ok := roundRobin.users[song.UserId]; ok == true {
		round = user_round + 1
	}

	// bump the user up to the current round if they're behind
	if round < roundRobin.round {
		round = roundRobin.round
	}

	roundRobin.users[song.UserId] = round

	sub := &submission{
		song:  song,
		round: round,
		time:  time.Now(),
	}

	roundRobin.queue = append(roundRobin.queue, sub)
	sort.Sort(byRoundRobin(roundRobin.queue))
}

func (roundRobin *RoundRobinQueuer) length() int {
	return len(roundRobin.queue)
}

func (roundRobin *RoundRobinQueuer) pop() *cmpb.Song {
	if roundRobin.length() > 0 {
		current_list := roundRobin.queue
		current_length := len(current_list)
		sub := current_list[0]
		current_list[0] = nil
		roundRobin.queue = current_list[1:current_length]
		sort.Sort(byRoundRobin(roundRobin.queue))

		// advance the round as songs are popped off
		roundRobin.round = sub.round

		return sub.song
	}

	return nil
}

func (roundRobin *RoundRobinQueuer) remove(songId uint32, userId uint32) error {
	for i, sub := range roundRobin.queue {
		if sub.song.SongId == songId && sub.song.UserId == userId {
			// move the song to be removed to the front and pop it off
			roundRobin.queue[i].round = markedForRemoval
			sort.Sort(byRoundRobin(roundRobin.queue))
			roundRobin.pop()
			// give the user deleting a song back one of their rounds
			roundRobin.users[userId]--
			return nil
		}
	}

	return errors.New(fmt.Sprintf("Song with id %d does not exist in the queue", songId))
}

func (roundRobin *RoundRobinQueuer) front() queueElement {
	if len(roundRobin.queue) > 0 {
		new_element := roundRobinElement{
			queue: roundRobin.queue,
			index: 0,
		}

		return new_element
	}

	return nil
}

type roundRobinElement struct {
	queue []*submission // song queue
	index int           // index of element
}

func (e roundRobinElement) value() *cmpb.Song {
	return e.queue[e.index].song
}

func (e roundRobinElement) next() queueElement {
	if e.index+1 < len(e.queue) {
		new_element := &roundRobinElement{
			queue: e.queue,
			index: e.index + 1,
		}

		return new_element
	}

	return nil
}

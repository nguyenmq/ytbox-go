/*
 * Manages all connected remote player clients. The manager acts a funnel where
 * new messages from player clients are fanned into the manager and new control
 * messages are sent out. The manager is responsible for popping songs off the
 * front of the playlist when all the remote player clients send in a ready
 * status.
 */

package backend

import (
	"log"
	"sync"

	sched "github.com/nguyenmq/ytbox-go/backend/scheduler"
	bepb "github.com/nguyenmq/ytbox-go/proto/backend"
)

const (
	// The ready status is AND'd together to get an all-ready status
	PLAYER_BUSY  = false
	PLAYER_READY = true
)

/*
 * An incoming Status message from the remote player with Id
 */
type playerMessage struct {
	Id     int
	Status bepb.PlayerStatus
}

/*
 * Manages communication between the backend server and remote player clients.
 */
type playerManager struct {
	fanIn      chan playerMessage
	fanOut     chan bepb.CommandType
	streams    map[int]bepb.YtbBePlayer_SongPlayerServer
	ready      map[int]bool
	playerLock sync.RWMutex
	streamIds  int
	queue      sched.QueueScheduler
}

/*
 * Initialize the player manager. It still needs to be started after being
 * initialized.
 */
func (mgr *playerManager) Init(queue sched.QueueScheduler) {
	mgr.fanIn = make(chan playerMessage)
	mgr.fanOut = make(chan bepb.CommandType)
	mgr.streams = make(map[int]bepb.YtbBePlayer_SongPlayerServer, 2)
	mgr.ready = make(map[int]bool, 2)
	mgr.streamIds = 0
	mgr.queue = queue
}

/*
 * Append a player stream for the player to keep track of
 */
func (mgr *playerManager) Append(out bepb.YtbBePlayer_SongPlayerServer) int {
	mgr.playerLock.Lock()
	defer mgr.playerLock.Unlock()
	mgr.streamIds++
	mgr.streams[mgr.streamIds] = out
	mgr.ready[mgr.streamIds] = PLAYER_BUSY
	log.Printf("New player %d", mgr.streamIds)
	return mgr.streamIds
}

/*
 * Fan in messages received from the remote players
 */
func (mgr *playerManager) FanIn() chan<- playerMessage {
	return mgr.fanIn
}

/*
 * Fan commands out to player clients
 */
func (mgr *playerManager) FanOut() chan<- bepb.CommandType {
	return mgr.fanOut
}

/*
 * Remove a player stream that the player manager was keeping track of
 */
func (mgr *playerManager) RemoveStream(id int) {
	mgr.playerLock.Lock()
	defer mgr.playerLock.Unlock()
	delete(mgr.streams, id)
	delete(mgr.ready, id)
	log.Printf("Removed player %d", id)
}

/*
 * Start the the player manager
 */
func (mgr *playerManager) Start() {
	go func() {
		nextSong := make(chan bepb.PlayerControl)

		for {
			select {
			case cmd, ok := <-mgr.fanOut:
				if !ok {
					close(nextSong)
					return
				}

				log.Printf("Sending out command: %v", cmd)
				mgr.playerLock.RLock()
				for id, out := range mgr.streams {
					go func() { out.Send(&bep) }()
				}
				mgr.playerLock.RUnlock()

			case msg, ok := <-mgr.fanIn:
				if !ok {
					close(nextSong)
					return
				}

				log.Printf("Player %d status: %v", msg.Id, msg.Status.GetCommand())
				if msg.Status.GetCommand() == bepb.CommandType_Ready {
					// Update the ready status of the current player
					mgr.playerLock.Lock()
					mgr.ready[msg.Id] = PLAYER_READY
					mgr.playerLock.Unlock()

					// Check if they're all ready
					if mgr.playersReady() {
						go mgr.getNextSong(nextSong)
					}
				}

			case control, ok := <-nextSong:
				// Send the song popped off the playlist to all the players and
				// then reset their ready flags
				if ok && control.GetCommand() == bepb.CommandType_Play {
					mgr.playerLock.Lock()
					for id, out := range mgr.streams {
						go func() { out.Send(&control) }()
						mgr.ready[id] = PLAYER_BUSY
					}
					mgr.playerLock.Unlock()
				}
			}
		}
	}()
}

/*
 * Get the next song from the playlist. This should run in a separate goroutine
 * because it will block and wait for more songs to be added to the playist if
 * the function is called while the playlist is empty.
 */
func (mgr *playerManager) getNextSong(nextSong chan<- bepb.PlayerControl) {
	// Wait for there to be at least one song in the playlist
	mgr.queue.WaitForMoreSongs()

	// Do a final check to see if all players are ready for the next song
	if mgr.playersReady() {
		song := mgr.queue.PopQueue()
		log.Println("Popped song")
		control := bepb.PlayerControl{}

		if song != nil {
			control.Command = bepb.CommandType_Play
			control.Song = song
		} else {
			control.Command = bepb.CommandType_None
		}

		nextSong <- control
	}
}

/*
 * Stop the player manager
 */
func (mgr *playerManager) Stop() {
	close(mgr.fanIn)
}

/*
 * Check to see if enough remote players have returned a ready status to play
 * the next song
 */
func (mgr *playerManager) playersReady() bool {
	mgr.playerLock.RLock()
	defer mgr.playerLock.RUnlock()

	if len(mgr.ready) == 0 {
		return false
	}

	allReady := true
	for _, ready := range mgr.ready {
		allReady = allReady && ready
	}

	return allReady
}

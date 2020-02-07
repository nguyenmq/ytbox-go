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

	queuer "github.com/nguyenmq/ytbox-go/backend/song_queuer"
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
	Status *bepb.PlayerStatus
}

/*
 * Keeps track of state data belonging to a player
 */
type playerState struct {
	out  bepb.YtbBePlayer_SongPlayerServer
	stop chan struct{}
}

/*
 * Manages communication between the backend server and remote player clients.
 */
type playerManager struct {
	fanIn      chan playerMessage
	fanOut     chan *bepb.PlayerControl
	streams    map[int]*playerState
	ready      map[int]bool
	playerLock sync.RWMutex
	streamIds  int
	queueMgr   *queuer.SongQueueManager
}

/*
 * Initialize the player manager. It still needs to be started after being
 * initialized.
 */
func (mgr *playerManager) init(queueMgr *queuer.SongQueueManager) {
	mgr.fanIn = make(chan playerMessage)
	mgr.fanOut = make(chan *bepb.PlayerControl)
	mgr.streams = make(map[int]*playerState, 2)
	mgr.ready = make(map[int]bool, 2)
	mgr.streamIds = 0
	mgr.queueMgr = queueMgr
}

/*
 * Add a player stream for the manager to keep track of. Returns a player id
 * and a channel to signal stop
 */
func (mgr *playerManager) add(out bepb.YtbBePlayer_SongPlayerServer) (int, chan struct{}) {
	mgr.playerLock.Lock()
	defer mgr.playerLock.Unlock()

	mgr.streamIds++
	state := new(playerState)
	state.out = out
	state.stop = make(chan struct{})

	mgr.streams[mgr.streamIds] = state
	mgr.ready[mgr.streamIds] = PLAYER_BUSY
	log.Printf("New player %d", mgr.streamIds)
	return mgr.streamIds, state.stop
}

/*
 * Receive messages from all remote players
 */
func (mgr *playerManager) receiveFromPlayers(id int, status *bepb.PlayerStatus) {
	mgr.fanIn <- playerMessage{Id: id, Status: status}
}

/*
 * Send commands to all player clients
 */
func (mgr *playerManager) sendToPlayers(control *bepb.PlayerControl) {
	mgr.fanOut <- control
}

/*
 * Remove a player stream that the player manager was keeping track of
 */
func (mgr *playerManager) remove(id int) int {
	mgr.playerLock.Lock()
	defer mgr.playerLock.Unlock()

	delete(mgr.streams, id)
	delete(mgr.ready, id)
	log.Printf("Removed player %d", id)
	return len(mgr.streams)
}

/*
 * Start the the player manager
 */
func (mgr *playerManager) start() {
	go func() {
		nextSong := make(chan bepb.PlayerControl)

		for {
			select {
			case control, ok := <-mgr.fanOut:
				if !ok {
					close(nextSong)
					return
				}

				log.Printf("Sending out command: %v", control.GetCommand())
				mgr.playerLock.RLock()
				for _, state := range mgr.streams {
					go sendToStream(control, state.out)
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
					for id, state := range mgr.streams {
						go sendToStream(&control, state.out)
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
	mgr.queueMgr.WaitForMoreSongs()

	// Do a final check to see if all players are ready for the next song
	if mgr.playersReady() {
		song := mgr.queueMgr.PopQueue()
		mgr.queueMgr.SavePlaylist(queuer.QueueSnapshot)
		log.Println("Popped song")
		control := bepb.PlayerControl{}

		if song != nil {
			control.Command = bepb.CommandType_Play
			control.Song = song
		} else {
			mgr.queueMgr.ClearNowPlaying()
			control.Command = bepb.CommandType_None
		}

		nextSong <- control
	}
}

/*
 * Stop the player manager and tell the goroutines listening on each stream to
 * stop
 */
func (mgr *playerManager) stop() {
	mgr.playerLock.Lock()
	defer mgr.playerLock.Unlock()

	for _, state := range mgr.streams {
		state.stop <- struct{}{}
	}

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

/*
 * Provides a goroutine for sending out the player control to the given stream
 */
func sendToStream(control *bepb.PlayerControl, out bepb.YtbBePlayer_SongPlayerServer) {
	out.Send(control)
}

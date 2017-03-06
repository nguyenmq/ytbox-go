/*
 * Fetches song data
 */

package backend

import (
	"errors"
	"log"
	"os/exec"
	"strings"

	sched "github.com/nguyenmq/ytbox-go/backend/scheduler"
)

/*
 * Fetch song data for the given link. Currently only YouTube links are
 * supported.
 */
func fetchSongData(link string, userId uint32) (*sched.SongData, error) {
	out, err := exec.Command("youtube-dl", "-e", "--get-id", link).Output()

	if err != nil {
		log.Printf("Failed to run youtube-dl with error: %v", err)
		return nil, errors.New("Failed to fetch song data")
	}

	parsed := strings.Split(string(out[:]), "\n")
	if len(parsed) < 2 {
		log.Printf("Unexpected response from YouTube: %v", parsed)
		return nil, errors.New("Failed to fetch song data")
	}

	var song *sched.SongData = &sched.SongData{
		Title:     parsed[0],
		SongId:    0,
		Service:   sched.ServiceYoutube,
		ServiceId: parsed[1],
		Username:  "",
		UserId:    sched.UserIdType(userId),
	}

	return song, nil
}

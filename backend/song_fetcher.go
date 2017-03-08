/*
 * Fetches song data
 */

package backend

import (
	"errors"
	"log"
	"os/exec"
	"strings"

	pb "github.com/nguyenmq/ytbox-go/proto/backend"
)

/*
 * Fetch song data for the given link. Currently only YouTube links are
 * supported.
 */
func fetchSongData(link string, userId uint32) (*pb.Song, error) {
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

	var song *pb.Song = &pb.Song{
		Title:     parsed[0],
		SongId:    1,
		Service:   pb.ServiceType_ServiceYoutube,
		ServiceId: parsed[1],
		Username:  "Kid A",
		UserId:    userId,
	}

	return song, nil
}

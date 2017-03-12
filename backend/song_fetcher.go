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
 * Fetch song data for the given link. This includes the song title, service
 * id, and service type. Currently only YouTube links are supported. Populates
 * the Song structure with the song data it retrieves. Returns an error status.
 */
func fetchSongData(link string, song *pb.Song) error {
	out, err := exec.Command("youtube-dl", "-e", "--get-id", link).Output()

	if err != nil {
		log.Printf("Failed to run youtube-dl with error: %v", err)
		return errors.New("Failed to fetch song data")
	}

	parsed := strings.Split(string(out[:]), "\n")
	if len(parsed) < 2 {
		log.Printf("Unexpected response from YouTube: %v", parsed)
		return errors.New("Failed to fetch song data")
	}

	song.Title = parsed[0]
	song.ServiceId = parsed[1]
	song.Service = pb.ServiceType_ServiceYoutube

	return nil
}

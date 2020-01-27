/*
 * Fetches song data
 */

package backend

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/dhowden/tag"
	cmpb "github.com/nguyenmq/ytbox-go/proto/common"
)

var (
	// match absolute paths to mp3 or flac files
	validFile = regexp.MustCompile(`(^\/).*\.(mp3|flac)$`)

	// match youtube links
	validYt = regexp.MustCompile(`^(https?://)(www\.)?(youtube\.com|youtu\.be)(\S+)$`)
)

func fetchSongData(link string, song *cmpb.Song) error {
	if validYt.MatchString(link) {
		return fetchYoutubeSongData(link, song)
	} else if validFile.MatchString(link) {
		return fetchLocalSongData(link, song)
	} else {
		err := errors.New(fmt.Sprintf("Unknown link submitted: %s", link))
		return err
	}
}

/*
 * Fetch song data for the given link. This includes the song title, service
 * id, and service type. Currently only YouTube links are supported. Populates
 * the Song structure with the song data it retrieves. Returns an error status.
 */
func fetchYoutubeSongData(link string, song *cmpb.Song) error {
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
	song.Service = cmpb.ServiceType_Youtube

	return nil
}

/*
 * Read the metadata out of a local mp3 or flac file
 */
func fetchLocalSongData(link string, song *cmpb.Song) error {
	file, err := os.Open(link)
	if err != nil {
		log.Printf("Failed to read file %s: %v", link, err)
		return err
	}
	defer file.Close()

	tags, err := tag.ReadFrom(file)
	if err != nil {
		log.Printf("Failed to parse tags: %v", err)
		return err
	}

	song.Title = fmt.Sprintf("%s - %s", tags.Artist(), tags.Title())
	song.ServiceId = link
	song.Service = cmpb.ServiceType_Local

	return nil
}

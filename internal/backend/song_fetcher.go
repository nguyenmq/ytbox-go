/*
 * Fetches song data
 */

package backend

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/dhowden/tag"
	cmpb "github.com/nguyenmq/ytbox-go/internal/proto/common"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var (
	// match absolute paths to mp3 or flac files
	validFile = regexp.MustCompile(`(^\/).*\.(mp3|flac)$`)

	// match youtube links
	validYt = regexp.MustCompile(`^(https?://)?(www\.)?(m\.)?(youtube\.com|youtu\.be)(\S+)$`)

	// match the full length youtube url
	fullYoutubeLink = regexp.MustCompile(`^(https?://)?(www\.)?(m\.)?youtube\.com/watch(\S+)$`)

	// match the shortened youtube url
	shortYoutubeLink = regexp.MustCompile(`^(https?://)?(www\.)?youtu\.be/(\S+)$`)

	// full length youtube url uses a query parameter
	videoQueryParam = regexp.MustCompile(`v=[A-Za-z0-9_\-]+`)

	// shortened youtube url uses a path parameter
	videoPathParam = regexp.MustCompile(`be/[A-Za-z0-9_\-]+`)
)

type SongFetcher struct {
	ytService *youtube.Service
}

func (fetcher *SongFetcher) init(apiKey string) {
	if apiKey == "" {
		fetcher.ytService = nil
	} else {
		fetcher.ytService, _ = youtube.NewService(context.Background(), option.WithAPIKey(apiKey))
	}
}

func (fetcher *SongFetcher) fetchSongData(link string, song *cmpb.Song) error {
	if validYt.MatchString(link) {
		if fetcher.ytService == nil {
			return fetcher.fetchYoutubeDlp(link, song)
		} else {
			return fetcher.fetchYoutubeSongData(link, song)
		}
	} else if validFile.MatchString(link) {
		return fetcher.fetchLocalSongData(link, song)
	} else {
		err := errors.New(fmt.Sprintf("Unknown link submitted: %s", link))
		return err
	}
}

func extractVideoId(link string) string {
	if fullYoutubeLink.MatchString(link) {
		return strings.TrimPrefix(videoQueryParam.FindString(link), "v=")
	} else if shortYoutubeLink.MatchString(link) {
		return strings.TrimPrefix(videoPathParam.FindString(link), "be/")
	} else {
		return ""
	}
}

/*
 * Fetch song data using yt-dlp
 */
func (fetcher *SongFetcher) fetchYoutubeDlp(link string, song *cmpb.Song) error {
	songId := extractVideoId(link)
	if len(songId) == 0 {
		log.Printf("Failed to extract id from link: %s\n", link)
		return errors.New("Failed to extract song id")
	}

	title, err := exec.Command("yt-dlp", "--print", "title", link).Output()

	if err != nil {
		log.Printf("Failed to run yt-dlp with error: %v", err)
		return errors.New("Failed to fetch song data")
	}

	song.Title = strings.TrimSpace(string(title[:]))
	song.ServiceId = songId
	song.Service = cmpb.ServiceType_Youtube
	song.Metadata = &cmpb.Metadata{
		Thumbnail: fmt.Sprintf("https://i.ytimg.com/vi/%s/mqdefault.jpg", songId),
		Duration:  "",
	}

	return nil
}

/*
 * Fetch song data for the given link. This includes the song title, service
 * id, and service type. Currently only YouTube links are supported. Populates
 * the Song structure with the song data it retrieves. Returns an error status.
 */
func (fetcher *SongFetcher) fetchYoutubeSongData(link string, song *cmpb.Song) error {
	songId := extractVideoId(link)
	if len(songId) == 0 {
		log.Printf("Failed to extract id from link: %s\n", link)
		return errors.New("Failed to extract song id")
	}

	args := []string{"snippet", "contentDetails"}
	request := fetcher.ytService.Videos.List(args)
	request.Id(songId)
	response, err := request.Do()

	if err != nil {
		log.Printf("Failed to fetch song data for %s with error: %s\n", songId, err.Error())
		return errors.New("Failed to fetch song metadata")
	}

	if len(response.Items) > 0 {
		item := response.Items[0]
		song.Title = item.Snippet.Title
		song.ServiceId = songId
		song.Service = cmpb.ServiceType_Youtube
		song.Metadata = &cmpb.Metadata{
			Thumbnail: fmt.Sprintf("https://i.ytimg.com/vi/%s/mqdefault.jpg", songId),
			Duration:  item.ContentDetails.Duration,
		}

		return nil
	}

	log.Printf("Did not get proper metadata from youtube: %v", response)
	return errors.New("Failed to fetch song metadata")
}

/*
 * Read the metadata out of a local mp3 or flac file
 */
func (fetcher *SongFetcher) fetchLocalSongData(link string, song *cmpb.Song) error {
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

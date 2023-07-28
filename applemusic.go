package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var wait bool = false

/* -- song data structures -- */
type AppleMusicSong struct {
    Data []AppleMusicSongData `json:"data"`
}

type AppleMusicSongData struct {
    Attributes    AppleMusicAttributes    `json:"attributes"`
    Relationships AppleMusicRelationships `json:"relationships"`
}

type AppleMusicAttributes struct {
    ArtistName     string  `json:"artistName"`
    Artwork        Artwork `json:"artwork"`
    URL            string  `json:"url"`
    Name           string  `json:"name"`
    ISRC           string  `json:"isrc"`
    TrackNumber    int     `json:"trackNumber"`
    AlbumName      string  `json:"albumName"`
    ContentRating *string  `json:"contentRating"`
}

type Artwork struct {
    URL string `json:"url"`
}

type AppleMusicRelationships struct {
    Albums AppleMusicAlbumData `json:"albums"`
}

type AppleMusicAlbumData struct {
    Data []RelationshipsData `json:"data"`
}

type RelationshipsData struct {
    ID string `json:"id"`
}

type AppleMusicSongSearch struct {
    Results AppleMusicSearchResults `json:"results"`
}

type AppleMusicSearchResults struct {
    Songs AppleMusicSong `json:"songs"`
}

/* -- song data structures -- */

/* -- album data structures -- */
type AppleMusicAlbum struct {
    Data []AppleMusicAlbumSearchData `json:"data"`
}

type AppleMusicAlbumSearchData struct {
    Attributes    AppleMusicAlbumAttributes    `json:"attributes"`
    Relationships AppleMusicAlbumRelationships `json:"relationships"`
}

type AppleMusicAlbumAttributes struct {
    ArtistName  string `json:"artistName"`
    URL         string `json:"url"`
    TrackCount  int    `json:"trackCount"`
    Name        string `json:"name"`
    RecordLabel string `json:"recordLabel"`
    UPC         string `json:"upc"`
}

type AppleMusicAlbumRelationships struct {
    Tracks AppleMusicTrackData `json:"tracks"`
}

type AppleMusicTrackData struct {
    Data []AppleMusicSongItem `json:"data"`
}

type AppleMusicSongItem struct {
    ID         string                       `json:"id"`
    Attributes AppleMusicSongItemAttributes `json:"attributes"`
}

type AppleMusicSongItemAttributes struct {
    ContentRating *string `json:"contentRating"`
}
/* -- album data structures -- */

/* -- artist data structures -- */
type AppleMusicArtist struct {
    Data []AppleMusicArtistData `json:"data"`
}

type AppleMusicArtistData struct {
    Attributes AppleMusicArtistAttributes `json:"attributes"`
}

type AppleMusicArtistAttributes struct {
    URL     string  `json:"url"`
    Name    string  `json:"name"`
    Artwork Artwork `json:"artwork"`
}

type AppleMusicArtistSearch struct {
    Results AppleMusicArtistSearchResults `json:"results"`
}

type AppleMusicArtistSearchResults struct {
    Artists AppleMusicArtistSearchArtists `json:"artists"`
}

type AppleMusicArtistSearchArtists struct {
    Data []AppleMusicArtistData `json:"data"`
}

/* -- artist data structures -- */

/* -- playlist data structures -- */
type AppleMusicPlaylist struct {
    Data []AppleMusicPlaylistData `json:"data"`
}

type AppleMusicPlaylistData struct {
    Attributes    AppleMusicPlaylistAttributes    `json:"attributes"`
    Relationships AppleMusicPlaylistRelationships `json:"relationships"`
}

type AppleMusicPlaylistAttributes struct {
    CuratorName string  `json:"curatorName"`
    Name        string  `json:"name"`
    Artwork     Artwork `json:"artwork"`
}

type AppleMusicPlaylistRelationships struct {
    Tracks AppleMusicPlaylistTracks `json:"tracks"`
}

type AppleMusicPlaylistTracks struct {
    Next *string                        `json:"next"`
    Data []AppleMusicPlaylistTracksData `json:"data"`
}

type AppleMusicPlaylistTracksData struct {
    ID string `json:"id"`
}
/* -- playlist data structures -- */

/*
    Checks to see if there is a wait time to be served b/c of rate limiting
*/
func appleMusicWaitIfLimited() {
    if wait {
        fmt.Println("Retrying after:", waitTimeSec, "seconds")
        time.Sleep(5 * time.Second)

        wait = false
    }
}

func getAppleMusicSongByID(id string, key string, appleMusicSong chan AppleMusicSong) {
    appleMusicWaitIfLimited()

    url := "https://api.music.apple.com/v1/catalog/us/songs/" + id
    authVal := "Bearer " + key

    client := &http.Client{}
    request, _ := http.NewRequest("GET", url, strings.NewReader(""))
    // set HTTP header values
    request.Header.Add("Content-Type", "application/json")
    request.Header.Add("Authorization", authVal)

    response, err := client.Do(request)

    if response.StatusCode == http.StatusTooManyRequests {
        fmt.Println("Too many requests")
        wait = true
    }

    if err != nil {
        fmt.Print(err.Error())
        os.Exit(1)
    }

    responseData, err := ioutil.ReadAll(response.Body)
    if err != nil {
        log.Fatal(err)
    }

    var responseObject AppleMusicSong

    json.Unmarshal(responseData, &responseObject)

    appleMusicSong <- responseObject
}

func getAppleMusicSongsBySearch(params string, key string, appleMusicSongSearch chan AppleMusicSongSearch) {
    appleMusicWaitIfLimited()

    url := "https://api.music.apple.com/v1/catalog/us/search?types=songs&term=" + params
    fmt.Println(url)
    authVal := "Bearer " + key

    client := &http.Client{}
    request, _ := http.NewRequest("GET", url, strings.NewReader(""))
    // set HTTP header values
    request.Header.Add("Content-Type", "application/json")
    request.Header.Add("Authorization", authVal)

    response, err := client.Do(request)

    if response.StatusCode == http.StatusTooManyRequests {
        fmt.Println("Too many requests")
        wait = true
    }

    if err != nil {
        fmt.Print(err.Error())
        os.Exit(1)
    }

    responseData, err := ioutil.ReadAll(response.Body)
    if err != nil {
        log.Fatal(err)
    }

    var responseObject AppleMusicSongSearch

    json.Unmarshal(responseData, &responseObject)

    appleMusicSongSearch <- responseObject
}

func getAppleMusicAlbumByID(id string, key string, appleMusicAlbum chan AppleMusicAlbum) {
    appleMusicWaitIfLimited()

    url := "https://api.music.apple.com/v1/catalog/us/albums/" + id
    authVal := "Bearer " + key

    client := &http.Client{}
    request, _ := http.NewRequest("GET", url, strings.NewReader(""))
    // set HTTP header values
    request.Header.Add("Content-Type", "application/json")
    request.Header.Add("Authorization", authVal)

    response, err := client.Do(request)

    if response.StatusCode == http.StatusTooManyRequests {
        fmt.Println("Too many requests")
        wait = true
    }

    if err != nil {
        fmt.Print(err.Error())
        os.Exit(1)
    }

    responseData, err := ioutil.ReadAll(response.Body)
    if err != nil {
        log.Fatal(err)
    }

    var responseObject AppleMusicAlbum

    json.Unmarshal(responseData, &responseObject)

    appleMusicAlbum <- responseObject
}

func getAppleMusicArtistByID(id string, key string, appleMusicArtist chan AppleMusicArtist) {
    appleMusicWaitIfLimited()

    url := "https://api.music.apple.com/v1/catalog/us/artists/" + id
    authVal := "Bearer " + key

    client := &http.Client{}
    request, _ := http.NewRequest("GET", url, strings.NewReader(""))
    // set HTTP header values
    request.Header.Add("Content-Type", "application/json")
    request.Header.Add("Authorization", authVal)

    response, err := client.Do(request)

    if response.StatusCode == http.StatusTooManyRequests {
        fmt.Println("Too many requests")
        wait = true
    }


    if err != nil {
        fmt.Print(err.Error())
        os.Exit(1)
    }

    responseData, err := ioutil.ReadAll(response.Body)
    if err != nil {
        log.Fatal(err)
    }

    var responseObject AppleMusicArtist

    json.Unmarshal(responseData, &responseObject)

    appleMusicArtist <- responseObject
}

func getAppleMusicArtistsBySearch(params string, key string, appleMusicArtistSearch chan AppleMusicArtistSearch) {
    appleMusicWaitIfLimited()

    url := "https://api.music.apple.com/v1/catalog/us/search?types=artists&term=" + params
    fmt.Println(url)
    authVal := "Bearer " + key

    client := &http.Client{}
    request, _ := http.NewRequest("GET", url, strings.NewReader(""))
    // set HTTP header values
    request.Header.Add("Content-Type", "application/json")
    request.Header.Add("Authorization", authVal)

    response, err := client.Do(request)

    if response.StatusCode == http.StatusTooManyRequests {
        fmt.Println("Too many requests")
        wait = true
    }

    if err != nil {
        fmt.Print(err.Error())
        os.Exit(1)
    }

    responseData, err := ioutil.ReadAll(response.Body)
    if err != nil {
        log.Fatal(err)
    }

    var responseObject AppleMusicArtistSearch

    json.Unmarshal(responseData, &responseObject)

    appleMusicArtistSearch <- responseObject
}

func getAppleMusicPlaylistByID(id string, key string, appleMusicPlaylist chan AppleMusicPlaylist) {
    appleMusicWaitIfLimited()

    url := "https://api.music.apple.com/v1/catalog/us/playlists/" + id
    authVal := "Bearer " + key

    client := &http.Client{}
    request, _ := http.NewRequest("GET", url, strings.NewReader(""))
    // set HTTP header values
    request.Header.Add("Content-Type", "application/json")
    request.Header.Add("Authorization", authVal)

    response, err := client.Do(request)

    if response.StatusCode == http.StatusTooManyRequests {
        fmt.Println("Too many requests")
        wait = true
    }

    if err != nil {
        fmt.Print(err.Error())
        os.Exit(1)
    }

    responseData, err := ioutil.ReadAll(response.Body)
    if err != nil {
        log.Fatal(err)
    }

    var responseObject AppleMusicPlaylist

    json.Unmarshal(responseData, &responseObject)

    appleMusicPlaylist <- responseObject
}

func getNextAppleMusicPlaylist(nextURL string, key string,
    nextAppleMusicPlaylistTracks chan AppleMusicPlaylistTracks) {
    appleMusicWaitIfLimited()

    url := "https://api.music.apple.com" + nextURL

    authVal := "Bearer " + key

    client := &http.Client{}
    request, _ := http.NewRequest("GET", url, strings.NewReader(""))
    // set HTTP header values
    request.Header.Add("Content-Type", "application/json")
    request.Header.Add("Authorization", authVal)

    response, err := client.Do(request)

    if response.StatusCode == http.StatusTooManyRequests {
        fmt.Println("Too many requests")
        wait = true
    }

    if err != nil {
        fmt.Print(err.Error())
        os.Exit(1)
    }

    responseData, err := ioutil.ReadAll(response.Body)
    if err != nil {
        log.Fatal(err)
    }

    var responseObject AppleMusicPlaylistTracks

    json.Unmarshal(responseData, &responseObject)

    nextAppleMusicPlaylistTracks <- responseObject
}


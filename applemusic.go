package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "strings"
    "sync"
    "time"
)

// used to avoid going over rate limit
type AppleWaitContainer struct {
    mu   sync.Mutex
    wait bool
}

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
func appleMusicWaitIfLimited(w *AppleWaitContainer) {
    w.mu.Lock()

    if w.wait {
        appleWaitTime := 10
        fmt.Println("Apple: Retrying after:", appleWaitTime, "seconds")
        time.Sleep(time.Duration(appleWaitTime) * time.Second)

        w.wait = false
    }

    w.mu.Unlock()
}

func getAppleMusicSongByID(
    w *AppleWaitContainer,
    id string,
    key string,
    appleMusicSong chan AppleMusicSong,
) {
    appleMusicWaitIfLimited(w)

    url := "https://api.music.apple.com/v1/catalog/us/songs/" + id
    authVal := "Bearer " + key

    client := &http.Client{}
    request, _ := http.NewRequest("GET", url, strings.NewReader(""))
    // set HTTP header values
    request.Header.Add("Content-Type", "application/json")
    request.Header.Add("Authorization", authVal)

    w.mu.Lock()
    response, err := client.Do(request)
    w.mu.Unlock()

    // try request again after a delay if there is a 429 error
    attempt := 0
    for attempt < 2 {
        if response.StatusCode == http.StatusTooManyRequests {
            fmt.Println("Too many requests")

            w.mu.Lock()
            w.wait = true

            appleMusicWaitIfLimited(w)
            w.mu.Unlock()

            attempt++
            response, err = client.Do(request)
        }

        attempt = 2
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

func getAppleMusicSongsBySearch(
    w *AppleWaitContainer,
    params string,
    key string,
    appleMusicSongSearch chan AppleMusicSongSearch,
) {
    appleMusicWaitIfLimited(w)

    url := "https://api.music.apple.com/v1/catalog/us/search?types=songs&term=" + params
    fmt.Println(url)
    authVal := "Bearer " + key

    client := &http.Client{}
    request, _ := http.NewRequest("GET", url, strings.NewReader(""))
    // set HTTP header values
    request.Header.Add("Content-Type", "application/json")
    request.Header.Add("Authorization", authVal)

    w.mu.Lock()
    response, err := client.Do(request)
    w.mu.Unlock()

    // try request again after a delay if there is a 429 error
    attempt := 0
    for attempt < 2 {
        if response.StatusCode == http.StatusTooManyRequests {
            fmt.Println("Too many requests")

            w.mu.Lock()
            w.wait = true

            appleMusicWaitIfLimited(w)
            w.mu.Unlock()

            attempt++
            response, err = client.Do(request)
        }

        attempt = 2
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

func getAppleMusicAlbumByID(
    w *AppleWaitContainer,
    id string,
    key string,
    appleMusicAlbum chan AppleMusicAlbum,
) {
    appleMusicWaitIfLimited(w)

    url := "https://api.music.apple.com/v1/catalog/us/albums/" + id
    authVal := "Bearer " + key

    client := &http.Client{}
    request, _ := http.NewRequest("GET", url, strings.NewReader(""))
    // set HTTP header values
    request.Header.Add("Content-Type", "application/json")
    request.Header.Add("Authorization", authVal)

    w.mu.Lock()
    response, err := client.Do(request)
    w.mu.Unlock()

    // try request again after a delay if there is a 429 error
    attempt := 0
    for attempt < 2 {
        if response.StatusCode == http.StatusTooManyRequests {
            fmt.Println("Too many requests")

            w.mu.Lock()
            w.wait = true

            appleMusicWaitIfLimited(w)
            w.mu.Unlock()

            attempt++
            response, err = client.Do(request)
        }

        attempt = 2
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

func getAppleMusicArtistByID(
    w *AppleWaitContainer,
    id string,
    key string,
    appleMusicArtist chan AppleMusicArtist,
) {
    appleMusicWaitIfLimited(w)

    url := "https://api.music.apple.com/v1/catalog/us/artists/" + id
    authVal := "Bearer " + key

    client := &http.Client{}
    request, _ := http.NewRequest("GET", url, strings.NewReader(""))
    // set HTTP header values
    request.Header.Add("Content-Type", "application/json")
    request.Header.Add("Authorization", authVal)

    w.mu.Lock()
    response, err := client.Do(request)
    w.mu.Unlock()

    // try request again after a delay if there is a 429 error
    attempt := 0
    for attempt < 2 {
        if response.StatusCode == http.StatusTooManyRequests {
            fmt.Println("Too many requests")

            w.mu.Lock()
            w.wait = true

            appleMusicWaitIfLimited(w)
            w.mu.Unlock()

            attempt++
            response, err = client.Do(request)
        }

        attempt = 2
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

func getAppleMusicArtistsBySearch(
    w *AppleWaitContainer,
    params string,
    key string,
    appleMusicArtistSearch chan AppleMusicArtistSearch,
) {
    appleMusicWaitIfLimited(w)

    url := "https://api.music.apple.com/v1/catalog/us/search?types=artists&term=" + params
    fmt.Println(url)
    authVal := "Bearer " + key

    client := &http.Client{}
    request, _ := http.NewRequest("GET", url, strings.NewReader(""))
    // set HTTP header values
    request.Header.Add("Content-Type", "application/json")
    request.Header.Add("Authorization", authVal)

    w.mu.Lock()
    response, err := client.Do(request)
    w.mu.Unlock()

    // try request again after a delay if there is a 429 error
    attempt := 0
    for attempt < 2 {
        if response.StatusCode == http.StatusTooManyRequests {
            fmt.Println("Too many requests")

            w.mu.Lock()
            w.wait = true

            appleMusicWaitIfLimited(w)
            w.mu.Unlock()

            attempt++
            response, err = client.Do(request)
        }

        attempt = 2
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

func getAppleMusicPlaylistByID(
    w *AppleWaitContainer,
    id string,
    key string,
    appleMusicPlaylist chan AppleMusicPlaylist,
) {
    appleMusicWaitIfLimited(w)

    url := "https://api.music.apple.com/v1/catalog/us/playlists/" + id
    authVal := "Bearer " + key

    client := &http.Client{}
    request, _ := http.NewRequest("GET", url, strings.NewReader(""))
    // set HTTP header values
    request.Header.Add("Content-Type", "application/json")
    request.Header.Add("Authorization", authVal)

    w.mu.Lock()
    response, err := client.Do(request)
    w.mu.Unlock()

    // try request again after a delay if there is a 429 error
    attempt := 0
    for attempt < 2 {
        if response.StatusCode == http.StatusTooManyRequests {
            fmt.Println("Too many requests")

            w.mu.Lock()
            w.wait = true

            appleMusicWaitIfLimited(w)
            w.mu.Unlock()

            attempt++
            response, err = client.Do(request)
        }

        attempt = 2
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

func getNextAppleMusicPlaylist(
    w *AppleWaitContainer,
    nextURL string,
    key string,
    nextAppleMusicPlaylistTracks chan AppleMusicPlaylistTracks,
) {
    appleMusicWaitIfLimited(w)

    url := "https://api.music.apple.com" + nextURL

    authVal := "Bearer " + key

    client := &http.Client{}
    request, _ := http.NewRequest("GET", url, strings.NewReader(""))
    // set HTTP header values
    request.Header.Add("Content-Type", "application/json")
    request.Header.Add("Authorization", authVal)

    w.mu.Lock()
    response, err := client.Do(request)
    w.mu.Unlock()

    // try request again after a delay if there is a 429 error
    attempt := 0
    for attempt < 2 {
        if response.StatusCode == http.StatusTooManyRequests {
            fmt.Println("Too many requests")

            w.mu.Lock()
            w.wait = true

            appleMusicWaitIfLimited(w)
            w.mu.Unlock()

            attempt++
            response, err = client.Do(request)
        }

        attempt = 2
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


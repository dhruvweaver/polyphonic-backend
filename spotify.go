package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

/* -- song data structures -- */
type SpotifySong struct {
    Album       Album       `json:"album"`
    Artists   []Artist      `json:"artists"`
    Explicit    bool        `json:"explicit"`
    ExternalIDs ExternalIDs `json:"external_ids"`
    Name        string      `json:"name"`
    TrackNumber int         `json:"track_number"`
    URI         string      `json:"uri"`
}

type Album struct {
    ID       string     `json:"id"`
    Name     string     `json:"name"`
    Images []SongImages `json:"images"`
}

type SongImages struct {
    URL string `json:"url"`
}

type Artist struct {
    Name string `json:"name"`
}

type ExternalIDs struct {
    ISRC string `json:"isrc"`
}

type SpotifySongSearch struct {
    Tracks SpotifySongSearchTracks `json:"tracks"`
}

type SpotifySongSearchTracks struct {
    Items []SpotifySong `json:"items"`
}

/* -- song data structures -- */

/* -- album data structures -- */
type SpotifyAlbum struct {
    Artists   []Artist           `json:"artists"`
    ExternalIDs ExternalAlbumIDs `json:"external_ids"`
    Name        string           `json:"name"`
    Label       string           `json:"label"`
    ID          string           `json:"id"`
    Tracks      MusicItems       `json:"tracks"`
    TotalTracks int              `json:"total_tracks"`
}

type ExternalAlbumIDs struct {
    UPC string `json:"upc"`
}

type MusicItems struct {
    Items []Item `json:"items"`
}

type Item struct {
    Explicit bool   `json:"explicit"`
    ID       string `json:"id"`
}
/* -- album data structures -- */

/* -- artist data structures -- */
type SpotifyArtist struct {
    Name     string       `json:"name"`
    Images []ProfileImage `json:"images"`
    URI      string       `json:"uri"`
}

type ProfileImage struct {
    URL string `json:"url"`
}
/* -- artist data structures -- */

/* -- playlist data structures -- */
type SpotifyPlaylist struct {
    Name         string        `json:"name"`
    Images     []PlaylistImage `json:"images"`
    Owner        PlaylistOwner `json:"owner"`
    Tracks       Tracks        `json:"tracks"`
    ExternalURLs ExternalURLs  `json:"external_urls"`
    ID           string        `json:"id"`
}

type PlaylistImage struct {
    URL string `json:"url"`
}

type PlaylistOwner struct {
    DisplayName string `json:"display_name"`
}

type Tracks struct {
    Items []PlaylistItem `json:"items"`
    Next   *string       `json:"next"`
}

type PlaylistItem struct {
    Track SpotifySong `json:"track"`
}

type ExternalURLs struct {
    Spotify string `json:"spotify"`
}
/* -- playlist data structures -- */

// gets Spotify auth key from local environment variables
// and returns the key and expiration time (from now)
func getSpotifyAuthKey(key chan string, exp chan int64) {
    type Response struct {
        AccessToken string `json:"access_token"`
        ExpiresIn   int64  `json:"expires_in"`
    }

    // generate authorization value from env variables
    clientID := os.Getenv("SPOTIFY_CLIENT_ID")
    clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
    authCode := clientID + ":" + clientSecret
    authCode = base64.StdEncoding.EncodeToString([]byte(authCode))
    authVal := "Basic " + authCode

    // set HTTP body values
    params := url.Values{}
    params.Add("grant_type","client_credentials")

    url := "https://accounts.spotify.com/api/token"

    client := &http.Client{}
    request, _ := http.NewRequest("POST", url, strings.NewReader(params.Encode()))
    // set HTTP header values
    request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    request.Header.Add("Authorization", authVal)

    response, err := client.Do(request)

    if response.StatusCode == http.StatusTooManyRequests {
        fmt.Println("Too many requests")
        fmt.Println("Retry after:", response.Header.Values("retry-after"), "seconds")
    }

    if err != nil {
        fmt.Print(err.Error())
        os.Exit(1)
    }

    responseData, err := ioutil.ReadAll(response.Body)
    if err != nil {
        log.Fatal(err)
    }

    var responseObject Response

    json.Unmarshal(responseData, &responseObject)

    key <- responseObject.AccessToken
    exp <- responseObject.ExpiresIn
}

func getSpotifySongByID(id string, key string, spotifySong chan SpotifySong) {
    url := "https://api.spotify.com/v1/tracks/" + id
    authVal := "Bearer " + key

    client := &http.Client{}
    request, _ := http.NewRequest("GET", url, strings.NewReader(""))
    // set HTTP header values
    request.Header.Add("Content-Type", "application/json")
    request.Header.Add("Authorization", authVal)

    response, err := client.Do(request)

    if response.StatusCode == http.StatusTooManyRequests {
        fmt.Println("Too many requests")
        fmt.Println("Retry after:", response.Header.Values("retry-after"), "seconds")
    }

    if err != nil {
        fmt.Print(err.Error())
        os.Exit(1)
    }

    responseData, err := ioutil.ReadAll(response.Body)
    if err != nil {
        log.Fatal(err)
    }

    var responseObject SpotifySong

    json.Unmarshal(responseData, &responseObject)

    spotifySong <- responseObject
}

func getSpotifySongsBySearch(params string, key string, spotifySongSearch chan SpotifySongSearch) {
    url := "https://api.spotify.com/v1/search?q=" + params
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
        fmt.Println("Retry after:", response.Header.Values("retry-after"), "seconds")
    }

    if err != nil {
        fmt.Print(err.Error())
        os.Exit(1)
    }

    responseData, err := ioutil.ReadAll(response.Body)
    if err != nil {
        log.Fatal(err)
    }

    var responseObject SpotifySongSearch

    json.Unmarshal(responseData, &responseObject)

    fmt.Println(responseObject)
    spotifySongSearch <- responseObject
}

func getSpotifyAlbumByID(id string, key string, spotifyAlbum chan SpotifyAlbum) {
    url := "https://api.spotify.com/v1/albums/" + id
    authVal := "Bearer " + key

    client := &http.Client{}
    request, _ := http.NewRequest("GET", url, strings.NewReader(""))
    // set HTTP header values
    request.Header.Add("Content-Type", "application/json")
    request.Header.Add("Authorization", authVal)

    response, err := client.Do(request)

    if response.StatusCode == http.StatusTooManyRequests {
        fmt.Println("Too many requests")
        fmt.Println("Retry after:", response.Header.Values("retry-after"), "seconds")
    }

    if err != nil {
        fmt.Print(err.Error())
        os.Exit(1)
    }

    responseData, err := ioutil.ReadAll(response.Body)
    if err != nil {
        log.Fatal(err)
    }

    var responseObject SpotifyAlbum

    json.Unmarshal(responseData, &responseObject)

    spotifyAlbum <- responseObject
}

func getSpotifyArtistByID(id string, key string, spotifyArtist chan SpotifyArtist) {
    url := "https://api.spotify.com/v1/artists/" + id
    authVal := "Bearer " + key

    client := &http.Client{}
    request, _ := http.NewRequest("GET", url, strings.NewReader(""))
    // set HTTP header values
    request.Header.Add("Content-Type", "application/json")
    request.Header.Add("Authorization", authVal)

    response, err := client.Do(request)

    if response.StatusCode == http.StatusTooManyRequests {
        fmt.Println("Too many requests")
        fmt.Println("Retry after:", response.Header.Values("retry-after"), "seconds")
    }


    if err != nil {
        fmt.Print(err.Error())
        os.Exit(1)
    }

    responseData, err := ioutil.ReadAll(response.Body)
    if err != nil {
        log.Fatal(err)
    }

    var responseObject SpotifyArtist

    json.Unmarshal(responseData, &responseObject)

    spotifyArtist <- responseObject
}

func getSpotifyArtistsBySearch(params string, key string, spotifyArtists chan []SpotifyArtist) {
    type Artists struct {
        Items []SpotifyArtist `json:"items"`
    }

    type SpotifyArtistSearch struct {
        Artists Artists `json:"artists"`
    }

    url := "https://api.spotify.com/v1/search?q=" + params
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
        fmt.Println("Retry after:", response.Header.Values("retry-after"), "seconds")
    }

    if err != nil {
        fmt.Print(err.Error())
        os.Exit(1)
    }

    responseData, err := ioutil.ReadAll(response.Body)
    if err != nil {
        log.Fatal(err)
    }

    var responseObject SpotifyArtistSearch

    json.Unmarshal(responseData, &responseObject)

    spotifyArtists <- responseObject.Artists.Items
}

func getSpotifyPlaylistByID(id string, key string, spotifyPlaylist chan SpotifyPlaylist) {
    url := "https://api.spotify.com/v1/playlists/" + id
    authVal := "Bearer " + key

    client := &http.Client{}
    request, _ := http.NewRequest("GET", url, strings.NewReader(""))
    // set HTTP header values
    request.Header.Add("Content-Type", "application/json")
    request.Header.Add("Authorization", authVal)

    response, err := client.Do(request)

    if response.StatusCode == http.StatusTooManyRequests {
        fmt.Println("Too many requests")
        fmt.Println("Retry after:", response.Header.Values("retry-after"), "seconds")
    }

    if err != nil {
        fmt.Print(err.Error())
        os.Exit(1)
    }

    responseData, err := ioutil.ReadAll(response.Body)
    if err != nil {
        log.Fatal(err)
    }

    var responseObject SpotifyPlaylist

    json.Unmarshal(responseData, &responseObject)

    spotifyPlaylist <- responseObject
}

func getNextSpotifyPlaylist(nextURL string, key string, nextSpotifyPlaylistTracks chan Tracks) {
    authVal := "Bearer " + key

    client := &http.Client{}
    request, _ := http.NewRequest("GET", nextURL, strings.NewReader(""))
    // set HTTP header values
    request.Header.Add("Content-Type", "application/json")
    request.Header.Add("Authorization", authVal)

    response, err := client.Do(request)

    if response.StatusCode == http.StatusTooManyRequests {
        fmt.Println("Too many requests")
        fmt.Println("Retry after:", response.Header.Values("retry-after"), "seconds")
    }

    if err != nil {
        fmt.Print(err.Error())
        os.Exit(1)
    }

    responseData, err := ioutil.ReadAll(response.Body)
    if err != nil {
        log.Fatal(err)
    }

    var responseObject Tracks

    json.Unmarshal(responseData, &responseObject)

    nextSpotifyPlaylistTracks <- responseObject
}


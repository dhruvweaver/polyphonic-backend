package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

var authSpotifyExp = time.Now().Unix() - 10 // initialize spotify auth time to be something that must be replaced
var authSpotifyKey string

var appleMusicKey = os.Getenv("APPLE_MUSIC_KEY")

type playlist struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Creator     string `json:"creator"`
    SongCount   int    `json:"song_count"`
    Platform    string `json:"platform"`
    OriginalURL string `json:"original_url"`
    Converted   bool   `json:"converted"`
}

type playlist_content struct {
    ID          string `json:"id"`
    KeyID       string `json:"key_id"`
    Title       string `json:"title"`
    PTrackNum   int    `json:"playlist_track_num"`
    ISRC        string `json:"isrc"`
    Artist      string `json:"artist"`
    Album       string `json:"album"`
    AlbumID     string `json:"album_id"`
    Explicit    bool   `json:"explicit"`
    OriginalURL string `json:"original_url"`
    ConvertURL  string `json:"converted_url"`
    Confidence  int    `json:"confidence"`
    TrackNum    int    `json:"track_num"`
}

type playlist_data struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Creator     string `json:"creator"`
    SongCount   int    `json:"song_count"`
    Platform    string `json:"platform"`
    OriginalURL string `json:"original_url"`
    Converted   bool   `json:"converted"`
    Content     []playlist_content `json:"content"`
}

// getPlaylistByID locates the playlist whose ID value matches the id
// parameter sent by the client, then returns that playlist as a response.
func getPlaylistByID(c *gin.Context) {
    id := c.Param("id")

    // Playlist related structs to hold data from the returned data.
    var playlistData playlist_data
    var playlist playlist
    var contents []playlist_content

    playlistsRow := db.QueryRow("SELECT * FROM playlists WHERE id = ?", id)
    if err := playlistsRow.Scan(
        &playlist.ID,
        &playlist.Name,
        &playlist.Creator,
        &playlist.SongCount,
        &playlist.Platform,
        &playlist.OriginalURL,
        &playlist.Converted); err != nil {

        log.Println(fmt.Errorf("playlistsById %v", err))
        if err == sql.ErrNoRows {
            c.IndentedJSON(http.StatusNotFound, gin.H{"message": "There is no such playlist"})
            return
        }
        c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Error getting playlist by ID"})
    }

    rows, err := db.Query("SELECT * FROM playlist_content WHERE id = ? ORDER BY playlist_track_num ASC", id)
    if err != nil {
        c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Error getting playlist content"})
    }
    defer rows.Close()
    // Loop through rows, using Scan to assign column data to struct fields.
    for rows.Next() {
        var content playlist_content
        if err := rows.Scan(
            &content.ID,
            &content.KeyID,
            &content.Title,
            &content.PTrackNum,
            &content.ISRC,
            &content.Artist,
            &content.Album,
            &content.AlbumID,
            &content.Explicit,
            &content.OriginalURL,
            &content.ConvertURL,
            &content.Confidence,
            &content.TrackNum); err != nil {

            log.Println(fmt.Errorf("playlistsById %v", err))
            c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Error getting playlist by ID"})
        }

        contents = append(contents, content)
    }
    if err := rows.Err(); err != nil {
        log.Println(fmt.Errorf("playlistsById %v", err))
        c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Error getting playlist by ID"})
    }

    playlistData.ID = playlist.ID
    playlistData.Name = playlist.Name
    playlistData.Creator = playlist.Creator
    playlistData.SongCount = playlist.SongCount
    playlistData.Platform = playlist.Platform
    playlistData.OriginalURL = playlist.OriginalURL
    playlistData.Converted = playlist.Converted
    playlistData.Content = contents

    c.IndentedJSON(http.StatusOK, playlistData)

}

// postPlaylists adds a playlist from JSON received in the request body.
func postPlaylists(c *gin.Context) {
    var newPlaylistData playlist_data

    // Call BindJSON to bind the received JSON to
    // newPlaylistData.
    if err := c.BindJSON(&newPlaylistData); err != nil {
        return
    }

    // TODO: check to see if ID already exists

    // Add the new album to the database.
    _, err := db.Exec("INSERT INTO playlists (id, name, creator, song_count, platform, original_url, converted) VALUES (?, ?, ?, ?, ?, ?, ?)",
    newPlaylistData.ID,
    newPlaylistData.Name,
    newPlaylistData.Creator,
    newPlaylistData.SongCount,
    newPlaylistData.Platform,
    newPlaylistData.OriginalURL,
    newPlaylistData.Converted)

    if err != nil {
        c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Error adding a new playlist (1)"})
    }

    for _, content := range newPlaylistData.Content {
        _, err := db.Exec("INSERT INTO playlist_content (id, key_id, title, playlist_track_num, isrc, artist, album, album_id, explicit, original_url, converted_url, confidence, track_num) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
        content.ID,
        content.KeyID,
        content.Title,
        content.PTrackNum,
        content.ISRC,
        content.Artist,
        content.Album,
        content.AlbumID,
        content.Explicit,
        content.OriginalURL,
        content.ConvertURL,
        content.Confidence,
        content.TrackNum)

        if err != nil {
            c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Error adding a new playlist (2)"})
        }
    }

    c.IndentedJSON(http.StatusCreated, newPlaylistData)
}

/*
    checkSpotifyAuth checks to see if the Spotify authorization code is within
    30 seconds of expiring. If it is going to expire, it will get a new code.
    When a new code is received, the expiration date is updated.
*/
func checkSpotifyAuth() {
    var expIn int64

    // check that there is more than 30 seconds left before key expiration
    if time.Now().Unix() > authSpotifyExp - 30 {
        fmt.Println("Getting another API key from Spotify")
        key := make(chan string)
        exp := make(chan int64)
        go getSpotifyAuthKey(key, exp)

        authSpotifyKey = <- key
        expIn = <- exp
        authSpotifyExp = time.Now().Unix() + expIn
    }
    // fmt.Println(authSpotifyKey)
}

/*
    polyphonicGetSpotifySongByID gets a Spotify song's data from the Spotify
    API using the song's ID and then responds with only the required data
    for translation.
*/
func polyphonicGetSpotifySongByID(c *gin.Context) {
    id := c.Param("id")

    checkSpotifyAuth()

    /* Get song by ID */
    spotifySongChan := make(chan SpotifySong)
    go getSpotifySongByID(id, authSpotifyKey, spotifySongChan)

    spotifySong := <- spotifySongChan
    fmt.Println(spotifySong)
    fmt.Println("Song title:", spotifySong.Name, "by", spotifySong.Artists[0].Name)
    /* Get song by ID */

    c.IndentedJSON(http.StatusOK, spotifySong)
}

/*
    polyphonicGetSpotifySongByID gets a Spotify song's data from the Spotify
    API using search terms and then responds with only the required data
    for translation.

    Format of terms: "track:[track title] artist:[artist name]"
*/
func polyphonicGetSpotifySongsBySearch(c *gin.Context) {
    terms := c.Param("terms")

    checkSpotifyAuth()

    /* Get song by search */
    spotifySongSearchChan := make(chan SpotifySongSearch)

    params := url.QueryEscape(terms) + "&type=track"
    go getSpotifySongsBySearch(params, authSpotifyKey, spotifySongSearchChan)

    spotifySongSearch := <- spotifySongSearchChan
    fmt.Println("First song title:", spotifySongSearch.Tracks.Items[0].Name, "by", spotifySongSearch.Tracks.Items[0].Artists[0])
    /* Get song by search */

    c.IndentedJSON(http.StatusOK, spotifySongSearch)
}

/*
    polyphonicGetSpotifyAlbumByID gets a Spotify album's data from the Spotify
    API using the album's ID and then responds with only the required data
    for translation.
*/
func polyphonicGetSpotifyAlbumByID(c *gin.Context) {
    id := c.Param("id")

    checkSpotifyAuth()

    /* Get album by ID */
    spotifyAlbumChan := make(chan SpotifyAlbum)
    go getSpotifyAlbumByID(id, authSpotifyKey, spotifyAlbumChan)

    spotifyAlbum := <- spotifyAlbumChan
    fmt.Println("Album title:", spotifyAlbum.Name, "by", spotifyAlbum.Artists[0].Name)
    /* Get album by ID */

    c.IndentedJSON(http.StatusOK, spotifyAlbum)
}

/*
    polyphonicGetSpotifyArtistByID gets a Spotify artist's data from the Spotify
    API using the artist's ID and then responds with only the required data
    for translation.
*/
func polyphonicGetSpotifyArtistByID(c *gin.Context) {
    id := c.Param("id")

    checkSpotifyAuth()

    /* Get artist by ID */
    spotifyArtistChan := make(chan SpotifyArtist)
    go getSpotifyArtistByID(id, authSpotifyKey, spotifyArtistChan)

    spotifyArtist := <- spotifyArtistChan
    fmt.Println("Artist name:", spotifyArtist.Name)
    /* Get artist by ID */

    c.IndentedJSON(http.StatusOK, spotifyArtist)
}

/*
    polyphonicGetSpotifyArtistBySearch gets a Spotify artist's data from the Spotify
    API using search terms and then responds with only the required data
    for translation.

    Format of terms: "[artist name]"
*/
func polyphonicGetSpotifyArtistBySearch(c *gin.Context) {
    terms := c.Param("terms")

    checkSpotifyAuth()

    /* Get artist by search */
    spotifyArtistSearchChan := make(chan SpotifyArtistSearch)

    params := url.QueryEscape(terms) + "&type=artist"
    go getSpotifyArtistsBySearch(params, authSpotifyKey, spotifyArtistSearchChan)

    spotifyArtistSearch := <- spotifyArtistSearchChan
    fmt.Println("First artist name:", spotifyArtistSearch.Artists.Items[0].Name)
    /* Get artist by search */

    c.IndentedJSON(http.StatusOK, spotifyArtistSearch)
}

/*
    polyphonicGetSpotifyPlaylistByID gets a Spotify playlist's data from the
    Spotify API using the playlist's ID and then responds with only the required
    data for translation.
*/
func polyphonicGetSpotifyPlaylistByID(c *gin.Context) {
    id := c.Param("id")

    checkSpotifyAuth()

    /* Get Playlist by ID */
    spotifyPlaylistChan := make(chan SpotifyPlaylist)
    go getSpotifyPlaylistByID(id, authSpotifyKey, spotifyPlaylistChan)

    spotifyPlayist := <- spotifyPlaylistChan

    var next = "Had more to get"
    if spotifyPlayist.Tracks.Next == nil {
        next = "No next"
    } else {
        var getMore bool = true
        var nextURL string = *spotifyPlayist.Tracks.Next

        for getMore {
            nextSpotifyPlaylistTracksChan := make(chan Tracks)
            go getNextSpotifyPlaylist(nextURL, authSpotifyKey, nextSpotifyPlaylistTracksChan)

            nextSpotifyPlaylistTracks := <- nextSpotifyPlaylistTracksChan
            fmt.Println("tracks from next section:", len(nextSpotifyPlaylistTracks.Items))

            spotifyPlayist.Tracks.Items = append(spotifyPlayist.Tracks.Items, nextSpotifyPlaylistTracks.Items...)

            if nextSpotifyPlaylistTracks.Next == nil {
                fmt.Println("done")
                getMore = false
            } else {
                nextURL = *nextSpotifyPlaylistTracks.Next
            }
        }
    }

    fmt.Println("Playlist name:", spotifyPlayist.Name, "track count:", len(spotifyPlayist.Tracks.Items),
    "next?", next)
    /* Get Playlist by ID */

    c.IndentedJSON(http.StatusOK, spotifyPlayist)
}

/*
    polyphonicGetAppleSongByID gets an Apple Music song's data from the Apple
    Music API using the song's ID and then responds with only the required data
    for translation.
*/
func polyphonicGetAppleSongByID(c *gin.Context) {
    id := c.Param("id")

    /* Get song by ID */
    appleMusicSongChan := make(chan AppleMusicSong)
    go getAppleMusicSongByID(id, appleMusicKey, appleMusicSongChan)

    appleMusicSong := <- appleMusicSongChan
    fmt.Println("Song title:", appleMusicSong.Data[0].Attributes.Name, "by", appleMusicSong.Data[0].Attributes.ArtistName)
    /* Get song by ID */

    c.IndentedJSON(http.StatusOK, appleMusicSong)
}

/*
    polyphonicGetAppleSongsBySearch gets an Apple Music song's data from the
    Apple Music API using search terms and then responds with only the required
    data for translation.

    Format of terms: "[track title] [artist name]"
*/
func polyphonicGetAppleSongsBySearch(c *gin.Context) {
    terms := c.Param("terms")

    /* Get song by search */
    appleMusicSongSearchChan := make(chan AppleMusicSongSearch)

    go getAppleMusicSongsBySearch(terms, appleMusicKey, appleMusicSongSearchChan)

    appleMusicSongSearch := <- appleMusicSongSearchChan
    fmt.Println("First Song title:", appleMusicSongSearch.Results.Songs.Data[0].Attributes.Name, "by", appleMusicSongSearch.Results.Songs.Data[0].Attributes.ArtistName)
    /* Get song by search */

    c.IndentedJSON(http.StatusOK, appleMusicSongSearch)
}

/*
    polyphonicGetAppleAlbumByID gets an Apple Music album's data from the Apple
    Music API using the album's ID and then responds with only the required data
    for translation.
*/
func polyphonicGetAppleAlbumByID(c *gin.Context) {
    id := c.Param("id")

    /* Get album by ID */
    appleMusicAlbumChan := make(chan AppleMusicAlbum)
    go getAppleMusicAlbumByID(id, appleMusicKey, appleMusicAlbumChan)

    appleMusicAlbum := <- appleMusicAlbumChan
    fmt.Println("Album title:", appleMusicAlbum.Data[0].Attributes.Name, "by", appleMusicAlbum.Data[0].Attributes.ArtistName)
    /* Get album by ID */

    c.IndentedJSON(http.StatusOK, appleMusicAlbum)
}

/*
    polyphonicGetAppleArtistByID gets an Apple Music artist's data from the
    Apple Music API using the artist's ID and then responds with only the
    required data for translation.
*/
func polyphonicGetAppleArtistByID(c *gin.Context) {
    id := c.Param("id")

    /* Get artist by ID */
    appleMusicArtistChan := make(chan AppleMusicArtist)
    go getAppleMusicArtistByID(id, appleMusicKey, appleMusicArtistChan)

    appleMusicArtist := <- appleMusicArtistChan
    fmt.Println("Artist name:", appleMusicArtist.Data[0].Attributes.Name)
    /* Get artist by ID */

    c.IndentedJSON(http.StatusOK, appleMusicArtist)
}

/*
    polyphonicGetAppleArtistBySearch gets an Apple Music artist's data from the
    Apple Music API using search terms and then responds with only the required
    data for translation.

    Format of terms: "[artist name]"
*/
func polyphonicGetAppleArtistBySearch(c *gin.Context) {
    terms := c.Param("terms")

    /* Get artist by search */
    appleMusicArtistSearchChan := make(chan AppleMusicArtistSearch)

    params := url.QueryEscape(terms)
    go getAppleMusicArtistsBySearch(params, appleMusicKey, appleMusicArtistSearchChan)

    appleMusicArtistSearch := <- appleMusicArtistSearchChan
    fmt.Println("First artist name:", appleMusicArtistSearch.Results.Artists.Data[0].Attributes.Name)
    /* Get Artist by search */

    c.IndentedJSON(http.StatusOK, appleMusicArtistSearch)
}

/*
    polyphonicGetApplePlaylistByID gets an Apple Music playlist's data from the
    Apple Music API using the playlist's ID and then responds with only the
    required data for translation.
*/
func polyphonicGetApplePlaylistByID(c *gin.Context) {
    id := c.Param("id")

    /* Get Playlist by ID */
    appleMusicPlaylistChan := make(chan AppleMusicPlaylist)
    go getAppleMusicPlaylistByID(id, appleMusicKey, appleMusicPlaylistChan)

    appleMusicPlaylist := <- appleMusicPlaylistChan

    next := "Had more to get"
    if appleMusicPlaylist.Data[0].Relationships.Tracks.Next == nil {
        next = "No next"
    } else {
        var getMore bool = true
        var nextURL string = *appleMusicPlaylist.Data[0].Relationships.Tracks.Next

        for getMore {
            nextAppleMusicPlaylistTracksChan := make(chan AppleMusicPlaylistTracks)
            go getNextAppleMusicPlaylist(nextURL, appleMusicKey, nextAppleMusicPlaylistTracksChan)

            nextAppleMusicPlaylistTracks := <- nextAppleMusicPlaylistTracksChan
            fmt.Println("tracks from next section:", len(nextAppleMusicPlaylistTracks.Data))

            appleMusicPlaylist.Data[0].Relationships.Tracks.Data = append(
                appleMusicPlaylist.Data[0].Relationships.Tracks.Data, nextAppleMusicPlaylistTracks.Data...)

            if nextAppleMusicPlaylistTracks.Next == nil {
                fmt.Println("done")
                getMore = false
            } else {
                nextURL = *nextAppleMusicPlaylistTracks.Next
            }
        }
    }

    fmt.Println("Playlist name:", appleMusicPlaylist.Data[0].Attributes.Name, "track count:", len(appleMusicPlaylist.Data[0].Relationships.Tracks.Data),
    "next?", next)

    c.IndentedJSON(http.StatusOK, appleMusicPlaylist)
}

func main() {
    // Capture connection properties.
    cfg := mysql.Config{
        User:   "root",
        Passwd: os.Getenv("DBPASS"),
        Net:    "tcp",
        // Addr:   "docker.for.mac.host.internal:3306",
        Addr:   "127.0.0.1:3306",
        DBName: "polyphonic",
    }

    // Get a database handle.
    var err error
    db, err = sql.Open("mysql", cfg.FormatDSN())
    if err != nil {
        log.Fatal(err)
    }

    pingErr := db.Ping()
    if pingErr != nil {
        log.Fatal(pingErr)
    }
    fmt.Println("Connected!")

    router := gin.Default()

    router.GET("/playlist/:id", getPlaylistByID)
    router.POST("/playlist", postPlaylists)

    /* Spotify API interfacing */
    router.GET("/spotify/song/id/:id",          polyphonicGetSpotifySongByID)
    router.GET("/spotify/song/search/:terms",   polyphonicGetSpotifySongsBySearch)

    router.GET("/spotify/album/id/:id",         polyphonicGetSpotifyAlbumByID)

    router.GET("/spotify/artist/id/:id",        polyphonicGetSpotifyArtistByID)
    router.GET("/spotify/artist/search/:terms", polyphonicGetSpotifyArtistBySearch)

    router.GET("/spotify/playlist/id/:id",      polyphonicGetSpotifyPlaylistByID)
    /* Spotify API interfacing */

    /* Apple Music API interfacing */
    router.GET("/apple/song/id/:id",          polyphonicGetAppleSongByID)
    router.GET("/apple/song/search/:terms",   polyphonicGetAppleSongsBySearch)

    router.GET("/apple/album/id/:id",         polyphonicGetAppleAlbumByID)

    router.GET("/apple/artist/id/:id",        polyphonicGetAppleArtistByID)
    router.GET("/apple/artist/search/:terms", polyphonicGetAppleArtistBySearch)

    router.GET("/apple/playlist/id/:id",      polyphonicGetApplePlaylistByID)
    /* Apple Music API interfacing */

    router.Run("0.0.0.0:7659")
}


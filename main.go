package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

type playlist struct {
    ID         string `json:"id"`
    Name       string `json:"name"`
    Creator    string `json:"creator"`
    SongCount  int    `json:"song_count"`
    Platform   string `json:"platform"`
    Converted  bool   `json:"converted"`
}

type playlist_content struct {
    ID          string `json:"id"`
    Title       string `json:"title"`
    PTrackNum   int    `json:"playlist_track_num"`
    ISRC        string `json:"isrc"`
    Artist      string `json:"artist"`
    Album       string `json:"album"`
    AlbumID     string `json:"album_id"`
    Explicit    bool   `json:"explicit"`
    ConvertURL  string `json:"converted_url"`
    TrackNum    int    `json:"track_num"`
}

type playlist_data struct {
    ID         string `json:"id"`
    Name       string `json:"name"`
    Creator    string `json:"creator"`
    SongCount  int    `json:"song_count"`
    Platform   string `json:"platform"`
    Converted  bool   `json:"converted"`
    Content    []playlist_content `json:"content"`
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
        &playlist.Converted); err != nil {

        log.Println(fmt.Errorf("playlistsById %v", err))
        if err == sql.ErrNoRows {
            c.IndentedJSON(http.StatusNotFound, gin.H{"message": "There is no such playlist"})
            return
        }
        c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Error getting playlist by ID"})
    }

    rows, err := db.Query("SELECT * FROM playlist_content WHERE id = ?", id)
    if err != nil {
        c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Error getting playlist content"})
    }
    defer rows.Close()
    // Loop through rows, using Scan to assign column data to struct fields.
    for rows.Next() {
        var content playlist_content
        if err := rows.Scan(
            &content.ID,
            &content.Title,
            &content.PTrackNum,
            &content.ISRC,
            &content.Artist,
            &content.Album,
            &content.AlbumID,
            &content.Explicit,
            &content.ConvertURL,
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
    _, err := db.Exec("INSERT INTO playlists (id, name, creator, song_count, platform, converted) VALUES (?, ?, ?, ?, ?, ?)",
    newPlaylistData.ID,
    newPlaylistData.Name,
    newPlaylistData.Creator,
    newPlaylistData.SongCount,
    newPlaylistData.Platform,
    newPlaylistData.Converted)

    if err != nil {
        c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Error adding a new playlist"})
    }

    for _, content := range newPlaylistData.Content {
        _, err := db.Exec("INSERT INTO playlist_content (id, title, playlist_track_num, isrc, artist, album, album_id, explicit, converted_url, track_num) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
        content.ID,
        content.Title,
        content.PTrackNum,
        content.ISRC,
        content.Artist,
        content.Album,
        content.AlbumID,
        content.Explicit,
        content.ConvertURL,
        content.TrackNum)

        if err != nil {
            c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Error adding a new playlist"})
        }
    }

    c.IndentedJSON(http.StatusCreated, newPlaylistData)
}

func main() {
    // Capture connection properties.
    cfg := mysql.Config{
        User:   "root",
        Passwd: os.Getenv("DBPASS"),
        Net:    "tcp",
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

    router.Run("localhost:8080")
}


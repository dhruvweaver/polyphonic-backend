# polyphonic-backend
Backend for [Polyphonic](https://github.com/dhruvweaver/Polyphonic) app:
- Provides communation between app and Spotify's and Apple Music's servers so that rate limit is controlled
- Supplies playlist sharing capability

Project still in early stages, the setup instructions may change over time.

## Setup Instructions
***Important:***
**You must add the following variables to your environment:**

```zsh
export SPOTIFY_CLIENT_ID=spotify-client-id
export SPOTIFY_CLIENT_SECRET=spotify-client-secret
export APPLE_MUSIC_KEY=apple-music-api-key
```

_Optional if you want to use an SSL key:_

```zsh
export POLYPHONIC_SSL_CERT_PATH=/path/to/cert.pem
export POLYPHONIC_SSL_KEY_PATH=/path/to/key.pem
```

If you aren't using SSL, change the following line (from main.go):
```
router.RunTLS("0.0.0.0:7659", certPath, keyPath)
```
to:
```
router.Run("0.0.0.0:7659")
```

### Tools you will need:
- mysql
- go

_**Note:** Windows instructions not included._

_The steps will likely be similar but these instructions are written with macOS and Linux in mind._

### Install mysql
Use your package manager to install the MySQL CLI.

_macOS_
```zsh
$ brew install mysql
```
_Ubuntu_
```bash
$ sudo apt-get install mysql-server
```
_Fedora/RHEL_
```bash
$ sudo dnf install community-mysql-server
```

### Initialize database
***Start the mysql server:***

_macOS_
```zsh
$ brew services start mysql
```
_Ubuntu_
```bash
$ sudo systemctl start mysql
```
_Fedora/RHEL_
```bash
$ sudo systemctl start mysqld
```

***Log in to the MySQL manager:***

```zsh
$ mysql -u root
```

***Create database:***

```shell
mysql> CREATE DATABASE polyphonic;
mysql> use polyphonic;
```

***Setup tables:***

You can either enter this SQL script directly into the mysql prompt, or place it into a file and source it:
```
DROP TABLE IF EXISTS playlists;
CREATE TABLE playlists (
  id         VARCHAR(255) NOT NULL,
  name       VARCHAR(255) NOT NULL,
  creator    VARCHAR(255) NOT NULL,
  song_count INT  NOT NULL,
  platform   VARCHAR(255) NOT NULL,
  original_url VARCHAR(255) NOT NULL,
  converted  BOOLEAN NOT NULL,
  PRIMARY KEY (`id`)
);

DROP TABLE IF EXISTS playlist_content;
CREATE TABLE playlist_content (
  id         VARCHAR(255) NOT NULL,
  key_id     VARCHAR(255) NOT NULL,
  title      VARCHAR(255) NOT NULL,
  playlist_track_num  INT NOT NULL,
  isrc       VARCHAR(255) NOT NULL,
  artist     VARCHAR(255) NOT NULL,
  album      VARCHAR(255) NOT NULL,
  album_id   VARCHAR(255) NOT NULL,
  explicit   BOOLEAN NOT NULL,
  original_url VARCHAR(255) NOT NULL,
  converted_url VARCHAR(255),
  confidence INT NOT NULL,
  track_num  INT NOT NULL,
  PRIMARY KEY (`key_id`)
);
```
If you are using the file sourcing method, enter the following command:
```shell
mysql> source /path/to/file.sql
```

### Run server
**Docker**

If you are using Docker to run the server, change the database IP address in main.go to "host.docker.internal"

_Build:_
```bash
$ docker build -t polyphonic-backend .
```
_Run:_
```bash
$ docker run --network="host" --add-host host.docker.internal:127.0.0.1 polyphonic-backend
```

**Run natively**

***Install go***

_macOS_
```zsh
$ brew install go
```
_Ubuntu_
```bash
$ sudo apt-get install golang
```
_Fedora/RHEL_
```bash
$ sudo dnf install golang
```

***Run go program***
Within the polyphonic-backend directory:
```zsh
$ go run .
```


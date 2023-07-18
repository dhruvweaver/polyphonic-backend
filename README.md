# polyphonic-backend
Backend for [Polyphonic](https://github.com/dhruvweaver/Polyphonic) app, supplying playlist sharing capability.

Project still in early stages, the setup instructions may change over time.

## Setup Instructions
Tools you will need:
- mysql
- go

_**Note:** Windows instructions not included.

The steps will be similar but these instructions are written with macOS and Linux in mind._

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
  converted  BOOLEAN NOT NULL,
  PRIMARY KEY (`id`)
);

DROP TABLE IF EXISTS playlist_content;
CREATE TABLE playlist_content (
  id         VARCHAR(255) NOT NULL,
  title      VARCHAR(255) NOT NULL,
  playlist_track_num INT NOT NULL,
  isrc       VARCHAR(255) NOT NULL,
  artist     VARCHAR(255) NOT NULL,
  album      VARCHAR(255) NOT NULL,
  album_id   VARCHAR(255) NOT NULL,
  explicit   BOOLEAN NOT NULL,
  converted_url VARCHAR(255),
  track_num  INT NOT NULL
);
```
If you are using the file sourcing method, enter the following command:
```shell
mysql> source /path/to/file.sql
```

### Run server
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


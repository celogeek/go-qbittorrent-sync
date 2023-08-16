# go-qbittorrent-sync
QBittorrent &amp; Rsync auto sync tools writing in go

This tools will automatically run an Rsync command of completed torrent in qbittorrent tags with Sync. 

While the transfer is in progress, it update the tag with Progress:X%. 

And at the end it can remove or set a tag like Synced.

## Installation
First ensure to have a working version of GO: [Installation](https://go.dev/doc/install)

Then install the last version of the tool:
```
$ go install github.com/celogeek/go-qbittorrent-sync@latest
```

## Usage

```
$ go-qbittorrent-sync -h

Usage of go-qbittorrent-sync:
  -pool-time int
    	Number of second to check new files to sync (default 30)
  -qbittorrent-password string
    	Password of qbittorrent
  -qbittorrent-sync-tag string
    	Tag of qbittorrent to copy (default "Sync")
  -qbittorrent-synced-tag string
    	Tag of qbittorrent when copy finished
  -qbittorrent-uri string
    	URI of qbittorrent (default "http://localhost:8080")
  -qbittorrent-username string
    	Username of qbittorrent
  -rsync-destination string
    	Rsync Destination directory (default ".")
  -rsync-hostname string
    	Rsync host
  -rsync-rsh string
    	Rsync rsh command (default ".")
  -rsync-username string
    	Rsync username
```

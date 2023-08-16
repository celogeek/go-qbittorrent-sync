package main

import (
	"flag"
	"log"
	"time"
)

func main() {
	qbitoptions := &QBitTorrentOptions{}
	rsyncoptions := &RsyncOptions{}
	var poolTime int
	flag.StringVar(&qbitoptions.Uri, "qbittorrent-uri", "http://localhost:8080", "URI of qbittorrent")
	flag.StringVar(&qbitoptions.Username, "qbittorrent-username", "", "Username of qbittorrent")
	flag.StringVar(&qbitoptions.Password, "qbittorrent-password", "", "Password of qbittorrent")
	flag.StringVar(&qbitoptions.SyncTag, "qbittorrent-sync-tag", "Sync", "Tag of qbittorrent to copy")
	flag.StringVar(&qbitoptions.SyncedTag, "qbittorrent-synced-tag", "", "Tag of qbittorrent when copy finished")
	flag.StringVar(&rsyncoptions.Hostname, "rsync-hostname", "", "Rsync host")
	flag.StringVar(&rsyncoptions.Username, "rsync-username", "", "Rsync username")
	flag.StringVar(&rsyncoptions.Destination, "rsync-destination", ".", "Rsync Destination directory")
	flag.StringVar(&rsyncoptions.Rsh, "rsync-rsh", ".", "Rsync rsh command")
	flag.IntVar(&poolTime, "pool-time", 30, "Number of second to check new files to sync")
	flag.Parse()

	if qbitoptions.Uri == "" ||
		qbitoptions.Username == "" ||
		qbitoptions.Password == "" ||
		qbitoptions.SyncTag == "" {
		log.Fatal("missing qbittorrent parameters")
	}

	if rsyncoptions.Hostname == "" {
		log.Fatal("missing rsync parameters")
	}

	log.Print("[QBit] Authentication")
	qcli, err := NewQBittorrentCli(qbitoptions)
	if err != nil {
		log.Fatalf("[QBit] Init Error: %v", err)
	}
	defer qcli.Logout()

	t := time.NewTicker(time.Second * time.Duration(poolTime))
	defer t.Stop()

	log.Print("[QBit] Pooling")
	for {
		torrents, err := qcli.List()
		if err != nil {
			log.Printf("[QBit] List Error: %v", err)
		}
		if len(torrents) > 0 {
			log.Printf("[QBit] Found %d torrents to sync", len(torrents))
		}

		for _, t := range torrents {
			log.Printf("[Rsync] Synching %s", t.Name)
			rtask := NewRsync(
				&RsyncOptions{
					Username:    rsyncoptions.Username,
					Hostname:    rsyncoptions.Hostname,
					Destination: rsyncoptions.Destination,
					Rsh:         rsyncoptions.Rsh,
					Path:        t.Path,
					OnProgress: func(p int) {
						err := qcli.SetProgress(t, p)
						if err != nil {
							log.Printf("[QBit] SetProgress Error: %v", err)
						}
					},
				})

			if err := rtask.Run(); err != nil {
				qcli.ClearTags()
				log.Printf("[Rsync] Error: %v", err)
			} else {
				qcli.SetDone(t)
				log.Printf("[Rsync] Synching %s done", t.Name)
			}
		}

		<-t.C
	}

}

func init() {
	log.SetFlags(0)
}

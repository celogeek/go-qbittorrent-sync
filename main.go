package main

import (
	"flag"
	"log"
	"time"
)

type Options struct {
	Qbittorent QBitTorrentOptions
}

func main() {
	options := &Options{}
	flag.StringVar(&options.Qbittorent.Uri, "qbittorrent-uri", "http://localhost:8080", "URI of qbittorrent")
	flag.StringVar(&options.Qbittorent.Username, "qbittorrent-username", "", "Username of qbittorrent")
	flag.StringVar(&options.Qbittorent.Password, "qbittorrent-password", "", "Password of qbittorrent")
	flag.StringVar(&options.Qbittorent.SyncTag, "qbittorrent-sync-tag", "Sync", "Tag of qbittorrent to copy")
	flag.StringVar(&options.Qbittorent.SyncedTag, "qbittorrent-synced-tag", "", "Tag of qbittorrent when copy finished")
	flag.Parse()

	if options.Qbittorent.Uri == "" ||
		options.Qbittorent.Username == "" ||
		options.Qbittorent.Password == "" ||
		options.Qbittorent.SyncTag == "" {
		log.Fatal("missing qbittorrent parameters")
	}

	qcli, err := NewQBittorrentCli(&options.Qbittorent)
	if err != nil {
		log.Fatal(err)
	}
	defer qcli.Logout()

	torrents, err := qcli.List()
	if err != nil {
		log.Fatal(err)
	}

	for _, t := range torrents {
		for p := 0; p <= 100; p += 20 {
			qcli.ClearTags()
			qcli.SetProgress(&t, p)
			time.Sleep(time.Second)
		}
		qcli.SetDone(&t)
	}
}

func init() {
	log.SetFlags(0)
}

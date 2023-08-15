package main

import (
	"flag"
	"fmt"
	"log"
)

type RsyncOptions struct {
	Username string
	Hostname string
}

func (r *RsyncOptions) Uri(path string) string {
	result := fmt.Sprintf("%s:%s", r.Hostname, path)
	if r.Username == "" {
		return result
	}
	return fmt.Sprintf("%s@%s", r.Username, result)
}

func main() {
	qbitoptions := &QBitTorrentOptions{}
	rsyncoptions := &RsyncOptions{}
	dest := ""
	flag.StringVar(&qbitoptions.Uri, "qbittorrent-uri", "http://localhost:8080", "URI of qbittorrent")
	flag.StringVar(&qbitoptions.Username, "qbittorrent-username", "", "Username of qbittorrent")
	flag.StringVar(&qbitoptions.Password, "qbittorrent-password", "", "Password of qbittorrent")
	flag.StringVar(&qbitoptions.SyncTag, "qbittorrent-sync-tag", "Sync", "Tag of qbittorrent to copy")
	flag.StringVar(&qbitoptions.SyncedTag, "qbittorrent-synced-tag", "", "Tag of qbittorrent when copy finished")
	flag.StringVar(&rsyncoptions.Hostname, "rsync-hostname", "", "Rsync host")
	flag.StringVar(&rsyncoptions.Username, "rsync-username", "", "Rsync username")
	flag.StringVar(&dest, "dest", ".", "Destination directory")
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

	qcli, err := NewQBittorrentCli(qbitoptions)
	if err != nil {
		log.Fatal(err)
	}
	defer qcli.Logout()

	torrents, err := qcli.List()
	if err != nil {
		log.Fatal(err)
	}

	for _, t := range torrents {
		rtask := NewRsync(
			rsyncoptions.Uri(t.Path),
			dest,
			func(p int) {
				qcli.SetProgress(t, p)
			},
		)

		if err := rtask.Run(); err != nil {
			qcli.ClearTags()
			log.Fatal(err)
		}

		qcli.SetDone(t)
	}
}

func init() {
	log.SetFlags(0)
}

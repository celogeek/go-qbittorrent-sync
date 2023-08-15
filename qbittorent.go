package main

import (
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
)

type QBitTorrentOptions struct {
	Uri       string
	Username  string
	Password  string
	SyncTag   string
	SyncedTag string
}

type QBittorrentCli struct {
	SyncTag   string
	SyncedTag string
	cli       *resty.Client
}

type Torrent struct {
	Name string `json:"name"`
	Path string `json:"content_path"`
	Hash string `json:"hash"`

	Progress int `json:"-"`
}

func NewQBittorrentCli(options *QBitTorrentOptions) (*QBittorrentCli, error) {
	r := resty.New().SetBaseURL(fmt.Sprintf("%s/api/v2", options.Uri))

	_, err := r.
		R().
		SetFormData(map[string]string{
			"username": options.Username,
			"password": options.Password,
		}).
		Post("/auth/login")

	if err != nil {
		return nil, err
	}

	cli := &QBittorrentCli{
		SyncTag:   options.SyncTag,
		SyncedTag: options.SyncedTag,
		cli:       r,
	}

	cli.ClearTags()

	return cli, nil
}

func (c *QBittorrentCli) Logout() error {
	_, err := c.cli.R().Post("/auth/logout")
	return err
}

func (c *QBittorrentCli) List() ([]*Torrent, error) {
	result := make([]*Torrent, 0)

	_, err := c.cli.R().
		SetQueryParam("filter", "completed").
		SetQueryParam("tag", c.SyncTag).
		SetResult(&result).
		Get("/torrents/info")

	if err != nil {
		return nil, err
	}

	for _, t := range result {
		t.Progress = -1
	}

	return result, nil
}

func (c *QBittorrentCli) ClearTags() error {
	// fetch tags
	tags := []string{}
	_, err := c.cli.
		R().
		SetResult(&tags).
		Get("/torrents/tags")

	if err != nil {
		return err
	}

	// check existing tags
	hasSync := false
	toDeleteTags := []string{}
	for _, t := range tags {
		if t == c.SyncTag {
			hasSync = true
		}
		if strings.HasPrefix(t, "Progress:") {
			toDeleteTags = append(toDeleteTags, t)
		}
	}

	// create missing sync
	if !hasSync {
		_, err := c.cli.
			R().
			SetFormData(map[string]string{
				"tags": c.SyncTag,
			}).
			Post("/torrents/createTags")

		if err != nil {
			return err
		}
	}

	// remove progress
	if len(toDeleteTags) > 0 {
		_, err := c.cli.
			R().
			SetFormData(map[string]string{
				"tags": strings.Join(toDeleteTags, ","),
			}).
			Post("/torrents/deleteTags")

		if err != nil {
			return err
		}
	}

	return nil
}

func (c *QBittorrentCli) SetTag(t *Torrent, tag string) error {
	_, err := c.cli.
		R().
		SetFormData(map[string]string{
			"hashes": t.Hash,
			"tags":   tag,
		}).
		Post("/torrents/addTags")

	return err
}

func (c *QBittorrentCli) SetProgress(t *Torrent, p int) error {
	if t.Progress == p {
		return nil
	}

	err := c.ClearTags()
	if err != nil {
		return err
	}

	c.SetTag(t, fmt.Sprintf("Progress:%d%%", p))
	if err != nil {
		return err
	}

	t.Progress = p
	return nil
}

func (c *QBittorrentCli) SetDone(t *Torrent) error {
	err := c.ClearTags()
	if err != nil {
		return err
	}

	_, err = c.cli.
		R().
		SetFormData(map[string]string{
			"hashes": t.Hash,
			"tags":   c.SyncTag,
		}).
		Post("/torrents/deleteTags")

	if err != nil {
		return err
	}

	if c.SyncedTag != "" {
		return c.SetTag(t, c.SyncedTag)
	}
	return nil
}

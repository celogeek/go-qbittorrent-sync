package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"sync"
)

type RsyncOptions struct {
	Username    string
	Hostname    string
	Path        string
	Destination string
	Rsh         string
	OnProgress  func(p int)
}

func (r *RsyncOptions) Uri() string {
	result := fmt.Sprintf("%s:%s", r.Hostname, r.Path)
	if r.Username == "" {
		return result
	}
	return fmt.Sprintf("%s@%s", r.Username, result)
}

type Rsync struct {
	Source      string
	Destination string
	Rsh         string
	OnProgress  func(p int)

	progress int
}

func NewRsync(options *RsyncOptions) *Rsync {
	return &Rsync{
		Source:      options.Uri(),
		Destination: options.Destination,
		Rsh:         options.Rsh,
		OnProgress:  options.OnProgress,
		progress:    -1,
	}
}

func ScanCR(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\r'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

func (r *Rsync) Run() error {
	args := []string{
		"--archive",
		"--partial",
		"--inplace",
		"--no-inc-recursive",
		"--info=progress2",
		"-s",
	}
	if r.Rsh != "" {
		args = append(args, "--rsh", r.Rsh)
	}
	args = append(args, r.Source, r.Destination)
	cmd := exec.Command(
		"rsync",
		args...,
	)
	cmd.Stderr = os.Stderr
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	defer out.Close()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(out)
		scanner.Split(ScanCR)
		progressMatch := regexp.MustCompile(`(\d+)%`)
		for scanner.Scan() {
			progress := scanner.Text()
			m := progressMatch.FindStringSubmatch(progress)
			if len(m) == 2 {
				if p, err := strconv.Atoi(m[1]); err == nil {
					r.OnProgress(p)
				}
			}
		}
	}()

	if err := cmd.Start(); err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	wg.Wait()
	return nil
}

package hlsdl

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sort"

	"github.com/grafov/m3u8"
	"github.com/sunshineplan/gohttp"
	"github.com/sunshineplan/utils/progressbar"
	"github.com/sunshineplan/utils/retry"
	"github.com/sunshineplan/utils/workers"
)

const defaultName = "output.ts"

type Downloader struct {
	m3u8    string
	workers int
	results []errResult
}

type errResult struct {
	id  uint64
	err error
}

func SetAgent(ua string) {
	gohttp.SetAgent(ua)
}

func NewTask(m3u8 string) *Downloader {
	return &Downloader{m3u8: m3u8}
}

func (d *Downloader) SetWorkers(n int) *Downloader {
	if n > 0 {
		d.workers = n
	}
	return d
}

func (d *Downloader) dlSegment(s *m3u8.MediaSegment, path, output string) {
	output = filepath.Join(path, output+".tmp", fmt.Sprintf("%d.ts", s.SeqId))

	if err := retry.Do(
		func() error {
			resp, err := gohttp.Get(s.URI, nil)
			if err != nil {
				return err
			}
			if resp.StatusCode != 200 {
				return fmt.Errorf("no StatusOK response from %s", s.URI)
			}
			_, err = resp.Save(output)
			return err
		}, 5, 5,
	); err != nil {
		d.results = append(d.results, errResult{s.SeqId, err})
		os.WriteFile(output, nil, 0666)
	}
}

func (d *Downloader) dlSegments(s []*m3u8.MediaSegment, path, output string) {
	pb := progressbar.New(len(s))
	pb.Start()
	defer pb.Done()

	workers.RunSlice(d.workers, s, func(_ int, segment *m3u8.MediaSegment) {
		defer pb.Add(1)
		d.dlSegment(segment, path, output)
	})
}

func (d *Downloader) Run(path, output string) error {
	if output == "" {
		output = defaultName
	}

	var playlist *m3u8.MediaPlaylist
	u, err := url.Parse(d.m3u8)
	if err != nil {
		return fmt.Errorf("invalid m3u8")
	} else {
		if u.Scheme == "http" || u.Scheme == "https" {
			u, playlist, err = FetchM3U8MediaPlaylist(u)
			if err != nil {
				return err
			}
		} else {
			u, playlist, err = LoadM3U8MediaPlaylist(d.m3u8)
			if err != nil {
				return err
			}
		}
	}

	tmp := filepath.Join(path, output+".tmp")
	if err := os.MkdirAll(tmp, 0755); err != nil {
		return err
	}

	log.Println("Downloading from", u)

	d.dlSegments(playlist.Segments, path, output)

	log.Print("Merging segments...")

	f, err := os.Create(filepath.Join(path, output))
	if err != nil {
		return err
	}
	defer f.Close()

	sort.Slice(playlist.Segments, func(i, j int) bool {
		return playlist.Segments[i].SeqId < playlist.Segments[j].SeqId
	})

	for _, segment := range playlist.Segments {
		file := filepath.Join(tmp, fmt.Sprintf("%d.ts", segment.SeqId))
		data, err := read(segment, file)
		if err != nil {
			return err
		}

		if _, err := f.Write(data); err != nil {
			return err
		}

		if err := os.Remove(file); err != nil {
			return err
		}
	}
	if len(d.results) > 0 {
		fmt.Printf("Total %d Error:\n", len(d.results))
		for _, i := range d.results {
			fmt.Printf("id: %d, error: %s\n", i.id, i.err)
		}
	}

	if err := os.Remove(tmp); err != nil {
		return err
	}

	log.Print("All Done.")

	return nil
}

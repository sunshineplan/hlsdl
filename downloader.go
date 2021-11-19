package hlsdl

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"

	"github.com/grafov/m3u8"
	"github.com/sunshineplan/gohttp"
	"github.com/sunshineplan/utils"
	"github.com/sunshineplan/utils/progressbar"
	"github.com/sunshineplan/utils/workers"
)

const defaultName = "output.ts"

type Downloader struct {
	m3u8    string
	workers *workers.Workers
	results []errResult
}

type errResult struct {
	id  uint64
	err error
}

func SetAgent(ua string) {
	gohttp.SetAgent(ua)
}

func NewTask(url string) *Downloader {
	return &Downloader{m3u8: url, workers: workers.New(runtime.NumCPU())}
}

func (d *Downloader) SetWorkers(n int) *Downloader {
	if n > 0 {
		d.workers = workers.New(n)
	}
	return d
}

func (d *Downloader) dlSegment(s *m3u8.MediaSegment, path, output string) {
	output = filepath.Join(path, output+".tmp", fmt.Sprintf("%d.ts", s.SeqId))

	if err := utils.Retry(
		func() error {
			res := gohttp.Get(s.URI, nil)
			if res.Error != nil {
				return res.Error
			}
			if res.StatusCode != 200 {
				return fmt.Errorf("no StatusOK response from %s", s.URI)
			}

			return res.Save(output)
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

	d.workers.Slice(s, func(_ int, item interface{}) {
		defer pb.Add(1)
		d.dlSegment(item.(*m3u8.MediaSegment), path, output)
	})
}

func (d *Downloader) Run(path, output string) error {
	if output == "" {
		output = defaultName
	}

	u, err := url.Parse(d.m3u8)
	if err != nil {
		return fmt.Errorf("invalid m3u8 url")
	}

	segments, err := getSegments(u)
	if err != nil {
		return err
	}

	tmp := filepath.Join(path, output+".tmp")
	if err := os.MkdirAll(tmp, 0755); err != nil {
		return err
	}

	d.dlSegments(segments, path, output)

	log.Print("Merging segments...")

	f, err := os.Create(filepath.Join(path, output))
	if err != nil {
		return err
	}
	defer f.Close()

	sort.Slice(segments, func(i, j int) bool {
		return segments[i].SeqId < segments[j].SeqId
	})

	for _, segment := range segments {
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

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

const path = "output"

type Downloader struct {
	m3u8    string
	workers *workers.Workers
}

type errResult struct {
	id  uint64
	err error
}

var results []errResult

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

func dlSegment(s *m3u8.MediaSegment, output string) {
	output = filepath.Join(path, output+".tmp", fmt.Sprintf("%d.ts", s.SeqId))

	var res *gohttp.Response
	if err := utils.Retry(
		func() error {
			res = gohttp.Get(s.URI, nil)
			if res.StatusCode != 200 {
				return fmt.Errorf("no StatusOK response from %s", s.URI)
			}
			if res.Error != nil {
				return res.Error
			}

			return res.Save(output)
		}, 5, 5,
	); err != nil {
		results = append(results, errResult{id: s.SeqId, err: err})

		os.OpenFile(output, os.O_RDONLY|os.O_CREATE, 0644)
	}
}

func (d *Downloader) dlSegments(s []*m3u8.MediaSegment, output string) error {
	pb := progressbar.New(len(s))
	pb.Start()

	if err := d.workers.Slice(s, func(_ int, item interface{}) {
		defer pb.Add(1)

		dlSegment(item.(*m3u8.MediaSegment), output)
	}); err != nil {
		return err
	}
	<-pb.Done

	return nil
}

func (d *Downloader) Run(output string) error {
	if output == "" {
		output = "output.mp4"
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
	if err := os.MkdirAll(tmp, 0644); err != nil {
		return err
	}

	if err := d.dlSegments(segments, output); err != nil {
		return err
	}

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
		d, err := save(segment, file)
		if err != nil {
			return err
		}

		if _, err := f.Write(d); err != nil {
			return err
		}

		if err := os.Remove(file); err != nil {
			return err
		}
	}
	if len(results) > 0 {
		fmt.Printf("Total %d Error:\n", len(results))
		for _, i := range results {
			fmt.Printf("id: %d, error: %s\n", i.id, i.err)
		}
	}

	if err := os.Remove(tmp); err != nil {
		return err
	}

	log.Print("All Done.")

	return nil
}

package hlsdl

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net/url"
	"os"
	"sort"
	"time"

	"github.com/grafov/m3u8"
	"github.com/sunshineplan/gohttp"
	"github.com/sunshineplan/utils/cache"
)

var c = cache.New(false)

func getSegments(u *url.URL) ([]*m3u8.MediaSegment, error) {
	mediaList, err := getM3u8MediaPlaylist(u)
	if err != nil {
		return nil, err
	}

	segments := []*m3u8.MediaSegment{}
	for _, s := range mediaList.Segments {
		if s == nil {
			continue
		}

		if s.Key != nil && s.Key.URI != "" {
			u, err := u.Parse(s.Key.URI)
			if err != nil {
				return nil, err
			}
			s.Key.URI = u.String()
		}

		if s.Discontinuity {
			mediaList.Key = s.Key
		} else {
			if s.Key == nil && mediaList.Key != nil {
				s.Key = mediaList.Key
			}
		}

		u, err := u.Parse(s.URI)
		if err != nil {
			return nil, err
		}
		s.URI = u.String()

		segments = append(segments, s)
	}

	return segments, nil
}

func read(s *m3u8.MediaSegment, file string) ([]byte, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}

	if s.Key != nil && s.Key.URI != "" && s.Key.Method == "AES-128" {
		key, err := getKey(s.Key.URI)
		if err != nil {
			return nil, err
		}

		iv := []byte(s.Key.IV)
		if len(iv) == 0 {
			iv = make([]byte, 16)
			binary.BigEndian.PutUint64(iv[8:], s.SeqId)
		}

		data, err = decryptAES128(data, key, iv)
		if err != nil {
			return nil, err
		}
	}

	for j := 0; j < len(data); j++ {
		// look for sync byte
		if data[j] == 0x47 {
			data = data[j:]
			break
		}
	}

	return data, nil
}

func getM3u8MediaPlaylist(u *url.URL) (*m3u8.MediaPlaylist, error) {
	res := gohttp.Get(u.String(), nil)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("no StatusOK response from %s", u)
	}

	playlist, _, err := m3u8.DecodeFrom(bytes.NewBuffer(res.Bytes()), false)
	if err != nil {
		return nil, err
	}
	if media, ok := playlist.(*m3u8.MediaPlaylist); ok {
		log.Println("Downloading from", u)
		return media, nil
	}
	if master, ok := playlist.(*m3u8.MasterPlaylist); ok {
		for _, i := range master.Variants {
			u, err := u.Parse(i.URI)
			if err != nil {
				continue
			}
			i.URI = u.String()
		}
		sort.SliceStable(master.Variants, func(i, j int) bool {
			return master.Variants[i].Bandwidth > master.Variants[j].Bandwidth
		})
		if len(master.Variants) != 0 {
			log.Print("Parse from master playlist:")
			fmt.Println(master)

			u, err = u.Parse(master.Variants[0].URI)
			if err != nil {
				return nil, err
			}
			return getM3u8MediaPlaylist(u)
		}
	}
	return nil, fmt.Errorf("unknown playlist type")
}

func getKey(url string) (b []byte, err error) {
	value, ok := c.Get(url)
	if ok {
		b = value.([]byte)
		return
	}

	res := gohttp.Get(url, nil)
	if res.Error != nil {
		err = res.Error
		return
	}
	if res.StatusCode != 200 {
		err = fmt.Errorf("no StatusOK response from %s", url)
		return
	}

	b = res.Bytes()

	c.Set(url, b, time.Hour, nil)

	return
}

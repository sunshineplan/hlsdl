package hlsdl

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"sort"
	"time"

	"github.com/grafov/m3u8"
	"github.com/sunshineplan/gohttp"
	"github.com/sunshineplan/utils/cache"
)

var c = cache.New(false)

func parse(url *url.URL, playlist m3u8.Playlist) (*url.URL, *m3u8.MediaPlaylist, error) {
	switch playlist := playlist.(type) {
	case *m3u8.MediaPlaylist:
		var segments []*m3u8.MediaSegment
		for _, s := range playlist.Segments {
			if s == nil {
				continue
			}

			if s.Key != nil && s.Key.URI != "" {
				u, err := url.Parse(s.Key.URI)
				if err != nil {
					return nil, nil, err
				}
				s.Key.URI = u.String()
			}

			if s.Discontinuity {
				playlist.Key = s.Key
			} else {
				if s.Key == nil && playlist.Key != nil {
					s.Key = playlist.Key
				}
			}

			u, err := url.Parse(s.URI)
			if err != nil {
				return nil, nil, err
			}
			s.URI = u.String()

			segments = append(segments, s)
		}
		playlist.Segments = segments

		return url, playlist, nil
	case *m3u8.MasterPlaylist:
		for _, i := range playlist.Variants {
			u, err := url.Parse(i.URI)
			if err != nil {
				continue
			}
			i.URI = u.String()
		}
		sort.SliceStable(playlist.Variants, func(i, j int) bool {
			return playlist.Variants[i].Bandwidth > playlist.Variants[j].Bandwidth
		})
		if len(playlist.Variants) != 0 {
			slog.Debug("Parse from master playlist:\n" + playlist.String())
			u, err := url.Parse(playlist.Variants[0].URI)
			if err != nil {
				return nil, nil, err
			}
			return FetchM3U8MediaPlaylist(u)
		} else {
			return nil, nil, fmt.Errorf("empty master playlist")
		}
	}

	return nil, nil, fmt.Errorf("unknown playlist type")
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

func getKey(url string) (b []byte, err error) {
	value, ok := c.Get(url)
	if ok {
		b = value.([]byte)
		return
	}

	resp, err := gohttp.Get(url, nil)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("no StatusOK response from %s", url)
		return
	}

	b = resp.Bytes()

	c.Set(url, b, time.Hour, nil)

	return
}

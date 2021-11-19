package hlsdl

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/grafov/m3u8"
	"github.com/sunshineplan/gohttp"
	"github.com/sunshineplan/utils/cache"
)

var c = cache.New(false)

func getSegments(u *url.URL) ([]*m3u8.MediaSegment, error) {
	p, t, err := getM3u8ListType(u.String())
	if err != nil {
		return nil, err
	}
	if t != m3u8.MEDIA {
		return nil, fmt.Errorf("unsupported m3u8 type: %d", t)
	}

	mediaList := p.(*m3u8.MediaPlaylist)
	segments := []*m3u8.MediaSegment{}
	for _, s := range mediaList.Segments {
		if s == nil {
			continue
		}

		if !strings.HasPrefix(s.URI, "http") {
			segmentURL, err := u.Parse(s.URI)
			if err != nil {
				return nil, err
			}

			s.URI = segmentURL.String()
		}

		if s.Discontinuity {
			mediaList.Key = s.Key
		} else {
			if s.Key == nil && mediaList.Key != nil {
				s.Key = mediaList.Key
			}
		}

		if s.Key != nil && s.Key.URI != "" && !strings.HasPrefix(s.Key.URI, "http") {
			keyURL, err := u.Parse(s.Key.URI)
			if err != nil {
				return nil, err
			}

			s.Key.URI = keyURL.String()
		}

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

func getM3u8ListType(url string) (m3u8.Playlist, m3u8.ListType, error) {
	res := gohttp.Get(url, nil)
	if res.Error != nil {
		return nil, 0, res.Error
	}
	if res.StatusCode != 200 {
		return nil, 0, fmt.Errorf("no StatusOK response from %s", url)
	}

	return m3u8.DecodeFrom(bytes.NewBuffer(res.Bytes()), false)
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

package hlsdl

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/grafov/m3u8"
	"github.com/sunshineplan/chrome"
	"github.com/sunshineplan/gohttp"
)

var urlParse = url.Parse

func LoadM3U8MediaPlaylist(file string, debug bool) (*url.URL, *m3u8.MediaPlaylist, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, nil, err
	}
	playlist, _, err := m3u8.DecodeFrom(f, false)
	if err != nil {
		return nil, nil, err
	}
	return parse(&url.URL{Host: file}, playlist, debug)
}

func FetchM3U8MediaPlaylist(u *url.URL, debug bool) (*url.URL, *m3u8.MediaPlaylist, error) {
	res := gohttp.Get(u.String(), nil)
	if res.Error != nil {
		return nil, nil, res.Error
	}
	if res.StatusCode != 200 {
		return nil, nil, fmt.Errorf("no StatusOK response from %s", u)
	}

	var r io.Reader
	playlist, _, err := m3u8.DecodeFrom(bytes.NewReader(res.Bytes()), false)
	if err != nil {
		if debug {
			log.Println("Analyzing", u)
		}
		r, u, err = fetchURL(u.String())
		if err != nil {
			return nil, nil, err
		}
		if debug {
			log.Println("Found", u)
		}
		playlist, _, err = m3u8.DecodeFrom(r, false)
		if err != nil {
			return nil, nil, err
		}
	}

	return parse(u, playlist, debug)
}

func fetchURL(url string) (io.Reader, *url.URL, error) {
	c := chrome.Headless(false)
	if _, _, err := c.WithTimeout(15 * time.Second); err != nil {
		return nil, nil, err
	}
	defer c.Close()

	if err := c.EnableFetch(func(ev *fetch.EventRequestPaused) bool {
		return ev.ResourceType == network.ResourceTypeImage ||
			ev.ResourceType == network.ResourceTypeStylesheet ||
			ev.ResourceType == network.ResourceTypeMedia
	}); err != nil {
		return nil, nil, err
	}

	done := c.ListenEvent(chrome.URLContains(".m3u8"), "GET", true)
	if err := chromedp.Run(c, chromedp.Navigate(url)); err != nil {
		return nil, nil, err
	}
	select {
	case <-c.Done():
		return nil, nil, c.Err()
	case e := <-done:
		u, _ := urlParse(e.Response.Response.URL)
		return bytes.NewReader(e.Bytes), u, nil
	}
}

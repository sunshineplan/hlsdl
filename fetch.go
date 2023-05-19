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
	"github.com/sunshineplan/useragent"
)

var ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36"

func DefaultUserAgent() string { return ua }

func init() {
	ua = useragent.UserAgent(ua)
	SetAgent(ua)
}

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
	resp, err := gohttp.Get(u.String(), nil)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode != 200 {
		return nil, nil, fmt.Errorf("no StatusOK response from %s", u)
	}

	var r io.Reader
	playlist, _, err := m3u8.DecodeFrom(bytes.NewReader(resp.Bytes()), false)
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

func fetchURL(address string) (io.Reader, *url.URL, error) {
	c := chrome.Headless().UserAgent(ua).DisableAutomationControlled()
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
	if err := chromedp.Run(c, chromedp.Navigate(address)); err != nil {
		return nil, nil, err
	}
	select {
	case <-c.Done():
		return nil, nil, c.Err()
	case e := <-done:
		u, _ := url.Parse(e.Response.Response.URL)
		return bytes.NewReader(e.Bytes), u, nil
	}
}

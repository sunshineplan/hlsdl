package hlsdl

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/grafov/m3u8"
	"github.com/sunshineplan/gohttp"
)

var urlParse = url.Parse

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
		r, u, err = fetchM3U8(u.String())
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

	switch playlist := playlist.(type) {
	case *m3u8.MediaPlaylist:
		var segments []*m3u8.MediaSegment
		for _, s := range playlist.Segments {
			if s == nil {
				continue
			}

			if s.Key != nil && s.Key.URI != "" {
				u, err := u.Parse(s.Key.URI)
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

			u, err := u.Parse(s.URI)
			if err != nil {
				return nil, nil, err
			}
			s.URI = u.String()

			segments = append(segments, s)
		}
		playlist.Segments = segments

		return u, playlist, nil
	case *m3u8.MasterPlaylist:
		for _, i := range playlist.Variants {
			u, err := u.Parse(i.URI)
			if err != nil {
				continue
			}
			i.URI = u.String()
		}
		sort.SliceStable(playlist.Variants, func(i, j int) bool {
			return playlist.Variants[i].Bandwidth > playlist.Variants[j].Bandwidth
		})
		if len(playlist.Variants) != 0 {
			if debug {
				log.Print("Parse from master playlist:")
				fmt.Println(playlist)
			}

			u, err = u.Parse(playlist.Variants[0].URI)
			if err != nil {
				return nil, nil, err
			}
			return FetchM3U8MediaPlaylist(u, debug)
		} else {
			return nil, nil, fmt.Errorf("empty master playlist")
		}
	}

	return nil, nil, fmt.Errorf("unknown playlist type")
}

func fetchM3U8(url string) (r io.Reader, u *url.URL, err error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var id network.RequestID
	done := make(chan struct{})
	chromedp.ListenTarget(ctx, func(v interface{}) {
		switch ev := v.(type) {
		case *network.EventRequestWillBeSent:
			if strings.Contains(ev.Request.URL, ".m3u8") && id == "" {
				u, _ = urlParse(ev.Request.URL)
				id = ev.RequestID
			}
		case *network.EventLoadingFinished:
			if ev.RequestID == id {
				close(done)
			}
		case *fetch.EventRequestPaused:
			go func() {
				c := chromedp.FromContext(ctx)
				ctx := cdp.WithExecutor(ctx, c.Target)

				if ev.ResourceType == network.ResourceTypeImage ||
					ev.ResourceType == network.ResourceTypeStylesheet ||
					ev.ResourceType == network.ResourceTypeMedia {
					fetch.FailRequest(ev.RequestID, network.ErrorReasonBlockedByClient).Do(ctx)
				} else {
					fetch.ContinueRequest(ev.RequestID).Do(ctx)
				}
			}()
		}
	})

	if err = chromedp.Run(ctx, fetch.Enable(), chromedp.Navigate(url)); err != nil {
		return
	}

	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	case <-done:
	}

	var body []byte
	if err = chromedp.Run(
		ctx,
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			body, err = network.GetResponseBody(id).Do(ctx)
			return
		}),
	); err != nil {
		return nil, nil, err
	}

	return bytes.NewReader(body), u, nil
}

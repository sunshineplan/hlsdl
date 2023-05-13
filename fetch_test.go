package hlsdl

import (
	"net/url"
	"testing"
)

func TestFetch(t *testing.T) {
	m3u8URL := "https://s3.amazonaws.com/qa.jwplayer.com/hlsjs/muxed-fmp4/hls.m3u8"
	m3u8 := `#EXTM3U
#EXT-X-VERSION:7
#EXT-X-MAP:URI="init.mp4"
#EXT-X-MEDIA-SEQUENCE:11
#EXT-X-TARGETDURATION:3
#EXTINF:1.001,
https://s3.amazonaws.com/qa.jwplayer.com/hlsjs/muxed-fmp4/hls11.m4s
#EXTINF:2.002,
https://s3.amazonaws.com/qa.jwplayer.com/hlsjs/muxed-fmp4/hls12.m4s
#EXTINF:2.002,
https://s3.amazonaws.com/qa.jwplayer.com/hlsjs/muxed-fmp4/hls13.m4s
#EXTINF:2.002,
https://s3.amazonaws.com/qa.jwplayer.com/hlsjs/muxed-fmp4/hls14.m4s
#EXTINF:2.002,
https://s3.amazonaws.com/qa.jwplayer.com/hlsjs/muxed-fmp4/hls15.m4s
#EXTINF:2.002,
https://s3.amazonaws.com/qa.jwplayer.com/hlsjs/muxed-fmp4/hls16.m4s
#EXTINF:2.002,
https://s3.amazonaws.com/qa.jwplayer.com/hlsjs/muxed-fmp4/hls17.m4s
#EXT-X-ENDLIST
`
	u, _ := url.Parse(m3u8URL)
	u, playlist, err := FetchM3U8MediaPlaylist(u, true)
	if err != nil {
		t.Fatal(err)
	}
	if u.String() != m3u8URL {
		t.Errorf("expected %s; got %s", m3u8URL, u)
	}
	if playlist.String() != m3u8 {
		t.Errorf("expected %s; got %s", m3u8, playlist)
	}
}

func TestM3U8(t *testing.T) {
	m3u8URL := "https://s3.amazonaws.com/qa.jwplayer.com/hlsjs/muxed-fmp4/hls.m3u8"
	m3u8Segments := []string{
		"https://s3.amazonaws.com/qa.jwplayer.com/hlsjs/muxed-fmp4/hls11.m4s",
		"https://s3.amazonaws.com/qa.jwplayer.com/hlsjs/muxed-fmp4/hls12.m4s",
		"https://s3.amazonaws.com/qa.jwplayer.com/hlsjs/muxed-fmp4/hls13.m4s",
		"https://s3.amazonaws.com/qa.jwplayer.com/hlsjs/muxed-fmp4/hls14.m4s",
		"https://s3.amazonaws.com/qa.jwplayer.com/hlsjs/muxed-fmp4/hls15.m4s",
		"https://s3.amazonaws.com/qa.jwplayer.com/hlsjs/muxed-fmp4/hls16.m4s",
		"https://s3.amazonaws.com/qa.jwplayer.com/hlsjs/muxed-fmp4/hls17.m4s",
	}
	u, _ := url.Parse(m3u8URL)
	u, playlist, err := FetchM3U8MediaPlaylist(u, true)
	if err != nil {
		t.Fatal(err)
	}
	segments := playlist.Segments

	if u.String() != m3u8URL {
		t.Errorf("expected %s; got %s", m3u8URL, u)
	}
	for i, s := range segments {
		if s.URI != m3u8Segments[i] {
			t.Errorf("#%d expected %s; got %s", i, m3u8Segments[i], s.URI)
		}
	}
}

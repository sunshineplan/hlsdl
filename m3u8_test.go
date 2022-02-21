package hlsdl

import (
	"testing"
)

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
	u, _ := urlParse(m3u8URL)
	u, segments, err := getSegments(u)
	if err != nil {
		t.Fatal(err)
	}
	if u.String() != m3u8URL {
		t.Errorf("expected %s; got %s", m3u8URL, u)
	}
	for i, s := range segments {
		if s.URI != m3u8Segments[i] {
			t.Errorf("#%d expected %s; got %s", i, m3u8Segments[i], s.URI)
		}
	}
}

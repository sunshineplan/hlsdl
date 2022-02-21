package hlsdl

import (
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
hls11.m4s
#EXTINF:2.002,
hls12.m4s
#EXTINF:2.002,
hls13.m4s
#EXTINF:2.002,
hls14.m4s
#EXTINF:2.002,
hls15.m4s
#EXTINF:2.002,
hls16.m4s
#EXTINF:2.002,
hls17.m4s
#EXT-X-ENDLIST
`
	u, _ := urlParse(m3u8URL)
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

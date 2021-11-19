package hlsdl

import (
	"testing"
)

func TestDownload(t *testing.T) {
	task := NewTask("https://s3.amazonaws.com/qa.jwplayer.com/hlsjs/muxed-fmp4/hls.m3u8")
	if err := task.Run("output", ""); err != nil {
		t.Fatal(err)
	}
}

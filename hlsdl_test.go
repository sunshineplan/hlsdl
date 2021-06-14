package hlsdl

import (
	"testing"
)

func TestDownload(t *testing.T) {
	task := NewTask("https://www.baobuzz.com/m3u8/test.m3u8")
	if err := task.Run(""); err != nil {
		t.Fatal(err)
	}
}

# HLS downloader
This is a library to support downloading video according online m3u8 format file. All segments will be downloaded into a temporary folder then be joined into a single file.

Default output directory is `output`


## Features:
* Concurrent download segments with multiple http connections
* Decrypt HLS encoded segments
* Auto retry download
* Display downloading progress bar


## How to use this library
```
package main

import "github.com/sunshineplan/hlsdl"

func main() {
    m3u8 := "https://s3.amazonaws.com/qa.jwplayer.com/hlsjs/muxed-fmp4/hls.m3u8"
	task := hlsdl.NewTask(m3u8).SetWorkers(10)
	if err := task.Run("video.ts"); err != nil {
		panic(err)
	}
}
```


## See also

  * [grafov/m3u8](https://github.com/grafov/m3u8)
  * [canhlinh/hlsdl](https://github.com/canhlinh/hlsdl) (provide command line tool)

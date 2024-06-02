# HLS downloader
This is a library to support downloading video according online m3u8 format file. All segments will be downloaded into a temporary folder then be joined into a single file.


## Features:
* Concurrent download segments with multiple http connections
* Decrypt HLS encoded segments
* Auto retry download
* Display downloading progress bar
* Command Line tool


## How to use this library
```go
package main

import "github.com/sunshineplan/hlsdl"

func main() {
	m3u8 := "https://s3.amazonaws.com/qa.jwplayer.com/hlsjs/muxed-fmp4/hls.m3u8"
	task := hlsdl.NewTask(m3u8).SetWorkers(10)
	if err := task.Run("output", "video.ts"); err != nil {
		panic(err)
	}
}
```


## How to use command line tool
```
Usage: hlsdl [options...] <url>
  --path <string>
    	Output Path
  --output <string>
    	Output File Name (default "output.ts")
  --workers <number>
    	Workers
  --ua <string>
    	User Agent String
```
```
./hlsdl -workers 10 -path output -output video.ts https://s3.amazonaws.com/qa.jwplayer.com/hlsjs/muxed-fmp4/hls.m3u8
```
Get prebuild binary: https://github.com/sunshineplan/hlsdl/releases

## See also

  * [grafov/m3u8](https://github.com/grafov/m3u8)
  * [canhlinh/hlsdl](https://github.com/canhlinh/hlsdl) (provide command line tool)

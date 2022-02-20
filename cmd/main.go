package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/sunshineplan/hlsdl"
)

var url string

var (
	path      = flag.String("path", "", "Output Path")
	output    = flag.String("output", "output.ts", "Output File Name")
	workers   = flag.Int("workers", 0, "Workers")
	userAgent = flag.String("ua", "", "User Agent String")
)

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), `Usage: %s [options...] <url>
  --path <string>
    	Output Path
  --output <string>
    	Output File Name (default "output.ts")
  --workers <number>
    	Workers
  --ua <string>
    	User Agent String
`, os.Args[0])
}

func main() {
	defer func() {
		fmt.Println("Press enter key to exit . . .")
		fmt.Scanln()
	}()

	flag.Usage = usage
	flag.Parse()
	for len(flag.Args()) != 0 {
		url = flag.Args()[0]
		os.Args = append(os.Args[:1], flag.Args()[1:]...)
		flag.Parse()
	}

	if url == "" {
		fmt.Print("Please input m3u8 url: ")
		fmt.Scanln(&url)
	}
	if url == "" {
		log.Print("No m3u8 url provided.")
		return
	}

	if *userAgent != "" {
		hlsdl.SetAgent(*userAgent)
	}

	task := hlsdl.NewTask(url).SetWorkers(*workers)
	if err := task.Run(*path, *output); err != nil {
		log.Print(err)
	}
}

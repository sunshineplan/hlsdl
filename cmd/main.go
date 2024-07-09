package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/sunshineplan/hlsdl"
)

var (
	path    = flag.String("path", "", "Output Path")
	output  = flag.String("output", "output.ts", "Output File Name")
	workers = flag.Int64("workers", 0, "Workers")
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

	var input string
	for len(flag.Args()) != 0 {
		input = flag.Args()[0]
		os.Args = append(os.Args[:1], flag.Args()[1:]...)
		flag.Parse()
	}

	if input == "" {
		fmt.Print("Please input m3u8 url or file path: ")
		fmt.Scanln(&input)
	}
	if input == "" {
		log.Print("No m3u8 provided.")
		return
	}

	if err := hlsdl.NewTask(input).SetWorkers(*workers).Run(*path, *output); err != nil {
		log.Print(err)
	}
}

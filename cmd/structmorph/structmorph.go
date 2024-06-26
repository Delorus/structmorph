package main

import (
	"flag"
	"log"
	"log/slog"
	"os"
	"structmorph"
)

var (
	from = flag.String("src", "", "Source struct name")
	to   = flag.String("dst", "", "Destination struct name")
	root = flag.String("root", "", "Root directory")
)

func main() {
	parseArgs()
	slog.Info("Parsed arguments", "src", from, "dst", to)

	var opts []structmorph.GenerationConfigOption
	if *root != "" {
		opts = append(opts, structmorph.WithProjectRoot(*root))
	}

	if err := structmorph.Generate(*from, *to, opts...); err != nil {
		log.Fatalf("Error generating code: %v", err)
	}
}

func parseArgs() {
	flag.Parse()

	if *from == "" || *to == "" {
		slog.Error("Usage: structmorph --from=domain.Person --to=main.PersonDTO")
		os.Exit(1)
	}
}

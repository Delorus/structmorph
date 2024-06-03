package main

import (
	"log"
	"log/slog"
	"os"
	"strings"
	"structmorph"
)

func main() {
	args := parseArgs()
	slog.Info("Parsed arguments", "from", args.from, "to", args.to)

	if err := structmorph.Generate(args.from, args.to); err != nil {
		log.Fatalf("Error generating code: %v", err)
	}
}

type Args struct {
	from string
	to   string
}

func parseArgs() Args {
	if len(os.Args) < 2 {
		slog.Error("Usage: structmorph --from=domain.Person --to=main.PersonDTO")
		os.Exit(1)
	}

	var args Args
	for _, arg := range os.Args[1:] {
		switch {
		case strings.HasPrefix(arg, "--from="):
			args.from = arg[len("--from="):]
		case strings.HasPrefix(arg, "--to="):
			args.to = arg[len("--to="):]
		}
	}

	if args.from == "" || args.to == "" {
		slog.Error("Usage: structmorph --from=domain.Person --to=main.PersonDTO")
		os.Exit(1)
	}

	return args
}

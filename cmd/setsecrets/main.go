package main

import (
	"context"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/oakcask/setsecrets"
	"github.com/oakcask/setsecrets/exec"
	"github.com/oakcask/setsecrets/gcp"
)

type parserState int

const initial = parserState(0)
const readingCommandLine = parserState(2)

func main() {
	provider := ""
	state := initial
	commandLine := []string{}
	mapping := map[string]string{}
	var e error

parseCommandLine:
	for idx, arg := range os.Args {
		if idx == 0 {
			continue
		}
		switch state {
		case initial:
			if strings.HasPrefix(arg, "--") {
				name := arg[2:]
				if name == "" {
					state = readingCommandLine
					break
				}
				provider = name
				break
			}
			eqlIndex := strings.Index(arg, "=")

			if eqlIndex >= len(arg) {
				log.Fatalf("key not found while expecting env=key pair in %v", arg)
			}
			if eqlIndex > 0 {
				environ := arg[:eqlIndex-1]
				key := arg[eqlIndex+1:]
				mapping[key] = environ
				break
			}
			if eqlIndex == 0 {
				log.Fatalf("env not found while expecting env=key pair in %v", arg)
			}
			state = readingCommandLine
			fallthrough
		case readingCommandLine:
			commandLine = os.Args[idx:]
			break parseCommandLine
		}
	}
	if provider == "" {
		log.Fatalf("provider should be specified")
	}
	if len(commandLine) < 1 {
		log.Fatalf("no command specified")
	}

	var providerImpl setsecrets.Provider
	ctx := context.Background()
	switch provider {
	case "gcp":
		providerImpl, e = gcp.New(ctx)
		if e != nil {
			log.Fatal(e)
		}

	default:
		log.Fatalf("unsupported provider: %v", provider)
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	e = setsecrets.SetSecrets(ctx, providerImpl, mapping)
	if e != nil {
		log.Fatal(e)
	}
	e = exec.Exec(commandLine[0], commandLine, os.Environ())
	if e != nil {
		log.Fatal(e)
	}
}

package cli

import (
	"flag"
	"fmt"
	"io"
	"strings"
)

// ErrHelp indicates that command-line help was requested.
var ErrHelp = flag.ErrHelp

// Usage returns the command-line usage text.
func Usage() string {
	return `Usage:
    fish-audio-cli [options]

Options:
    --config PATH   JSON configuration file
                    default: config/config.json

    --text TEXT     text to synthesize
                    standard input is used when omitted

    --format FORMAT output format: wav, mp3, opus or ogg

    --output PATH   destination audio file

    --help          show this help
`
}

// Options contains command-line arguments accepted by fish-audio-cli.
type Options struct {
	ConfigPath string
	Text       string
	OutputPath string
	Format     string
}

// ParseOptions parses and validates command-line arguments.
func ParseOptions(args []string) (Options, error) {
	flagSet := flag.NewFlagSet("fish-audio-cli", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	var options Options
	options.ConfigPath = "config/config.json"

	flagSet.StringVar(&options.ConfigPath, "config", options.ConfigPath, "path to the JSON configuration file")
	flagSet.StringVar(&options.Text, "text", "", "text to process; stdin is used when omitted")
	flagSet.StringVar(&options.OutputPath, "output", "", "path to the output audio file")
	flagSet.StringVar(&options.Format, "format", "", "audio format: wav, mp3, opus or ogg")

	if err := flagSet.Parse(args); err != nil {
		return Options{}, fmt.Errorf("parse arguments: %w", err)
	}

	if positional := flagSet.Args(); len(positional) > 0 {
		return Options{}, fmt.Errorf(
			"unexpected positional arguments: %q",
			positional,
		)
	}

	if options.OutputPath == "" {
		return Options{}, fmt.Errorf("--output is required")
	}

	options.Format = strings.ToLower(options.Format)

	switch options.Format {
	case "wav", "mp3", "opus":
	case "ogg":
		options.Format = "opus"
	default:
		return Options{}, fmt.Errorf(
			"unsupported format %q: expected wav, mp3, opus or ogg",
			options.Format,
		)
	}

	return options, nil
}

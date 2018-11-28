package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/cosiner/flag"
)

type config struct {
	Concurrency          int    `names:"--concurrency, -c" usage:"Amount of routines to spawn at the same time for checksum verification"`
	Complete             bool   `names:"--complete" usage:"Completion for shell"`
	DataDir              string `names:"--datadir, -d" usage:"Data directory path to run new files scan"`
	DbPath               string `names:"--database, -D" usage:"Database file path (required)"`
	GenerateChecksumOnly bool   `names:"--generate-checksums-only" usage:"Skip verification step and add new files only"`
	Pattern              string `names:"--pattern, -p" usage:"Pattern to match filenames which checking for new files"`
	SkipFailed           bool   `names:"--skip-failed, --sf" usage:"Skip FAIL verification results from output"`
	SkipMissed           bool   `names:"--skip-missed, --sm" usage:"Skip MISS verification results from output"`
	SkipOk               bool   `names:"--skip-ok, --so" usage:"Skip OK verification results from output"`
	Progressbar          bool   `names:"--progressbar" usage:"Show progress bar instead of printing handled files"`
	Version              bool   `names:"--version, -V" usage:"Print application and Golang versions"`
}

func newConfig() *config {
	var c config

	set := flag.NewFlagSet(flag.Flag{})
	set.ParseStruct(&c, os.Args...)

	if c.DbPath == "" {
		set.Help(true)
		os.Exit(1)
	}

	if c.Concurrency == 0 {
		c.Concurrency = runtime.NumCPU()
	}

	return &c
}

func (c *config) Metadata() map[string]flag.Flag {
	var (
		usage = `Utility to verify files consistency with length, SHA1 and SHA256`

		version = fmt.Sprintf(`
			version: %s
			built with: %s
		`, Version, runtime.Version())

		desc = `
		checksum creates database (actually just a JSON file) to store file length, SHA1, SHA256 
		to verify file consistency and report if something goes wrong.
		`
	)
	return map[string]flag.Flag{
		"": {
			Usage:   usage,
			Version: version,
			Desc:    desc,
		},
		"--concurrency": {
			Desc: fmt.Sprintf("Default value is %d for your system", runtime.NumCPU()),
		},
		"--pattern": {
			Desc: fmt.Sprintf("Default is `%s`", rawFilePattern),
		},
	}
}
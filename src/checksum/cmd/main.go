package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fatih/color"

	"checksum/database"
)

var wg sync.WaitGroup

// Version - variable to store current commit,tag,whatever
var Version = "No version specified(probably trunk build)"

var db *database.Database

var filePattern *regexp.Regexp

var cntAdded uint64
var cntFailed uint64
var cntMissed uint64
var cntPassed uint64

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage: %s [OPTION]...\n", os.Args[0])
		fmt.Printf("OPTIONS:\n")
		fmt.Printf("  -concurrency <int>\n")
		fmt.Printf("    Amount of routines to spawn at the same time for checksum verification(%v by default for your system)\n", runtime.NumCPU())
		fmt.Printf("  -database <string>\n")
		fmt.Printf("    Specify database path\n")
		fmt.Printf("  -datadir <string>\n")
		fmt.Printf("    Specify data directory\n")
		fmt.Printf("  -pattern <string>\n")
		fmt.Printf("    Pattern to match files in filewalk mode(default is `.(ar2|arw|cr2|crw|nef)$`)\n")
		fmt.Printf("  -version\n")
		fmt.Printf("    Print checksum version\n\n")
		fmt.Printf("Examples:\n")
		fmt.Printf("  %s -database /tmp/db.json -datadir /Volumes/Storage/Photos\n", os.Args[0])
	}

	concurrency := flag.Int("concurrency", runtime.NumCPU(), "")
	version := flag.Bool("version", false, "")
	datadir := flag.String("datadir", "", "")
	dbPath := flag.String("database", "", "")
	pattern := flag.String("pattern", ".(ar2|arw|cr2|crw|nef)$", "")

	flag.Parse()

	if (*datadir == "" || *dbPath == "") && !*version {
		flag.Usage()
		os.Exit(1)
	}

	if *version == true {
		printVersion()
		return
	}

	var err error

	db = database.NewDatabase(*dbPath)
	filePattern, err = regexp.Compile(*pattern)
	if err != nil {
		log.Fatalf("Error compiling pattern: %s", err)
	}
	sem := make(chan bool, *concurrency)

	for file, obj := range db.MapObjects() {
		sem <- true
		wg.Add(1)
		go func() {
			if _, err := os.Stat(file); os.IsNotExist(err) {
				atomic.AddUint64(&cntMissed, 1)
				return
			}

			res := verify(file, obj.Sha256)

			if res {
				fmt.Printf("%s %s\n", color.GreenString("[ OK ]"), file)
				atomic.AddUint64(&cntPassed, 1)
			} else {
				fmt.Printf("%s %s\n", color.RedString("[FAIL]"), file)
				atomic.AddUint64(&cntFailed, 1)
			}

			defer func() {
				<-sem
			}()
		}()
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
	wg.Wait()

	log.Printf("First pass done. Starting filewalk on path: %s", *datadir)

	err = filepath.Walk(*datadir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if isApplicable(path) {
			db.WriteOne(path, database.Data{
				Sha256:   sha256file(path),
				Modified: time.Now(),
			})
			fmt.Printf("%s %s\n", color.YellowString("[CALCULATED]"), path)
			atomic.AddUint64(&cntAdded, 1)
		}
		return nil
	})
	if err != nil {
		panic(fmt.Sprintf("Error walking through files: %s", err))
	}

	err = db.Commit()
	if err != nil {
		panic(fmt.Sprintf("Error commiting the data: %s", err))
	}

	fmt.Printf("Job is done:\n")
	fmt.Printf("  Added: %d\n", cntAdded)
	fmt.Printf("  Missed: %d\n", cntMissed)
	fmt.Printf("  Failed: %d\n", cntFailed)
	fmt.Printf("  Passed: %d\n", cntPassed)
}

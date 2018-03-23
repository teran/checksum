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
	"gopkg.in/cheggaaa/pb.v1"

	"checksum/database"
)

var wg sync.WaitGroup

// Version - variable to store current commit,tag,whatever
var Version = "No version specified(probably trunk build)"

var db *database.Database

var filePattern *regexp.Regexp

var (
	cntAdded  uint64
	cntFailed uint64
	cntMissed uint64
	cntPassed uint64
)

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
		fmt.Printf("    Pattern to match filenames which checking for new files(default is `.(3fr|ari|arw|bay|crw|cr2|cap|data|dcs|dcr|drf|eip|erf|fff|gpr|iiq|k25|kdc|mdc|mef|mos|mrw|nef|nrw|obm|orf|pef|ptx|pxn|r3d|raf|raw|rwl|rw2|rwz|sr2|srf|srw|x3f)$`)\n")
		fmt.Printf("  -progressbar\n")
		fmt.Printf("    Show progress bar instead of printing handled files(the same as `-skipfailed`, `-skipmissed`, `-skipok` but with pretty progress bar)\n")
		fmt.Printf("  -skipfailed\n")
		fmt.Printf("    Skip FAIL verification results from output\n")
		fmt.Printf("  -skipmissed\n")
		fmt.Printf("    Skip MISS verification results from output\n")
		fmt.Printf("  -skipok\n")
		fmt.Printf("    Skip OK verification results from output\n")
		fmt.Printf("  -version\n")
		fmt.Printf("    Print application and Golang versions\n\n")
		fmt.Printf("Examples:\n")
		fmt.Printf("  %s -database /tmp/db.json -datadir /Volumes/Storage/Photos\n", os.Args[0])
	}

	concurrency := flag.Int("concurrency", runtime.NumCPU(), "")
	complete := flag.Bool("complete", false, "")
	version := flag.Bool("version", false, "")
	datadir := flag.String("datadir", "", "")
	dbPath := flag.String("database", "", "")
	pattern := flag.String("pattern", ".(3fr|ari|arw|bay|crw|cr2|cap|data|dcs|dcr|drf|eip|erf|fff|gpr|iiq|k25|kdc|mdc|mef|mos|mrw|nef|nrw|obm|orf|pef|ptx|pxn|r3d|raf|raw|rwl|rw2|rwz|sr2|srf|srw|x3f)$", "")
	skipfailed := flag.Bool("skipfailed", false, "")
	skipmissed := flag.Bool("skipmissed", false, "")
	skipok := flag.Bool("skipok", false, "")
	progressbar := flag.Bool("progressbar", false, "")

	flag.Parse()

	if *version == true {
		printVersion()
		return
	}

	if *complete == true {
		completeArgs(flag.Arg(1))
		return
	}

	if *datadir == "" || *dbPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *progressbar == true {
		*skipfailed = true
		*skipmissed = true
		*skipok = true
	}

	var err error

	db = database.NewDatabase(*dbPath)
	filePattern, err = regexp.Compile(*pattern)
	if err != nil {
		log.Fatalf("Error compiling pattern: %s", err)
	}
	sem := make(chan bool, *concurrency)
	var bar *pb.ProgressBar
	if *progressbar {
		bar = pb.New(db.Count())
		bar.ShowCounters = true
		bar.SetRefreshRate(time.Second)
		bar.ShowPercent = true
		bar.ShowBar = true
		bar.ShowTimeLeft = true
		bar.ShowSpeed = true
		bar.Start()
	}

	for file, obj := range db.MapObjects() {
		sem <- true
		wg.Add(1)
		go func(file string, obj database.Data) {
			if *progressbar {
				defer func() {
					bar.Increment()
				}()
			}
			defer func() {
				<-sem
			}()
			defer wg.Done()

			if _, err := os.Stat(file); os.IsNotExist(err) {
				if !*skipmissed {
					fmt.Printf("%s %s\n", color.RedString("[MISS]"), file)
				}
				atomic.AddUint64(&cntMissed, 1)
				return
			}

			res := verify(file, obj.Sha256)

			if res {
				if !*skipok {
					fmt.Printf("%s %s\n", color.GreenString("[ OK ]"), file)
				}
				atomic.AddUint64(&cntPassed, 1)
				return
			}
			if !*skipfailed {
				fmt.Printf("%s %s\n", color.RedString("[FAIL]"), file)
			}
			atomic.AddUint64(&cntFailed, 1)
		}(file, obj)
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
	wg.Wait()

	if *progressbar {
		bar.Finish()
	}

	fmt.Printf("%s Verification step complete. Checking for new files on %s\n", color.CyanString("[INFO]"), *datadir)

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

	fmt.Printf("------------\n")
	fmt.Printf("Job is done:\n")
	fmt.Printf("  Added: %d\n", cntAdded)
	fmt.Printf("  Missed: %d\n", cntMissed)
	fmt.Printf("  Failed: %d\n", cntFailed)
	fmt.Printf("  Passed: %d\n", cntPassed)
}

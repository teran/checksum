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

	"github.com/teran/checksum/database"
)

var (
	wg sync.WaitGroup

	// Version - variable to store current commit,tag,whatever
	Version = "No version specified(probably trunk build)"

	db             *database.Database
	filePattern    *regexp.Regexp
	rawFilePattern = ".(3fr|ari|arw|bay|crw|cr2|cr3|cap|data|dcs|dcr|drf|eip|erf|fff|gpr|iiq|k25|kdc|mdc|mef|mos|mrw|nef|nrw|obm|orf|pef|ptx|pxn|r3d|raf|raw|rwl|rw2|rwz|sr2|srf|srw|x3f)$"

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
		fmt.Printf("  -generate-checksum-only\n")
		fmt.Printf("    Skip step of file verification and only check for new files and generate checksums for them\n")
		fmt.Printf("  -pattern <string>\n")
		fmt.Printf("    Pattern to match filenames which checking for new files(default is `" + rawFilePattern + "`)\n")
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
	generateChecksumOnly := flag.Bool("generate-checksum-only", false, "")
	pattern := flag.String("pattern", rawFilePattern, "")
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
		*skipmissed = true
		*skipok = true
	}

	var err error

	db = database.NewDatabase(*dbPath)
	filePattern, err = regexp.Compile(*pattern)
	if err != nil {
		log.Fatalf("Error compiling pattern: %s", err)
	}

	if !*generateChecksumOnly {
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

				isChanged := false

				if obj.Length == 0 {
					obj.Length = flength(file)
					isChanged = true
				}

				if obj.SHA1 == "" {
					obj.SHA1 = sha1file(file)
					isChanged = true
				}

				if obj.SHA256 == "" {
					obj.SHA256 = sha256file(file)
					isChanged = true
				}

				res := verify(file, obj.Length, obj.SHA1, obj.SHA256)

				if isChanged {
					db.WriteOne(file, database.Data{
						Length:   obj.Length,
						SHA1:     obj.SHA1,
						SHA256:   obj.SHA256,
						Modified: time.Now().UTC(),
					})
				}

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
	}

	err = filepath.Walk(*datadir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if isApplicable(path) {
			db.WriteOne(path, database.Data{
				Length:   flength(path),
				SHA1:     sha1file(path),
				SHA256:   sha256file(path),
				Modified: time.Now().UTC(),
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

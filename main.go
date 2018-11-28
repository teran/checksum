package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
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
	cfg := newConfig()

	if cfg.Version == true {
		printVersion()
		return
	}

	if cfg.Complete == true {
		completeArgs(flag.Arg(1))
		return
	}

	if cfg.Progressbar == true {
		cfg.SkipOk = true
	}

	var err error

	db = database.NewDatabase(cfg.DbPath)
	filePattern, err = regexp.Compile(cfg.Pattern)
	if err != nil {
		log.Fatalf("Error compiling pattern: %s", err)
	}

	if !cfg.GenerateChecksumOnly {
		sem := make(chan bool, cfg.Concurrency)
		var bar *pb.ProgressBar
		if cfg.Progressbar {
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
				if cfg.Progressbar {
					defer func() {
						bar.Increment()
					}()
				}
				defer func() {
					<-sem
				}()
				defer wg.Done()

				if _, err := os.Stat(file); os.IsNotExist(err) {
					if !cfg.SkipMissed {
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
					if !cfg.SkipOk {
						fmt.Printf("%s %s\n", color.GreenString("[ OK ]"), file)
					}
					atomic.AddUint64(&cntPassed, 1)
					return
				}
				if !cfg.SkipFailed {
					fmt.Printf("%s %s\n", color.RedString("[FAIL]"), file)
				}
				atomic.AddUint64(&cntFailed, 1)
			}(file, obj)
		}

		for i := 0; i < cap(sem); i++ {
			sem <- true
		}
		wg.Wait()

		if cfg.Progressbar {
			bar.Finish()
		}

		fmt.Printf("%s Verification step complete\n", color.CyanString("[INFO]"))
	}

	if cfg.DataDir != "" {
		fmt.Printf("%s Checking for new files on %s", color.CyanString("[INFO]"), cfg.DataDir)

		err = filepath.Walk(cfg.DataDir, func(path string, info os.FileInfo, err error) error {
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
	} else {
		fmt.Printf("%s data directory is not specified. Skipping new files check\n", color.CyanString("[INFO]"))
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

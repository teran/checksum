package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fatih/color"
	"gopkg.in/cheggaaa/pb.v1"

	database "github.com/teran/checksum/database/legacy"
)

var (
	wg sync.WaitGroup

	version = "No version specified(probably trunk build)"
	commit  = "master"
	date    = "0000-00-00T00:00:00Z"

	db          *database.Database
	filePattern *regexp.Regexp

	cntAdded   uint64
	cntDeleted uint64
	cntFailed  uint64
	cntMissed  uint64
	cntPassed  uint64
)

func main() {
	cfg := newConfig()

	if cfg.Complete == true {
		completeArgs(flag.Arg(1))
		return
	}

	if cfg.Progressbar == true {
		cfg.SkipOk = true
	}

	var err error

	db, err = database.NewDatabase(cfg.DbPath)
	if err != nil {
		log.Fatalf("error opening database: %s", err)
	}

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

		objects := db.MapObjects()
		keys := make([]string, 0, len(objects))
		for k := range objects {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			sem <- true
			wg.Add(1)
			go func(file string, obj *database.DataObject) {
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

					if cfg.DeleteMissed {
						fmt.Printf("%s DeleteMissed requested: deleting file `%s` from database\n", color.BlueString("[NOTE]"), file)
						db.DeleteOne(file)
						atomic.AddUint64(&cntDeleted, 1)
					}

					atomic.AddUint64(&cntMissed, 1)
					return
				}

				isChanged := false

				if obj.Length == 0 {
					obj.Length = flength(file)
					isChanged = true
				}

				data, err := readFile(file)
				if err != nil {
					log.Fatalf("error reading data: %s", err)
				}

				if obj.SHA1 == "" {
					obj.SHA1, err = SHA1(bytes.NewReader(data))
					if err != nil {
						log.Fatalf("error calculating SHA1: %s", err)
					}

					isChanged = true
				}

				if obj.SHA256 == "" {
					obj.SHA256, err = SHA256(bytes.NewReader(data))
					if err != nil {
						log.Fatalf("error calculating SHA256: %s", err)
					}

					isChanged = true
				}

				res := verify(file, obj.Length, obj.SHA1, obj.SHA256)

				if isChanged {
					db.WriteOne(file, &database.DataObject{
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
			}(key, objects[key])
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
		fmt.Printf("%s Checking for new files on %s\n", color.CyanString("[INFO]"), cfg.DataDir)

		err = filepath.Walk(cfg.DataDir, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if isApplicable(path) {
				data, err := readFile(path)
				if err != nil {
					log.Fatalf("error reading file: %s", err)
				}

				sha1, err := SHA1(bytes.NewReader(data))
				if err != nil {
					log.Fatalf("error calculating SHA1: %s", err)
				}

				sha256, err := SHA256(bytes.NewReader(data))
				if err != nil {
					log.Fatalf("error calculating SHA256: %s", err)
				}

				db.WriteOne(path, &database.DataObject{
					Length:   flength(path),
					SHA1:     sha1,
					SHA256:   sha256,
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
	fmt.Printf("  Deleted: %d\n", cntDeleted)
	fmt.Printf("  Missed: %d\n", cntMissed)
	fmt.Printf("  Failed: %d\n", cntFailed)
	fmt.Printf("  Passed: %d\n", cntPassed)
}

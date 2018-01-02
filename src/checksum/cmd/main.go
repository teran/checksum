package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"checksum/database"
)

var wg sync.WaitGroup

// Version - variable to store current commit,tag,whatever
var Version = "No version specified(probably trunk build)"

var db *database.Database

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage: %s [OPTION]... [FILE]...\n", os.Args[0])
		fmt.Println("  -database")
		fmt.Println("    Specify database path")
		fmt.Println("  -datadir")
		fmt.Println("    Specify data directory")
		// fmt.Printf("Print or check SHA256 hashes for files\n")
		// fmt.Printf("  -check\n")
		// fmt.Printf("    read hashes from the FILEs and check them\n")
		// fmt.Printf("  -concurrency\n")
		// fmt.Printf("    Amount of routines to spawn at the same time(%v by default for your system)\n", runtime.NumCPU())
		// fmt.Printf("  -version\n")
		// fmt.Printf("    Print checksum version\n\n")
		// fmt.Printf("Examples:\n")
		// fmt.Printf("  %s file.iso\n", os.Args[0])
		// fmt.Printf("  %s file.jpg | tee /tmp/database.txt\n", os.Args[0])
		// fmt.Printf("  %s -check /tmp/database.txt\n", os.Args[0])
	}

	concurrency := flag.Int("concurrency", runtime.NumCPU(), "")
	version := flag.Bool("version", false, "")
	datadir := flag.String("datadir", "", "")
	dbPath := flag.String("database", "", "")

	flag.Parse()

	db = database.NewDatabase(*dbPath)

	// if flag.NArg() < 1 && !*version {
	//	flag.Usage()
	//	os.Exit(1)
	// }

	if *version == true {
		printVersion()
		return
	}

	sem := make(chan bool, *concurrency)
	for _, file := range db.ListPaths() {
		sem <- true
		wg.Add(1)
		go func() {
			obj, ok := db.ReadOne(file)
			if !ok {
				log.Println("Error retrieving entry for file %s", file)
				return
			}
			res := verify(file, obj)

			if res {
				fmt.Printf("[ OK ] %s\n", file)
			} else {
				fmt.Printf("[FAIL] %s\n", file)
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

	err := filepath.Walk(*datadir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".cr2" || filepath.Ext(path) == ".nef" || filepath.Ext(path) == ".html" {
			db.WriteOne(path, database.Data{
				Sha256:   sha256file(path),
				Modified: time.Now(),
			})
			fmt.Printf("[CALCULATED] %s\n", path)
		}
		return nil
	})
	if err != nil {
		panic(fmt.Sprintf("Error walking through files: %s", err))
	}

	db.Commit()
}

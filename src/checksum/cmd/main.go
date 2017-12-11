package main

import (
  "crypto/sha256"
  "encoding/hex"
  "flag"
  "fmt"
  "io"
  "log"
  "os"
  "runtime"
  "sync"
)

var wg sync.WaitGroup

// Version - variable to store current commit,tag,whatever
var Version = "No version specified(probably trunk build)"

func calculate(filename string) {
  fp, err := os.Open(filename)
  if err != nil {
    log.Fatal(err)
  }
  defer fp.Close()

  sha256 := sha256.New()
  if _, err := io.Copy(sha256, fp); err != nil {
    log.Fatal(err)
  }

  fmt.Printf(
    "%s  %s\n", hex.EncodeToString(sha256.Sum(nil)), filename)
  wg.Done()
}

func verify() {

}

func printVersion() {
  fmt.Println(Version)
}

func main() {
  flag.Usage = func() {
		fmt.Printf("Usage: %s [OPTION]... [FILE]...\n", os.Args[0])
		fmt.Printf("Print or check SHA1, SHA256 and SHA512 hashes for files\n")
		fmt.Printf("  -check\n")
		fmt.Printf("    read hashes from the FILEs and check them\n")
		fmt.Printf("  -concurrency\n")
		fmt.Printf("    Amount of routines to spawn at the same time(%v by default for your system)\n", runtime.NumCPU())
		fmt.Printf("  -version\n")
		fmt.Printf("    Print checksum version\n\n")
		fmt.Printf("Examples:\n")
		fmt.Printf("  %s file.iso\n", os.Args[0])
		fmt.Printf("  %s file.jpg | tee /tmp/database.txt\n", os.Args[0])
		fmt.Printf("  %s -check /tmp/database.txt\n", os.Args[0])
	}

  check := flag.Bool("check", false, "")
  concurrency := flag.Int("concurrency", runtime.NumCPU(), "")
  version := flag.Bool("version", false, "")

  flag.Parse()

  if flag.NArg() < 1 && !*version {
    flag.Usage()
    os.Exit(1)
  }

  if *check == true {
    verify()
  } else if *version == true {
    printVersion()
  } else {
    sem := make(chan bool, *concurrency)
    for file := range flag.Args() {
      sem <- true
      wg.Add(1)
      go func() {
        calculate(flag.Arg(file))
        defer func() {
          <-sem
        }()
      }()
    }

    for i := 0; i < cap(sem); i++ {
  		sem <- true
  	}
  	wg.Wait()
  }
}

package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func completeArgs(word string) {
	fmt.Println(strings.Join([]string{
		"-concurrency",
		"-database",
		"-datadir",
		"-pattern",
		"-progressbar",
		"-skipfailed",
		"-skipmissed",
		"-skipok",
		"-version",
	}, " "))
}

func sha256file(filename string) string {
	fp, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()

	sha256 := sha256.New()
	if _, err := io.Copy(sha256, fp); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", sha256.Sum(nil))
}

func verify(path string, sha256 string) bool {
	s := sha256file(path)

	return s == sha256
}

func printVersion() {
	fmt.Printf("checksum version: %s\n", Version)
	fmt.Printf("Built wih Go version: %s\n", runtime.Version())
}

func isApplicable(path string) bool {
	_, ok := db.ReadOne(path)
	if filePattern.MatchString(filepath.Ext(path)) && !ok {
		return true
	}
	return false
}

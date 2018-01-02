package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"checksum/database"
)

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

func verify(path string, obj database.Data) bool {
	defer wg.Done()

	s := sha256file(path)
	o := obj.Sha256

	return s == o
}

func printVersion() {
	fmt.Println(Version)
}

func isApplicable(path string) bool {
	_, ok := db.ReadOne(path)
	if filePattern.MatchString(filepath.Ext(path)) && !ok {
		return true
	}
	return false
}

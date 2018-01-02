package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"

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

func calculate(filename string) {
	_ = sha256file(filename)

	wg.Done()
}

func verify(path string, obj database.Data) bool {
	s := sha256file(path)
	o := obj.Sha256

	return s == o
}

func printVersion() {
	fmt.Println(Version)
}

func calculateByPath(path string) {
}

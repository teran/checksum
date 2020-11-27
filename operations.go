package main

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func completeArgs(word string) {
	fmt.Println(strings.Join([]string{
		"--concurrency",
		"--database",
		"--datadir",
		"--pattern",
		"--progressbar",
		"--skipfailed",
		"--skipmissed",
		"--skipok",
		"--version",
	}, " "))
}

func readFile(fn string) ([]byte, error) {
	fp, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	return ioutil.ReadAll(fp)
}

// SHA256 ...
func SHA256(rd io.Reader) (string, error) {
	h := sha256.New()
	_, err := io.Copy(h, rd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// SHA1 ...
func SHA1(rd io.Reader) (string, error) {
	h := sha1.New()
	_, err := io.Copy(h, rd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func flength(filename string) int64 {
	stat, err := os.Stat(filename)
	if err != nil {
		log.Fatal(err)
	}

	return stat.Size()
}

func verify(path string, length int64, sha1, sha256 string) bool {
	data, err := readFile(path)
	if err != nil {
		log.Printf("error reading file: %s", err)
		return false
	}

	actSHA1, err := SHA1(bytes.NewReader(data))
	if err != nil {
		log.Printf("error calculating SHA1: %s", err)
		return false
	}

	actSHA256, err := SHA256(bytes.NewReader(data))
	if err != nil {
		log.Printf("error calculating SHA256: %s", err)
		return false
	}

	return flength(path) == length && actSHA1 == sha1 && actSHA256 == sha256
}

func printVersion() {
	fmt.Printf("checksum version: %s\n", version)
	fmt.Printf("Built wih Go version: %s\n", runtime.Version())
}

func isApplicable(path string) bool {
	_, ok := db.ReadOne(path)
	if filePattern.MatchString(filepath.Ext(path)) && !ok {
		return true
	}
	return false
}

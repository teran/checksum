package main

import (
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/teran/checksum/utils/concurrent"
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

func flength(filename string) int64 {
	stat, err := os.Stat(filename)
	if err != nil {
		log.Fatal(err)
	}

	return stat.Size()
}

func generateActualChecksum(filename string) (sha1sum string, sha256sum string, err error) {
	fi, err := os.Stat(filename)
	if err != nil {
		return "", "", err
	}

	fp, err := os.Open(filename)
	if err != nil {
		return "", "", err
	}
	defer fp.Close()

	sha1hasher := sha1.New()
	sha256hasher := sha256.New()

	w, err := concurrent.NewConcurrentMultiWriter(context.TODO(), sha1hasher, sha256hasher)
	if err != nil {
		return "", "", err
	}

	n, err := io.Copy(w, fp)
	if err != nil {
		return "", "", err
	}

	if n != fi.Size() {
		return "", "", io.ErrShortWrite
	}

	return hex.EncodeToString(sha1hasher.Sum(nil)), hex.EncodeToString(sha256hasher.Sum(nil)), nil
}

func verify(path string, length int64, sha1, sha256 string) bool {
	actSHA1, actSHA256, err := generateActualChecksum(path)
	if err != nil {
		return false
	}

	return flength(path) == length && actSHA1 == sha1 && actSHA256 == sha256
}

func printVersion() {
	fmt.Printf("checksum version: %s\n", appVersion)
	fmt.Printf("Built wih Go version: %s\n", runtime.Version())
}

func isApplicable(path string) bool {
	_, ok := db.ReadOne(path)
	if filePattern.MatchString(filepath.Ext(path)) && !ok {
		return true
	}
	return false
}

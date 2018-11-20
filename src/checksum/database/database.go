package database

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
)

// Data is a file object in JSON database
type Data struct {
	Length   int64     `json:"length"`
	SHA1     string    `json:"sha1"`
	SHA256   string    `json:"sha256"`
	Modified time.Time `json:"modified"`
}

// Schema is a container for file objects
type Schema struct {
	Data     map[string]Data `json:"data"`
	Modified time.Time       `json:"modified"`
}

// Database object struct
type Database struct {
	Path      string
	Schema    Schema
	IsChanged bool
}

var mutex = &sync.Mutex{}

// NewDatabase creates new Database object
func NewDatabase(path string) *Database {
	fmt.Printf("%s Opening database at %s\n", color.CyanString("[INFO]"), path)
	database := Database{
		Path:      path,
		IsChanged: false,
	}

	if database.Schema.Modified.IsZero() {
		database.IsChanged = true
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("%s Database at %s doesn't exist. Creating a new one\n", color.YellowString("[WARN]"), path)
		js, err := json.Marshal(Schema{
			Data: make(map[string]Data),
		})
		if err != nil {
			panic(fmt.Sprintf("Error marshaling initial JSON: %s", err))
		}

		err = ioutil.WriteFile(path, js, 0644)
		if err != nil {
			panic(fmt.Sprintf("Error creating schema: %s", err))
		}
	}

	fp, err := os.Open(path)
	if err != nil {
		panic(fmt.Sprintf("Error opening file: %s", err))
	}
	defer fp.Close()

	decoder := json.NewDecoder(fp)
	err = decoder.Decode(&database.Schema)
	if err != nil {
		panic(fmt.Sprintf("Error decoding JSON data: %s", err))
	}

	return &database
}

// ReadOne reads Data entry for specific file
func (d *Database) ReadOne(path string) (Data, bool) {
	mutex.Lock()
	defer mutex.Unlock()
	data, ok := d.Schema.Data[path]
	return data, ok
}

// WriteOne writes Data entry for specific file
func (d *Database) WriteOne(path string, data Data) (Data, bool) {
	mutex.Lock()
	defer mutex.Unlock()
	d.Schema.Data[path] = data
	d.IsChanged = true

	_, ok := d.Schema.Data[path]
	return data, ok
}

// Count returns count of elements in database
func (d *Database) Count() int {
	mutex.Lock()
	defer mutex.Unlock()
	return len(d.Schema.Data)
}

// ListPaths returns list of files present in database
func (d *Database) ListPaths() []string {
	mutex.Lock()
	defer mutex.Unlock()
	keys := make([]string, 0, d.Count())
	for k := range d.Schema.Data {
		keys = append(keys, k)
	}

	return keys
}

// MapObjects Returns objects map
func (d *Database) MapObjects() map[string]Data {
	mutex.Lock()
	defer mutex.Unlock()
	return d.Schema.Data
}

// Commit writes all the in-mem changes to disk
func (d *Database) Commit() error {
	mutex.Lock()
	defer mutex.Unlock()
	if !d.IsChanged {
		return nil
	}

	d.Schema.Modified = time.Now().UTC()

	js, err := json.Marshal(d.Schema)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(d.Path, js, 0644)
	return err
}

package database

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

// Data is a file object in JSON database
type Data struct {
	Sha256   string    `json:"sha256"`
	Modified time.Time `json:"modified"`
}

// Schema is a container for file objects
type Schema struct {
	Data map[string]Data `json:"data"`
}

// Database object struct
type Database struct {
	Path   string
	Schema Schema
}

// NewDatabase creates new Database object
func NewDatabase(path string) *Database {
	log.Printf("Opening database on path=%s", path)
	database := Database{
		Path: path,
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("Database on path=%s doesn't exists. Creating a new one with empty scheme: %s", path, err)
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
	data, ok := d.Schema.Data[path]
	return data, ok
}

// WriteOne writes Data entry for specific file
func (d *Database) WriteOne(path string, data Data) (Data, bool) {
	d.Schema.Data[path] = data
	_, ok := d.Schema.Data[path]
	return data, ok
}

// Count returns count of elements in database
func (d *Database) Count() int {
	return len(d.Schema.Data)
}

// ListPaths returns list of files present in database
func (d *Database) ListPaths() []string {
	keys := make([]string, 0, d.Count())
	for k := range d.Schema.Data {
		keys = append(keys, k)
	}

	return keys
}

// Commit writes all the in-mem changes to disk
func (d *Database) Commit() error {
	js, err := json.Marshal(d.Schema)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(d.Path, js, 0644)
	return err
}

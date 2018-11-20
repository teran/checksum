package database

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestDatabase(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "database")
	if err != nil {
		t.Errorf("Error creating tempfile: %s", err)
	}
	os.Remove(tmpfile.Name())

	d := NewDatabase(tmpfile.Name())

	dataObject := Data{
		Modified: time.Now(),
		Length:   10,
		SHA1:     "deadbeaf",
		SHA256:   "deadbeaf",
	}

	_, ok := d.WriteOne("testPath", dataObject)
	if !ok {
		t.Errorf("Error reported invoking WriteOne")
	}

	data, ok := d.ReadOne("testPath")
	if !ok {
		t.Errorf("Error reported invoking ReadOne")
	}

	if data != dataObject {
		t.Errorf("Wrong data object returned from ReadOne")
	}

	err = d.Commit()
	if err != nil {
		t.Errorf("Error while executing Commit(): %s", err)
	}

	d2 := NewDatabase(tmpfile.Name())
	do, ok := d2.ReadOne("testPath")
	if !ok {
		t.Errorf("Error reported invoking ReadOne on flushed database")
	}

	if do.Length != dataObject.Length {
		t.Errorf("Saved object differs with Length from it's origin: %#v != %#v", dataObject, do)
	}

	if do.SHA1 != dataObject.SHA1 {
		t.Errorf("Saved object differs with SHA1 from it's origin: %#v != %#v", dataObject, do)
	}

	if do.SHA256 != dataObject.SHA256 {
		t.Errorf("Saved object differs with SHA256 from it's origin: %#v != %#v", dataObject, do)
	}
}

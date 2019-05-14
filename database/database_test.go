package database

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type DatabaseTestSuite struct {
	suite.Suite

	tmpfile *os.File
	db      *Database
}

func (s *DatabaseTestSuite) SetupTest() {
	var err error
	s.tmpfile, err = ioutil.TempFile("", "database")
	s.Require().NoError(err)

	// We need file path only
	os.Remove(s.tmpfile.Name())

	s.db, err = NewDatabase(s.tmpfile.Name())
	s.Require().NoError(err)
}

func (s *DatabaseTestSuite) TestAll() {
	dataObject := &DataObject{
		Modified: time.Now(),
		Length:   10,
		SHA1:     "deadbeaf",
		SHA256:   "deadbeaf",
	}

	cnt := s.db.Count()
	s.Require().Equal(0, cnt)

	res, ok := s.db.WriteOne("test-path", dataObject)
	s.Require().True(ok)
	s.Require().NotNil(res)
	s.Require().Equal(dataObject, res)

	res, ok = s.db.WriteOne("test-path", dataObject)
	s.Require().True(ok)
	s.Require().NotNil(res)
	s.Require().Equal(dataObject, res)

	res, ok = s.db.ReadOne("not-existent-path")
	s.Require().False(ok)
	s.Require().Nil(res)

	res, ok = s.db.ReadOne("test-path")
	s.Require().True(ok)
	s.Require().NotNil(res)
	s.Require().Equal(dataObject, res)

	resMap := s.db.MapObjects()
	s.Require().NotNil(resMap)
	s.Require().Equal(map[string]*DataObject{
		"test-path": dataObject,
	}, resMap)

	resKeys := s.db.ListPaths()
	s.Require().NotNil(resKeys)
	s.Require().Len(resKeys, 1)
	s.Require().Equal([]string{"test-path"}, resKeys)

	cnt = s.db.Count()
	s.Require().Equal(1, cnt)

	ok = s.db.DeleteOne("not-existent-path")
	s.Require().False(ok)

	ok = s.db.DeleteOne("test-path")
	s.Require().True(ok)

	res, ok = s.db.ReadOne("not-existent-path")
	s.Require().False(ok)
	s.Require().Nil(res)

	err := s.db.Commit()
	s.Require().NoError(err)
}

func (s *DatabaseTestSuite) TearDownTest() {
	os.Remove(s.tmpfile.Name())
}

func TestDatabaseTestSuite(t *testing.T) {
	suite.Run(t, &DatabaseTestSuite{})
}

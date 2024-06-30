package concurrent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func (s *concurrentWriterTestSuite) TestWrite() {
	sample := func() []byte {
		out := make([]byte, 1024)
		for i := range out {
			out[i] = 0xaf
		}
		return out
	}()

	writers := func() []*bytes.Buffer {
		out := make([]*bytes.Buffer, 10)
		for i := range out {
			out[i] = &bytes.Buffer{}
		}
		return out
	}()

	w, err := NewConcurrentMultiWriter(context.TODO(), func(buffers ...*bytes.Buffer) (out []io.Writer) {
		for _, wr := range buffers {
			out = append(out, wr)
		}
		return
	}(writers...)...)
	s.Require().NoError(err)

	n, err := w.Write(sample)
	s.Require().NoError(err)
	s.Require().Equal(len(sample), n)

	for i, buf := range writers {
		s.T().Run(fmt.Sprintf("buffer #%d", i), func(t *testing.T) {
			r := require.New(t)

			r.Equal(sample, buf.Bytes())
		})
	}
}

// ========================================================================
// Test suite setup
// ========================================================================
type concurrentWriterTestSuite struct {
	suite.Suite
}

func TestConcurrentWriterTestSuite(t *testing.T) {
	suite.Run(t, &concurrentWriterTestSuite{})
}

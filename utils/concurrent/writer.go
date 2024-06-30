package concurrent

import (
	"context"
	"io"
	"runtime"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type writer struct {
	ctx     context.Context
	writers []io.Writer
}

func (cw *writer) Write(p []byte) (n int, err error) {
	g, _ := errgroup.WithContext(cw.ctx)
	g.SetLimit(runtime.NumCPU())

	for idx, w := range cw.writers {
		g.Go(func(idx int, w io.Writer) func() error {
			return func() error {
				n, err := w.Write(p)
				if err != nil {
					return errors.Wrapf(err, "error writing to channel #%d", idx)
				}

				if n != len(p) {
					return errors.Wrapf(io.ErrShortWrite, "error writing to channel #%d", idx)
				}
				return nil
			}
		}(idx, w))
	}

	return len(p), g.Wait()
}

func NewConcurrentMultiWriter(ctx context.Context, writers ...io.Writer) (io.Writer, error) {
	w := make([]io.Writer, len(writers))

	n := copy(w, writers)
	if n != len(writers) {
		return nil, errors.Errorf(
			"unexpected copy amount: expected %d copied %d. Looks like internal error or memory corruption",
			len(writers), n,
		)
	}

	return &writer{
		ctx:     ctx,
		writers: w,
	}, nil
}

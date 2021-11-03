package stream

import (
	"context"
	"io"
	"sync"

	"github.com/ONSdigital/log.go/v2/log"
)

type Transformer = func(ctx context.Context, r io.Reader, w io.Writer) error
type Consumer = func(ctx context.Context, r io.Reader) error

// Stream is a generic streamming method that creates 2 go-routines:
// - one transforms the provided body into another stream by calling the provided transform method
// - the other consumes the transformed stream using the provided consume method
// This method block until all work is complete, at which point all Readers and Writers are closed and any error is returned.
func Stream(ctx context.Context, body io.ReadCloser, transform Transformer, consume Consumer) (err error) {
	r, w := io.Pipe()
	wg := &sync.WaitGroup{}

	// Start go-routine to read response body, transform it 'on-the-fly' and write it to the pipe writer
	wg.Add(1)
	go func() {
		defer func() {
			closeResponseBody(ctx, body)
			w.Close()
			wg.Done()
		}()
		err = transform(ctx, body, w)
	}()

	// Start go-routine to read pipe reader (transformed) and call the consumer func
	wg.Add(1)
	go func() {
		defer func() {
			r.Close()
			wg.Done()
		}()
		err = consume(ctx, r)
	}()

	wg.Wait()
	return err
}

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(ctx context.Context, body io.Closer) {
	if body != nil {
		if err := body.Close(); err != nil {
			log.Error(ctx, "error closing http response body", err)
		}
	}
}

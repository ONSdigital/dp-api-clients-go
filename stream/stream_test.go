package stream

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	ctx                = context.Background()
	errTestTransformer = errors.New("transfomer error")
	errTestConsumer    = errors.New("consumer error")
	testInput          = ioutil.NopCloser(bytes.NewReader([]byte{1, 2, 3, 4, 5}))
	f                  = func(in []byte) []byte {
		out := make([]byte, 10)
		for i, b := range in {
			out[i] = 2 * b
			out[5+i] = b + 1
		}
		return out
	}
	expectedOutput = []byte{2, 4, 6, 8, 10, 2, 3, 4, 5, 6}
)

func TestStream(t *testing.T) {
	Convey("Given a valid Transformer an Consumer", t, func(c C) {
		output := make([]byte, 10)

		var transform Transformer = func(ctx context.Context, r io.Reader, w io.Writer) error {
			rb := make([]byte, 5)
			n, err := r.Read(rb)
			c.So(n, ShouldEqual, 5)
			c.So(err, ShouldBeNil)

			bw := f(rb)

			n, err = w.Write(bw)
			c.So(n, ShouldEqual, 10)
			c.So(err, ShouldBeNil)

			return nil
		}

		var consume Consumer = func(ctx context.Context, r io.Reader) error {
			n, err := r.Read(output)
			c.So(n, ShouldEqual, 10)
			c.So(err, ShouldBeNil)
			return nil
		}

		Convey("Then calling Stream is successful and obtains the expected output", func() {
			err := Stream(ctx, testInput, transform, consume)
			So(err, ShouldBeNil)
			So(output, ShouldResemble, expectedOutput)
		})
	})

	Convey("Given an erroring Transformer and a valid Consumer", t, func(c C) {
		output := make([]byte, 10)

		var errTransform Transformer = func(ctx context.Context, r io.Reader, w io.Writer) error {
			return errTestTransformer
		}

		var consume = func(ctx context.Context, r io.Reader) error {
			n, err := r.Read(output)
			c.So(n, ShouldEqual, 0)
			c.So(err, ShouldResemble, errTestTransformer)
			return nil
		}

		Convey("Then calling Stream fails with the expected error and the output is not written", func() {
			err := Stream(ctx, testInput, errTransform, consume)
			So(err, ShouldResemble, fmt.Errorf("transform error: %w", errTestTransformer))
			So(output, ShouldResemble, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
		})
	})

	Convey("Given a valid Transformer and an erroring Consumer", t, func(c C) {
		output := make([]byte, 10)

		var transform Transformer = func(ctx context.Context, r io.Reader, w io.Writer) error {
			rb := make([]byte, 5)
			n, err := r.Read(rb)
			c.So(n, ShouldEqual, 0)
			c.So(err, ShouldResemble, errors.New("EOF"))
			return nil
		}

		var errConsume Consumer = func(ctx context.Context, r io.Reader) error {
			return errTestConsumer
		}

		Convey("Then calling Stream fails with the expected error and the output is not written", func() {
			err := Stream(ctx, testInput, transform, errConsume)
			So(err, ShouldResemble, fmt.Errorf("consumer error: %w", errTestConsumer))
			So(output, ShouldResemble, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
		})
	})

	Convey("Given an erroring Transformer and Consumer", t, func(c C) {
		output := make([]byte, 10)

		var transform Transformer = func(ctx context.Context, r io.Reader, w io.Writer) error {
			return errTestTransformer
		}

		var errConsume Consumer = func(ctx context.Context, r io.Reader) error {
			return errTestConsumer
		}

		Convey("Then calling Stream fails with the expected error and the output is not written", func() {
			err := Stream(ctx, testInput, transform, errConsume)
			So(err, ShouldResemble, fmt.Errorf("transform error: %v, consumer error: %v", errTestTransformer, errTestConsumer))
			So(output, ShouldResemble, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
		})
	})
}

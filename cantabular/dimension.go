package cantabular

import (
	"context"
	"errors"
	"fmt"
)

type (
	// Dimension represents the 'dimension' field from a GraphQL
	// query dataset response
	Dimension struct {
		Count      int          `json:"count"`
		Categories []Category   `json:"categories"`
		Variable   VariableBase `json:"variable"`
	}

	// Dimensions describes the structure of a table
	Dimensions []Dimension

	// Iterator facilitates reading the coordinates of each cell in row-major order
	Iterator struct {
		ctx        context.Context
		dims       Dimensions
		dimIndices []int
	}
)

// NewIterator creates an iterator over a table on these Dimensions
func (dims Dimensions) NewIterator(ctx context.Context) *Iterator {
	return &Iterator{
		ctx:        ctx,
		dims:       dims,
		dimIndices: make([]int, len(dims)),
	}
}

// End returns true if there are no more cells in the table
func (it *Iterator) End() bool {
	return it.dimIndices[0] >= it.dims[0].Count
}

// Next advances to the next table cell. It should not be called if End() would return true.
// If the Iterator contet is done, an error will be returned.
func (it *Iterator) Next() error {
	select {
	case <-it.ctx.Done():
		return fmt.Errorf("context is done: %w", it.ctx.Err())
	default:
		if err := it.checkNotAtEnd(); err != nil {
			return err
		}
		for j := len(it.dimIndices) - 1; j >= 0; j -= 1 {
			if it.dimIndices[j] += 1; it.dimIndices[j] < it.dims[j].Count || j == 0 {
				break
			}
			it.dimIndices[j] = 0
		}
	}
	return nil
}

// CategoryAtColumn returns the i-th coordinate of the current cell
func (it *Iterator) CategoryAtColumn(i int) Category {
	it.checkNotAtEnd()
	return it.dims[i].Categories[it.dimIndices[i]]
}

func (it *Iterator) checkNotAtEnd() error {
	if it.End() {
		return errors.New("after end of table")
	}
	return nil
}

package cantabular

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
		dims       Dimensions
		dimIndices []int
	}
)

// NewIterator creates an iterator over a table on these Dimensions
func (dims Dimensions) NewIterator() *Iterator {
	return &Iterator{
		dims:       dims,
		dimIndices: make([]int, len(dims)),
	}
}

// End returns true if there are no more cells in the table
func (ti *Iterator) End() bool {
	return ti.dimIndices[0] >= ti.dims[0].Count
}

// Next advances to the next table cell. It should not be called if End() would return true.
func (ti *Iterator) Next() {
	ti.checkNotAtEnd()
	for j := len(ti.dimIndices) - 1; j >= 0; j -= 1 {
		if ti.dimIndices[j] += 1; ti.dimIndices[j] < ti.dims[j].Count || j == 0 {
			break
		}
		ti.dimIndices[j] = 0
	}
}

// CategoryAtColumn returns the i-th coordinate of the current cell
func (ti *Iterator) CategoryAtColumn(i int) Category {
	ti.checkNotAtEnd()
	return ti.dims[i].Categories[ti.dimIndices[i]]
}

func (ti *Iterator) checkNotAtEnd() {
	if ti.End() {
		panic("after end of table")
	}
}

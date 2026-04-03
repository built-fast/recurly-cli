package output

// TypedColumn defines a type-safe table column with a header label and a
// function to extract the cell value from a strongly-typed data item.
type TypedColumn[T any] struct {
	Header  string
	Extract func(T) string
}

// ToColumns converts a slice of TypedColumn[T] into a slice of Column by
// wrapping each typed extractor with a type assertion. The assertion panics on
// mismatch, which is intentional — generics prevent mismatches at compile time,
// so a panic here indicates a programming error.
func ToColumns[T any](typed []TypedColumn[T]) []Column {
	cols := make([]Column, len(typed))
	for i, tc := range typed {
		extract := tc.Extract
		cols[i] = Column{
			Header: tc.Header,
			Extract: func(v any) string {
				return extract(v.(T)) //nolint:errcheck // type guaranteed by generics
			},
		}
	}
	return cols
}

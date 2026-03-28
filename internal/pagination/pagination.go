package pagination

// Lister is a generic interface matching the SDK's paginated list types.
// Each SDK lister (AccountLister, SubscriptionLister, etc.) implements
// Fetch(), Data(), and HasMore() with its own concrete type.
type Lister[T any] interface {
	Fetch() error
	Data() []T
	HasMore() bool
}

// Result holds the items collected from a paginated lister along with
// whether more results exist beyond what was collected.
type Result[T any] struct {
	Items   []T
	HasMore bool
}

// Collect iterates through a paginated SDK lister and returns collected results.
// If all is true, it fetches every page. Otherwise, it returns up to limit results.
// Default limit is 20 if limit <= 0 and all is false.
func Collect[T any](lister Lister[T], limit int, all bool) (Result[T], error) {
	if !all && limit <= 0 {
		limit = 20
	}

	var results []T
	hasMore := false

	for lister.HasMore() {
		if err := lister.Fetch(); err != nil {
			return Result[T]{}, err
		}

		page := lister.Data()
		if all {
			results = append(results, page...)
			continue
		}

		remaining := limit - len(results)
		if len(page) >= remaining {
			results = append(results, page[:remaining]...)
			hasMore = len(page) > remaining || lister.HasMore()
			break
		}
		results = append(results, page...)
	}

	return Result[T]{Items: results, HasMore: hasMore}, nil
}

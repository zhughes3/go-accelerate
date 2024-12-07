package slices

func Map[T, S any](vals []T, mapper func(T) S) []S {
	out := make([]S, 0, len(vals))
	for _, v := range vals {
		out = append(out, mapper(v))
	}

	return out
}

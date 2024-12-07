package maps

func Inverse[M ~map[K]V, K comparable, V comparable](m M) map[V]K {
	return InverseWithValueMapper(m, func(v V) V { return v })
}

func InverseWithValueMapper[M ~map[K]V, K comparable, V comparable](m M, mapper func(V) V) map[V]K {
	flipped := make(map[V]K, len(m))
	for k, v := range m {
		flipped[mapper(v)] = k
	}

	return flipped
}

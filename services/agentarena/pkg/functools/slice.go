package functools

func Map[T, U any](a []T, mapper func(T) U) []U {
	out := make([]U, len(a))
	for i, e := range a {
		out[i] = mapper(e)
	}
	return out
}

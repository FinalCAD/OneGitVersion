package uarray

func Contains[K comparable](a []K, element K) bool {
	for _, x := range a {
		if x == element {
			return true
		}
	}
	return false
}

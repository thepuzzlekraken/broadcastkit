package sony

import (
	"fmt"
	"strconv"
	"strings"
)

func commaJoin(s ...string) string {
	return strings.Join(s, ",")
}

func commaSplit(s string) []string {
	return strings.Split(s, ",")
}

func itoa(i int) string {
	return strconv.Itoa(i)
}

func atoi(s string) (int, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	if c := itoa(i); s != c {
		return 0, fmt.Errorf("loosy int format %s != %s", s, c)
	}
	return i, nil
}

func castGeneric[T Parameter](ss []T) []Parameter {
	gs := make([]Parameter, len(ss))
	for i, p := range ss {
		gs[i] = p
	}
	return gs
}

func castSpecific[T Parameter](gs []Parameter) []T {
	ss := make([]T, len(gs))
	for i, p := range ss {
		ss[i] = p
	}
	return ss
}

package sony

import (
	"fmt"
	"net/url"
	"slices"
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
		gs[i] = Parameter(p)
	}
	return gs
}

func castSpecific[T Parameter](gs []Parameter) []T {
	ss := make([]T, 0, len(gs))
	for _, p := range gs {
		if s, ok := p.(T); ok {
			ss = append(ss, s)
		}
	}
	return ss
}

func urlEncode(v url.Values) string {
	// This function is copied from the standard library.
	// The Sony FR-7 cameras do NOT accept space as "+" in the query strings.
	// This function has a workaround to standard-breakingly leave spaces alone.
	if len(v) == 0 {
		return ""
	}
	var buf strings.Builder
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		vs := v[k]
		keyEscaped := url.QueryEscape(k)
		keyEscaped = strings.ReplaceAll(keyEscaped, "+", " ")
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(keyEscaped)
			buf.WriteByte('=')
			val := url.QueryEscape(v)
			val = strings.ReplaceAll(val, "+", " ")
			buf.WriteString(val)
		}
	}
	return buf.String()
}

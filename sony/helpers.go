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
	if len(s) == 0 {
		return []string{}
	}
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

func hexDecode(s string) (int, error) {
	i := 0
	for j := 0; j < len(s); j++ {
		i <<= 4
		c := s[j]
		if c <= '9' && c >= '0' {
			i |= int(c - '0')
			continue
		}
		if c <= 'f' && c >= 'a' {
			i |= int(c - 'a' + 10)
			continue
		}
		return 0, fmt.Errorf("invalid ascii char in hex: %x", c)
	}
	return i, nil
}

func hex20Decoder(h string) (int, error) {
	if len(h) != 5 {
		return 0, fmt.Errorf("invalid hex20 length: %d", len(h))
	}
	i, err := hexDecode(h)
	// fix signage bit underflow
	if (i & 0x80000) != 0 {
		i |= ^int(0x7ffff)
	}
	return i, err
}

func hex16Decoder(h string) (int, error) {
	if len(h) != 4 {
		return 0, fmt.Errorf("invalid hex16 length: %d", len(h))
	}
	i, err := hexDecode(h)
	// hex16 is unsigned, leave as-is
	return i, err
}

const intSize = 32 << (^uint(0) >> 63) // see stdlib math/const.go

func hexEncoder(i int) string {
	b := make([]byte, intSize/4)
	for j := intSize / 4; j > 0; j-- {
		c := i & 0xf
		if c <= 9 {
			c += '0'
		} else {
			c += 'a' - 10
		}
		b[j-1] = byte(c)
		i >>= 4
	}
	return string(b)
}

func hex20Encoder(i int) string {
	e := hexEncoder(i)
	return e[len(e)-5:]
}

func hex16Encoder(i int) string {
	e := hexEncoder(i)
	return e[len(e)-4:]
}

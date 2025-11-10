package protomaskerpkg

import (
	"strings"
	"unicode"
)

const (
	maskRune     = '*'
	passwordMask = "********"
)

func MaskAsGeneric(s string) string {
	rr := []rune(s)

	for i, r := range rr {
		if !unicode.IsSpace(r) {
			rr[i] = maskRune
		}
	}

	return string(rr)
}

func MaskAsPassword(s string) string {
	if s == "" {
		return ""
	}

	return passwordMask
}

func MaskAsEmail(s string) string {
	if at := strings.Index(s, "@"); at == -1 {
		return MaskAsGeneric(s)
	} else {
		sb := strings.Builder{}
		sb.Grow(len(s))
		showFirst := min(10, at/4)
		for i, r := range s {
			if i < showFirst {
				sb.WriteRune(r)
			} else if i < at {
				sb.WriteRune(maskRune)
			} else {
				sb.WriteRune(r)
			}
		}
		return sb.String()
	}
}

func MaskAsPhone(s string) string {
	rr := []rune(s)

	// Analyze first.
	maskFrom, maskTo := analyzeSpans(rr, 4, 4, func(r rune) bool {
		return '0' <= r && r <= '9'
	})

	for i := maskFrom; i < maskTo; i++ {
		if '0' <= rr[i] && rr[i] <= '9' {
			rr[i] = maskRune
		}
	}

	return string(rr)
}

func MaskAsCardPAN(s string) string {
	rr := []rune(s)

	// Analyze first.
	maskFrom, maskTo := analyzeSpans(rr, 6, 4, func(r rune) bool {
		return '0' <= r && r <= '9'
	})

	for i := maskFrom; i < maskTo; i++ {
		if '0' <= rr[i] && rr[i] <= '9' {
			rr[i] = maskRune
		}
	}

	return string(rr)
}

func analyzeSpans(rr []rune, showFirst, showLast int, shouldMask func(rune) bool) (maskFrom, maskTo int) {
	maskFrom = -1
	for i, r := range rr {
		if shouldMask(r) {
			showFirst--
		}
		if showFirst <= 0 {
			maskFrom = i + 1
			break
		}
	}

	if maskFrom < 0 {
		return 0, len(rr)
	}

	maskTo = -1
	for i := len(rr) - 1; i > showFirst; i-- {
		if shouldMask(rr[i]) {
			showLast--
		}
		if showLast <= 0 {
			maskTo = i
			break
		}
	}

	if maskTo < 0 || maskTo-maskFrom <= 0 {
		return 0, len(rr)
	}

	return maskFrom, maskTo
}

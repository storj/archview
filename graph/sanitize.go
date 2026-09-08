package graph

import "strings"

func sanitize(name string) string {
	// Escape underscores as well as punctuation so distinct qualified names
	// cannot collapse to the same identifier. The prefix keeps IDs valid even
	// when a package path starts with a digit.
	var out strings.Builder
	out.WriteString("n_")
	const hex = "0123456789abcdef"
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch {
		case 'a' <= c && c <= 'z', 'A' <= c && c <= 'Z', '0' <= c && c <= '9':
			out.WriteByte(c)
		default:
			out.WriteByte('_')
			out.WriteByte(hex[c>>4])
			out.WriteByte(hex[c&15])
		}
	}
	return out.String()
}

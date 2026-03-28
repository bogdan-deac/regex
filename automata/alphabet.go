package automata

import (
	"cmp"
	"fmt"
	"slices"
)

type Symbol = rune

type StateLike interface {
	cmp.Ordered
	fmt.Stringer
}
type ASCIISet []rune

const Wildcard = -1

// ASCIIChars contains all ASCII characters (0–127).
var ASCIIChars = ASCIISet{
	'\x00', '\x01', '\x02', '\x03', '\x04', '\x05', '\x06', '\x07',
	'\x08', '\x09', '\x0A', '\x0B', '\x0C', '\x0D', '\x0E', '\x0F',
	'\x10', '\x11', '\x12', '\x13', '\x14', '\x15', '\x16', '\x17',
	'\x18', '\x19', '\x1A', '\x1B', '\x1C', '\x1D', '\x1E', '\x1F',
	' ', '!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/',
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	':', ';', '<', '=', '>', '?', '@',
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
	'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T',
	'U', 'V', 'W', 'X', 'Y', 'Z',
	'[', '\\', ']', '^', '_', '`',
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
	'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't',
	'u', 'v', 'w', 'x', 'y', 'z',
	'{', '|', '}', '~',
	'\x7F',
}

func (r ASCIISet) Without(a rune) ASCIISet {
	return slices.DeleteFunc(r, func(el rune) bool { return el == a })
}

func (r ASCIISet) WithoutRange(start, end rune) ASCIISet {
	return slices.DeleteFunc(r, func(el rune) bool { return start <= el && el <= end })

}

func (r ASCIISet) Add(el rune) ASCIISet {
	if !slices.Contains(r, el) {
		return append(r, el)
	}
	return r
}

func (r ASCIISet) Delete(el rune) ASCIISet {
	return slices.DeleteFunc(r, func(t rune) bool { return t == el })
}

func (r ASCIISet) Union(start, end rune) ASCIISet {
	var toAdd ASCIISet
	for el := start; el <= end; el++ {
		if !slices.Contains(r, el) {
			toAdd = append(toAdd, el)
		}
	}
	return append(r, toAdd...)
}

func (r ASCIISet) Difference(start, end rune) ASCIISet {
	var newRange ASCIISet
	for _, el := range r {
		if el < start || el > end {
			newRange = append(newRange, el)
		}
	}
	return newRange
}

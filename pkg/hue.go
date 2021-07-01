/*
MIT License

Copyright (c) 2021 Rally Health, Inc.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package pkg

import (
	"fmt"
	"io"
	"regexp"
)

// Foreground color codes
const (
	Black = iota + 30
	Red
	Green
	Brown // Looks more like yellow on most terminal emulators
	Blue
	Magenta
	Cyan
	White
	Underline
	Default
)

// For iterating through all color codes
const (
	First = Black
	Last  = White
)

const (
	// ASCIIFmt is a format specifer for a ECMA-48 color string
	ASCIIFmt = "\033[%d;%dm"
	// ASCIIReset is a reset code that restores the colors to their defaults
	ASCIIReset = "\033[0m"
	// ASCIIFmtReset is a combination of ASCIIFmt and ASCIIReset with a value is sandwiched in between.
	ASCIIFmtReset = "\033[%d;%dm%v\033[0m"

	// fmt.Sprintf can't accept arguments in the form fmt.Sprintf("%s %s %s %s", "a", "b", []interface{ "c", "d"}...)
	// see Encode for the workaround that uses this constant
	ASCIIFmtFmtReset = "\033[%d;%dm%%v\033[0m"
)

// StringToHue is a map of hue strings to color codes
var StringToHue = map[string]int{
	"black":   Black,
	"blue":    Blue,
	"brown":   Brown,
	"cyan":    Cyan,
	"default": Default,
	"green":   Green,
	"magenta": Magenta,
	"red":     Red,
	"white":   White,
}

// HueToString is a map of hue color codes to their names
var HueToString = map[int]string{
	Black:   "black",
	Blue:    "blue",
	Brown:   "brown",
	Cyan:    "cyan",
	Default: "default",
	Green:   "green",
	Magenta: "magenta",
	Red:     "red",
	White:   "white",
}

// SetHue sets the Writer's hue
func (w *Writer) SetHue(h *Hue) {
	w.Hue = h
}

// NewWriter returns a new Writer with the hue 'h'
func NewWriter(w io.Writer, h *Hue) *Writer {
	n := new(Writer)
	n.wrapped = w
	n.SetHue(h)
	return n
}

// Write colorizes and writes the contents of p to the underlying
// writer object.
func (w Writer) Write(p []byte) (n int, err error) {
	return w.wrapped.Write([]byte(w.Sprintf("%s", string(p))))
}

// WriteString colorizes and writes the string s to the
// underlying writer object
func (w Writer) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

// Writer implements colorization for an underlying io.Writer object
type Writer struct {
	*Hue
	wrapped io.Writer
}

// String is a string containing ECMA-48 color codes. Its purpose is to
// remind the user at compile time that it differs from the string builtin.
type String string

// New creates a new hue object with foreground and background colors specified.
func New(fg, bg int) *Hue {
	h := new(Hue)
	h.SetFg(fg)
	h.SetBg(bg)
	return h
}

func (h *Hue) SetFg(c int) {
	h.fg = c
}

func (h *Hue) SetBg(c int) {
	h.bg = c + 10
}

func (h *Hue) Fg() int {
	return h.fg
}

func (h *Hue) Bg() int {
	return h.bg
}

// Hue holds the foreground color and background color as integers
type Hue struct {
	fg, bg int
}

// Decode strips all color data from the String object
// and returns a standard string
func (hs String) Decode() (s string) {
	if l := len(hs); l < 12 {
		panic(fmt.Sprintf("Can't decode hue.String: \"%s\" because it's length is \"%d\" (minimum length is 12)", hs, l))
	}

	b := []byte(hs)
	return string(b[8 : len(b)-4])
}

// Encode encapsulates interface a's string representation
// with the ECMA-40 color codes stored in the hue structure.
func Encode(h *Hue, a ...interface{}) String {
	finalFmt := fmt.Sprintf(ASCIIFmtFmtReset, h.Fg(), h.Bg())
	return String(fmt.Sprintf(finalFmt, a...))

	//return String(fmt.Sprintf(ASCIIFmtReset, h.Fg(), h.Bg(), a))
}

// Sprintf behaves like fmt.Sprintf, except it colorizes the output String
func (h *Hue) Sprintf(format string, a ...interface{}) String {
	return String(fmt.Sprintf(string(Encode(h, format)), a...))
}

// Printf behaves like fmt.Printf, except it colorizes the output
func (h *Hue) Printf(format string, a ...interface{}) {
	fmt.Printf(string(Encode(h, format)), a...)
}

// Print behaves like fmt.Print, except it colorizes the output
func (h *Hue) Print(a ...interface{}) {
	fmt.Print(Encode(h, a...))
}

// Println behaves like fmt.Println, except it colorizes the output
func (h *Hue) Println(a ...interface{}) {
	fmt.Println(Encode(h, a...))
}

// RegexpWriter implements colorization for a io.Writer object by processing
// a set of rules. Rules are hue objects assocated with regular expressions.
type RegexpWriter struct {
	rules   []rule
	wrapped io.Writer
}

type rule struct {
	*Hue
	*regexp.Regexp
}

// NewRegexpWriter returns a new RegexpWriter
func NewRegexpWriter(w io.Writer) *RegexpWriter {
	n := new(RegexpWriter)
	n.wrapped = w
	return n
}

// AddRuleStringPOSIX binds a hue to the POSIX regexp in the string 's'.
// Similar to AddRule, except the caller passes in an uncompiled POSIX regexp.
func (w *RegexpWriter) AddRuleStringPOSIX(h *Hue, s string) {
	re := regexp.MustCompilePOSIX(s)
	w.AddRule(h, re)
}

// AddRuleString binds a hue to the regexp in the string 's'.
// Similar to AddRule, except the caller passes in an uncompiled regexp.
func (w *RegexpWriter) AddRuleString(h *Hue, s string) {
	re := regexp.MustCompile(s)
	w.AddRule(h, re)
}

// AddRule binds a hue to a regular expression.
func (w *RegexpWriter) AddRule(h *Hue, re *regexp.Regexp) {
	//w.rules.PushBack( rule{h, re} )
	w.rules = append(w.rules, rule{h, re})
}

// FlushRules deletes all rules added with AddRule from Writer
func (w *RegexpWriter) FlushRules() {
	w.rules = nil
}

// PrintRules prints out the rules
func (w *RegexpWriter) PrintRules() {
	for _, v := range w.rules {
		fmt.Println(v)
	}
}

// WriteString is similar to Write, except it writes a string to the underlying
// buffer instead of a byte slice.
func (w RegexpWriter) WriteString(s string) (n int, err error) {
	return w.Write([]byte(s))
}

// Write writes the contents of p into the buffer after processesing the regexp
// rules added to Writer with AddRule. Write colorizes the contents as it writes
// to the underlying writer object.
func (w RegexpWriter) Write(p []byte) (n int, err error) {
	huemap := make([]byte, len(p))
	rulemap := make([]*Hue, len(w.rules)+1)
	rulemap[0] = &Hue{}

	for i := 1; i < len(rulemap); i++ {
		r := w.rules[i-1]
		x := r.FindAllIndex(p, -1)

		rulemap[i] = r.Hue
		for _, w := range x {
			for j := w[0]; j < w[1]; j++ {
				huemap[j] = byte(i)
			}
		}
	}

	var hue byte
	for i := range p {
		if huemap[i] != hue {
			hue = huemap[i]
			th := rulemap[hue]

			nb, err := fmt.Fprintf(w.wrapped, ASCIIFmt, th.Fg(), th.Bg())
			if err != nil {
				return n, err
			}
			n += nb
		}

		nb, err := fmt.Fprintf(w.wrapped, "%c", p[i])
		if err != nil {
			return n, err
		}
		n += nb
	}

	for i := range p {
		if huemap[i] != 0 {
			fmt.Print(ASCIIReset)
			break
		}
	}

	return n, err
}

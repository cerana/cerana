package configutil

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/spf13/pflag"
)

var (
	flagSep     = regexp.MustCompile(`[\s._-]+`)
	specialCaps = regexp.MustCompile("(?i)^(url|cpu|ip|id)$")
)

// Normalize sets a normalization function for the flagset and updates
// the flagset's usage output to reflect that.
func Normalize(f *pflag.FlagSet) {
	f.SetNormalizeFunc(NormalizeFunc)
	UsageNormalizedNote(f)
}

// NormalizeFunc translates flag names to camelCase using whitespace, periods,
// underscores, and dashes as word boundaries. All-caps words are preserved.
// Special words are capitalized if not the first word (URL, CPU, IP, ID).
func NormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	// Standardize separators to a single space and trim leading/trailing spaces
	name = strings.TrimSpace(flagSep.ReplaceAllString(name, " "))

	nameParts := strings.Split(name, " ")
	if len(nameParts) > 1 {
		for i, part := range nameParts {
			if part != strings.ToUpper(part) {
				// Convert to title case (lower case with leading caps, preserved all caps)
				part = strings.Title(strings.ToLower(part))
			}

			// Some words should always be all caps or all lower case (e.g. Interval)
			caseFn := strings.ToUpper
			if i == 0 {
				caseFn = strings.ToLower
			}

			nameParts[i] = specialCaps.ReplaceAllStringFunc(part, caseFn)
		}
	}

	// Split on space and examine the first part
	first := nameParts[0]
	if utf8.RuneCountInString(first) == 1 || first != strings.ToUpper(first) {
		// Lowercase the first letter if it is not an all-caps word
		r, n := utf8.DecodeRuneInString(first)
		nameParts[0] = string(unicode.ToLower(r)) + first[n:]
	}

	return pflag.NormalizedName(strings.Join(nameParts, ""))
}

// UsageNormalizedNote sets an Usage function on the flagset with a note about normalized fields.
func UsageNormalizedNote(f *pflag.FlagSet) {
	var origUsage func()

	usage := func() {
		origUsage()
		fmt.Fprintln(os.Stderr, "Note: Long flag names can be specified in either fooBar or foo[_-.]bar form.")
	}

	if f == pflag.CommandLine {
		origUsage = pflag.Usage
		pflag.Usage = usage
	} else {
		origUsage = f.Usage
		if f.Usage == nil {
			origUsage = func() {
				fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
				f.PrintDefaults()
			}
		}
		f.Usage = usage
	}
}

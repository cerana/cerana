# configutil

[![configutil](https://godoc.org/github.com/cerana/cerana/pkg/configutil?status.svg)](https://godoc.org/github.com/cerana/cerana/pkg/configutil)



## Usage

#### func  Normalize

```go
func Normalize(f *pflag.FlagSet)
```
Normalize sets a normalization function for the flagset and updates the
flagset's usage output to reflect that.

#### func  NormalizeFunc

```go
func NormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName
```
NormalizeFunc translates flag names to camelCase using whitespace, periods,
underscores, and dashes as word boundaries. All-caps words are preserved.
Special words are capitalized if not the first word (URL, CPU, IP, ID).

#### func  UsageNormalizedNote

```go
func UsageNormalizedNote(f *pflag.FlagSet)
```
UsageNormalizedNote sets an Usage function on the flagset with a note about
normalized fields.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*

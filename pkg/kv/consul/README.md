# consul

[![consul](https://godoc.org/github.com/cerana/cerana/pkg/kv/consul?status.svg)](https://godoc.org/github.com/cerana/cerana/pkg/kv/consul)



## Usage

#### func  New

```go
func New(addr string) (kv.KV, error)
```
New instantiates a consul kv implementation. The parameter addr may be the empty
string or a valid URL. If addr is not empty it must be a valid URL with schemes
http, https or consul; consul is synonymous with http. If addr is the empty
string the consul client will connect to the default address, which may be
influenced by the environment.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*

# etcd

[![etcd](https://godoc.org/github.com/cerana/cerana/pkg/kv/etcd?status.svg)](https://godoc.org/github.com/cerana/cerana/pkg/kv/etcd)



## Usage

#### func  New

```go
func New(addr string) (kv.KV, error)
```
New instantiates an etcd kv implementation. The parameter addr may be the empty
string or a valid URL. If addr is not empty it must be a valid URL with schemes
http, https or etcd; etcd is synonymous with http. If addr is the empty string
the etcd client will connect to the default address.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*

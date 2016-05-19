package kv

// Cookie holds a unique value that is used as a reference to server side storage.
type Cookie struct {
	Cookie uint64 `json:"cookie"`
}

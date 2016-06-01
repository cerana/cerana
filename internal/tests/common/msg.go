package common

import "fmt"

// Messager generates a function for creating a string message with a prefix.
func (s *Suite) Messager(prefix string) func(...interface{}) string {
	return func(val ...interface{}) string {
		if len(val) == 0 {
			return prefix
		}
		msgPrefix := prefix + " : "
		if len(val) == 1 {
			return msgPrefix + val[0].(string)
		}
		return msgPrefix + fmt.Sprintf(val[0].(string), val[1:]...)
	}
}

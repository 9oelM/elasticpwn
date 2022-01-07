package EPUtils

import (
	"strings"
	"sync"
)

func ConvertSyncMapToMap(sMap sync.Map) map[string]interface{} {
	tmpMap := make(map[string]interface{})
	sMap.Range(func(k, v interface{}) bool {
		tmpMap[k.(string)] = v
		return true
	})
	return tmpMap
}

func containsXWith(cb func(a string, b string) bool) func(x string, wordlist []string) (idx int) {
	return func(x string, wordlist []string) (idx int) {
		for i, v := range wordlist {
			if cb(x, v) {
				return i
			}
		}
		return -1
	}
}

func Unique(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func equals(x string, y string) bool {
	return x == y
}

// returns -1 if a string does not end with any word in wordlist
var ContainsEndsWith = containsXWith(strings.HasSuffix)

// returns -1 if a string does not start with any word in wordlist
var ContainsStartsWith = containsXWith(strings.HasPrefix)

// returns -1 if none of the strings exactly matches any word in wordlist
var ContainsExactlyMatchesWith = containsXWith(equals)

// returns -1 if a string does not match any word in wordlist
var Contains = containsXWith(strings.Contains)

func ExitOnError(e error) {
	if e != nil {
		panic(e)
	}
}

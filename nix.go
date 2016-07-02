package main

import (
	"encoding/base32"
	"fmt"
	"os"
	"regexp"
)

var nixStorePath = os.Getenv("NIX_STORE")

func init() {
	if nixStorePath == "" {
		panic("NIX_STORE is not set")
	}
}

const nixHashSize = 20

var nixBase32 = base32.NewEncoding("0123456789abcdfghijklmnpqrsvwxyz")
var nixStorePathRe = regexp.MustCompile(fmt.Sprintf(
	`^%s([0-9a-z]{%d})-([a-zA-Z0-9+\-?=][.a-zA-Z0-9+\-?=]*)$`,
	regexp.QuoteMeta(nixStorePath+string(os.PathSeparator)),
	nixBase32.EncodedLen(nixHashSize),
))

func parseNixPath(path string) (hash [nixHashSize]byte, name string, ok bool) {
	m := nixStorePathRe.FindStringSubmatch(path)
	if m == nil {
		return
	}
	nixBase32.Decode(hash[:], []byte(m[1]))
	reverse(hash[:])
	name = string(m[2])
	return hash, name, true
}

func reverse(buf []byte) {
	n := len(buf)
	for i := 0; i < n/2; i += 1 {
		j := n - i - 1
		buf[i], buf[j] = buf[j], buf[i]
	}
}

package banning

import (
	"errors"
	"github.com/bitmagnet-io/bitmagnet/internal/protocol/metainfo"
	"unicode/utf8"
)

type utf8Checker struct {
}

func (c utf8Checker) Check(info metainfo.Info) error {
	checkUtf8Strings := make([]string, 0, len(info.Files)+1)
	checkUtf8Strings = append(checkUtf8Strings, info.BestName())
	for _, file := range info.Files {
		checkUtf8Strings = append(checkUtf8Strings, file.DisplayPath(&info))
	}
	for _, str := range checkUtf8Strings {
		if !utf8.ValidString(str) {
			return errors.New("meta info contains an invalid utf8 string")
		}
	}
	return nil
}

package metainfo

import (
	"errors"
	"fmt"
	"github.com/anacrolix/torrent/bencode"
	mi "github.com/anacrolix/torrent/metainfo"
	"github.com/bitmagnet-io/bitmagnet/internal/protocol"
)

func ParseMetaInfoBytes(infoHash protocol.ID, metaInfoBytes []byte) (Info,MetaInfo, error) {
	if protocol.ID(mi.HashBytes(metaInfoBytes)) != infoHash {
		return Info{}, MetaInfo{}, errors.New("info bytes have wrong hash")
	}
	var info Info
	if unmarshalErr := bencode.Unmarshal(metaInfoBytes, &info); unmarshalErr != nil {
		return Info{}, MetaInfo{}, fmt.Errorf("error unmarshaling info bytes: %s", unmarshalErr)
	}
	var metainfo MetaInfo
	if unmarshalErr := bencode.Unmarshal(metaInfoBytes, &metainfo); unmarshalErr != nil {
		return Info{}, MetaInfo{}, fmt.Errorf("error unmarshaling info bytes: %s", unmarshalErr)
	}
	if metainfo.CreationDate > 0 {
		fmt.Println(fmt.Errorf("parse %s %d", info.Name, metainfo.CreationDate))
	}
	return info, metainfo, nil
}

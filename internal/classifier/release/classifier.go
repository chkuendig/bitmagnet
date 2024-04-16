package release

import (
	"context"
	"fmt"

	"github.com/bitmagnet-io/bitmagnet/internal/classifier"
	"github.com/bitmagnet-io/bitmagnet/internal/model"
)

type releaseClassifier struct {
}

func (c releaseClassifier) Key() string {
	return "release"
}

func (c releaseClassifier) Priority() int {
	return 10
}

func (c releaseClassifier) Classify(_ context.Context, t model.Torrent) (classifier.Classification, error) {
	fmt.Println("release classifier:" + t.Name)
	if !t.Hint.IsNil() || t.FilesStatus == model.FilesStatusNoInfo || t.FilesStatus == model.FilesStatusOverThreshold {
		return classifier.Classification{}, classifier.ErrNoMatch
	}
	if t.FilesStatus == model.FilesStatusSingle {
		if t.Extension.Valid {
			ct := model.ContentTypeFromExtension(t.Extension.String)
			if ct.Valid {
				return classifier.Classification{
					ContentType: ct,
				}, nil
			}
		}
		return classifier.Classification{}, classifier.ErrNoMatch
	}
	var unknownSize uint64
	sizeMap := make(map[model.ContentType]uint64)
	for _, f := range t.Files {
		if f.Size == 0 {
			unknownSize++
			continue
		}
		ct := model.ContentTypeFromExtension(f.Extension.String)
		if ct.Valid {
			sizeMap[ct.ContentType] += f.Size
		} else {
			unknownSize += f.Size
		}
	}
	var maxSize uint64
	var maxType model.ContentType
	for k, v := range sizeMap {
		if v > maxSize {
			maxSize = v
			maxType = k
		}
	}
	if maxSize > 0 && maxSize > unknownSize {
		return classifier.Classification{
			ContentType: model.NewNullContentType(maxType),
		}, nil
	}
	return classifier.Classification{}, classifier.ErrNoMatch
}

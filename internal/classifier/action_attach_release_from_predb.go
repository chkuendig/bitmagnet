package classifier

import (
	"fmt"

	"github.com/bitmagnet-io/bitmagnet/internal/classifier/classification"
)

const attachReleaseFromPreDBName = "attach_release_from_predb"

type attachReleaseFromPreDBAction struct{}

func (attachReleaseFromPreDBAction) name() string {
	return attachReleaseFromPreDBName
}

var attachReleaseFromPreDBPayloadSpec = payloadLiteral[string]{
	literal:     attachReleaseFromPreDBName,
	description: "Attempt to attach a predb release with a search on the torrent name",
}

func (a attachReleaseFromPreDBAction) compileAction(ctx compilerContext) (action, error) {
	if _, err := attachReleaseFromPreDBPayloadSpec.Unmarshal(ctx); err != nil {
		return action{}, ctx.error(err)
	}
	return action{
		run: func(ctx executionContext) (classification.Result, error) {
			cl := ctx.result
			fmt.Println("Looking for release for ", ctx.torrent.Name)
			if !cl.ContentType.Valid || !cl.BaseTitle.Valid {
				return cl, classification.ErrUnmatched
			}
			content, err := ctx.search.ContentBySearch(ctx.Context, cl.ContentType.ContentType, cl.BaseTitle.String, cl.Date.Year)
			if err != nil {
				return cl, err
			}
			cl.AttachContent(&content)
			return cl, nil
		},
	}, nil
}

func (attachReleaseFromPreDBAction) JsonSchema() JsonSchema {
	return attachReleaseFromPreDBPayloadSpec.JsonSchema()
}

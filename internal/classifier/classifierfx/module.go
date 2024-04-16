package classifierfx

import (
	"github.com/bitmagnet-io/bitmagnet/internal/boilerplate/config/configfx"
	"github.com/bitmagnet-io/bitmagnet/internal/classifier"
	"github.com/bitmagnet-io/bitmagnet/internal/classifier/release"
	"github.com/bitmagnet-io/bitmagnet/internal/classifier/video/videofx"
	"go.uber.org/fx"
)

func New() fx.Option {
	return fx.Module(
		"workflow",
		configfx.NewConfigModule[classifier.Config]("classifier", classifier.NewDefaultConfig()),
		fx.Provide(
			classifier.New,
			extension.New,
		),
	)
}

package fx

import "go.uber.org/fx"

func AsProvider(provider any, providerInterface interface{}) fx.Option {
	return fx.Provide(fx.Annotate(provider, fx.As(providerInterface)))
}

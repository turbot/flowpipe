package config

import (
	"context"
)

type configContextKey struct{}

/*
type viperContextKey struct{}

func ContextWithViper(ctx context.Context) context.Context {
	return context.WithValue(ctx, viperContextKey{}, viper.New())
}

func Viper(ctx context.Context) *viper.Viper {
	return ctx.Value(viperContextKey{}).(*viper.Viper)
}
*/

func ContextWithConfig(ctx context.Context) (context.Context, error) {
	//cfg, err := NewConfig(ctx, WithFlags())
	cfg, err := NewConfig(ctx)
	if err != nil {
		return nil, err
	}
	return context.WithValue(ctx, configContextKey{}, cfg), nil
}

func Config(ctx context.Context) *Configuration {
	return ctx.Value(configContextKey{}).(*Configuration)
}

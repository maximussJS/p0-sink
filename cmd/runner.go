package main

import (
	"go.uber.org/fx"
	"p0-sink/internal/bootstrap"
)

func main() {
	fx.New(bootstrap.CreateApp()).Run()
}

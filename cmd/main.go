// cmd/main.go
package main

import (
	"shop-microservice/internal/di"

	"go.uber.org/fx"
)

func main() {
	fx.New(di.Module).Run()
}

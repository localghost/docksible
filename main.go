package main

import (
	"github.com/localghost/docksible/builder"
	"github.com/localghost/docksible/product"
)

func main() {
	builder := builder.New()
	builder.Bootstrap()
	conatinerId := product.New().Run()
	builder.ProvisionContainer(conatinerId)
}

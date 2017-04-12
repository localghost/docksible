package main

import (
	"log"
	"github.com/localghost/docksible/cmd"
	"github.com/localghost/docksible/builder"
	"github.com/localghost/docksible/product"
)

func main() {
	if err := cmd.CreateRootCommand().Execute(); err != nil {
		log.Fatal(err)
	}
	builder := builder.New()
	builder.Bootstrap()
	builder.ProvisionContainer(product.New().Run())
}

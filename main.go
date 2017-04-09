package main

import (
	"github.com/localghost/docksible/builder"
	"github.com/localghost/docksible/docker"
	"github.com/localghost/docksible/product"
)

func main() {
	builder := builder.New()
	builder.Bootstrap()
	containerId := product.New().Run()
	builder.ProvisionContainer(docker.NewContainer(containerId, nil))
}

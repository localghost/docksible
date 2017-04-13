package main

import (
	"github.com/localghost/docksible/cmd"
	//"github.com/localghost/docksible/builder"
	//"github.com/localghost/docksible/product"
	"fmt"
	"os"
)

func main() {
	if err := cmd.CreateRootCommand().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//builder := builder.New()
	//builder.Bootstrap()
	//builder.ProvisionContainer(product.New().Run())
}

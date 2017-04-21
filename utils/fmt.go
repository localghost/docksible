package utils

import "fmt"

func Println(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

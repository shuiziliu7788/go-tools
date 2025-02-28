package main

import (
	"fmt"
	"github.com/shuiziliu7788/go-tools/utils"
)

func main() {
	dialer, err := utils.LoadLocalDialer(true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(dialer)
}

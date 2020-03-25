package main

import (
	"fmt"
	"github.com/atotto/clipboard"
	"testing"
)

func Test0(t *testing.T) {
	msg, err := clipboard.ReadAll()
	fmt.Println(msg,err)
}
package main

import "fmt"

type BaseConfig struct {
}

func (c *BaseConfig) Init(s string) {
	fmt.Println("BaseConfig", "-", s)
}

type ChildConfig struct {
	*BaseConfig
}

func (c *ChildConfig) Init(s string) {
	fmt.Println("ChildConfig", "-", s)
}

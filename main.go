package main

import "fmt"

func main() {
	list := make([]string, 10)

	list[0] = "hebert santos"

	user := make(map[string]string, 1)

	user["nome"] = "hebert"

	fmt.Println(user["nome"])
}

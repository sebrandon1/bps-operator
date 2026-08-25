package main

import (
	"fmt"
	"sort"

	"github.com/redhat-best-practices-for-k8s/checks"
	checksall "github.com/redhat-best-practices-for-k8s/checks/all"
)

func main() {
	checksall.Register()
	names := make([]string, 0)
	for _, c := range checks.All() {
		names = append(names, c.Name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Println(name)
	}
}

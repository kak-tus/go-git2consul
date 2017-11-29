package main

import (
	"fmt"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

var data = `
---
a: 1
b: 2
`

func main() {
	path := os.Getenv("G2C_PATH")
	keys := buildTree(path)
	fmt.Println(keys)
}

func buildTree(path string) map[string]interface{} {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	stat, err := file.Stat()
	if err != nil {
		panic(err)
	}

	tree := make(map[string]interface{})

	if stat.IsDir() == false {
		if filepath.Ext(file.Name()) == ".yml" {
			err = yaml.Unmarshal([]byte(data), &tree)
			if err != nil {
				panic(err)
			}
		}

		return tree
	}

	names, err := file.Readdirnames(0)
	for _, v := range names {
		val := buildTree(filepath.Join(path, v))
		if len(val) > 0 {
			ext := filepath.Ext(v)
			name := v[0 : len(v)-len(ext)]

			tree[name] = val
		}
	}

	return tree
}

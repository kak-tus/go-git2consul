package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"

	yaml "gopkg.in/yaml.v2"
)

func main() {
	path := os.Getenv("G2C_PATH")

	tree := buildTree(path)
	keys := genKeys("", tree)
	for k, v := range keys {
		fmt.Println(k)
		fmt.Println(v)
	}
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
			var data = make([]byte, stat.Size())
			_, err := file.Read(data)
			if err != nil {
				panic(err)
			}

			err = yaml.Unmarshal(data, &tree)
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

func genKeys(path string, tree interface{}) map[string]string {
	var res = make(map[string]string)

	treeVal := reflect.ValueOf(tree)

	switch reflect.TypeOf(tree).Kind() {
	case reflect.String:
		res[path] = treeVal.String()
	case reflect.Int:
		res[path] = strconv.FormatInt(treeVal.Int(), 10)
	case reflect.Map:
		for _, k := range treeVal.MapKeys() {
			subPath := path + "/" + k.Interface().(string)
			subVal := treeVal.MapIndex(k).Interface()
			subRes := genKeys(subPath, subVal)
			if len(subRes) > 0 {
				for k, v := range subRes {
					res[k] = v
				}
			}
		}
	}

	return res
}

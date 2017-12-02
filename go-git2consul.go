package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	consul "github.com/hashicorp/consul/api"
	git "gopkg.in/src-d/go-git.v4"
	yaml "gopkg.in/yaml.v2"
)

var kv *consul.KV
var logger = log.New(os.Stdout, "", log.LstdFlags)
var errLogger = log.New(os.Stderr, "", log.LstdFlags)
var repo *git.Repository

func main() {
	initKV()

	dir := clone()
	defer os.RemoveAll(dir)

	update(dir)

	period, err := strconv.Atoi(os.Getenv("G2C_PERIOD"))
	if err != nil {
		errLogger.Println(err)
		period = 300
	}

	timer := time.Tick(time.Duration(period) * time.Second)
	for range timer {
		update(dir)
	}
}

func clone() string {
	dir, err := ioutil.TempDir("", "g2c")
	if err != nil {
		errLogger.Panicln(err)
	}

	repoURL := os.Getenv("G2C_REPO")

	repo, err = git.PlainClone(dir, false, &git.CloneOptions{URL: repoURL})
	if err != nil {
		errLogger.Panicln(err)
	}

	return dir
}

func buildTree(path string) map[string]interface{} {
	file, err := os.Open(path)
	if err != nil {
		errLogger.Panicln(err)
	}

	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		errLogger.Panicln(err)
	}

	tree := make(map[string]interface{})

	if stat.IsDir() == false {
		if filepath.Ext(file.Name()) == ".yml" {
			var data = make([]byte, stat.Size())
			_, err := file.Read(data)
			if err != nil {
				errLogger.Panicln(err)
			}

			err = yaml.Unmarshal(data, &tree)
			if err != nil {
				errLogger.Panicln(err)
			}
		}

		return tree
	}

	names, err := file.Readdirnames(0)
	for _, v := range names {
		val := buildTree(filepath.Join(path, v))
		if len(val) == 0 {
			continue
		}

		ext := filepath.Ext(v)
		name := v[0 : len(v)-len(ext)]

		tree[name] = val
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
	case reflect.Float64:
		res[path] = strconv.FormatFloat(treeVal.Float(), 'f', -1, 64)
	case reflect.Map:
		for _, k := range treeVal.MapKeys() {
			subPath := k.Interface().(string)
			if len(path) > 0 {
				subPath = path + "/" + subPath
			}
			subVal := treeVal.MapIndex(k).Interface()

			subRes := genKeys(subPath, subVal)

			if len(subRes) == 0 {
				continue
			}

			for k, v := range subRes {
				res[k] = v
			}
		}
	default:
		errLogger.Println("Unsupported type: ", reflect.TypeOf(tree).Kind())
	}

	return res
}

func initKV() {
	client, err := consul.NewClient(consul.DefaultConfig())
	if err != nil {
		errLogger.Panicln(err)
	}

	kv = client.KV()
}

func getCurrentKeys() map[string]string {
	res := make(map[string]string)

	key := os.Getenv("G2C_TARGET")

	list, _, err := kv.List(key, nil)
	if err != nil {
		errLogger.Panicln(err)
	}

	for _, v := range list {
		if strings.HasSuffix(v.Key, "/") {
			continue
		}

		res[v.Key] = string(v.Value)
	}

	return res
}

func deleteKeys(keys map[string]string, currKeys map[string]string) {
	for k := range currKeys {
		_, exists := keys[k]
		if !exists {
			logger.Println("Delete " + k)
			_, err := kv.Delete(k, nil)
			if err != nil {
				errLogger.Println(err)
			}
		}
	}
}

func addKeys(keys map[string]string, currKeys map[string]string) {
	for k := range keys {
		_, exists := currKeys[k]
		if !exists {
			logger.Println("Add " + k)
			put := &consul.KVPair{Key: k, Value: []byte(keys[k])}
			_, err := kv.Put(put, nil)
			if err != nil {
				errLogger.Println(err)
			}
		}
	}
}

func updateKeys(keys map[string]string, currKeys map[string]string) {
	for k := range keys {
		_, exists := currKeys[k]
		if exists && currKeys[k] != keys[k] {
			logger.Println("Update " + k)
			put := &consul.KVPair{Key: k, Value: []byte(keys[k])}
			_, err := kv.Put(put, nil)
			if err != nil {
				errLogger.Println(err)
			}
		}
	}
}

func update(dir string) {
	wtree, err := repo.Worktree()
	if err != nil {
		errLogger.Panicln(err)
	}

	err = wtree.Pull(&git.PullOptions{})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		errLogger.Panicln(err)
	}

	tree := buildTree(dir)
	keys := genKeys("", tree)

	currKeys := getCurrentKeys()

	deleteKeys(keys, currKeys)
	addKeys(keys, currKeys)
	updateKeys(keys, currKeys)
}

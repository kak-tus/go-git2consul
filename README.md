# git2consul written on Go

Original git2consul is not suitable in my case: in "expand_keys_diff" mode it stole keys in K\V. Without this mode - it rewrite all Consul K\V tree, so services, based on consul-template reloaded frequently.

This go-git2consul so simple, so it can: it get git repository name, clone it and put all yaml files (add/update/delete keys) from it to Consul in mode:

```
"expand_keys": true
"expand_keys_diff": true,
"include_branch_name": false,
"ignore_repo_name": true,
"ignore_file_extension": true,
```

## Configuration with environment variables

CONSUL_HTTP_ADDR - Consul HTTP API

Example

```
CONSUL_HTTP_ADDR=consul.service.consul:8500
```

G2C_REPO - repo with configs

Example

```
G2C_REPO=https://github.com/kak-tus/go-git2consul.git
```

G2C_TARGET - limit top level key in Consul. All keys, under this key will be deleted if not found in config files.

Example

```
G2C_TARGET=config
```

G2C_PERIOD - period to update repository and update Consul

```
G2C_PERIOD=300
```

## Test run

```
docker run --rm -e CONSUL_HTTP_ADDR=<your consul> -e G2C_REPO=https://github.com/kak-tus/go-git2consul.git -e G2C_TARGET=config-example kaktuss/go-git2consul
```

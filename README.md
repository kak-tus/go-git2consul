git2consul written on Go.

Original git2consul is not suitable in my case: in "expand_keys_diff" mode it stole keys in K\V. Without this mode - it rewrite all Consul K\V tree, so services, based on consul-template reloaded frequently.

This go-git2consul so simple, so it can: it get git repository name, clone it and put all yaml files from it to Consul in mode:

```
"expand_keys": true
"expand_keys_diff": true,
"include_branch_name": false,
"ignore_repo_name": true,
"ignore_file_extension": true,
```

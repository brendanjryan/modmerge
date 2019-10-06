`modmerge`
---

A command line tool for merging multiple `go.mod` files into a single file. 

Useful when combining projects or making an "uber module" for projects which use the [multiple module repository](https://github.com/golang/go/wiki/Modules#what-are-multi-module-repositories) pattern.

## Installation

```console
go get -u github.com/brendanjryan/modmerge
```

## Usage

Default:
```console
modmerge <go.mod files...>
``` 

To specify an output file:

```console
modmerge -o go.mod.new go.mod services/user/go.mod services/tweets/go.mod
```


## Example 

```console
$ modmerge go.mod project/go.mod
2019/10/06 14:34:29 reading module files...
2019/10/06 14:34:29 reading file: go.mod
2019/10/06 14:34:29 reading file: project/go.mod
2019/10/06 14:34:29 merging module files...
2019/10/06 14:34:29 writng final result to go.mod.new...
2019/10/06 14:34:29 Successfully wrote merged modules to go.mod.new
```

```console
$ modmerge -o go.mod.merged go.mod project/go.mod
2019/10/06 14:34:29 reading module files...
2019/10/06 14:34:29 reading file: go.mod
2019/10/06 14:34:29 reading file: project/go.mod
2019/10/06 14:34:29 merging module files...
2019/10/06 14:34:29 writng final result to go.mod.merged...
2019/10/06 14:34:29 Successfully wrote merged modules to go.mod.merged
```

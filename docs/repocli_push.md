## repocli push

Push file to remote

### Synopsis

The `lpath` parameter can be '-' to read from stdin.

```
repocli push lpath rpath [flags]
```

### Examples

```
# Push a file to remote
repocli push local/path/a.txt remote/path/a.txt

# Upload data from stdin
repocli push - remote/path/somefile.txt
```

### Options

```
  -h, --help   help for push
```

### Options inherited from parent commands

```
  -c, --conf string   config file (default "/etc/repocli/repocli.conf")
```

### SEE ALSO

* [repocli](repocli.md)	 - `repocli` is a command line tool for managing remote storages.


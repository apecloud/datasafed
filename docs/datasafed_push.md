## datasafed push

Push file to remote

### Synopsis

The `lpath` parameter can be '-' to read from stdin.

```
datasafed push lpath rpath [flags]
```

### Examples

```
# Push a file to remote
datasafed push local/path/a.txt remote/path/a.txt

# Upload data from stdin
datasafed push - remote/path/somefile.txt
```

### Options

```
  -h, --help   help for push
```

### Options inherited from parent commands

```
  -c, --conf string   config file (default "/etc/datasafed/datasafed.conf")
```

### SEE ALSO

* [datasafed](datasafed.md)	 - `datasafed` is a command line tool for managing remote storages.


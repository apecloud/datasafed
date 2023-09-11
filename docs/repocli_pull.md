## repocli pull

Pull remote file

### Synopsis

The `lpath` parameter can be "-" to write to stdout.

```
repocli pull rpath lpath [flags]
```

### Examples

```
# Pull the file and save it to a local path
repocli pull some/path/file.txt /tmp/file.txt

# Pull the file and print it to stdout
repocli pull some/path/file.txt - | wc -l
```

### Options

```
  -h, --help   help for pull
```

### Options inherited from parent commands

```
  -c, --conf string   config file (default "/etc/repocli/repocli.conf")
```

### SEE ALSO

* [repocli](repocli.md)	 - `repocli` is a command line tool for managing remote storages.


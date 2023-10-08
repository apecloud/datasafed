## datasafed pull

Pull remote file

### Synopsis

The `lpath` parameter can be "-" to write to stdout.

```
datasafed pull rpath lpath [flags]
```

### Examples

```
# Pull the file and save it to a local path
datasafed pull some/path/file.txt /tmp/file.txt

# Pull the file and print it to stdout
datasafed pull some/path/file.txt - | wc -l
```

### Options

```
  -h, --help   help for pull
```

### Options inherited from parent commands

```
  -c, --conf string   config file (default "/etc/datasafed/datasafed.conf")
```

### SEE ALSO

* [datasafed](datasafed.md)	 - `datasafed` is a command line tool for managing remote storages.


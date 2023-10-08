## datasafed stat

Stat a remote path to get the total size and number of entries.

### Synopsis

It counts files and dirs in the path and calculates the total size recursively.

```
datasafed stat [--json] rpath [flags]
```

### Examples

```
# Stat a file
datasafed stat path/to/file.txt

# Stat a directory with json output
datasafed stat -json path/to/dir
```

### Options

```
  -h, --help   help for stat
      --json   output in json format
```

### Options inherited from parent commands

```
  -c, --conf string   config file (default "/etc/datasafed/datasafed.conf")
```

### SEE ALSO

* [datasafed](datasafed.md)	 - `datasafed` is a command line tool for managing remote storages.


## repocli stat

Stat a remote path to get the total size and number of entries.

### Synopsis

It counts files and dirs in the path and calculates the total size recursively.

```
repocli stat [--json] rpath [flags]
```

### Examples

```
# Stat a file
repocli stat path/to/file.txt

# Stat a directory with json output
repocli stat -json path/to/dir
```

### Options

```
  -h, --help   help for stat
      --json   output in json format
```

### Options inherited from parent commands

```
  -c, --conf string   config file (default "/etc/repocli/repocli.conf")
```

### SEE ALSO

* [repocli](repocli.md)	 - `repocli` is a command line tool for managing remote storages.


## repocli rm

Remove one remote file, or all files in a remote directory.

```
repocli rm [-r] rpath [flags]
```

### Examples

```
# Remove a single file
repocli rm some/path/to/file.txt

# Recursively remove a directory
repocli rm -r some/path/to/dir
```

### Options

```
  -h, --help        help for rm
  -r, --recursive   remove recursively
```

### Options inherited from parent commands

```
  -c, --conf string   config file (default "/etc/repocli/repocli.conf")
```

### SEE ALSO

* [repocli](repocli.md)	 - `repocli` is a command line tool for managing remote storages.


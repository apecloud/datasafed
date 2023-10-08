## datasafed rm

Remove one remote file, or all files in a remote directory.

```
datasafed rm [-r] rpath [flags]
```

### Examples

```
# Remove a single file
datasafed rm some/path/to/file.txt

# Recursively remove a directory
datasafed rm -r some/path/to/dir
```

### Options

```
  -h, --help        help for rm
  -r, --recursive   remove recursively
```

### Options inherited from parent commands

```
  -c, --conf string   config file (default "/etc/datasafed/datasafed.conf")
```

### SEE ALSO

* [datasafed](datasafed.md)	 - `datasafed` is a command line tool for managing remote storages.


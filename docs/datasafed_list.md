## datasafed list

List contents of a remote directory or file.

```
datasafed list [-d|-f] [-r] [--max-depth depth] [-s sortBy] [--reverse] [--newer-than time] [--older-than time] [--name pattern] [-o outputFormat] rpath [flags]
```

### Examples

```
# List the root directory
datasafed list /

# List one file and extract its size
datasafed list somefile.txt -o long | awk '{print $2}'

# List all files under the directory
datasafed list -r -f /some/dir

# List files modified within 1 hour and sort the result by size
datasafed list -r -f -s size --newer-than $(( $(date +%s) - 3600 )) /some/dir

# List files with the name pattern
datasafed list --name "*.txt" /some/dir
```

### Options

```
  -d, --dirs-only              list directories only
  -f, --files-only             list files only
  -h, --help                   help for list
      --max-depth int          max depth when listing recursively
      --name string            list only entries whose name matches the specified pattern (https://pkg.go.dev/path/filepath#Match)
      --newer-than int         list only entries whose last modification time is newer than the specified unix timestamp (exclusive)
      --older-than int         list only entries whose last modification time is older than the specified unix timestamp (exclusive)
  -o, --output-format string   output format, choices: ["short" "long" "json"] (default "short")
  -r, --recursive              list recursively
      --reverse                reverse order
  -s, --sort string            sort by which field, choices: ["path" "size" "mtime"] (default "path")
```

### Options inherited from parent commands

```
  -c, --conf string   config file (default "/etc/datasafed/datasafed.conf")
```

### SEE ALSO

* [datasafed](datasafed.md)	 - `datasafed` is a command line tool for managing remote storages.


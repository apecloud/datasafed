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

# List all files under the directory (ends with '/')
datasafed list -r -f /some/dir/

# List files modified within 1 hour and sort the result by size
datasafed list -r -f -s size --newer-than $(( $(date +%s) - 3600 )) /some/dir/

# List files with the name pattern
datasafed list --name "*.txt" /some/dir/
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
  -s, --sort string            sort by which field, choices: ["path" "size" "mtime"], this option conflicts with --recursive
```

### Options inherited from parent commands

```
  -c, --conf string                       config file (default "/etc/datasafed/datasafed.conf")
      --console-log                       Enable console log
      --console-timestamps                Log timestamps to stderr. (default true)
      --disable-color                     Disable color output
      --file-log-level string             File log level (default "debug")
      --file-log-local-tz                 When logging to a file, use local timezone
      --force-color                       Force color output
      --json-log-console                  JSON log file
      --json-log-file                     JSON log file
      --log-dir string                    Directory where log files should be written.
      --log-dir-max-age duration          Maximum age of log files to retain (default 720h0m0s)
      --log-dir-max-files int             Maximum number of log files to retain (default 100)
      --log-dir-max-total-size-mb float   Maximum total size of log files to retain (default 1000)
      --log-file string                   Override log file.
      --log-level string                  Console log level (default "info")
      --max-log-file-segment-size int     Maximum size of a single log file segment (default 50000000)
```

### SEE ALSO

* [datasafed](datasafed.md)	 - `datasafed` is a command line tool for managing remote storages.


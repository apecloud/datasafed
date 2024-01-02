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
  -z, --compress string   compress the file using the specified algorithm before sending it to remote, choices: ["deflate-best-compression" "deflate-best-speed" "deflate-default" "gzip" "gzip-best-compression" "gzip-best-speed" "lz4" "pgzip" "pgzip-best-compression" "pgzip-best-speed" "s2-better" "s2-default" "s2-parallel-4" "s2-parallel-8" "zstd" "zstd-best-compression" "zstd-better-compression" "zstd-fastest"]
  -h, --help              help for push
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


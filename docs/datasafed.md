## datasafed

`datasafed` is a command line tool for managing remote storages.

### Options

```
  -c, --conf string                       config file (default "/etc/datasafed/datasafed.conf")
      --console-log                       Enable console log
      --console-timestamps                Log timestamps to stderr. (default true)
      --disable-color                     Disable color output
      --file-log-level string             File log level (default "debug")
      --file-log-local-tz                 When logging to a file, use local timezone
      --force-color                       Force color output
  -h, --help                              help for datasafed
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

* [datasafed getconf](datasafed_getconf.md)	 - Get the value of the configuration item.
* [datasafed list](datasafed_list.md)	 - List contents of a remote directory or file.
* [datasafed mkdir](datasafed_mkdir.md)	 - Create an empty remote directory.
* [datasafed pull](datasafed_pull.md)	 - Pull remote file
* [datasafed push](datasafed_push.md)	 - Push file to remote
* [datasafed rm](datasafed_rm.md)	 - Remove one remote file, or all files in a remote directory.
* [datasafed rmdir](datasafed_rmdir.md)	 - Remove an empty remote directory.
* [datasafed stat](datasafed_stat.md)	 - Stat a remote path to get the total size and number of entries.
* [datasafed version](datasafed_version.md)	 - Show version of datasafed.


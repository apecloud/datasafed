## repocli getconf

Get the value of the configuration item.

### Synopsis

The pattern of the `item` parameter is "section.field".

```
repocli getconf item [flags]
```

### Examples

```
# get the "type" field from the "storage" section
repocli getconf storage.type

# get access_key_id (only available for S3 backend)
repocli getconf storage.access_key_id
```

### Options

```
  -h, --help   help for getconf
```

### Options inherited from parent commands

```
  -c, --conf string   config file (default "/etc/repocli/repocli.conf")
```

### SEE ALSO

* [repocli](repocli.md)	 - `repocli` is a command line tool for managing remote storages.


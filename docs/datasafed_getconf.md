## datasafed getconf

Get the value of the configuration item.

### Synopsis

The pattern of the `item` parameter is "section.field".

```
datasafed getconf item [flags]
```

### Examples

```
# get the "type" field from the "storage" section
datasafed getconf storage.type

# get access_key_id (only available for S3 backend)
datasafed getconf storage.access_key_id
```

### Options

```
  -h, --help   help for getconf
```

### Options inherited from parent commands

```
  -c, --conf string   config file (default "/etc/datasafed/datasafed.conf")
```

### SEE ALSO

* [datasafed](datasafed.md)	 - `datasafed` is a command line tool for managing remote storages.


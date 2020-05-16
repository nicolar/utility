# Utility
Various utilities

## cups_filewrite

Simple backend for CUPS that write the printout to a simple file.

### Setup
- Create a new raw printer in CUPS
- Set the URI as: ```filewrite:/home/folder/prefix```

This will write files as:
- ```/home/folder/prefix-ddmmYYYY-HHMMSS``` where `ddmmYYYY` is a date and `HHMMSS` is a time

## smartftp

Enhanced FTP send/receive files script

This scripts send and get a pattern of files handling proper file rename to permits complete tranfers (avoid application to read files before tranfer ends)

```
Usage: smartftp.pl --action [put|get] --host <ftphost> [--port <ftpport> ]
                   [ --username <username> --password <password> ]
                   [ --rename_prefix <rename_prefix> --rename_suffix <rename_suffix>]
                   --file <file1> --file <file2>... --remote_dir <remote_dir>
```
checksum
========

Utility to store sh256 hashes of files in dedicated "database"(actually just a JSON file) to
verify it later as a part of consistency check.

Usage
-----

```
Usage: ./bin/checksum-darwin-amd64 [OPTION]...
OPTIONS:
  -concurrency <int>
    Amount of routines to spawn at the same time for checksum verification(8 by default for your system)
  -database <string>
    Specify database path
  -datadir <string>
    Specify data directory
  -pattern <string>
    Pattern to match files in filewalk mode(default is `.(ar2|arw|cr2|crw|nef)$`)
  -version
    Print checksum version

Examples:
  ./bin/checksum-darwin-amd64 -database /tmp/db.json -datadir /Volumes/Storage/Photos
```

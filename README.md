checksum
========

Utility to store sha256 hashes of files in dedicated "database"(actually just a JSON file) to
verify it later as a part of consistency check with automatic new file indexing.

Why not shasum/md5sum/etc.?
---------------------------

checksum provides straight workflow for verification and adding new files processes
to avoid remembering someting like `find $dir | xargs md5sum >> /tmp/database.txt`.

checkum automatically:

 * verifies files
 * adds new
 * report about fails and misses

How to install
--------------

macOS with Homebrew:
```
brew install https://raw.githubusercontent.com/teran/checksum/master/contrib/Homebrew/checksum.rb
```

Other distros:

just refer to releases page and download appropriate binary for your platform

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

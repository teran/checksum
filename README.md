checksum
========

Utility to store length, sha1, sha256 hashes of files in dedicated "database"(actually just a JSON file) to
verify it later as a part of consistency check with automatic new file indexing.

Usage
-----

```man
Utility to verify files consistency with length, SHA1 and SHA256

Usage:
    checksum [FLAG]...

Version:
    version: 7d4c766
    commit: 7d4c766
    built with: go1.12.5
    built at: 2019-05-14T23:51:55Z

Description:
    checksum creates database (actually just a JSON file) to store file length, SHA1, SHA256
    to verify file consistency and report if something goes wrong.

Flags:
    --concurrency, -c        Amount of routines to spawn at the same time for checksum verification (type: int)
        Default value is 8 for your system
    --complete               Completion for shell (type: bool)
    --datadir, -d            Data directory path to run new files scan (type: string)
    --database, -D           Database file path (required) (type: string)
    --delete-missed          Delete missed files from database (type: bool)
    --generate-checksums-only Skip verification step and add new files only (type: bool)
    --pattern, -p            Pattern to match filenames which checking for new files (type: string)
        Default is `.(3fr|ari|arw|bay|crw|cr2|cr3|cap|data|dcs|dcr|drf|eip|erf|fff|gpr|iiq|k25|kdc|mdc|mef|mos|mrw|nef|nrw|obm|orf|pef|ptx|pxn|r3d|raf|raw|rwl|rw2|rwz|sr2|srf|srw|x3f)$`
    --skip-failed, --sf      Skip FAIL verification results from output (type: bool)
    --skip-missed, --sm      Skip MISS verification results from output (type: bool)
    --skip-ok, --so          Skip OK verification results from output (type: bool)
    --progressbar            Show progress bar instead of printing handled files (type: bool)
    --version, -V            Print application and Golang versions (type: bool)
    -h, --help               show help (type: bool)
```


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

Just refer to [releases page](https://github.com/teran/checksum/releases) and download appropriate binary for your platform or build your own one right from master.

Build
-----

System-wide requirements:

* Go

Build:

```bash
make predependencies build
```

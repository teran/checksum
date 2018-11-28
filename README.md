checksum
========

Utility to store length, sha1, sha256 hashes of files in dedicated "database"(actually just a JSON file) to
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

```bash
brew install https://raw.githubusercontent.com/teran/checksum/master/contrib/Homebrew/checksum.rb
```

or to upgrade:

```bash
brew upgrade https://raw.githubusercontent.com/teran/checksum/master/contrib/Homebrew/checksum.rb
```

Other distros:

just refer to releases page and download appropriate binary for your platform

Usage
-----

```man
Utility to verify files consistency with length, SHA1 and SHA256

Usage:
      checksum-darwin-amd64 [FLAG]...

Version:
      version: 98029e9
      built with: go1.11.2

Description:
      checksum creates database (actually just a JSON file) to store file length, SHA1, SHA256
      to verify file consistency and report if something goes wrong.

Flags:
      --concurrency, -c        Amount of routines to spawn at the same time for checksum verification
            Default value is 8 for your system
      --complete               Completion for shell
      --datadir, -d            Data directory path to run new files scan
      --database, -D           Database file path (required)
      --generate-checksums-only Skip verification step and add new files only
      --pattern, -p            Pattern to match filenames which checking for new files
            Default is `.(3fr|ari|arw|bay|crw|cr2|cr3|cap|data|dcs|dcr|drf|eip|erf|fff|gpr|iiq|k25|kdc|mdc|mef|mos|mrw|nef|nrw|obm|orf|pef|ptx|pxn|r3d|raf|raw|rwl|rw2|rwz|sr2|srf|srw|x3f)$`
      --skip-failed, --sf      Skip FAIL verification results from output
      --skip-missed, --sm      Skip MISS verification results from output
      --skip-ok, --so          Skip OK verification results from output
      --progressbar            Show progress bar instead of printing handled files
      --version, -V            Print application and Golang versions
      -h, --help               show help
```

Build
-----

System-wide requirements:

* Go

Build:

```bash
make predependencies build
```

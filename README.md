# ghlatest

This is a tool for downloading the latest release of a package from GitHub and optionally extracting the contents.

**NOTE**: This project is not yet ready for production use!

## Installation

This app is statically compiled for various platforms, binaries can be downloaded from the [ghlatest releases](https://github.com/backplane/ghlatest/releases) page.

**Note:** Your system or container will need to have PKI trust roots of some kind in order to run `ghlatest`. One common package that provides these on unix systems is called `ca-certificates`.

## Usage

This is the general help text produced by the program. Each command has additional help text available, you can access this text with a command-line like: `ghlatest list -h`

### General Help

```
$ ghlatest -h
NAME:
   ghlatest - Release locator for software on github

USAGE:
   ghlatest [global options] command [command options] [arguments...]

VERSION:
   dev

COMMANDS:
   list, ls      list available releases
   download, dl  download the latest available release
   json, j       print json doc representing latest release from github api
   extract, x    Extract files from the given archive (supports zip, gzip, bzip2, xz, 7z, and tar formats)
   help, h       Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --verbosity value  Sets the verbosity level of the log messages printed by the program, should be one of the following:
      "debug", "error", "fatal", "info", "panic", "trace", or "warn"
   --help, -h     show help
   --version, -v  print the version
```

### Download Help

```
NAME:
   ghlatest download - download the latest available release

USAGE:
   ghlatest download [command options] [arguments...]

OPTIONS:
   --filter value, -f value [ --filter value, -f value ]  Filter release assets with the given regular expression
   --ifilter value, -i value                              Filter release assets with the given CASE-INSENSITIVE regular expression
   --current-arch                                         Filter release assets with a regex describing the current processor architecture (default: false)
   --current-os                                           Filter release assets with a regex describing the current operating system (default: false)
   --source, -s                                           List/download source zip files instead of released assets (default: false)
   --outputpath value, -o value                           The name of the file to write to
   --mode value, -m value                                 Set the output file's protection mode (ala chmod) (default: "0755")
   --extract, -x                                          Extract files from the downloaded archive (supports zip, gzip, bzip2, xz, 7z, and tar formats) (default: false)
   --keep value, -k value [ --keep value, -k value ]      When extracting, only keep the files matching this/these regex(s)
   --overwrite                                            When extracting, if one of the output files already exists, overwrite it (default: false)
   --remove-archive, --rm                                 After extracting the archive, delete it (default: false)
   --help, -h                                             show help
```

### List Help

```
$ ghlatest list -h
NAME:
   ghlatest list - list available releases

USAGE:
   ghlatest list [command options] [arguments...]

OPTIONS:
   --filter value, -f value [ --filter value, -f value ]  Filter release assets with the given regular expression
   --ifilter value, -i value                              Filter release assets with the given CASE-INSENSITIVE regular expression
   --current-arch                                         Filter release assets with a regex describing the current processor architecture (default: false)
   --current-os                                           Filter release assets with a regex describing the current operating system (default: false)
   --source, -s                                           List/download source zip files instead of released assets (default: false)
   --help, -h                                             show help
```

### Extract Help

```
NAME:
   ghlatest extract - Extract files from the given archive (supports zip, gzip, bzip2, xz, 7z, and tar formats)

USAGE:
   ghlatest extract [command options] [arguments...]

OPTIONS:
   --outputpath value, -o value                       The name of the file to write to
   --mode value, -m value                             Set the output file's protection mode (ala chmod) (default: "0755")
   --keep value, -k value [ --keep value, -k value ]  When extracting, only keep the files matching this/these regex(s)
   --overwrite                                        When extracting, if one of the output files already exists, overwrite it (default: false)
   --remove-archive, --rm                             After extracting the archive, delete it (default: false)
   --help, -h                                         show help
```

## Example Session

I want to find out the what files are available in the latest release of the repo [`glvnst/snakeeyes`](https://github.com/glvnst/snakeeyes).

```
$ ghlatest ls glvnst/snakeeyes
https://github.com/glvnst/snakeeyes/releases/download/v0.2.3/checksums.txt
https://github.com/glvnst/snakeeyes/releases/download/v0.2.3/snakeeyes_0.2.3_dragonfly_amd64.tar.gz
https://github.com/glvnst/snakeeyes/releases/download/v0.2.3/snakeeyes_0.2.3_freebsd_amd64.tar.gz
https://github.com/glvnst/snakeeyes/releases/download/v0.2.3/snakeeyes_0.2.3_freebsd_armv7.tar.gz
https://github.com/glvnst/snakeeyes/releases/download/v0.2.3/snakeeyes_0.2.3_linux_amd64.tar.gz
https://github.com/glvnst/snakeeyes/releases/download/v0.2.3/snakeeyes_0.2.3_linux_arm64.tar.gz
https://github.com/glvnst/snakeeyes/releases/download/v0.2.3/snakeeyes_0.2.3_linux_armv7.tar.gz
https://github.com/glvnst/snakeeyes/releases/download/v0.2.3/snakeeyes_0.2.3_macOS_all.tar.gz
https://github.com/glvnst/snakeeyes/releases/download/v0.2.3/snakeeyes_0.2.3_macOS_amd64.tar.gz
https://github.com/glvnst/snakeeyes/releases/download/v0.2.3/snakeeyes_0.2.3_netbsd_amd64.tar.gz
https://github.com/glvnst/snakeeyes/releases/download/v0.2.3/snakeeyes_0.2.3_netbsd_armv7.tar.gz
https://github.com/glvnst/snakeeyes/releases/download/v0.2.3/snakeeyes_0.2.3_openbsd_amd64.tar.gz
https://github.com/glvnst/snakeeyes/releases/download/v0.2.3/snakeeyes_0.2.3_openbsd_arm64.tar.gz
https://github.com/glvnst/snakeeyes/releases/download/v0.2.3/snakeeyes_0.2.3_openbsd_armv7.tar.gz
https://github.com/glvnst/snakeeyes/releases/download/v0.2.3/snakeeyes_0.2.3_windows_amd64.zip
https://github.com/glvnst/snakeeyes/releases/download/v0.2.3/snakeeyes_0.2.3_windows_armv7.zip
```

I only care about the ones for my current operating system (linux) so I'll filter for those:

```
$ ghlatest ls --current-os glvnst/snakeeyes
https://github.com/glvnst/snakeeyes/releases/download/v0.2.3/snakeeyes_0.2.3_linux_amd64.tar.gz
https://github.com/glvnst/snakeeyes/releases/download/v0.2.3/snakeeyes_0.2.3_linux_arm64.tar.gz
https://github.com/glvnst/snakeeyes/releases/download/v0.2.3/snakeeyes_0.2.3_linux_armv7.tar.gz
```

I also only want the release for my current processor architecture so I apply that filter as well:

```
$ ghlatest ls --current-os --current-arch glvnst/snakeeyes
https://github.com/glvnst/snakeeyes/releases/download/v0.2.3/snakeeyes_0.2.3_linux_arm64.tar.gz
```

Now that I have that down to one URL I can change `ls` to `dl` to download the release, I also want to extract it, so I'll add the `--extract flag`:

```
$ ghlatest dl --current-os --current-arch --extract glvnst/snakeeyes
INFO[0001] wrote 825399 bytes to snakeeyes_0.2.3_linux_arm64.tar.gz 
INFO[0001] extracting (tgz) snakeeyes_0.2.3_linux_arm64.tar.gz 
INFO[0001] created COPYING mode: 0644                   
INFO[0001] created README.md mode: 0644                 
INFO[0001] created snakeeyes mode: 0755                 
INFO[0001] extraction complete
$ ls -al
total 3156
drwxr-xr-x    6 user     user           192 Feb 20 09:23 .
drwxr-xr-x   21 user     user           672 Feb 20 09:23 ..
-rw-r--r--    1 user     user         34523 Feb 20 09:23 COPYING
-rw-r--r--    1 user     user          7080 Feb 20 09:23 README.md
-rwxr-xr-x    1 user     user       2359296 Feb 20 09:23 snakeeyes
-rwxr-xr-x    1 user     user        825399 Feb 20 09:23 snakeeyes_0.2.3_linux_arm64.tar.gz
```

That produced a lot of files that I don't want at the moment. So I'll add a `--keep` filter to only extract the binary, and I'll also add `--rm` to remove the downloaded archive after I'm done with it.

```
$ ghlatest dl --current-os --current-arch --extract --rm --keep snakeeyes glvnst/snakeeyes
INFO[0001] wrote 825399 bytes to snakeeyes_0.2.3_linux_arm64.tar.gz 
INFO[0001] extracting (tgz) snakeeyes_0.2.3_linux_arm64.tar.gz 
INFO[0001] created snakeeyes mode: 0755                 
INFO[0001] extraction complete                          
INFO[0001] removed "snakeeyes_0.2.3_linux_arm64.tar.gz" after extraction
$ $ ls -al
total 2304
drwxr-xr-x    3 user     user            96 Feb 20 09:26 .
drwxr-xr-x   21 user     user           672 Feb 20 09:23 ..
-rwxr-xr-x    1 user     user       2359296 Feb 20 09:26 snakeeyes
```

Now we have a command which produces the file that I want from the latest release of the given GitHub repo, it can be used in scripting contexts or in container infrastructure, such as this `Dockerfile`:

```Dockerfile
FROM backplane/ghlatest as downloader
RUN ghlatest dl --current-os --current-arch --extract --rm --keep snakeeyes glvnst/snakeeyes

FROM alpine:3
COPY --from=downloader /work/snakeeyes /bin/
ENTRYPOINT ["/bin/snakeeyes"]
```
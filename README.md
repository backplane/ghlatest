# WIP: ghlatest

**NOTE**: This project is not yet ready for production use!

This is a tool for downloading the latest release of a package from github and optionally extracting the contents.

## Installation

TBD This app is written in go and cross-compiled for various common platforms, probably downloading a binary release is the right strategy.

**Note:** Your system will need to have PKI trust roots of some kind in order to run ghlatest. One common package that provides these on unix systems is called `ca-certificates`.

## Usage

This is the general help test from the program. Each command has additional help text available, you can access this text with a command-line like: `ghlatest list -h`

```
$ ghlatest -h
NAME:
   ghlatest - Release locator for software on github

USAGE:
   ghlatest [global options] command [command options] [arguments...]

VERSION:
   v0.1.1

COMMANDS:
     list, l, ls      list available releases
     download, d, dl  download the latest available release
     json, j          print json doc representing latest release from github api
     help, h          Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

See the help test for each command, for example:

```
$ ghlatest download -h
NAME:
   ghlatest download - download the latest available release

USAGE:
   ghlatest download [command options] [arguments...]

OPTIONS:
   --filter value, -f value      Filter release assets with the given regular expression (default: "^")
   --source, -s                  List/download source zip files instead of released assets
   --outputpath value, -o value  The name of the file to write to
   --mode value, -m value        Set the output file's protection mode (ala chmod) (default: "0755")
   --extract, -x                 Unzip the downloaded file
   
```

## Changelog

### v0.1.5

* update deps

### v0.1.4

* release infrastructure updates, minor documentation tweaks

### v0.1.3

* release infrastructure updates, minor documentation tweaks

### v0.1.2

* release infrastructure updates, minor documentation tweaks

### v0.1.1

* argument change:

  * `chmod` -> `mode` - preferring mode to chmod to more accurately reflect that the file is being created with these permissions
  * `json` - add option
  * `source` - add option to include source zip file listing / downloading
  * `extract` - add download option to extract downloaded archives
  * `namefilter` -> `filter` - preferring shorter argument name
  * `outputfile` -> `outputpath` - preferring this because with extract in place the output could be a directory


### v0.1.0

* Fixes error handling bug in latestReleasedAssets - removing null pointer deref
* Adds README note about CA certificates requirement


## Type sniffing

### zipball

`Content-Type: application/zip`

see here

```
> GET /kolide/launcher/legacy.zip/0.5.0 HTTP/1.1
> Host: codeload.github.com
> User-Agent: curl/7.54.0
> Accept: */*
> 
< HTTP/1.1 200 OK
< Transfer-Encoding: chunked
< Access-Control-Allow-Origin: https://render.githubusercontent.com
< Content-Security-Policy: default-src 'none'; style-src 'unsafe-inline'; sandbox
< Strict-Transport-Security: max-age=31536000
< Vary: Authorization,Accept-Encoding
< X-Content-Type-Options: nosniff
< X-Frame-Options: deny
< X-XSS-Protection: 1; mode=block
< ETag: "6f697beacb9e530a97376c9eed45c9013564e946"
< Content-Type: application/zip
< Content-Disposition: attachment; filename=kolide-launcher-0.5.0-0-g6f697be.zip
< X-Geo-Block-List: 
< Date: Sun, 01 Jul 2018 06:47:12 GMT
< X-GitHub-Request-Id: C61E:0A11:113231:28BA33:5B3878F0

```


## Download Logic

### Filenames

Note that the content-disposition includes a filename!

for source:

`Content-Disposition: attachment; filename=kolide-launcher-0.5.0-0-g6f697be.zip`

for released files:

`Content-Disposition: attachment; filename=launcher_0.5.0.zip`

See <https://golang.org/pkg/mime/multipart/#Part.FileName> for how to parse that correctly

See <https://tools.ietf.org/html/rfc6266> for the standard "Use of the Content-Disposition Header Field in the Hypertext Transfer Protocol (HTTP)"

### Extraction


 Download Contents                            | Action
----------------------------------------------|-----------
 single uncompressed file direct download     | writeFileWithNameAndMode()
 single file at root of archive               | writeFileWithNameAndMode(extract())
 single dir at root of archive                | writeDirWithNameAndMode(extract())
 multiple files/dirs at root of archive       | mkdirWithNameAndMode()

questions to answer at runtime

is the download actually compressed? 

is output path a file or a directory?

do we need to create the output directory?

does the zip contain one file or more than one file?



new idea for extraction logic:

* does the outputpath argument end in a slash?

    * if so, always write files into a directory
    * if the directory exists, write into it



## Repos with unusual behavior

* wikimedia/mediawiki - no releases, working with tags https://api.github.com/repos/wikimedia/mediawiki/tags
* ether/etherpad-lite - no assets released

## Github API Releases Endpoint

This app uses the github API's releases endpoint:

`$ curl https://api.github.com/repos/kolide/launcher/releases/latest`

The result looks like this:

```json
{
  "url": "https://api.github.com/repos/kolide/launcher/releases/9880525",
  "assets_url": "https://api.github.com/repos/kolide/launcher/releases/9880525/assets",
  "upload_url": "https://uploads.github.com/repos/kolide/launcher/releases/9880525/assets{?name,label}",
  "html_url": "https://github.com/kolide/launcher/releases/tag/0.5.0",
  "id": 9880525,
  "node_id": "MDc6UmVsZWFzZTk4ODA1MjU=",
  "tag_name": "0.5.0",
  "target_commitish": "master",
  "name": "0.5.0 ",
  "draft": false,
  "author": {
    "login": "groob",
    "id": 1526945,
    "node_id": "MDQ6VXNlcjE1MjY5NDU=",
    "avatar_url": "https://avatars3.githubusercontent.com/u/1526945?v=4",
    "gravatar_id": "",
    "url": "https://api.github.com/users/groob",
    "html_url": "https://github.com/groob",
    "followers_url": "https://api.github.com/users/groob/followers",
    "following_url": "https://api.github.com/users/groob/following{/other_user}",
    "gists_url": "https://api.github.com/users/groob/gists{/gist_id}",
    "starred_url": "https://api.github.com/users/groob/starred{/owner}{/repo}",
    "subscriptions_url": "https://api.github.com/users/groob/subscriptions",
    "organizations_url": "https://api.github.com/users/groob/orgs",
    "repos_url": "https://api.github.com/users/groob/repos",
    "events_url": "https://api.github.com/users/groob/events{/privacy}",
    "received_events_url": "https://api.github.com/users/groob/received_events",
    "type": "User",
    "site_admin": false
  },
  "prerelease": false,
  "created_at": "2018-02-28T19:16:36Z",
  "published_at": "2018-02-28T19:20:40Z",
  "assets": [
    {
      "url": "https://api.github.com/repos/kolide/launcher/releases/assets/6357751",
      "id": 6357751,
      "node_id": "MDEyOlJlbGVhc2VBc3NldDYzNTc3NTE=",
      "name": "launcher_0.5.0.zip",
      "label": null,
      "uploader": {
        "login": "groob",
        "id": 1526945,
        "node_id": "MDQ6VXNlcjE1MjY5NDU=",
        "avatar_url": "https://avatars3.githubusercontent.com/u/1526945?v=4",
        "gravatar_id": "",
        "url": "https://api.github.com/users/groob",
        "html_url": "https://github.com/groob",
        "followers_url": "https://api.github.com/users/groob/followers",
        "following_url": "https://api.github.com/users/groob/following{/other_user}",
        "gists_url": "https://api.github.com/users/groob/gists{/gist_id}",
        "starred_url": "https://api.github.com/users/groob/starred{/owner}{/repo}",
        "subscriptions_url": "https://api.github.com/users/groob/subscriptions",
        "organizations_url": "https://api.github.com/users/groob/orgs",
        "repos_url": "https://api.github.com/users/groob/repos",
        "events_url": "https://api.github.com/users/groob/events{/privacy}",
        "received_events_url": "https://api.github.com/users/groob/received_events",
        "type": "User",
        "site_admin": false
      },
      "content_type": "application/zip",
      "state": "uploaded",
      "size": 24866298,
      "download_count": 658,
      "created_at": "2018-02-28T19:21:42Z",
      "updated_at": "2018-02-28T19:21:50Z",
      "browser_download_url": "https://github.com/kolide/launcher/releases/download/0.5.0/launcher_0.5.0.zip"
    }
  ],
  "tarball_url": "https://api.github.com/repos/kolide/launcher/tarball/0.5.0",
  "zipball_url": "https://api.github.com/repos/kolide/launcher/zipball/0.5.0",
  "body": "* Add a local debug server. (#187)\r\n--TRUNCATED FOR README DOCS --"
}
```



## Building

Run `make help` to get info on builds. The general process for building on your own is just to run `make`. 

```
$ make help
AUTHORS                        Generate the AUTHORS file from the git log
all                            Runs a clean, build, fmt, lint, test, staticcheck, vet and install
build                          Builds a dynamic executable or package (default target)
clean                          Cleanup any build binaries or packages
cover                          Runs go test with coverage
fmt                            Verifies all files have been `gofmt`ed
install                        Installs the executable or package
lint                           Verifies `golint` passes
release                        Build cross-compiled binaries for target architectures
static                         Builds a static executable
staticcheck                    Verifies `staticcheck` passes
tag                            Create a new git tag to prepare to build a release
test                           Runs the go tests
version-bump-major             Increment the major version number in VERSION.txt, e.g. v1.2.3 -> v2.2.3
version-bump-minor             Increment the minor version number in VERSION.txt, e.g. v1.2.3 -> v1.3.3
version-bump-patch             Increment the patch version number in VERSION.txt, e.g. v1.2.3 -> v1.2.4
vet                            Verifies `go vet` passes
```

### Releases

If you're maintaining your own releases you'll need to update `.travis.yml` with your `api_key` (see travis-ci docs, you'll need the travis CLI and you'll want to run `travis setup releases`). Then use the following process for a patch-level version update:

```
make version-bump-patch # this increments the appropriate number in VERSION.txt
# git add and push your changes to the branch
# merge the branch to master
git checkout master
make tag
# you will need to unlock your pgp keychain for signing purposes.
# run the "git push" command that is printed by make
```

At this point a tagged version release will be created on gitlab and travis-ci will automatically push the build artifacts to github.

## Resources

* <https://stackoverflow.com/a/50540591> - stackoverflow answer pointing to github api
* <https://developer.github.com/v3/repos/releases/#get-the-latest-release>
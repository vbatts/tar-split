# go-mtree

`mtree` is a filesystem hierarchy validation tooling and format.
This is a library and simple cli tool for [mtree(8)][mtree(8)] support.

While the traditional `mtree` cli utility is primarily on BSDs (FreeBSD,
openBSD, etc), but even broader support for the `mtree` specification format is
provided with libarchive ([libarchive-formats(5)][libarchive-formats(5)]).

There is also an [mtree port for Linux][archiecobbs/mtree-port] though it is
not widely packaged for Linux distributions.


## Format

The format of hierarchy specification is consistent with the `# mtree v2.0`
format.  Both the BSD `mtree` and libarchive ought to be interoperable with it
with only one definite caveat.  On Linux, extended attributes (`xattr`) on
files are often a critical aspect of the file, holding ACLs, capabilities, etc.
While FreeBSD filesystem do support `extattr`, this feature has not made its
way into their `mtree`.

This implementation of mtree supports an additional "keyword" of `xattr`. If
you include this keyword, then the FreeBSD `mtree` will fail as it is an
unknown keyword to that implementation.


### Typical form

With the standard keywords, plus say `sha256digest`, the hierarchy
specification looks like:

```mtree
# .
/set type=file nlink=1 mode=0664 uid=1000 gid=100
. size=4096 type=dir mode=0755 nlink=6 time=1459370393.273231538
    LICENSE size=1502 mode=0644 time=1458851690.0 sha256digest=ef4e53d83096be56dc38dbf9bc8ba9e3068bec1ec37c179033d1e8f99a1c2a95
    README.md size=2820 mode=0644 time=1459370256.316148361 sha256digest=d9b955134d99f84b17c0a711ce507515cc93cd7080a9dcd50400e3d993d876ac

[...]
```

See the directory presently in, and the files present. Along with each
path, is provided the keywords and the unique values for each path. Any common
keyword and values are established in the `/set` command.


### Extended attributes form

```mtree
# .
/set type=file nlink=1 mode=0664 uid=1000 gid=1000
. size=4096 type=dir mode=0775 nlink=6 time=1459370191.11179595 xattr.security.selinux=6b53fb56e2e61a6c6d672817791db03ebe693748
    LICENSE size=1502 time=1458851690.583562292 xattr.security.selinux=6b53fb56e2e61a6c6d672817791db03ebe693748
    README.md size=2366 mode=0644 time=1459369604.0 xattr.security.selinux=6b53fb56e2e61a6c6d672817791db03ebe693748

[...]
```

See the keyword prefixed with `xattr.` followed by the extended attribute's
namespace and keyword. This setup is consistent for use with Linux extended
attributes as well as FreeBSD extended attributes.

Since extended attributes are an unordered hashmap, this approach allows for
checking each `<namespace>.<key>` individually.

The value is the [SHA1 digest][sha1] of the value of the particular extended
attribute. Since the values themselves could be raw bytes, this approach both
avoids issues with encoding, as well as issues of information leaking. The
designation of SHA1 is arbitrary and seen as a general "good enough" assertion
of the value.


## Usage

To use the Go programming language library, see [the docs][godoc].

To use the command line tool, first [build it](#Building), then the following.


### Create a manifest

This will also include the sha512 digest of the files.

```bash
gomtree -c -K sha512digest -p . > /tmp/mtree.txt
```

### Validate a manifest

```bash
gomtree -p . -f /tmp/mtree.txt
```

### See the supported keywords

```bash
gomtree -l
```


## Building

Either:

```bash
go get github.com/vbatts/go-mtree/cmd/gomtree
```

or

```bash
git clone git://github.com/vbatts/go-mtree.git
cd ./go-mtree/cmd/gomtree
go build .
```


[mtree(8)]: https://www.freebsd.org/cgi/man.cgi?mtree(8)
[libarchive-formats(5)]: https://www.freebsd.org/cgi/man.cgi?query=libarchive-formats&sektion=5&n=1
[archiecobbs/mtree-port]: https://github.com/archiecobbs/mtree-port
[godoc]: https://godoc.org/github.com/vbatts/go-mtree
[sha1]: https://tools.ietf.org/html/rfc3174

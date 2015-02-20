Overview
========

Extend the upstream golang stdlib `archive/tar` library, to expose the raw
bytes of the TAR, rather than just the marshalled headers and file stream.

The goal being that by preserving the raw bytes of each header, padding bytes,
and the raw file payload, one could reassemble the original archive.


Caveat
------

Eventually this should detect TARs that this is not possible with.

For example stored sparse files that have "holes" in them, will be read as a
contiguous file, though the archive contents may be recorded in sparse format.
Therefore when adding the file payload to a reassembled tar, to achieve
identical output, the file payload would need be precisely re-sparsified. This
is not something I seek to fix imediately, but would rather have an alert that
precise reassembly is not possible.
(see more http://www.gnu.org/software/tar/manual/html_node/Sparse-Formats.html)


Contract
--------

Do not break the API of stdlib `archive/tar`


Example
-------

First we'll get an archive to work with. For repeatability, we'll make an
archive from what you've just cloned:

```
git archive --format=tar -o tar-split.tar HEAD .
```

Then build the example main.go:

```
go build ./main.go
```

Now run the example over the archive:

```
$ ./main tar-split.tar
2015/02/20 15:00:58 writing "tar-split.tar" to "tar-split.tar.out"
pax_global_header pre: 512 read: 52 post: 0
LICENSE pre: 972 read: 1075 post: 0
README.md pre: 973 read: 1004 post: 0
archive/ pre: 532 read: 0 post: 0
archive/tar/ pre: 512 read: 0 post: 0
archive/tar/common.go pre: 512 read: 7790 post: 0
archive/tar/example_test.go pre: 914 read: 1659 post: 0
archive/tar/reader.go pre: 901 read: 25303 post: 0
archive/tar/reader_test.go pre: 809 read: 17513 post: 0
archive/tar/stat_atim.go pre: 919 read: 414 post: 0
archive/tar/stat_atimespec.go pre: 610 read: 414 post: 0
archive/tar/stat_unix.go pre: 610 read: 716 post: 0
archive/tar/tar_test.go pre: 820 read: 6673 post: 0
archive/tar/testdata/ pre: 1007 read: 0 post: 0
archive/tar/testdata/gnu.tar pre: 512 read: 3072 post: 0
archive/tar/testdata/nil-uid.tar pre: 512 read: 1024 post: 0
archive/tar/testdata/pax.tar pre: 512 read: 10240 post: 0
archive/tar/testdata/small.txt pre: 512 read: 5 post: 0
archive/tar/testdata/small2.txt pre: 1019 read: 11 post: 0
archive/tar/testdata/sparse-formats.tar pre: 1013 read: 17920 post: 0
archive/tar/testdata/star.tar pre: 512 read: 3072 post: 0
archive/tar/testdata/ustar.tar pre: 512 read: 2048 post: 0
archive/tar/testdata/v7.tar pre: 512 read: 3584 post: 0
archive/tar/testdata/writer-big-long.tar pre: 512 read: 4096 post: 0
archive/tar/testdata/writer-big.tar pre: 512 read: 4096 post: 0
archive/tar/testdata/writer.tar pre: 512 read: 3584 post: 0
archive/tar/testdata/xattrs.tar pre: 512 read: 5120 post: 0
archive/tar/writer.go pre: 512 read: 11867 post: 0
archive/tar/writer_test.go pre: 933 read: 12436 post: 0
main.go pre: 876 read: 1568 post: 0
old.go pre: 992 read: 4918 post: 0
Size: 174080; Sum: 174080
```

Ideally the input tar and output `*.out`, will match:

```
$ sha1sum tar-split.tar*
ca9e19966b892d9ad5960414abac01ef585a1e22  tar-split.tar
ca9e19966b892d9ad5960414abac01ef585a1e22  tar-split.tar.out
```

What's Next?
------------

* Add tests for different types of tar options/extensions
* Package for convenience handling around collecting the RawBytes()
* Marshalling and storing index, ordering, file size and perhaps relative path of extracted files
 - perhaps have an API to allow user to provided a `hash.Hash` to checksum and store for the file payloads
 - though not enabled by default
 - this way, users wanting to implement an on disk tree validation could do so
 - but otherwise, we rely on the resulting re-assembled tar be validated
* Using stored index information, make an API for providing `io.Reader` and perhaps `tar.Reader` from re-assembled tar

License
-------

See LICENSE



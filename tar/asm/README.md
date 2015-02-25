asm
===

This library for assembly and disassembly of tar archives, facilitated by
`github.com/vbatts/tar-split/tar/storage`.


Concerns
--------

For completely safe assembly/disassembly, there will need to be a Content
Addressable Storage (CAS) directory, that maps to a checksum in the
`storage.Entity` of `storage.FileType`.

This is due to the fact that tar archives _can_ allow multiple records for the
same path, but the last one effectively wins. Even if the prior records had a
different payload. 

In this way, when assembling an archive from relative paths, if the archive has
multiple entries for the same path, then all payloads read in from a relative
path would be identical.


Thoughts
--------

While the initial implementation is based on a relative path, I'm thinking the
next step is to have something like a FileGetter interface, of which a path
based getter is just one type.

Then you could pass a path based Getter and an Unpacker, and receive a
io.Reader that is your tar stream.


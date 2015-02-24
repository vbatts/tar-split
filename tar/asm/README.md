asm
===

This library for assembly and disassembly of tar archives, facilitated by
`github.com/vbatts/tar-split/tar/storage`.



Thoughts
--------

While the initial implementation is based on a relative path, I'm thinking the
next step is to have something like a FileGetter interface, of which a path
based getter is just one type.

Then you could pass a path based Getter and an Unpacker, and receive a
io.Reader that is your tar stream.


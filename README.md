Overview
--------

Extend the upstream golang stdlib `archive/tar` library, to expose the raw
bytes of the TAR, rather than just the marshalled headers and file stream.

The goal being that by preserving the raw bytes of each header, padding bytes,
and the raw file payload, one could reassemble the original archive.


Caveat
======

Eventually this should detect TARs that this is not possible with.

For example stored sparse files that have "holes" in them, will be read as a
contiguous file, though the archive contents may be recorded in sparse format.
Therefore when adding the file payload to a reassembled tar, to achieve
identical output, the file payload would need be precisely re-sparsified. This
is not something I seek to fix imediately, but would rather have an alert that
precise reassembly is not possible.
(see more http://www.gnu.org/software/tar/manual/html_node/Sparse-Formats.html)


Contract
========

Do not break the API of stdlib `archive/tar`


License
=======

See LICENSE



# ASN.1 tool

A tool for the investigation of data encoded in ASN.1 (DER). It takes an example
of encoded data and produces an interactive HTML document showing the structure
of the data along with values contained within.

## FAQ

*Q*: what is ASN.1?

*A*: abstract syntax notation one, a flexible way of defining an encoding
structure for data that, at least in theory, is very amenable to machine
processing.

*Q*: what is DER?

*A*: ASN.1 allows data with the same structure to be encoded with different rule
sets. DER are the distinguised encoding rules, which are really just a subset of
BER (basic encoding rules). DER defines policies which mean that there is only
ever a single way of encoding a given construct, whereas BER might allow that
construct to be encoded in different ways. DER is thus appropriate for use where
data might be encoded multiple times, and the resultant bytes need to be equal
in each instance. DER encoding is quite often used in cryptographic protocols,
due to its reproduceability. It is thus the target of this tool.

*Q*: should I use ASN.1 for a new application?

*A*: no; use [Protocol Buffers](https://developers.google.com/protocol-buffers/)
instead. ASN.1 was very powerful and flexible for its time, but three decades of
experience have produced something much more simple and interoperable while
retaining all of the same power.

A lot of existing standards use ASN.1 in practice, and so it is important to be
able to understand and process such data if required. This tool is an aid to
investigating instances of such data.

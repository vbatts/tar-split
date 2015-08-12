package asm

// Defaults that matched existing behavior
var (
	DefaultOutputOptions = OptFileCheck | OptSegment
	DefaultInputOptions  = OptFileCheck | OptSegment
)

// Options for processing the tar stream with additional options. Like
// including entries for on-disk verification.
type Options int

// The options include the FileCheckEntry, SegmentEntry, and for VerficationEntry
const (
	OptFileCheck Options = 1 << iota
	OptSegment
	OptVerify
)

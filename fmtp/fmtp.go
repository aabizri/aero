// Package fmtp provides Flight Message Transfer Protocol (FMTP) support.
// It currently supports v2.0
package fmtp

// The following constants define the type of the message being carried
const (
	_ = iota
	Operational
	Operator
	identification
	system
)

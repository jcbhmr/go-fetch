// go-fetch-specific code related to RFC 5234. This is not a complete implementation of RFC 5234.
package rfc5234

import "regexp"

/*
# B.1.  Core Rules

Certain basic rules are in uppercase, such as SP, HTAB, CRLF, DIGIT,
ALPHA, etc.

https://www.rfc-editor.org/rfc/rfc5234.html#appendix-B.1
*/

// ALPHA          =  %x41-5A / %x61-7A   ; A-Z / a-z
//
// https://www.rfc-editor.org/rfc/rfc5234.html#appendix-B.1
var ALPHA = regexp.MustCompile(`^[A-Za-z]$`)

// BIT            =  "0" / "1"
//
// https://www.rfc-editor.org/rfc/rfc5234.html#appendix-B.1
var BIT = regexp.MustCompile(`^[01]$`)

// CHAR           =  %x01-7F
//
//	; any 7-bit US-ASCII character,
//	;  excluding NUL
//
// https://www.rfc-editor.org/rfc/rfc5234.html#appendix-B.1
var CHAR = regexp.MustCompile(`^[\x01-\x7F]$`)

// CR             =  %x0D
//
//	; carriage return
//
// https://www.rfc-editor.org/rfc/rfc5234.html#appendix-B.1
var CR = regexp.MustCompile(`^\x0D$`)

// CRLF           =  CR LF
//
//	; Internet standard newline
//
// https://www.rfc-editor.org/rfc/rfc5234.html#appendix-B.1
var CRLF = regexp.MustCompile(`^\x0D\x0A$`)

// CTL            =  %x00-1F / %x7F
//
//	; controls
//
// https://www.rfc-editor.org/rfc/rfc5234.html#appendix-B.1
var CTL = regexp.MustCompile(`^[\x00-\x1F\x7F]$`)

// DIGIT          =  %x30-39
//
//	; 0-9
//
// https://www.rfc-editor.org/rfc/rfc5234.html#appendix-B.1
var DIGIT = regexp.MustCompile(`^\d$`)

// DQUOTE         =  %x22
//
//	; " (Double Quote)
//
// https://www.rfc-editor.org/rfc/rfc5234.html#appendix-B.1
var DQUOTE = regexp.MustCompile(`^"$`)

// HEXDIG         =  DIGIT / "A" / "B" / "C" / "D" / "E" / "F"
//
// https://www.rfc-editor.org/rfc/rfc5234.html#appendix-B.1
var HEXDIG = regexp.MustCompile(`^[\dA-F]$`)

// HTAB           =  %x09
//
//	; horizontal tab
//
// https://www.rfc-editor.org/rfc/rfc5234.html#appendix-B.1
var HTAB = regexp.MustCompile(`^\x09$`)

// LF             =  %x0A
//
//	; linefeed
//
// https://www.rfc-editor.org/rfc/rfc5234.html#appendix-B.1
var LF = regexp.MustCompile(`^\x0A$`)

// LWSP           =  *(WSP / CRLF WSP)
//
//	; Use of this linear-white-space rule
//	;  permits lines containing only white
//	;  space that are no longer legal in
//	;  mail headers and have caused
//	;  interoperability problems in other
//	;  contexts.
//	; Do not use when defining mail
//	;  headers and use with caution in
//	;  other contexts.
//
// https://www.rfc-editor.org/rfc/rfc5234.html#appendix-B.1
var LWSP = regexp.MustCompile(`^(\s|\x0D\x0A\s)*$`)

// OCTET          =  %x00-FF
//
//	; 8 bits of data
//
// https://www.rfc-editor.org/rfc/rfc5234.html#appendix-B.1
var OCTET = regexp.MustCompile(`^[\x00-\xFF]$`)

// SP             =  %x20
//
// https://www.rfc-editor.org/rfc/rfc5234.html#appendix-B.1
var SP = regexp.MustCompile(`^\x20$`)

// VCHAR          =  %x21-7E
//
//	; visible (printing) characters
//
// https://www.rfc-editor.org/rfc/rfc5234.html#appendix-B.1
var VCHAR = regexp.MustCompile(`^[\x21-\x7E]$`)

// WSP            =  SP / HTAB
//
//	; white space
//
// https://www.rfc-editor.org/rfc/rfc5234.html#appendix-B.1
var WSP = regexp.MustCompile(`^[\x20\x09]$`)

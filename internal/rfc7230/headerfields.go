package rfc7230

import "regexp"

/*
# 3.2.3.  Whitespace

This specification uses three rules to denote the use of linear
whitespace: OWS (optional whitespace), RWS (required whitespace), and
BWS ("bad" whitespace).

https://www.rfc-editor.org/rfc/rfc7230.html#section-3.2.3
*/

// OWS            = *( SP / HTAB )
//
//	; optional whitespace
//
// The OWS rule is used where zero or more linear whitespace octets
// might appear.  For protocol elements where optional whitespace is
// preferred to improve readability, a sender SHOULD generate the
// optional whitespace as a single SP; otherwise, a sender SHOULD NOT
// generate optional whitespace except as needed to white out invalid or
// unwanted protocol elements during in-place message filtering.
//
// https://www.rfc-editor.org/rfc/rfc7230.html#section-3.2.3
var OWS = regexp.MustCompile(`^[ \t]*$`)

// The RWS rule is used when at least one linear whitespace octet is
// required to separate field tokens.  A sender SHOULD generate RWS as a
// single SP.
//
// RWS            = 1*( SP / HTAB )
//
//	; required whitespace
var RWS = regexp.MustCompile(`^[ \t]+$`)

// The BWS rule is used where the grammar allows optional whitespace
// only for historical reasons.  A sender MUST NOT generate BWS in
// messages.  A recipient MUST parse for such bad whitespace and remove
// it before interpreting the protocol element.
//
// BWS            = OWS
//
//	; "bad" whitespace
var BWS = regexp.MustCompile(`^[ \t]*$`)

/*
# 3.2.6.  Field Value Components

Most HTTP header field values are defined using common syntax
components (token, quoted-string, and comment) separated by
whitespace or specific delimiting characters.  Delimiters are chosen
from the set of US-ASCII visual characters not allowed in a token
(DQUOTE and "(),/:;<=>?@[\]{}").

https://www.rfc-editor.org/rfc/rfc7230.html#section-3.2.6
*/

// 	token          = 1*tchar
var Token = regexp.MustCompile(`^[!#$%&'*+\-.^_` + "`" + `|~0-9A-Za-z]+$`)

// 	tchar          = "!" / "#" / "$" / "%" / "&" / "'" / "*"
// 				/ "+" / "-" / "." / "^" / "_" / "`" / "|" / "~"
// 				/ DIGIT / ALPHA
// 				; any VCHAR, except delimiters
var TChar = regexp.MustCompile(`^[!#$%&'*+\-.^_` + "`" + `|~0-9A-Za-z]$`)

// A string of text is parsed as a single value if it is quoted using
// double-quote marks.
//
// 	quoted-string  = DQUOTE *( qdtext / quoted-pair ) DQUOTE
var QuotedString = regexp.MustCompile(`^"[ \t\x21\x23-\x5B\x5D-\x7E\x80-\xFF]*"$`)

// 	qdtext         = HTAB / SP /%x21 / %x23-5B / %x5D-7E / obs-text
var QDText = regexp.MustCompile(`^[\x09\x20\x21\x23-\x5B\x5D-\x7E\x80-\xFF]$`)

// 	obs-text       = %x80-FF
var ObsText = regexp.MustCompile(`^[\x80-\xFF]$`)

// Comments can be included in some HTTP header fields by surrounding
// the comment text with parentheses.  Comments are only allowed in
// fields containing "comment" as part of their field value definition.
//
// 	comment        = "(" *( ctext / quoted-pair / comment ) ")"
var Comment = regexp.MustCompile(`^\([ \t\x21-\x27\x2A-\x5B\x5D-\x7E\x80-\xFF]*\)$`)

// 	ctext          = HTAB / SP / %x21-27 / %x2A-5B / %x5D-7E / obs-text
var CText = regexp.MustCompile(`^[\x09\x20\x21-\x27\x2A-\x5B\x5D-\x7E\x80-\xFF]$`)

// The backslash octet ("\") can be used as a single-octet quoting
// mechanism within quoted-string and comment constructs.  Recipients
// that process the value of a quoted-string MUST handle a quoted-pair
// as if it were replaced by the octet following the backslash.
//
// 	quoted-pair    = "\" ( HTAB / SP / VCHAR / obs-text )
//
// A sender SHOULD NOT generate a quoted-pair in a quoted-string except
// where necessary to quote DQUOTE and backslash octets occurring within
// that string.  A sender SHOULD NOT generate a quoted-pair in a comment
// except where necessary to quote parentheses ["(" and ")"] and
// backslash octets occurring within that comment.
var QuotedPair = regexp.MustCompile(`^\\[\x09\x20\x21-\x7E\x80-\xFF]$`)

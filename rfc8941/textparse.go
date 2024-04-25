// go-fetch-specific code related to RFC 8941. This is not a complete implementation of RFC 8941.
package rfc8941

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/barweiss/go-tuple"
	"github.com/jcbhmr/go-fetch/rfc7230"
	"golang.org/x/exp/utf8string"
)

/*
# 4.2. Parsing Structured Fields

When a receiving implementation parses HTTP fields that are known to be Structured Fields, it is important that care be taken, as there are a number of edge cases that can cause interoperability or even security problems. This section specifies the algorithm for doing so.

https://httpwg.org/specs/rfc8941.html#text-parse
*/

type StructuredFieldValue = any

// Given an array of bytes as input_bytes that represent the chosen field's field-value (which is empty if that field is not present) and field_type (one of "dictionary", "list", or "item"), return the parsed header value.
//
// For Lists and Dictionaries, this has the effect of correctly concatenating all of the field's lines, as long as individual members of the top-level data structure are not split across multiple header instances. The parsing algorithms for both types allow tab characters, since these might be used to combine field lines by some implementations.
//
// Strings split across multiple field lines will have unpredictable results, because one or more commas (with optional whitespace) will become part of the string output by the parser. Since concatenation might be done by an upstream intermediary, the results are not under the control of the serializer or the parser, even when they are both under the control of the same party.
//
// Tokens, Integers, Decimals, and Byte Sequences cannot be split across multiple field lines because the inserted commas will cause parsing to fail.
//
// Parsers MAY fail when processing a field value spread across multiple field lines, when one of those lines does not parse as that field. For example, a parsing handling an Example-String field that's defined as an sf-string is allowed to fail when processing this field section:
//
// Example-String: "foo
// Example-String: bar"
// If parsing fails -- including when calling another algorithm -- the entire field value MUST be ignored (i.e., treated as if the field were not present in the section). This is intentionally strict, to improve interoperability and safety, and specifications referencing this document are not allowed to loosen this requirement.
//
// Note that this requirement does not apply to an implementation that is not parsing the field; for example, an intermediary is not required to strip a failing field from a message before forwarding it.
//
// https://httpwg.org/specs/rfc8941.html#text-parse
func TextParse(inputBytes []byte, fieldType string) (StructuredFieldValue, error) {
	// 1. Convert input_bytes into an ASCII string input_string; if conversion fails, fail parsing.
	inputString := string(inputBytes)
	if !utf8string.NewString(inputString).IsASCII() {
		return nil, fmt.Errorf("parsing failed: %s", string(inputBytes))
	}

	// 2. Discard any leading SP characters from input_string.
	inputString = strings.TrimLeft(inputString, " ")

	// 3. If field_type is "list", let output be the result of running Parsing a List (Section 4.2.1) with input_string.
	var output any
	var err error
	if fieldType == "list" {
		output, err = ParseList(&inputString)
	} else if fieldType == "dictionary" {
		// 4. If field_type is "dictionary", let output be the result of running Parsing a Dictionary (Section 4.2.2) with input_string.
		output, err = ParseDictionary(&inputString)
	} else if fieldType == "item" {
		// 5. If field_type is "item", let output be the result of running Parsing an Item (Section 4.2.3) with input_string.
		output, err = ParseItem(&inputString)
	} else {
		panic(fmt.Errorf(`fieldType must be "list"|"dictionary"|"item" got %#v`, fieldType))
	}
	if err != nil {
		return nil, err
	}

	// 6. Discard any leading SP characters from input_string.
	inputString = strings.TrimLeft(inputString, " ")

	// 7. If input_string is not empty, fail parsing.
	if inputString != "" {
		return nil, fmt.Errorf("parsing failed: %s", string(inputBytes))
	}

	// 8. Otherwise, return output.
	return output, nil
}

/*
# 4.2.1. Parsing a List

https://httpwg.org/specs/rfc8941.html#parse-list
*/

// Given an ASCII string as input_string, return an array of (item_or_inner_list, parameters) tuples. input_string is modified to remove the parsed value.
//
// https://httpwg.org/specs/rfc8941.html#parse-list
func ParseList(inputString *string) ([]tuple.T2[ItemOrInnerList, Parameters], error) {

	// 1.  Let members be an empty array.
	members := []tuple.T2[ItemOrInnerList, Parameters]{}

	// 2.  While input_string is not empty:
	for *inputString != "" {

		// 1.  Append the result of running Parsing an Item or Inner List (Section 4.2.1.1) with input_string to members.
		res, err := ParseItemOrList(inputString)
		if err != nil {
			return nil, err
		}
		members = append(members, res)

		// 2.  Discard any leading OWS characters from input_string.
		*inputString = strings.TrimLeft(*inputString, " \t")

		// 3.  If input_string is empty, return members.
		if *inputString == "" {
			return members, nil
		}

		// 4.  Consume the first character of input_string; if it is not ",", fail parsing.
		firstChar := (*inputString)[0]
		*inputString = (*inputString)[1:]
		if firstChar != ',' {
			return nil, fmt.Errorf("parsing failed: %s", *inputString)
		}

		// 5.  Discard any leading OWS characters from input_string.
		*inputString = strings.TrimLeft(*inputString, " \t")

		// 6.  If input_string is empty, there is a trailing comma; fail parsing.
		if *inputString == "" {
			return nil, fmt.Errorf("parsing failed: %s", *inputString)
		}
	}

	// 3.  No structured data has been found; return members (which is empty).
	return members, nil
}

/*
# 4.2.1.1. Parsing an Item or Inner List

https://httpwg.org/specs/rfc8941.html#parse-item-or-list
*/

// Given an ASCII string as input_string, return the tuple (item_or_inner_list, parameters), where item_or_inner_list can be either a single bare item or an array of (bare_item, parameters) tuples. input_string is modified to remove the parsed value.
//
// https://httpwg.org/specs/rfc8941.html#parse-item-or-list
func ParseItemOrList(inputString *string) (tuple.T2[ItemOrInnerList, Parameters], error) {
	// 1.  If the first character of input_string is "(", return the result of running Parsing an Inner List (Section 4.2.1.2) with input_string.
	if (*inputString)[0] == '(' {
		res, err := ParseInnerList(inputString)
		if err != nil {
			return tuple.New2[ItemOrInnerList, Parameters](nil, nil), err
		}
		return tuple.New2[ItemOrInnerList, Parameters](res.V1, res.V2), nil
	}

	// 2.  Return the result of running Parsing an Item (Section 4.2.3) with input_string.
	return ParseItem(inputString)
}

/*
# 4.2.1.2. Parsing an Inner List

https://httpwg.org/specs/rfc8941.html#parse-innerlist
*/

type InnerList = []tuple.T2[BareItem, Parameters]

// Given an ASCII string as input_string, return the tuple (inner_list, parameters), where inner_list is an array of (bare_item, parameters) tuples. input_string is modified to remove the parsed value.
//
// https://httpwg.org/specs/rfc8941.html#parse-innerlist
func ParseInnerList(inputString *string) (tuple.T2[InnerList, Parameters], error) {
	// 1.  Consume the first character of input_string; if it is not "(", fail parsing.
	firstChar := (*inputString)[0]
	*inputString = (*inputString)[1:]
	if firstChar != '(' {
		return tuple.New2[InnerList, Parameters](nil, nil), fmt.Errorf("parsing failed: %s", *inputString)
	}

	// 2.  Let inner_list be an empty array.
	innerList := []tuple.T2[any, Parameters]{}

	// 3.  While input_string is not empty:
	for *inputString != "" {

		//     1.  Discard any leading SP characters from input_string.
		*inputString = strings.TrimLeft(*inputString, " ")

		//     2.  If the first character of input_string is ")":
		if (*inputString)[0] == ')' {

			//         1.  Consume the first character of input_string.
			*inputString = (*inputString)[1:]

			//         2.  Let parameters be the result of running Parsing Parameters (Section 4.2.3.2) with input_string.
			parameters, err := ParseParam(inputString)
			if err != nil {
				return tuple.New2[InnerList, Parameters](nil, nil), err
			}

			//         3.  Return the tuple (inner_list, parameters).
			return tuple.New2(innerList, parameters), nil
		}

		//     3.  Let item be the result of running Parsing an Item (Section 4.2.3) with input_string.
		item, _ := ParseItem(inputString)

		//     4.  Append item to inner_list.
		innerList = append(innerList, item)

		//     5.  If the first character of input_string is not SP or ")", fail parsing.
		if (*inputString)[0] != ' ' && (*inputString)[0] != ')' {
			return tuple.New2[InnerList, Parameters](nil, nil), fmt.Errorf("parsing failed: %s", *inputString)

		}
	}

	// 4.  The end of the Inner List was not found; fail parsing.
	return tuple.New2[InnerList, Parameters](nil, nil), fmt.Errorf("parsing failed: %s", *inputString)
}

/*
# 4.2.2. Parsing a Dictionary

https://httpwg.org/specs/rfc8941.html#parse-dictionary
*/

type ItemOrInnerList = any

// Given an ASCII string as input_string, return an ordered map whose values are (item_or_inner_list, parameters) tuples. input_string is modified to remove the parsed value.
//
// Note that when duplicate Dictionary keys are encountered, all but the last instance are ignored.
//
// https://httpwg.org/specs/rfc8941.html#parse-dictionary
func ParseDictionary(inputString *string) (map[string]tuple.T2[ItemOrInnerList, Parameters], error) {
	// 1. Let dictionary be an empty, ordered map.
	dictionary := map[string]tuple.T2[ItemOrInnerList, Parameters]{}
	// 2. While input_string is not empty:
	for *inputString != "" {

		// 1. Let this_key be the result of running Parsing a Key (Section 4.2.3.3) with input_string.
		thisKey, err := ParseKey(inputString)
		if err != nil {
			return nil, err
		}

		var member tuple.T2[any, Parameters]
		var value any

		// 2. If the first character of input_string is "=":
		if (*inputString)[0] == '=' {
			// 1. Consume the first character of input_string.
			*inputString = (*inputString)[1:]
			// 2. Let member be the result of running Parsing an Item or Inner List (Section 4.2.1.1) with input_string.
			var err error
			member, err = ParseItemOrList(inputString)
			if err != nil {
				return nil, err
			}
		} else {
			// 3. Otherwise:

			// 1. Let value be Boolean true.
			value = true

			// 2. Let parameters be the result of running Parsing Parameters (Section 4.2.3.2) with input_string.
			parameters, err := ParseParam(inputString)
			if err != nil {
				return nil, err
			}

			// 3. Let member be the tuple (value, parameters).
			member = tuple.New2(value, parameters)
		}

		// 4. If dictionary already contains a key this_key (comparing character for character), overwrite its value with member.
		if _, ok := dictionary[thisKey]; ok {
			dictionary[thisKey] = member
		} else {
			// 5. Otherwise, append key this_key with value member to dictionary.
			dictionary[thisKey] = member
		}

		// 6. Discard any leading OWS characters from input_string.
		*inputString = strings.TrimLeft(*inputString, " \t")

		// 7. If input_string is empty, return dictionary.
		if *inputString == "" {
			return dictionary, nil
		}

		// 8. Consume the first character of input_string; if it is not ",", fail parsing.
		firstChar := (*inputString)[0]
		*inputString = (*inputString)[1:]
		if firstChar != ',' {
			return nil, fmt.Errorf("parsing failed: %s", *inputString)
		}

		// 9. Discard any leading OWS characters from input_string.
		*inputString = strings.TrimLeft(*inputString, " \t")

		// 10. If input_string is empty, there is a trailing comma; fail parsing.
		if *inputString == "" {
			return nil, fmt.Errorf("parsing failed: %s", *inputString)
		}
	}

	// 3. No structured data has been found; return dictionary (which is empty).
	return dictionary, nil
}

/*
# 4.2.3. Parsing an Item

https://httpwg.org/specs/rfc8941.html#parse-item
*/

// Given an ASCII string as input_string, return a (bare_item, parameters) tuple. input_string is modified to remove the parsed value.
//
// https://httpwg.org/specs/rfc8941.html#parse-item
func ParseItem(inputString *string) (tuple.T2[BareItem, Parameters], error) {
	// 1. Let bare_item be the result of running Parsing a Bare Item (Section 4.2.3.1) with input_string.
	bareItem, err := ParseBareItem(inputString)
	if err != nil {
		return tuple.New2[any, Parameters](nil, nil), err
	}

	// 2. Let parameters be the result of running Parsing Parameters (Section 4.2.3.2) with input_string.
	parameters, err := ParseParam(inputString)
	if err != nil {
		return tuple.New2[any, Parameters](nil, nil), err
	}

	// 3. Return the tuple (bare_item, parameters).
	return tuple.New2(bareItem, parameters), nil
}

/*
# 4.2.3.1. Parsing a Bare Item

https://httpwg.org/specs/rfc8941.html#parse-bare-item
*/

type BareItem = any

// Given an ASCII string as input_string, return a bare Item. input_string is modified to remove the parsed value.
//
// https://httpwg.org/specs/rfc8941.html#parse-bare-item
func ParseBareItem(inputString *string) (BareItem, error) {
	// 1. If the first character of input_string is a "-" or a DIGIT, return the result of running Parsing an Integer or Decimal (Section 4.2.4) with input_string.
	if regexp.MustCompile(`^[-\d]`).MatchString(*inputString) {
		return ParseNumber(inputString)
	} else if (*inputString)[0] == '"' {
		// 2. If the first character of input_string is a DQUOTE, return the result of running Parsing a String (Section 4.2.5) with input_string.
		return ParseString(inputString)
	} else if regexp.MustCompile(`^[a-z*]`).MatchString(*inputString) {
		// 3. If the first character of input_string is an ALPHA or "*", return the result of running Parsing a Token (Section 4.2.6) with input_string.
		return ParseToken(inputString)
	} else if (*inputString)[0] == ':' {
		// 4. If the first character of input_string is ":", return the result of running Parsing a Byte Sequence (Section 4.2.7) with input_string.
		return ParseBinary(inputString)
	} else if (*inputString)[0] == '?' {
		// 5. If the first character of input_string is "?", return the result of running Parsing a Boolean (Section 4.2.8) with input_string.
		return ParseBoolean(inputString)
	} else {
		// 6. Otherwise, the item type is unrecognized; fail parsing.
		return nil, fmt.Errorf("parsing failed: %s", *inputString)
	}
}

/*
4.2.3.2. Parsing Parameters

https://httpwg.org/specs/rfc8941.html#parse-param
*/

type Parameters = []tuple.T2[string, BareItem]

// Given an ASCII string as input_string, return an ordered map whose values are bare Items. input_string is modified to remove the parsed value.
//
// Note that when duplicate parameter keys are encountered, all but the last instance are ignored.
//
// https://httpwg.org/specs/rfc8941.html#parse-param
func ParseParam(inputString *string) (Parameters, error) {
	// 1. Let parameters be an empty, ordered map.
	parameters := []tuple.T2[string, any]{}

	// 2. While input_string is not empty:
	for *inputString != "" {

		// 1. If the first character of input_string is not ";", exit the loop.
		if (*inputString)[0] != ';' {
			break
		}

		// 2. Consume the ";" character from the beginning of input_string.
		*inputString = (*inputString)[1:]

		// 3. Discard any leading SP characters from input_string.
		*inputString = strings.TrimLeft(*inputString, " ")

		// 4. Let param_key be the result of running Parsing a Key (Section 4.2.3.3) with input_string.
		paramKey, err := ParseKey(inputString)
		if err != nil {
			return nil, err
		}

		// 5. Let param_value be Boolean true.
		var paramValue any = true
		// 6. If the first character of input_string is "=":
		if (*inputString)[0] == '=' {

			// 1. Consume the "=" character at the beginning of input_string.
			*inputString = (*inputString)[1:]
			// 2. Let param_value be the result of running Parsing a Bare Item (Section 4.2.3.1) with input_string.
			paramValue, err = ParseBareItem(inputString)
		}
		// 7. If parameters already contains a key param_key (comparing character for character), overwrite its value with param_value.
		index := -1
		for i, p := range parameters {
			if p.V1 == paramKey {
				index = i
				break
			}
		}
		if index != -1 {
			parameters[index] = tuple.New2(paramKey, paramValue)
		} else {
			// 8. Otherwise, append key param_key with value param_value to parameters.
			parameters = append(parameters, tuple.New2(paramKey, paramValue))
		}
	}
	// 3. Return parameters.
	return parameters, nil
}

/*
# 4.2.3.3. Parsing a Key

https://httpwg.org/specs/rfc8941.html#parse-key
*/

type Key = string

// Given an ASCII string as input_string, return a key. input_string is modified to remove the parsed value.
//
// https://httpwg.org/specs/rfc8941.html#parse-key
func ParseKey(inputString *string) (Key, error) {
	// 1. If the first character of input_string is not lcalpha or "*", fail parsing.
	if !regexp.MustCompile(`^[a-z*]`).MatchString(*inputString) {
		return "", fmt.Errorf("parsing failed: %s", *inputString)
	}

	// 2. Let output_string be an empty string.
	outputString := ""
	// 3. While input_string is not empty:
	for *inputString != "" {

		// 1. If the first character of input_string is not one of lcalpha, DIGIT, "_", "-", ".", or "*", return output_string.
		if !regexp.MustCompile(`^[a-z0-9_\-.]`).MatchString(*inputString) {
			return outputString, nil
		}
		// 2. Let char be the result of consuming the first character of input_string.
		char := (*inputString)[0]
		*inputString = (*inputString)[1:]
		// 3. Append char to output_string.
		outputString += string(char)
	}
	// 4. Return output_string.
	return outputString, nil
}

/*
# 4.2.4. Parsing an Integer or Decimal

https://httpwg.org/specs/rfc8941.html#parse-number
*/

type IntegerOrDecimal = any

// Given an ASCII string as input_string, return an Integer or Decimal. input_string is modified to remove the parsed value.
//
// NOTE: This algorithm parses both Integers (Section 3.3.1) and Decimals (Section 3.3.2), and returns the corresponding structure.
//
// https://httpwg.org/specs/rfc8941.html#parse-number
func ParseNumber(inputString *string) (IntegerOrDecimal, error) {
	// 1. Let type be "integer".
	type_ := "integer"
	// 2. Let sign be 1.
	sign := 1
	// 3. Let input_number be an empty string.
	inputNumber := ""

	// 4. If the first character of input_string is "-", consume it and set sign to -1.
	if (*inputString)[0] == '-' {
		*inputString = (*inputString)[1:]
		sign = -1
	}

	// 5. If input_string is empty, there is an empty integer; fail parsing.
	if *inputString == "" {
		return nil, fmt.Errorf("parsing failed: %s", *inputString)
	}

	// 6. If the first character of input_string is not a DIGIT, fail parsing.
	if !regexp.MustCompile(`^\d`).MatchString(*inputString) {
		return nil, fmt.Errorf("parsing failed: %s", *inputString)
	}

	// 7. While input_string is not empty:
	for *inputString != "" {

		// 1. Let char be the result of consuming the first character of input_string.
		char := (*inputString)[0]
		*inputString = (*inputString)[1:]

		// 2. If char is a DIGIT, append it to input_number.
		if regexp.MustCompile(`^\d`).MatchString(string(char)) {
			inputNumber += string(char)
		} else if char == '.' {
			// 3. Else, if type is "integer" and char is ".":

			// 1. If input_number contains more than 12 characters, fail parsing.
			if len(inputNumber) > 12 {
				return nil, fmt.Errorf("parsing failed: %s", *inputString)
			} else {
				// 2. Otherwise, append char to input_number and set type to "decimal".
				inputNumber += string(char)
				type_ = "decimal"
			}
		} else {
			// 4. Otherwise, prepend char to input_string, and exit the loop.
			*inputString = string(char) + *inputString
			break
		}

		// 5. If type is "integer" and input_number contains more than 15 characters, fail parsing.
		if type_ == "integer" && len(inputNumber) > 15 {
			return nil, fmt.Errorf("parsing failed: %s", *inputString)
		}

		// 6. If type is "decimal" and input_number contains more than 16 characters, fail parsing.
		if type_ == "decimal" && len(inputNumber) > 16 {
			return nil, fmt.Errorf("parsing failed: %s", *inputString)
		}
	}

	// 8. If type is "integer":
	var outputNumber any
	if type_ == "integer" {

		// 1. Parse input_number as an integer and let output_number be the product of the result and sign.
		num, err := strconv.ParseInt(inputNumber, 10, 64)
		if err != nil {
			return nil, err
		}
		outputNumber = num * int64(sign)

	} else {
		// 9. Otherwise:

		// 1. If the final character of input_number is ".", fail parsing.
		if inputNumber[len(inputNumber)-1] == '.' {
			return nil, fmt.Errorf("parsing failed: %s", *inputString)
		}

		// 2. If the number of characters after "." in input_number is greater than three, fail parsing.
		if len(inputNumber)-strings.Index(inputNumber, ".") > 3 {
			return nil, fmt.Errorf("parsing failed: %s", *inputString)
		}

		// 3. Parse input_number as a decimal number and let output_number be the product of the result and sign.
		num, err := strconv.ParseFloat(inputNumber, 64)
		if err != nil {
			return nil, err
		}
		outputNumber = num * float64(sign)
	}

	// 10. Return output_number.
	return outputNumber, nil
}

/*
# 4.2.5. Parsing a String

https://httpwg.org/specs/rfc8941.html#parse-string
*/

// Given an ASCII string as input_string, return an unquoted String. input_string is modified to remove the parsed value.
//
// https://httpwg.org/specs/rfc8941.html#parse-string
func ParseString(inputString *string) (string, error) {
	// 1. Let output_string be an empty string.
	outputString := ""

	// 2. If the first character of input_string is not DQUOTE, fail parsing.
	if (*inputString)[0] != '"' {
		return "", fmt.Errorf("parsing failed: %s", *inputString)
	}

	// 3. Discard the first character of input_string.
	*inputString = (*inputString)[1:]

	// 4. While input_string is not empty:
	for *inputString != "" {
		// 1. Let char be the result of consuming the first character of input_string.
		char := (*inputString)[0]
		*inputString = (*inputString)[1:]

		// 2. If char is a backslash ("\"):
		if char == '\\' {
			// 1. If input_string is now empty, fail parsing.
			if *inputString == "" {
				return "", fmt.Errorf("parsing failed: %s", *inputString)
			}

			// 2. Let next_char be the result of consuming the first character of input_string.
			nextChar := (*inputString)[0]
			*inputString = (*inputString)[1:]

			// 3. If next_char is not DQUOTE or "\", fail parsing.
			if nextChar != '"' && nextChar != '\\' {
				return "", fmt.Errorf("parsing failed: %s", *inputString)
			}

			// 4. Append next_char to output_string.
			outputString += string(nextChar)
		} else if char == '"' {
			// 3. Else, if char is DQUOTE, return output_string.
			return outputString, nil
		} else if char <= 0x1f || (char >= 0x7f && char <= 0xff) {
			// 4. Else, if char is in the range %x00-1f or %x7f-ff (i.e., it is not in VCHAR or SP), fail parsing.
			return "", fmt.Errorf("parsing failed: %s", *inputString)
		} else {
			// 5. Else, append char to output_string.
			outputString += string(char)
		}
	}

	// 5. Reached the end of input_string without finding a closing DQUOTE; fail parsing.
	return "", fmt.Errorf("parsing failed: %s", *inputString)
}

/*
# 4.2.6. Parsing a Token

https://httpwg.org/specs/rfc8941.html#parse-token
*/

type Token = string

// Given an ASCII string as input_string, return a Token. input_string is modified to remove the parsed value.
//
// https://httpwg.org/specs/rfc8941.html#parse-token
func ParseToken(inputString *string) (Token, error) {
	// 1. If the first character of input_string is not ALPHA or "*", fail parsing.
	if !regexp.MustCompile(`^[A-Za-z*]`).MatchString(*inputString) {
		return "", fmt.Errorf("parsing failed: %s", *inputString)
	}

	// 2. Let output_string be an empty string.
	outputString := ""

	// 3. While input_string is not empty:
	for *inputString != "" {

		// 1. If the first character of input_string is not in tchar, ":", or "/", return output_string.
		if !(rfc7230.TChar.MatchString((*inputString)[:1]) || (*inputString)[0] == ':' || (*inputString)[0] == '/') {
			return outputString, nil
		}

		// 2. Let char be the result of consuming the first character of input_string.
		char := (*inputString)[0]
		*inputString = (*inputString)[1:]

		// 3. Append char to output_string.
		outputString += string(char)
	}

	// 4. Return output_string.
	return outputString, nil
}

/*
# 4.2.7. Parsing a Byte Sequence

https://httpwg.org/specs/rfc8941.html#parse-binary
*/

// Given an ASCII string as input_string, return a Byte Sequence. input_string is modified to remove the parsed value.
//
// Because some implementations of base64 do not allow rejection of encoded data that is not properly "=" padded (see [RFC4648], Section 3.2), parsers SHOULD NOT fail when "=" padding is not present, unless they cannot be configured to do so.
//
// Because some implementations of base64 do not allow rejection of encoded data that has non-zero pad bits (see [RFC4648], Section 3.5), parsers SHOULD NOT fail when non-zero pad bits are present, unless they cannot be configured to do so.
//
// This specification does not relax the requirements in [RFC4648], Sections 3.1 and 3.3; therefore, parsers MUST fail on characters outside the base64 alphabet and on line feeds in encoded data.
//
// https://httpwg.org/specs/rfc8941.html#parse-binary
func ParseBinary(inputString *string) ([]byte, error) {
	// 1. If the first character of input_string is not ":", fail parsing.
	if (*inputString)[0] != ':' {
		return nil, fmt.Errorf("parsing failed: %s", *inputString)
	}

	// 2. Discard the first character of input_string.
	*inputString = (*inputString)[1:]

	// 3. If there is not a ":" character before the end of input_string, fail parsing.
	if !strings.Contains(*inputString, ":") {
		return nil, fmt.Errorf("parsing failed: %s", *inputString)
	}

	// 4. Let b64_content be the result of consuming content of input_string up to but not including the first instance of the character ":".
	colonIndex := strings.Index(*inputString, ":")
	b64Content := (*inputString)[:colonIndex]
	*inputString = (*inputString)[colonIndex:]

	// 5. Consume the ":" character at the beginning of input_string.
	*inputString = (*inputString)[1:]

	// 6. If b64_content contains a character not included in ALPHA, DIGIT, "+", "/", and "=", fail parsing.
	if regexp.MustCompile(`[^A-Za-z0-9+/=]`).MatchString(b64Content) {
		return nil, fmt.Errorf("parsing failed: %s", *inputString)
	}

	// 7. Let binary_content be the result of base64-decoding [RFC4648] b64_content, synthesizing padding if necessary (note the requirements about recipient behavior below). If base64 decoding fails, parsing fails.
	binaryContent, err := base64.StdEncoding.DecodeString(b64Content)
	if err != nil {
		return nil, fmt.Errorf("parsing failed: %s", *inputString)
	}

	// 8. Return binary_content.
	return binaryContent, nil
}

/*
# 4.2.8. Parsing a Boolean

https://httpwg.org/specs/rfc8941.html#parse-boolean
*/

// Given an ASCII string as input_string, return a Boolean. input_string is modified to remove the parsed value.
//
// https://httpwg.org/specs/rfc8941.html#parse-boolean
func ParseBoolean(inputString *string) (bool, error) {

	// 1. If the first character of input_string is not "?", fail parsing.
	if (*inputString)[0] != '?' {
		return false, fmt.Errorf("parsing failed: %s", *inputString)
	}

	// 2. Discard the first character of input_string.
	*inputString = (*inputString)[1:]

	// 3. If the first character of input_string matches "1", discard the first character, and return true.
	if (*inputString)[0] == '1' {
		*inputString = (*inputString)[1:]
		return true, nil
	}

	// 4. If the first character of input_string matches "0", discard the first character, and return false.
	if (*inputString)[0] == '0' {
		*inputString = (*inputString)[1:]
		return false, nil
	}

	// 5. No value has matched; fail parsing.
	return false, fmt.Errorf("parsing failed: %s", *inputString)
}

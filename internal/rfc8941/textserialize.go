package rfc8941

import (
	"encoding/base64"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/jcbhmr/go-fetch/internal/rfc5234"
	"github.com/jcbhmr/go-fetch/internal/rfc7230"
	"golang.org/x/exp/utf8string"
	"github.com/samber/lo"
)

/*
# 4. Working with Structured Fields in HTTP

This section defines how to serialize and parse Structured Fields in textual
HTTP field values and other encodings compatible with them (e.g., in HTTP/2
[RFC7540] before compression with HPACK [RFC7541]).

https://httpwg.org/specs/rfc8941.html#text
*/

/*
# 4.1. Serializing Structured Fields

https://httpwg.org/specs/rfc8941.html#text-serialize
*/

type List = []lo.Tuple2[ItemOrInnerList, Parameters]
type Item = lo.Tuple2[BareItem, Parameters]

// Given a structure defined in this specification, return an ASCII string suitable for use in an HTTP field value.
//
// https://httpwg.org/specs/rfc8941.html#text-serialize
func TextSerialize(input any) ([]byte, error) {
	// 1. If the structure is a Dictionary or List and its value is empty (i.e., it has no members), do not serialize the field at all (i.e., omit both the field-name and field-value).
	if _, ok := input.(Dictionary); ok {
		if len(input.(Dictionary)) == 0 {
			return nil, nil
		}
	}
	if _, ok := input.(List); ok {
		if len(input.(List)) == 0 {
			return nil, nil
		}
	}

	var outputString string
	// 2. If the structure is a List, let output_string be the result of running Serializing a List (Section 4.1.1) with the structure.
	if v, ok := input.(List); ok {
		res, err := SerList(v)
		if err != nil {
			return nil, err
		}
		outputString = res
	} else if v, ok := input.(Dictionary); ok {
		// 3. Else, if the structure is a Dictionary, let output_string be the result of running Serializing a Dictionary (Section 4.1.2) with the structure.
		res, err := SerDictionary(v)
		if err != nil {
			return nil, err
		}
		outputString = res
	} else if v, ok := input.(Item); ok {
		// 4. Else, if the structure is an Item, let output_string be the result of running Serializing an Item (Section 4.1.3) with the structure.
		res, err := SerItem(v.A, v.B)
		if err != nil {
			return nil, err
		}
		outputString = res
	} else {
		// 5. Else, fail serialization.
		return nil, fmt.Errorf("serialization failed: %#+v", input)
	}

	// 6. Return output_string converted into an array of bytes, using ASCII encoding [RFC0020].
	return []byte(outputString), nil
}

/*
# 4.1.1. Serializing a List

https://httpwg.org/specs/rfc8941.html#ser-list
*/

// Given an array of (member_value, parameters) tuples as input_list, return an ASCII string suitable for use in an HTTP field value.
//
// https://httpwg.org/specs/rfc8941.html#ser-list
func SerList(inputList []lo.Tuple2[ItemOrInnerList, Parameters]) (string, error) {
	// 1. Let output be an empty string.
	output := ""
	// 2. For each (member_value, parameters) of input_list:
	for i, keyValue := range inputList {
		memberValue := keyValue.A
		parameters := keyValue.B

		// 1. If member_value is an array, append the result of running Serializing an Inner List (Section 4.1.1.1) with (member_value, parameters) to output.
		if _, ok := memberValue.([]any); ok {
			innerListStr, err := SerInnerList(memberValue.(InnerList), parameters)
			if err != nil {
				return "", err
			}
			output += innerListStr
		} else {
			// 2. Otherwise, append the result of running Serializing an Item (Section 4.1.3) with (member_value, parameters) to output.
			itemStr, err := SerItem(memberValue, parameters)
			if err != nil {
				return "", err
			}
			output += itemStr

			// 3. If more member_values remain in input_list:
			if i < len(inputList)-1 {
				// 1. Append "," to output.
				output += ","
				// 2. Append a single SP to output.
				output += " "
			}
		}
	}
	// 3. Return output.
	return output, nil
}

/*
# 4.1.1.1. Serializing an Inner List

https://httpwg.org/specs/rfc8941.html#ser-innerlist
*/

// Given an array of (member_value, parameters) tuples as inner_list, and parameters as list_parameters, return an ASCII string suitable for use in an HTTP field value.
//
// https://httpwg.org/specs/rfc8941.html#ser-innerlist
func SerInnerList(innerList InnerList, listParameters Parameters) (string, error) {
	// 1. Let output be the string "(".
	output := "("
	// 2. For each (member_value, parameters) of inner_list:
	for i, keyValue := range innerList {
		memberValue := keyValue.A
		parameters := keyValue.B

		// 1. Append the result of running Serializing an Item (Section 4.1.3) with (member_value, parameters) to output.
		itemStr, err := SerItem(memberValue, parameters)
		if err != nil {
			return "", err
		}
		output += itemStr

		// 2. If more values remain in inner_list, append a single SP to output.
		if i < len(innerList)-1 {
			output += " "
		}
	}

	// 3. Append ")" to output.
	output += ")"
	// 4. Append the result of running Serializing Parameters (Section 4.1.1.2) with list_parameters to output.
	listParametersStr, err := SerParams(listParameters)
	if err != nil {
		return "", err
	}
	output += listParametersStr
	// 5. Return output.
	return output, nil
}

/*
# 4.1.1.2. Serializing Parameters

https://httpwg.org/specs/rfc8941.html#ser-params
*/

// Given an ordered Dictionary as input_parameters (each member having a param_key and a param_value), return an ASCII string suitable for use in an HTTP field value.
//
// https://httpwg.org/specs/rfc8941.html#ser-params
func SerParams(inputParameters Parameters) (string, error) {
	// 1. Let output be an empty string.
	output := ""
	// 2. For each param_key with a value of param_value in input_parameters:
	for _, keyValue := range inputParameters {
		paramKey := keyValue.A
		paramValue := keyValue.B

		// 3. Append ";" to output.
		output += ";"

		// 4. Append the result of running Serializing a Key (Section 4.1.1.3) with param_key to output.
		paramKeyStr, err := SerKey(paramKey)
		if err != nil {
			return "", err
		}
		output += paramKeyStr

		// 5. If param_value is not Boolean true:
		if value, ok := paramValue.(bool); !ok || !value {
			// 1. Append "=" to output.
			output += "="
			// 2. Append the result of running Serializing a bare Item (Section 4.1.3.1) with param_value to output.
			paramValueStr, err := SerBareItem(paramValue)
			if err != nil {
				return "", err
			}
			output += paramValueStr
		}
	}

	// 3. Return output.
	return output, nil
}

/*
# 4.1.1.3. Serializing a Key

https://httpwg.org/specs/rfc8941.html#ser-key
*/

// Given a key as input_key, return an ASCII string suitable for use in an HTTP field value.
//
// https://httpwg.org/specs/rfc8941.html#ser-key
func SerKey(inputKey Key) (string, error) {
	// 1. Convert input_key into a sequence of ASCII characters; if conversion fails, fail serialization.
	value := string(inputKey)
	if !utf8string.NewString(value).IsASCII() {
		return "", fmt.Errorf("serialization failed: %#+v", inputKey)
	}

	// 2. If input_key contains characters not in lcalpha, DIGIT, "_", "-", ".", or "*", fail serialization.
	for _, r := range value {
		if regexp.MustCompile(`[^a-z0-9_\-.*]`).MatchString(string(r)) {
			return "", fmt.Errorf("serialization failed: %#+v", inputKey)
		}
	}

	// 3. If the first character of input_key is not lcalpha or "*", fail serialization.
	if !regexp.MustCompile(`^[a-z*]`).MatchString(value[:1]) {
		return "", fmt.Errorf("serialization failed: %#+v", inputKey)
	}

	// 4. Let output be an empty string.
	output := ""
	// 5. Append input_key to output.
	output += value
	// 6. Return output.
	return output, nil
}

/*
# 4.1.2. Serializing a Dictionary

https://httpwg.org/specs/rfc8941.html#ser-dictionary
*/

// Given an ordered Dictionary as input_dictionary (each member having a member_key and a tuple value of (member_value, parameters)), return an ASCII string suitable for use in an HTTP field value.
//
// https://httpwg.org/specs/rfc8941.html#ser-dictionary
func SerDictionary(inputDictionary Dictionary) (string, error) {
	// 1. Let output be an empty string.
	output := ""
	// 2. For each member_key with a value of (member_value, parameters) in input_dictionary:
	for i, keyValue := range inputDictionary {
		memberKey := keyValue.A
		memberValue := keyValue.B.A
		parameters := keyValue.B.B

		// 1. Append the result of running Serializing a Key (Section 4.1.1.3) with member's member_key to output.
		memberKeyStr, err := SerKey(memberKey)
		if err != nil {
			return "", err
		}
		output += memberKeyStr

		// 2. If member_value is Boolean true:
		if value, ok := memberValue.(bool); ok && value {
			// 1. Append the result of running Serializing Parameters (Section 4.1.1.2) with parameters to output.
			parametersStr, err := SerParams(parameters)
			if err != nil {
				return "", err
			}
			output += parametersStr
		} else {
			// 3. Otherwise:
			// 1. Append "=" to output.
			output += "="
			// 2. If member_value is an array, append the result of running Serializing an Inner List (Section 4.1.1.1) with (member_value, parameters) to output.
			if v, ok := memberValue.(InnerList); ok {
				innerListStr, err := SerInnerList(v, parameters)
				if err != nil {
					return "", err
				}
				output += innerListStr
			} else {
				// 3. Otherwise, append the result of running Serializing an Item (Section 4.1.3) with (member_value, parameters) to output.
				itemStr, err := SerItem(memberValue, parameters)
				if err != nil {
					return "", err
				}
				output += itemStr
			}
		}

		// 4. If more members remain in input_dictionary:
		if i < len(inputDictionary)-1 {
			// 1. Append "," to output.
			output += ","
			// 2. Append a single SP to output.
			output += " "
		}
	}
	// 3. Return output.
	return output, nil
}

/*
# 4.1.3. Serializing an Item

https://httpwg.org/specs/rfc8941.html#ser-item
*/

// Given an Item as bare_item and Parameters as item_parameters, return an ASCII string suitable for use in an HTTP field value.
//
// https://httpwg.org/specs/rfc8941.html#ser-item
func SerItem(bareItem any, itemParameters Parameters) (string, error) {
	// 1. Let output be an empty string.
	output := ""
	// 2. Append the result of running Serializing a Bare Item (Section 4.1.3.1) with bare_item to output.
	bareItemStr, err := SerBareItem(bareItem)
	if err != nil {
		return "", err
	}
	output += bareItemStr

	// 3. Append the result of running Serializing Parameters (Section 4.1.1.2) with item_parameters to output.
	itemParametersStr, err := SerParams(itemParameters)
	if err != nil {
		return "", err
	}
	output += itemParametersStr

	// 4. Return output.
	return output, nil
}

/*
# 4.1.3.1. Serializing a Bare Item

https://httpwg.org/specs/rfc8941.html#ser-bare-item
*/

// Given an Item as input_item, return an ASCII string suitable for use in an HTTP field value.
//
// https://httpwg.org/specs/rfc8941.html#ser-bare-item
func SerBareItem(inputItem any) (string, error) {
	// 1. If input_item is an Integer, return the result of running Serializing an Integer (Section 4.1.4) with input_item.
	if _, ok := inputItem.(int64); ok {
		return SerInteger(inputItem)
	}

	// 2. If input_item is a Decimal, return the result of running Serializing a Decimal (Section 4.1.5) with input_item.
	if _, ok := inputItem.(float64); ok {
		return SerDecimal(inputItem)
	}

	// 3. If input_item is a String, return the result of running Serializing a String (Section 4.1.6) with input_item.
	if _, ok := inputItem.(string); ok {
		return SerString(inputItem)
	}

	// 4. If input_item is a Token, return the result of running Serializing a Token (Section 4.1.7) with input_item.
	if _, ok := inputItem.(Token); ok {
		return SerToken(inputItem.(Token))
	}

	// 5. If input_item is a Byte Sequence, return the result of running Serializing a Byte Sequence (Section 4.1.8) with input_item.
	if _, ok := inputItem.([]byte); ok {
		return SerByteSequence(inputItem)
	}

	// 6. If input_item is a Boolean, return the result of running Serializing a Boolean (Section 4.1.9) with input_item.
	if _, ok := inputItem.(bool); ok {
		return SerBoolean(inputItem)
	}

	// 7. Otherwise, fail serialization.
	return "", fmt.Errorf("serialization failed: %#+v", inputItem)
}

/*
# 4.1.4. Serializing an integer

https://httpwg.org/specs/rfc8941.html#ser-integer
*/

// Given an Integer as input_integer, return an ASCII string suitable for use in an HTTP field value.
//
// https://httpwg.org/specs/rfc8941.html#ser-integer
func SerInteger(inputInteger any) (string, error) {
	// 1. If input_integer is not an integer in the range of -999,999,999,999,999 to 999,999,999,999,999 inclusive, fail serialization.
	value, ok := inputInteger.(int64)
	if !ok {
		return "", fmt.Errorf("serialization failed: %#+v", inputInteger)
	}
	if value < -999999999999999 || value > 999999999999999 {
		return "", fmt.Errorf("serialization failed: %#+v", inputInteger)
	}

	// 2. Let output be an empty string.
	output := ""
	// 3. If input_integer is less than (but not equal to) 0, append "-" to output.
	if value < 0 {
		output += "-"
	}
	// 4. Append input_integer's numeric value represented in base 10 using only decimal digits to output.
	output += strconv.FormatInt(int64(math.Abs(float64(value))), 10)
	// 5. Return output.
	return output, nil
}

/*
# 4.1.5. Serializing a Decimal

https://httpwg.org/specs/rfc8941.html#ser-decimal
*/

// Given a decimal number as input_decimal, return an ASCII string suitable for use in an HTTP field value.
//
// https://httpwg.org/specs/rfc8941.html#ser-decimal
func SerDecimal(inputDecimal any) (string, error) {
	// 1. If input_decimal is not a decimal number, fail serialization.
	value, ok := inputDecimal.(float64)
	if !ok {
		return "", fmt.Errorf("serialization failed: %#+v", inputDecimal)
	}

	// 2. If input_decimal has more than three significant digits to the right of the decimal point, round it to three decimal places, rounding the final digit to the nearest value, or to the even value if it is equidistant.
	value = math.RoundToEven(value*1000) / 1000

	// 3. If input_decimal has more than 12 significant digits to the left of the decimal point after rounding, fail serialization.
	if value > 999999999999 {
		return "", fmt.Errorf("serialization failed: %#+v", inputDecimal)
	}

	// 4. Let output be an empty string.
	output := ""
	// 5. If input_decimal is less than (but not equal to) 0, append "-" to output.
	if value < 0 {
		output += "-"
	}
	// 6. Append input_decimal's integer component represented in base 10 (using only decimal digits) to output; if it is zero, append "0".
	integer, fractional := math.Modf(value)
	integer = math.Abs(integer)
	fractional = math.Abs(fractional)
	if integer == 0 {
		output += "0"
	} else {
		output += strconv.FormatInt(int64(integer), 10)
	}
	// 7. Append "." to output.
	output += "."
	// 8. If input_decimal's fractional component is zero, append "0" to output.
	if fractional == 0 {
		output += "0"
	} else {
		// 9. Otherwise, append the significant digits of input_decimal's fractional component represented in base 10 (using only decimal digits) to output.
		output += strings.TrimRight(strconv.FormatFloat(fractional, 'f', -1, 64), "0")
	}

	// 10. Return output.
	return output, nil
}

/*
# 4.1.6. Serializing a String

https://httpwg.org/specs/rfc8941.html#ser-string
*/

// Given a String as input_string, return an ASCII string suitable for use in an HTTP field value.
//
// https://httpwg.org/specs/rfc8941.html#ser-string
func SerString(inputString any) (string, error) {
	// 1. Convert input_string into a sequence of ASCII characters; if conversion fails, fail serialization.
	value, ok := inputString.(string)
	if !ok {
		return "", fmt.Errorf("serialization failed: %#+v", inputString)
	}
	if !utf8string.NewString(value).IsASCII() {
		return "", fmt.Errorf("serialization failed: %#+v", inputString)
	}

	// 2. If input_string contains characters in the range %x00-1f or %x7f-ff (i.e., not in VCHAR or SP), fail serialization.
	for _, r := range value {
		rint := int(r)
		if (rint >= 0x00 && rint <= 0x1f) || (rint >= 0x7f && rint <= 0xff) {
			return "", fmt.Errorf("serialization failed: %#+v", inputString)
		}
	}

	// 3. Let output be the string DQUOTE.
	output := "\""

	// 4. For each character char in input_string:
	for _, r := range value {
		// 1. If char is "\" or DQUOTE:
		if r == '\\' || r == '"' {
			// 2. Append "\" to output.
			output += "\\"
		}
		// 2. Append char to output.
		output += string(r)
	}
	// 5. Append DQUOTE to output.
	output += "\""
	// 6. Return output.
	return output, nil
}

/*
# 4.1.7. Serializing a Token

https://httpwg.org/specs/rfc8941.html#ser-token
*/

// Given a Token as input_token, return an ASCII string suitable for use in an HTTP field value.
//
// https://httpwg.org/specs/rfc8941.html#ser-token
func SerToken(inputToken Token) (string, error) {
	// 1. Convert input_token into a sequence of ASCII characters; if conversion fails, fail serialization.
	value := string(inputToken)
	if !utf8string.NewString(value).IsASCII() {
		return "", fmt.Errorf("serialization failed: %#+v", inputToken)
	}

	// 2. If the first character of input_token is not ALPHA or "*", or the remaining portion contains a character not in tchar, ":", or "/", fail serialization.
	if !(rfc5234.ALPHA.MatchString(value[:1]) || value[0] == '*') {
		return "", fmt.Errorf("serialization failed: %#+v", inputToken)
	}
	for _, r := range value[1:] {
		if !(rfc7230.TChar.MatchString(string(r)) || r == ':' || r == '/') {
			return "", fmt.Errorf("serialization failed: %#+v", inputToken)
		}
	}

	// 3. Let output be an empty string.
	output := ""
	// 4. Append input_token to output.
	output += value
	// 5. Return output.
	return output, nil
}

/*
# 4.1.8. Serializing a Byte Sequence

https://httpwg.org/specs/rfc8941.html#ser-binary
*/

// Given a Byte Sequence as input_bytes, return an ASCII string suitable for use
// in an HTTP field value.
//
// The encoded data is required to be padded with "=", as per [RFC4648], Section
// 3.2.
//
// Likewise, encoded data SHOULD have pad bits set to zero, as per [RFC4648],
// Section 3.5, unless it is not possible to do so due to implementation
// constraints.
//
// https://httpwg.org/specs/rfc8941.html#ser-binary
func SerByteSequence(inputBytes any) (string, error) {
	// 1. If input_bytes is not a sequence of bytes, fail serialization.
	value, ok := inputBytes.([]byte)
	if !ok {
		return "", fmt.Errorf("serialization failed: %#+v", inputBytes)
	}

	// 2. Let output be an empty string.
	output := ""
	// 3. Append ":" to output.
	output += ":"
	// 4. Append the result of base64-encoding input_bytes as per [RFC4648], Section 4, taking account of the requirements below.
	output += base64.StdEncoding.EncodeToString(value)
	// 5. Append ":" to output.
	output += ":"
	// 6. Return output.
	return output, nil
}

/*
# 4.1.9. Serializing a Boolean

https://httpwg.org/specs/rfc8941.html#ser-boolean
*/

// Given a Boolean as input_boolean, return an ASCII string suitable for use in an HTTP field value.
//
// https://httpwg.org/specs/rfc8941.html#ser-boolean
func SerBoolean(inputBoolean any) (string, error) {
	// 1. If input_boolean is not a boolean, fail serialization.
	value, ok := inputBoolean.(bool)
	if !ok {
		return "", fmt.Errorf("serialization failed: %#+v", inputBoolean)
	}

	// 2. Let output be an empty string.
	output := ""
	// 3. Append "?" to output.
	output += "?"
	// 4. If input_boolean is true, append "1" to output.
	if value {
		output += "1"
	}
	// 5. If input_boolean is false, append "0" to output.
	if !value {
		output += "0"
	}
	// 6. Return output.
	return output, nil
}

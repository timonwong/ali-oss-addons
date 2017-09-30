package oss_addons

import (
	"encoding/base64"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// For JSON-escaping; see safeAppendString below.
const _hex = "0123456789abcdef"

// expirationDateFormat date format for expiration key in json policy.
const expirationDateFormat = "2006-01-02T15:04:05.999Z"

// policyCondition explanation:
// https://help.aliyun.com/document_detail/31988.html
// https://yq.aliyun.com/articles/58524
//
// Example:
//
//   policyCondition {
//       matchType: "$eq",
//       key: "$Content-Type",
//       value: "image/png",
//   }
//
type policyCondition struct {
	matchType string
	condition string
	value     string
}

// PostPolicy - Provides strict static type conversion and validation
// for Aliyun OSS POST policy JSON string.
type PostPolicy struct {
	// Expiration date and time of the POST policy.
	expiration time.Time
	// Collection of different policy conditions.
	conditions []policyCondition
	// ContentLengthRange minimum and maximum allowable size for the
	// uploaded content.
	contentLengthRange struct {
		min int64
		max int64
	}

	// Post form data.
	formData map[string]string
}

// NewPostPolicy - Instantiate new post policy.
func NewPostPolicy() *PostPolicy {
	p := &PostPolicy{}
	p.conditions = make([]policyCondition, 0)
	p.formData = make(map[string]string)
	return p
}

// SetExpires - Sets expiration time for the new policy.
func (p *PostPolicy) SetExpires(t time.Time) error {
	if t.IsZero() {
		return NewInvalidArgumentError("no expiry time set")
	}
	p.expiration = t
	return nil
}

// SetKey - Sets an object name for the policy based upload.
func (p *PostPolicy) SetKey(key string) error {
	if strings.TrimSpace(key) == "" || key == "" {
		return NewInvalidArgumentError("object name is empty")
	}
	policyCond := policyCondition{
		matchType: "eq",
		condition: "$key",
		value:     key,
	}
	if err := p.addNewPolicy(policyCond); err != nil {
		return err
	}
	p.formData["key"] = key
	return nil
}

// SetKeyStartsWith - Sets an object name that an policy based upload
// can start with.
func (p *PostPolicy) SetKeyStartsWith(keyStartsWith string) error {
	if strings.TrimSpace(keyStartsWith) == "" || keyStartsWith == "" {
		return NewInvalidArgumentError("object prefix is empty")
	}
	policyCond := policyCondition{
		matchType: "starts-with",
		condition: "$key",
		value:     keyStartsWith,
	}
	if err := p.addNewPolicy(policyCond); err != nil {
		return err
	}
	p.formData["key"] = keyStartsWith
	return nil
}

// SetBucket - Sets bucket at which objects will be uploaded to.
func (p *PostPolicy) SetBucket(bucketName string) error {
	if strings.TrimSpace(bucketName) == "" || bucketName == "" {
		return NewInvalidArgumentError("bucket name is empty")
	}
	policyCond := policyCondition{
		matchType: "eq",
		condition: "$bucket",
		value:     bucketName,
	}
	if err := p.addNewPolicy(policyCond); err != nil {
		return err
	}
	p.formData["bucket"] = bucketName
	return nil
}

// SetContentType - Sets content-type of the object for this policy
// based upload.
func (p *PostPolicy) SetContentType(contentType string) error {
	if strings.TrimSpace(contentType) == "" || contentType == "" {
		return NewInvalidArgumentError("no content type specified")
	}
	policyCond := policyCondition{
		matchType: "eq",
		condition: "$Content-Type",
		value:     contentType,
	}
	if err := p.addNewPolicy(policyCond); err != nil {
		return err
	}
	p.formData["Content-Type"] = contentType
	return nil
}

// SetContentLengthRange - Set new min and max content length
// condition for all incoming uploads.
func (p *PostPolicy) SetContentLengthRange(min, max int64) error {
	if min > max {
		return NewInvalidArgumentError("minimum limit is larger than maximum limit")
	}
	if min < 0 {
		return NewInvalidArgumentError("minimum limit cannot be negative")
	}
	if max < 0 {
		return NewInvalidArgumentError("maximum limit cannot be negative")
	}
	p.contentLengthRange.min = min
	p.contentLengthRange.max = max
	return nil
}

// SetSuccessStatusAction - Sets the status success code of the object for this policy
// based upload.
func (p *PostPolicy) SetSuccessStatusAction(status string) error {
	if strings.TrimSpace(status) == "" || status == "" {
		return NewInvalidArgumentError("status is empty")
	}
	policyCond := policyCondition{
		matchType: "eq",
		condition: "$success_action_status",
		value:     status,
	}
	if err := p.addNewPolicy(policyCond); err != nil {
		return err
	}
	p.formData["success_action_status"] = status
	return nil
}

// addNewPolicy - internal helper to validate adding new policies.
func (p *PostPolicy) addNewPolicy(policyCond policyCondition) error {
	if policyCond.matchType == "" || policyCond.condition == "" || policyCond.value == "" {
		return NewInvalidArgumentError("policy fields are empty")
	}
	p.conditions = append(p.conditions, policyCond)
	return nil
}

// Stringer interface for printing policy in json formatted string.
func (p PostPolicy) String() string {
	return string(p.marshalJSON())
}

// marshalJSON - Provides Marshaled JSON in bytes.
func (p PostPolicy) marshalJSON() []byte {
	buf := make([]byte, 0, 1024) // reserve 1k buffer

	// Expiration
	buf = append(buf, `{"expiration":"`...)
	buf = p.expiration.AppendFormat(buf, expirationDateFormat)
	buf = append(buf, `","conditions":[`...)

	// Conditions
	insertComma := false
	// Content-Length-Range
	if p.contentLengthRange.min != 0 || p.contentLengthRange.max != 0 {
		buf = append(buf, `["content-length-range",`...)
		buf = strconv.AppendInt(buf, p.contentLengthRange.min, 10)
		buf = append(buf, ',')
		buf = strconv.AppendInt(buf, p.contentLengthRange.max, 10)
		buf = append(buf, ']')
		insertComma = true
	}

	for _, po := range p.conditions {
		if insertComma {
			buf = append(buf, ',')
		}
		buf = append(buf, `["`...)
		buf = append(buf, po.matchType...)
		buf = append(buf, `","`...)
		buf = safeAppendString(buf, po.condition)
		buf = append(buf, `","`...)
		buf = safeAppendString(buf, po.value)
		buf = append(buf, `"]`...)
	}
	buf = append(buf, `]}`...)
	return buf
}

// base64 - Produces base64 of PostPolicy's Marshaled json.
func (p PostPolicy) base64() string {
	return base64.StdEncoding.EncodeToString(p.marshalJSON())
}

// safeAppendString JSON-escapes a string and appends it to the buffer.
// Unlike the standard library's encoder, it doesn't attempt to protect the
// user from browser vulnerabilities or JSONP-related problems.
func safeAppendString(buf []byte, s string) []byte {
	var ok bool
	for i := 0; i < len(s); {
		buf, ok = tryAddRuneSelf(buf, s[i])
		if ok {
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		buf, ok = tryAppendRuneError(buf, r, size)
		if ok {
			i++
			continue
		}
		buf = append(buf, s[i:i+size]...)
		i += size
	}
	return buf
}

// tryAddRuneSelf appends b if it is valid UTF-8 character represented in a single byte.
func tryAddRuneSelf(buf []byte, b byte) ([]byte, bool) {
	if b >= utf8.RuneSelf {
		return buf, false
	}
	if 0x20 <= b && b != '\\' && b != '"' {
		buf = append(buf, b)
		return buf, true
	}
	switch b {
	case '\\', '"':
		buf = append(buf, '\\')
		buf = append(buf, b)
	case '\n':
		buf = append(buf, '\\', 'n')
	case '\r':
		buf = append(buf, '\\', 'r')
	case '\t':
		buf = append(buf, '\\', 't')
	default:
		// Encode bytes < 0x20, except for the escape sequences above.
		buf = append(buf, `\u00`...)
		buf = append(buf, _hex[b>>4], _hex[b&0xF])
	}
	return buf, true
}

func tryAppendRuneError(buf []byte, r rune, size int) ([]byte, bool) {
	if r == utf8.RuneError && size == 1 {
		buf = append(buf, `\ufffd`...)
		return buf, true
	}
	return buf, false
}

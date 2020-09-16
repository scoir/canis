package ursa

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"
)

type CredentialValues struct {
	attrs map[string]CredentialAttributeValue
}

type CredentialAttributeValue struct {
	Raw     interface{} `json:"raw"`
	Encoded string      `json:"encoded"`
}

func NewValues() *CredentialValues {
	return &CredentialValues{
		attrs: map[string]CredentialAttributeValue{},
	}
}

func (r *CredentialValues) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.attrs)
}

func (r *CredentialValues) AddValue(name string, raw interface{}) {
	enc := r.encodeValue(raw)
	r.attrs[name] = CredentialAttributeValue{
		Raw:     raw,
		Encoded: enc,
	}
}

func (r *CredentialValues) encodeValue(raw interface{}) string {
	var enc string

	switch v := raw.(type) {
	case nil:
		enc = r.toEncodedNumber("None")
	case string:
		i, err := strconv.Atoi(v)
		if err == nil && (i <= math.MaxInt32 && i >= math.MinInt32) {
			enc = v
		} else {
			enc = r.toEncodedNumber(v)
		}
	case bool:
		if v {
			enc = "1"
		} else {
			enc = "0"
		}
	case int32:
		enc = strconv.Itoa(int(v))
	case int64:
		if v <= math.MaxInt32 && v >= math.MinInt32 {
			enc = strconv.Itoa(int(v))
		} else {
			enc = r.toEncodedNumber(strconv.Itoa(int(v)))
		}
	case int:
		if v <= math.MaxInt32 && v >= math.MinInt32 {
			enc = strconv.Itoa(v)
		} else {
			enc = r.toEncodedNumber(strconv.Itoa(v))
		}
	case float64:
		if v == 0 {
			enc = r.toEncodedNumber("0.0")
		} else {
			enc = r.toEncodedNumber(fmt.Sprintf("%f", v))
		}
	default:
		//Not sure what to do with Go and unknown types...  this works for now
		enc = r.toEncodedNumber(fmt.Sprintf("%v", v))
	}

	return enc
}

func (r *CredentialValues) toEncodedNumber(raw string) string {
	b := []byte(raw)
	hasher := sha256.New()
	hasher.Write(b)

	sh := hasher.Sum(nil)
	i := new(big.Int)
	i.SetBytes(sh)

	return i.String()
}

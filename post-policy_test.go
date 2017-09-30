package oss_addons

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type policyJSON struct {
	Expiration time.Time       `json:"expiration"`
	Conditions [][]interface{} `json:"conditions"`
}

func TestPostPolicy(t *testing.T) {
	expiresAt := time.Date(2017, 1, 23, 4, 5, 6, 0, time.UTC)
	contentLengthMin := rand.Int63n(1000)
	contentLengthMax := contentLengthMin + rand.Int63n(999999) + 1

	policy := NewPostPolicy()
	policy.SetExpires(expiresAt)
	policy.SetContentLengthRange(contentLengthMin, contentLengthMax)
	policy.SetBucket("test-bucket")
	policy.SetKey(`"test-object-name"`)

	jsonData := policy.marshalJSON()
	t.Logf("The output post policy is: %s", string(jsonData))

	var o policyJSON
	dec := json.NewDecoder(bytes.NewReader(jsonData))
	dec.UseNumber()
	err := dec.Decode(&o)
	if !assert.NoError(t, err, "The post policy should be a valid JSON string") {
		return
	}

	assert.Equal(t, expiresAt, o.Expiration)
	for _, pcond := range o.Conditions {
		if len(pcond) != 3 {
			t.Errorf("Unknown condition: %v", pcond)
			continue
		}

		if !assert.IsType(t, "", pcond[0]) {
			continue
		}

		matchType := pcond[0].(string)
		if matchType == "content-length-range" {
			min, _ := pcond[1].(json.Number).Int64()
			max, _ := pcond[2].(json.Number).Int64()
			assert.Equal(t, contentLengthMin, min)
			assert.Equal(t, contentLengthMax, max)
		} else {
			assert.Equal(t, "eq", matchType)
			condition := pcond[1].(string)
			switch condition {
			case "$bucket":
				assert.Equal(t, "test-bucket", pcond[2])
			case "$key":
				assert.Equal(t, `"test-object-name"`, pcond[2])
			default:
				t.Errorf("Unknown condition: %s", condition)
			}
		}
	}
}

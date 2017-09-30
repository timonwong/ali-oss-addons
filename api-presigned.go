package oss_addons

import (
	"errors"
	"net/url"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/timonwong/ali-oss-addons/signer"
)

// PresignedPostPolicyV1 returns POST urlString, form data to upload an object.
func PresignedPostPolicyV1(c *oss.Client, p *PostPolicy) (u *url.URL, formData map[string]string, err error) {
	// Validate input arguments.
	if p.expiration.IsZero() {
		return nil, nil, errors.New("expiration time must be specified")
	}
	if _, ok := p.formData["key"]; !ok {
		return nil, nil, errors.New("object key must be specified")
	}
	if _, ok := p.formData["bucket"]; !ok {
		return nil, nil, errors.New("bucket name must be specified")
	}

	bucketName := p.formData["bucket"]

	// Build target url
	u, err = url.Parse(c.Config.Endpoint)
	if err != nil {
		return nil, nil, err
	}

	if !c.Config.IsCname {
		u.Path = "/" + bucketName
	}

	policyBase64 := p.base64()
	p.formData["policy"] = policyBase64
	p.formData["OSSAccessKeyId"] = c.Config.AccessKeyID
	// Sign the policy.
	p.formData["signature"] = signer.PostPresignSignatureV1(policyBase64, c.Config.AccessKeySecret)
	return u, p.formData, nil
}

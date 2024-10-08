package s3

import (
	"github.com/ks3sdklib/aws-sdk-go/aws"
)

type BucketMirror struct {
	Version          *string            `json:"version" type:"string" locationName:"version"`
	UseDefaultRobots *bool              `json:"use_default_robots" locationName:"use_default_robots"`
	AsyncMirrorRule  *AsyncMirrorRule   `json:"async_mirror_rule,omitempty" type:"structure" locationName:"async_mirror_rule"`
	SyncMirrorRules  []*SyncMirrorRules `json:"sync_mirror_rules,omitempty" type:"list" locationName:"sync_mirror_rules"`
}
type SavingSetting struct {
	ACL *string `json:"acl,omitempty"  required:"true" locationName:"acl"`
}
type AsyncMirrorRule struct {
	MirrorUrls    []*string      `json:"mirror_urls,omitempty" required:"true" locationName:"mirror_urls"`
	SavingSetting *SavingSetting `json:"saving_setting,omitempty" required:"true" locationName:"saving_setting"`
}
type MatchCondition struct {
	HTTPCodes   []*string `json:"http_codes" locationName:"http_codes"`
	KeyPrefixes []*string `json:"key_prefixes" locationName:"key_prefixes"`
}
type SetHeaders struct {
	Key   *string `json:"key,omitempty" locationName:"key"`
	Value *string `json:"value,omitempty" locationName:"value"`
}
type RemoveHeaders struct {
	Key *string `json:"key,omitempty" locationName:"key"`
}
type PassHeaders struct {
	Key *string `json:"key,omitempty" locationName:"key"`
}
type HeaderSetting struct {
	SetHeaders    []*SetHeaders    `json:"set_headers,omitempty" locationName:"set_headers"`
	RemoveHeaders []*RemoveHeaders `json:"remove_headers,omitempty" locationName:"remove_headers"`
	PassAll       *bool            `json:"pass_all,omitempty" locationName:"pass_all"`
	PassHeaders   []*PassHeaders   `json:"pass_headers,omitempty" locationName:"pass_headers"`
}
type MirrorRequestSetting struct {
	PassQueryString *bool          `json:"pass_query_string,omitempty" locationName:"pass_query_string"`
	Follow3Xx       *bool          `json:"follow3xx,omitempty" locationName:"follow3xx"`
	HeaderSetting   *HeaderSetting `json:"header_setting,omitempty" locationName:"header_setting"`
}
type SyncMirrorRules struct {
	MatchCondition       MatchCondition        `json:"match_condition" locationName:"match_condition"`
	MirrorURL            *string               `json:"mirror_url,omitempty" locationName:"mirror_url"`
	MirrorRequestSetting *MirrorRequestSetting `json:"mirror_request_setting,omitempty" locationName:"mirror_request_setting"`
	SavingSetting        *SavingSetting        `json:"saving_setting,omitempty" locationName:"saving_setting"`
}

var opPutBucketMirror *aws.Operation

type PutBucketMirrorInput struct {
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	BucketMirror *BucketMirror `locationName:"BucketMirror" json:"-" type:"structure"`

	ContentType *string `location:"header" locationName:"Content-Type" type:"string"`

	metadataPutBucketMirrorInput `json:"-" xml:"-"`
}

type metadataPutBucketMirrorInput struct {
	SDKShapeTraits bool `type:"structure" payload:"BucketMirror"`
}

type PutBucketMirrorOutput struct {
	Metadata map[string]*string `location:"headers"  type:"map"`

	StatusCode *int64 `location:"statusCode" type:"integer"`
}

var opGetBucketMirror *aws.Operation

type GetBucketMirrorInput struct {
	Bucket      *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`
	ContentType *string `location:"header" locationName:"Content-Type" type:"string"`
}
type GetBucketMirrorOutput struct {
	BucketMirror *BucketMirror `locationName:"BucketMirror" type:"structure"`

	Metadata map[string]*string `location:"headers"  type:"map"`

	StatusCode *int64 `location:"statusCode" type:"integer"`

	metadataGetBucketMirrorInput `json:"-" xml:"-"`
}

type metadataGetBucketMirrorInput struct {
	SDKShapeTraits bool `type:"structure" payload:"BucketMirror"`
}

var opDeleteBucketMirror *aws.Operation

type DeleteBucketMirrorInput struct {
	Bucket      *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`
	ContentType *string `location:"header" locationName:"Content-Type" type:"string"`
}
type DeleteBucketMirrorOutput struct {
	Metadata map[string]*string `location:"headers"  type:"map"`

	StatusCode *int64 `location:"statusCode" type:"integer"`
}

func (c *S3) PutBucketMirrorRequest(input *PutBucketMirrorInput) (req *aws.Request, output *PutBucketMirrorOutput) {
	oprw.Lock()
	defer oprw.Unlock()
	if opPutBucketMirror == nil {
		opPutBucketMirror = &aws.Operation{
			Name:       "PutBucketMirror",
			HTTPMethod: "PUT",
			HTTPPath:   "/{Bucket}?mirror",
		}
	}
	if input == nil {
		input = &PutBucketMirrorInput{}
	}
	req = c.newRequest(opPutBucketMirror, input, output)
	req.ContentType = "application/json"
	output = &PutBucketMirrorOutput{}
	req.Data = output
	return
}

func (c *S3) PutBucketMirror(input *PutBucketMirrorInput) (*PutBucketMirrorOutput, error) {
	req, out := c.PutBucketMirrorRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) PutBucketMirrorWithContext(ctx aws.Context, input *PutBucketMirrorInput) (*PutBucketMirrorOutput, error) {
	req, out := c.PutBucketMirrorRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

func (c *S3) GetBucketMirrorRequest(input *GetBucketMirrorInput) (req *aws.Request, output *GetBucketMirrorOutput) {
	oprw.Lock()
	defer oprw.Unlock()
	if opGetBucketMirror == nil {
		opGetBucketMirror = &aws.Operation{
			Name:       "GetBucketMirror",
			HTTPMethod: "GET",
			HTTPPath:   "/{Bucket}?mirror",
		}
	}
	if input == nil {
		input = &GetBucketMirrorInput{}
	}
	req = c.newRequest(opGetBucketMirror, input, output)
	req.ContentType = "application/json"
	output = &GetBucketMirrorOutput{
		BucketMirror: &BucketMirror{},
	}
	req.Data = output
	return
}

func (c *S3) GetBucketMirror(input *GetBucketMirrorInput) (*GetBucketMirrorOutput, error) {
	req, out := c.GetBucketMirrorRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) GetBucketMirrorWithContext(ctx aws.Context, input *GetBucketMirrorInput) (*GetBucketMirrorOutput, error) {
	req, out := c.GetBucketMirrorRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

func (c *S3) DeleteBucketMirrorRequest(input *DeleteBucketMirrorInput) (req *aws.Request, output *DeleteBucketMirrorOutput) {
	oprw.Lock()
	defer oprw.Unlock()
	if opDeleteBucketMirror == nil {
		opDeleteBucketMirror = &aws.Operation{
			Name:       "DeleteBucketMirror",
			HTTPMethod: "DELETE",
			HTTPPath:   "/{Bucket}?mirror",
		}
	}
	if input == nil {
		input = &DeleteBucketMirrorInput{}
	}
	req = c.newRequest(opDeleteBucketMirror, input, output)
	output = &DeleteBucketMirrorOutput{}
	req.Data = output
	return
}

func (c *S3) DeleteBucketMirror(input *DeleteBucketMirrorInput) (*DeleteBucketMirrorOutput, error) {
	req, out := c.DeleteBucketMirrorRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) DeleteBucketMirrorWithContext(ctx aws.Context, input *DeleteBucketMirrorInput) (*DeleteBucketMirrorOutput, error) {
	req, out := c.DeleteBucketMirrorRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

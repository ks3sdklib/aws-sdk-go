package s3

// HTTP headers
const (
	HTTPHeaderAcceptEncoding     string = "Accept-Encoding"
	HTTPHeaderAuthorization             = "Authorization"
	HTTPHeaderCacheControl              = "Cache-Control"
	HTTPHeaderContentDisposition        = "Content-Disposition"
	HTTPHeaderContentEncoding           = "Content-Encoding"
	HTTPHeaderContentLength             = "Content-Length"
	HTTPHeaderContentMD5                = "Content-MD5"
	HTTPHeaderContentType               = "Content-Type"
	HTTPHeaderContentLanguage           = "Content-Language"
	HTTPHeaderDate                      = "Date"
	HTTPHeaderEtag                      = "ETag"
	HTTPHeaderExpires                   = "Expires"
	HTTPHeaderHost                      = "Host"
	HTTPHeaderkssACL                    = "X-kss-Acl"

	ChannelBuf  int = 1000
	PartSize5MB     = 5 * 1024 * 1024 // part size, 5MB
	MinPartSize     = 100 * 1024      // Min part size, 100KB
)

// ACL
const (
	ACLPrivate         string = "private"
	ACLPublicRead      string = "public-read"
	ACLPublicReadWrite string = "public-read-write"
)

// StorageClass
const (
	StorageClassExtremePL3      string = "EXTREME_PL3"
	StorageClassExtremePL2      string = "EXTREME_PL2"
	StorageClassExtremePL1      string = "EXTREME_PL1"
	StorageClassStandard        string = "STANDARD"
	StorageClassIA              string = "STANDARD_IA"
	StorageClassDeepIA          string = "DEEP_IA"
	StorageClassArchive         string = "ARCHIVE"
	StorageClassDeepColdArchive string = "DEEP_COLD_ARCHIVE"
)

// BucketType
const (
	BucketTypeExtremePL3 string = "EXTREME_PL3"
	BucketTypeExtremePL2 string = "EXTREME_PL2"
	BucketTypeExtremePL1 string = "EXTREME_PL1"
	BucketTypeNormal     string = "NORMAL"
	BucketTypeIA         string = "IA"
	BucketTypeDeepIA     string = "DEEP_IA"
	BucketTypeArchive    string = "ARCHIVE"
)

type HTTPMethod string

const (
	PUT    HTTPMethod = "PUT"
	GET    HTTPMethod = "GET"
	DELETE HTTPMethod = "DELETE"
	HEAD   HTTPMethod = "HEAD"
	POST   HTTPMethod = "POST"
)

const AllUsersUri = "http://acs.amazonaws.com/groups/global/AllUsers"

type CannedAccessControlType int32

const (
	PublicReadWrite CannedAccessControlType = 0
	PublicRead      CannedAccessControlType = 1
	Private         CannedAccessControlType = 2
)

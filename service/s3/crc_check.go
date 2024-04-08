package s3

import (
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/internal/apierr"
	"hash"
	"strconv"
)

func CheckUploadCrc64(r *aws.Request) {
	clientCrc := r.Crc64.Sum64()
	serverCrc, _ := strconv.ParseUint(r.HTTPResponse.Header.Get("X-Amz-Checksum-Crc64ecma"), 10, 64)

	r.Config.WriteLog(aws.LogOn, "client crc:%d, server crc:%d\n", clientCrc, serverCrc)

	if r.HTTPResponse.Header.Get("X-Amz-Checksum-Crc64ecma") != "" && clientCrc != serverCrc {
		r.Error = apierr.New("CRCCheckError", "client crc and server crc do not match", nil)
		r.Config.WriteLog(aws.LogOn, "error:%s\n", r.Error.Error())
	}
}

func CheckDownloadCrc64(s3 *S3, res *GetObjectOutput, crc hash.Hash64) error {
	var err error
	clientCrc := crc.Sum64()
	var serverCrc uint64
	if res.Metadata["X-Amz-Checksum-Crc64ecma"] == nil {
		serverCrc = 0
	} else {
		serverCrc, _ = strconv.ParseUint(*res.Metadata["X-Amz-Checksum-Crc64ecma"], 10, 64)
	}

	s3.Config.WriteLog(aws.LogOn, "client crc:%d, server crc:%d\n", clientCrc, serverCrc)

	if serverCrc != 0 && clientCrc != serverCrc {
		err = apierr.New("CRCCheckError", "client crc and server crc do not match", nil)
		s3.Config.WriteLog(aws.LogOn, "error:%s\n", err.Error())
	}

	return err
}

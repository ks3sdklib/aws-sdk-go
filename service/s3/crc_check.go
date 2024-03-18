package s3

import (
	"fmt"
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/internal/apierr"
	"hash"
	"strconv"
)

func CheckUploadCrc64(r *aws.Request) {
	clientCrc := r.Crc64.Sum64()
	serverCrc, _ := strconv.ParseUint(r.HTTPResponse.Header.Get("X-Amz-Checksum-Crc64ecma"), 10, 64)

	if r.HTTPResponse.Header.Get("X-Amz-Checksum-Crc64ecma") != "" && clientCrc != serverCrc {
		r.Error = apierr.New("CRCCheckError", "client crc and server crc do not match", nil)
	}
	if r.Config.LogLevel > 0 {
		out := r.Config.Logger
		fmt.Fprintln(out, "---[ CHECK CRC64 ]--------------------------------")
		fmt.Fprintln(out, "client crc:", clientCrc, "server crc:", serverCrc)
		if r.Error != nil {
			fmt.Fprintln(out, r.Error.Error())
		}
		fmt.Fprintln(out, "-----------------------------------------------------")
	}
}

func CheckDownloadCrc64(c *S3, res *GetObjectOutput, crc hash.Hash64) error {
	var err error
	clientCrc := crc.Sum64()
	serverCrc, _ := strconv.ParseUint(*res.Metadata["X-Amz-Checksum-Crc64ecma"], 10, 64)

	if *res.Metadata["X-Amz-Checksum-Crc64ecma"] != "" && clientCrc != serverCrc {
		err = apierr.New("CRCCheckError", "client crc and server crc do not match", nil)
	}

	if c.Config.LogLevel > 0 {
		out := c.Config.Logger
		fmt.Fprintln(out, "---[ CHECK CRC64 ]--------------------------------")
		fmt.Fprintln(out, "client crc:", clientCrc, "server crc:", serverCrc)
		if err != nil {
			fmt.Fprintln(out, err.Error())
		}
		fmt.Fprintln(out, "-----------------------------------------------------")
	}

	return err
}

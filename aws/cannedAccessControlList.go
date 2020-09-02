package aws

import "aws-sdk-go/service/s3"
import "github.com/deckarep/golang-set"

type CannedAccessControlType int32
const (
	PublicReadWrite      CannedAccessControlType = 0
	PublicRead      CannedAccessControlType = 1
	Private      CannedAccessControlType = 2
)
const AllUsersUri = "http://acs.amazonaws.com/groups/global/AllUsers"

func GetAcl(resp s3.GetObjectACLOutput) (CannedAccessControlType)  {

	allUsersPermissions := mapset.NewSet()
	for _, value:= range resp.Grants {
		if value.Grantee.URI !=nil && *value.Grantee.URI == AllUsersUri{
			allUsersPermissions.Add(value.Permission)
		}
	}
	read := allUsersPermissions.Contains("READ");
	write := allUsersPermissions.Contains("WRITE");
	if (read && write) {
		return PublicReadWrite;
	} else if (read) {
		return PublicRead;
	} else {
		return Private;
	}
}

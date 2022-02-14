package qsgpm

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"
)

type User struct {
	types.User
	Namespace string
}

func (u *User) String() string {
	return fmt.Sprintf("user<%s %s, CustomPermission:%s>", u.IdentityType, *u.UserName, viewStarString(u.CustomPermissionsName))
}

func (u *User) SessionName() string {
	if u.IdentityType != types.IdentityTypeIam {
		return ""
	}
	parts := strings.SplitAfterN(*u.UserName, "/", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

func (u *User) IAMRoleName() string {
	if u.IdentityType != types.IdentityTypeIam {
		return ""
	}
	parts := strings.SplitAfterN(*u.UserName, "/", 2)
	if len(parts) == 2 {
		return strings.TrimRight(parts[0], "/")
	}
	return ""
}

func (u *User) IsNeedUpdateCustomPermission(customPermissionName *string) bool {
	if u.CustomPermissionsName == nil && customPermissionName == nil {
		return false
	}
	if u.CustomPermissionsName != nil && customPermissionName != nil {
		return *u.CustomPermissionsName != *customPermissionName
	}
	return true
}

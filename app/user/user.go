package user

import (
	"log"
)
type UserPermission int

const (
	View UserPermission  = 0
	Admin UserPermission = 99999
)

type UserInfo struct {
	Name       string
	Permission UserPermission
}

func DefaultUserInfo() UserInfo {
	return UserInfo{
		Name:       "anonymous",
		Permission: View,
	}
}

func (u UserInfo) HasDeletePerm() bool {
	return u.Permission == Admin
}

func (u UserInfo) CheckReportAccess(expectedAccessToken string, givenAccessToken string) bool {
	//this denies view access to all anon users for old reports that don't have an access token
	if givenAccessToken != expectedAccessToken || expectedAccessToken == "" {
		log.Printf("user %s: wrong or missing report access token (provided %s, should be %s) - checking admin perms", u.Name, givenAccessToken, expectedAccessToken)
		//TODO: allow user account access to specific reports depending on criteria (e.g. crash plugin)
		if u.Permission != Admin {
			log.Printf("user %s does not have permission to access report", u.Name)
			return false
		}
	}

	return true
}

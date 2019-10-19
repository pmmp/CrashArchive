package user

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
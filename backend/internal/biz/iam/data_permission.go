package iam

// DataPermissionScope 是按当前登录人计算出的数据权限过滤条件。
type DataPermissionScope struct {
	DataScope int8
	UserID    int64
	DeptID    int64
	DeptIDs   []int64
}

func widestDataScope(roles []*Role) int8 {
	if len(roles) == 0 {
		return DataScopeSelf
	}
	hasDeptSub := false
	hasDept := false
	for _, role := range roles {
		switch role.DataScope {
		case DataScopeAll:
			return DataScopeAll
		case DataScopeDeptSub:
			hasDeptSub = true
		case DataScopeDept:
			hasDept = true
		}
	}
	if hasDeptSub {
		return DataScopeDeptSub
	}
	if hasDept {
		return DataScopeDept
	}
	return DataScopeSelf
}

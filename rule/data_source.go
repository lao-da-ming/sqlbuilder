package rule

import (
	"errors"
	"github.com/duke-git/lancet/v2/slice"
)

// createBy维度
func createByDimensional(loginEmployee int64, Elements []Element) (userIds []int64, err error) {
	//每个元素之间是或关系，取并集
	for _, item := range Elements {
		var uids []int64
		switch item.Type {
		case "self": //员工自身
			uids, err = self(loginEmployee, item.Value)
		case "organization": //组织
			uids, err = organization(loginEmployee, item.Value)
		case "custom": //自定义
			uids, err = custom(loginEmployee, item.Value)
		}
		if err != nil {
			return nil, err
		}
		userIds = append(userIds, uids...)
	}
	return slice.Unique(userIds), nil
}

// 组织维度
func positionDimensional(loginEmployee int64, Elements []Element) (userIds []int64, err error) {
	//每个元素之间是或关系，取并集
	for _, item := range Elements {
		var uids []int64
		switch item.Type {
		case "self":
			uids, err = selfOrg(loginEmployee, item.Value)
		}
		if err != nil {
			return nil, err
		}
		userIds = append(userIds, uids...)
	}
	return slice.Unique(userIds), nil
}

// 本人所在组织
func selfOrg(loginEmployee int64, value interface{}) ([]int64, error) {
	if value.(bool) {
		return []int64{11, 22, 33, 44}, nil
	}
	return nil, nil
}

// 获取自定义权限人员定义
func custom(loginEmployee int64, value interface{}) ([]int64, error) {
	return []int64{11, 50, 80, 90}, nil
}

// 本人
func self(loginEmployee int64, value interface{}) ([]int64, error) {
	if value.(bool) {
		return []int64{loginEmployee}, nil
	}
	return nil, nil
}

// 所在组织||所在组织含下级
func organization(loginEmployee int64, value interface{}) ([]int64, error) {
	switch value.(string) {
	case "directly":
		//查询用户所在组织直接人员
		return []int64{1}, nil
	case "all":
		//查询用户所在组织（含下级）所有人员
		return []int64{1, 2, 3}, nil
	default:
		return nil, errors.New("invalid value of organization type")
	}
}

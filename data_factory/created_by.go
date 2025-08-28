package data_factory

import (
	"context"
	"errors"
	"github.com/duke-git/lancet/v2/slice"
)

type UserSource struct {
}

// createBy维度
func (u *UserSource) GetData(ctx context.Context, userId int64, Elements []Element) (result []any, err error) {
	//每个元素之间是或关系，取并集
	for _, item := range Elements {
		var uids []any
		switch item.Type {
		case "self": //员工自身
			uids, err = self(ctx, userId, item.Value)
		case "organization": //组织
			uids, err = organization(ctx, userId, item.Value)
		case "custom": //自定义
			uids, err = custom(ctx, userId, item.Value)
		}
		if err != nil {
			return nil, err
		}
		result = append(result, uids...)
	}
	return slice.Unique(result), nil
}

// 本人
func self(ctx context.Context, loginEmployee int64, value any) ([]any, error) {
	if value.(bool) {
		return []any{loginEmployee}, nil
	}
	return nil, nil
}

// 获取自定义权限人员定义
func custom(ctx context.Context, loginEmployee int64, value any) ([]any, error) {
	return []any{11, 50, 80, 90}, nil
}

// 所在组织||所在组织含下级
func organization(ctx context.Context, loginEmployee int64, value any) ([]any, error) {
	switch value.(string) {
	case "directly":
		//查询用户所在组织直接人员
		return []any{1}, nil
	case "all":
		//查询用户所在组织（含下级）所有人员
		return []any{1, 2, 3}, nil
	default:
		return nil, errors.New("invalid value of organization type")
	}
}

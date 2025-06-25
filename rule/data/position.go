package data

import (
	"context"
	"github.com/duke-git/lancet/v2/slice"
)

// 组织维度
func PositionDimensional(ctx context.Context, loginEmployee int64, Elements []Element) (positionIds []any, err error) {
	//每个元素之间是或关系，取并集
	for _, item := range Elements {
		var pids []any
		switch item.Type {
		case "self":
			pids, err = selfOrg(ctx, loginEmployee, item.Value)
		}
		if err != nil {
			return nil, err
		}
		positionIds = append(positionIds, pids...)
	}
	return slice.Unique(positionIds), nil
}

// 本人所在组织
func selfOrg(ctx context.Context, loginEmployee int64, value any) ([]any, error) {
	if value.(bool) {
		return []any{"11", "22", "33", "44"}, nil
	}
	return nil, nil
}

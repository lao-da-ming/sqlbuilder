package data_factory

import (
	"context"
	"github.com/duke-git/lancet/v2/slice"
)

type PositionSource struct {
}

// 组织维度
func (p *PositionSource) GetData(ctx context.Context, userId int64, Elements []Element) (result []any, err error) {
	//每个元素之间是或关系，取并集
	for _, item := range Elements {
		var pids []any
		switch item.Type {
		case "self":
			pids, err = selfOrg(ctx, userId, item.Value)
		}
		if err != nil {
			return nil, err
		}
		result = append(result, pids...)
	}
	return slice.Unique(result), nil
}

// 本人所在组织
func selfOrg(ctx context.Context, loginEmployee int64, value any) ([]any, error) {
	if value.(bool) {
		return []any{"11", "22", "33", "44"}, nil
	}
	return nil, nil
}

package data_factory

import (
	"context"
)

type EmployeePositionSource struct {
}

// 组织维度
func (ep *EmployeePositionSource) GetData(ctx context.Context, userId int64, Elements []Element) (result []any, err error) {
	return
}

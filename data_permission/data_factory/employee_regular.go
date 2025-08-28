package data_factory

import "context"

type EmployeeRegularSource struct {
}

// 组织维度
func (er *EmployeeRegularSource) GetData(ctx context.Context, userId int64, Elements []Element) (result []any, err error) {
	return
}

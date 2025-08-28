package data_factory

import "context"

type EmployeeStatusSource struct {
}

// 组织维度
func (es *EmployeeStatusSource) GetData(ctx context.Context, userId int64, Elements []Element) (result []any, err error) {
	return
}

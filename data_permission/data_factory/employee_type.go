package data_factory

import "context"

type EmployeeTypeSource struct {
}

// 组织维度
func (ep *EmployeeTypeSource) GetData(ctx context.Context, userId int64, Elements []Element) (result []any, err error) {
	return
}

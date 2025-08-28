package data_factory

import "context"

type EmployeeTraineeSource struct {
}

// 组织维度
func (et *EmployeeTraineeSource) GetData(ctx context.Context, userId int64, Elements []Element) (result []any, err error) {
	return
}

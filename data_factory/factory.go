package data_factory

import (
	"context"
	"errors"
)

// 最小单元结构
type Element struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// 数据字段
type DbField string

const (
	CreatedBy        DbField = "created_by"        //创建人
	Position         DbField = "position"          //岗位.
	EmployeePosition DbField = "employee_position" //任职岗位
	EmployeeRegular  DbField = "employee_regular"  //转正状态
	EmployeeType     DbField = "employee_type"     //用工类型
	EmployeeStatus   DbField = "employee_status"   //在职状态
	EmployeeTrainee  DbField = "employee_trainee"  //培训生状态
)

// 数据源实例字典，todo 新增的规则在这里添加
var mapDataSource = map[DbField]DataSource{
	CreatedBy:        &UserSource{},
	Position:         &PositionSource{},
	EmployeePosition: &EmployeePositionSource{},
	EmployeeRegular:  &EmployeeRegularSource{},
	EmployeeType:     &EmployeeTypeSource{},
	EmployeeStatus:   &EmployeeStatusSource{},
	EmployeeTrainee:  &EmployeeTraineeSource{},
}

// 获取数据源
type DataSource interface {
	GetData(ctx context.Context, userId int64, Elements []Element) (result []any, err error) //获取源数据
}

// new工厂实例
func NewDataSource(field DbField) (DataSource, error) {
	source, ok := mapDataSource[field]
	if !ok {
		return nil, errors.New("invalid db field")
	}
	return source, nil
}

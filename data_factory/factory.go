package data_factory

import (
	"context"
	"errors"
	"sync"
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

// 为 DbField 实现 String 方法
func (d DbField) String() string {
	return string(d)
}

// 定义错误类型
var (
	ErrInvalidDbField = errors.New("invalid db field")
)

// 数据源接口
type DataSource interface {
	GetData(ctx context.Context, userID int64, elements []Element) (result []any, err error)
}

// 私有化映射，避免外部直接修改
var mapDataSource sync.Map

// 初始化注册
func init() {
	//todo 后面追加规则在这里追加
	RegisterDataSource(CreatedBy, &UserSource{})
	RegisterDataSource(Position, &PositionSource{})
	RegisterDataSource(EmployeePosition, &EmployeePositionSource{})
	RegisterDataSource(EmployeeRegular, &EmployeeRegularSource{})
	RegisterDataSource(EmployeeType, &EmployeeTypeSource{})
	RegisterDataSource(EmployeeStatus, &EmployeeStatusSource{})
	RegisterDataSource(EmployeeTrainee, &EmployeeTraineeSource{})
}

// 注册数据源，可用于动态注入
func RegisterDataSource(field DbField, source DataSource) {
	if source == nil {
		return
	}
	//已经存在不允许修改覆盖
	_, ok := mapDataSource.Load(field)
	if ok {
		return
	}
	mapDataSource.Store(field, source)
}

// new工厂实例
func NewDataSource(field DbField) (DataSource, error) {
	source, ok := mapDataSource.Load(field)
	if !ok {
		return nil, ErrInvalidDbField
	}
	return source.(DataSource), nil
}

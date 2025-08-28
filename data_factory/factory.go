package data_factory

import "context"

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

// 获取数据源
type DataSource interface {
	GetData(ctx context.Context, userId int64, Elements []Element) (result []any, err error) //获取源数据
}

// new工厂实例
func NewDataSource(field DbField) DataSource {
	//todo 新增数据权限规则在这里新增
	switch field {
	case CreatedBy: //创建人
		return &UserSource{}
	case Position: //岗位
		return &PositionSource{}
	case EmployeePosition: //任职岗位
		return &EmployeePositionSource{}
	case EmployeeRegular: //转正状态
		return &EmployeeRegularSource{}
	case EmployeeType: //用工类型
		return &EmployeeTypeSource{}
	case EmployeeStatus: //在职状态
		return &EmployeeStatusSource{}
	case EmployeeTrainee: //培训生状态
		return &EmployeeTraineeSource{}
	default:
		return nil
	}
}

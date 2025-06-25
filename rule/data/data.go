package data

// 最小单元
type Element struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// 数据字段
type DbField string

const (
	CreatedBy DbField = "created_by"
	Position  DbField = "position"
)

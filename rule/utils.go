package rule

import (
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"reflect"
	"sqlbuilder/rule/data"
	"strings"
)

// 构建最终的sql
func buildSql(includeValues, excludeValues map[data.DbField][]any) (string, error) {
	lengthInclude := len(includeValues)
	lengthExclude := len(excludeValues)
	//拼接sql
	index := 0
	sqlBuilder := strings.Builder{}
	sqlBuilder.WriteString(" (")
	for field, value := range includeValues {
		index++
		sqlBuilder.WriteString(string(field))
		sqlBuilder.WriteString(" IN (")
		sqlBuilder.WriteString(joinArrayElementForSql(value, ","))
		sqlBuilder.WriteString(")")
		if index != lengthInclude {
			sqlBuilder.WriteString(" OR ")
		}
	}
	sqlBuilder.WriteString(") ")
	if lengthExclude == 0 {
		return sqlBuilder.String(), nil
	}
	sqlBuilder.WriteString("AND (")
	//初始化index
	index = 0
	for field, value := range excludeValues {
		index++
		sqlBuilder.WriteString(string(field))
		sqlBuilder.WriteString(" NOT IN (")
		sqlBuilder.WriteString(joinArrayElementForSql(value, ","))
		sqlBuilder.WriteString(")")
		if index != lengthExclude {
			sqlBuilder.WriteString(" OR ")
		}
	}
	sqlBuilder.WriteString(") ")
	return sqlBuilder.String(), nil
}

// 处理别名
func setAlias(includeValues, excludeValues map[data.DbField][]any, fieldAlias map[data.DbField]data.DbField) {
	for field, alias := range fieldAlias {
		if value, ok := includeValues[field]; ok {
			includeValues[alias] = value
			delete(includeValues, field)
		}
		if value, ok := excludeValues[field]; ok {
			excludeValues[alias] = value
			delete(excludeValues, field)
		}
	}
}

// 数组元素连接
func joinArrayElementForSql(values []any, separator string) string {
	lenValues := len(values)
	if lenValues == 0 {
		return ""
	}
	itemType := reflect.TypeOf(values[0]).String()
	strBuilder := strings.Builder{}
	switch itemType {
	case "string": //字符串类型
		for key, item := range values {
			strBuilder.WriteString(fmt.Sprintf("'%v'", item))
			if key != lenValues-1 {
				strBuilder.WriteString(separator)
			}
		}
	default: //其他类型型
		strBuilder.WriteString(slice.Join(values, separator))
	}
	return strBuilder.String()
}

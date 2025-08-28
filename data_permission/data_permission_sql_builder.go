package data_permission

import (
	"context"
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"reflect"
	"sqlbuilder/data_permission/data_factory"
	"strings"
)

type DataPermissionSqlBuilder struct {
}

func NewDataPermissionSqlBuilder() *DataPermissionSqlBuilder {
	return &DataPermissionSqlBuilder{}
}

// 获取可以访问数据的sql条件，返回空sql表示无任何权限
func (d *DataPermissionSqlBuilder) GetDataPermissionSql(ctx context.Context, loginEmployee int64, fieldAlias map[data_factory.DbField]string, include, exclude [][]map[data_factory.DbField][]data_factory.Element) (sql string, err error) {
	//无权限返回的sql
	denySql := " 1=2 "
	includeValues, err := d.getFieldValues(ctx, loginEmployee, include)
	if err != nil {
		return "", err
	}
	//log.Println("指定的:", includeValues)
	//包含的为空则无任何数据权限
	if len(includeValues) == 0 {
		return denySql, nil
	}
	//排除部分这层每个元素之间都是or关系
	excludeValues, err := d.getFieldValues(ctx, loginEmployee, exclude)
	if err != nil {
		return "", err
	}
	//log.Println("排除的:", excludeValues)
	//包含跟排除之间对应字段取差集(map引用无需返回值)
	for field, _ := range includeValues {
		//不存在排除，继续下一个
		if _, ok := excludeValues[field]; !ok {
			continue
		}
		//存在则取差集
		includeValues[field] = slice.Difference(includeValues[field], excludeValues[field])
		//and条件如果include差集为空，则全部不合适
		if len(includeValues[field]) == 0 {
			fmt.Println(fmt.Sprintf("字段:%s取差集后为空集合", field))
			return denySql, nil
		}
		//取了差集，删掉排除对应的字段
		delete(excludeValues, field)
	}
	//log.Printf("最终差集为指定:%v  排除:%v \r\n", includeValues, excludeValues)
	//处理别名(map引用无需返回值)
	d.setAlias(includeValues, excludeValues, fieldAlias)
	//拼接sql
	return d.buildSql(includeValues, excludeValues)
}

// 获取对应字段的拥有可查看的id值
func (d *DataPermissionSqlBuilder) getFieldValues(ctx context.Context, loginEmployee int64, fieldValues [][]map[data_factory.DbField][]data_factory.Element) (map[data_factory.DbField][]any, error) {
	if len(fieldValues) == 0 {
		return nil, nil
	}
	//每个元素之间是or关系，取并集
	mapFieldValues := make(map[data_factory.DbField][]any, len(fieldValues))
	for _, firstLevelItem := range fieldValues {
		secondLevelValues, err := d.secondLevel(ctx, loginEmployee, firstLevelItem)
		if err != nil {
			return nil, err
		}
		//这层是or，没有就跳过
		if len(secondLevelValues) == 0 {
			continue
		}
		//取对应字段并集
		for field, secondLevelValue := range secondLevelValues {
			//如果原来没有
			if _, ok := mapFieldValues[field]; !ok {
				mapFieldValues[field] = secondLevelValue
				continue
			}
			//有了就取并集
			mapFieldValues[field] = append(mapFieldValues[field], secondLevelValue...)
			mapFieldValues[field] = slice.Unique(mapFieldValues[field])
		}
	}
	return mapFieldValues, nil
}

// 第2层(元素之间与逻辑)
func (d *DataPermissionSqlBuilder) secondLevel(ctx context.Context, loginEmployee int64, firstLevelItem []map[data_factory.DbField][]data_factory.Element) (map[data_factory.DbField][]any, error) {
	//这层每个元素之间都是and关系,取交集
	mapFieldValues := make(map[data_factory.DbField][]any, len(firstLevelItem))
	for _, secondLevelItem := range firstLevelItem {
		thirdLevelValues, err := d.thirdLevel(ctx, loginEmployee, secondLevelItem)
		if err != nil {
			return nil, err
		}
		//and有不成立直接返回
		if len(thirdLevelValues) == 0 {
			return nil, nil
		}
		//取对应字段的交集(这里循环其实只有一个元素)
		for field, thirdLevelValue := range thirdLevelValues {
			_, ok := mapFieldValues[field]
			//不存在，则直接赋值
			if !ok {
				mapFieldValues[field] = thirdLevelValue
				continue
			}
			//取交集
			mapFieldValues[field] = slice.Intersection(mapFieldValues[field], thirdLevelValue)
			//去重
			mapFieldValues[field] = slice.Unique(mapFieldValues[field])
			//and关系判断交集,如果是空的则全部不成立
			if len(mapFieldValues[field]) == 0 {
				return nil, nil
			}
		}
	}
	return mapFieldValues, nil
}

// 第3层
func (d *DataPermissionSqlBuilder) thirdLevel(ctx context.Context, loginEmployee int64, secondLevelItem map[data_factory.DbField][]data_factory.Element) (map[data_factory.DbField][]any, error) {
	//map只有一个元素
	mapFieldValues := make(map[data_factory.DbField][]any, 1)
	//其实只有一个元素
	for field, value := range secondLevelItem {
		//根据field获取数据
		source, err := data_factory.NewDataSource(field)
		if err != nil {
			return nil, err
		}
		collections, err := source.GetData(ctx, loginEmployee, value)
		if err != nil {
			return nil, err
		}
		if len(collections) == 0 {
			return nil, nil
		}
		mapFieldValues[field] = collections
		break
	}
	return mapFieldValues, nil
}

// ##########################辅助函数#######################################
// 构建最终的sql
func (d *DataPermissionSqlBuilder) buildSql(includeValues, excludeValues map[data_factory.DbField][]any) (string, error) {
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
		sqlBuilder.WriteString(d.joinArrayElementForSql(value, ","))
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
		sqlBuilder.WriteString(d.joinArrayElementForSql(value, ","))
		sqlBuilder.WriteString(")")
		if index != lengthExclude {
			sqlBuilder.WriteString(" OR ")
		}
	}
	sqlBuilder.WriteString(") ")
	return sqlBuilder.String(), nil
}

// 处理别名
func (d *DataPermissionSqlBuilder) setAlias(includeValues, excludeValues map[data_factory.DbField][]any, fieldAlias map[data_factory.DbField]string) {
	for field, alias := range fieldAlias {
		if value, ok := includeValues[field]; ok {
			includeValues[data_factory.DbField(alias)] = value
			delete(includeValues, field)
		}
		if value, ok := excludeValues[field]; ok {
			excludeValues[data_factory.DbField(alias)] = value
			delete(excludeValues, field)
		}
	}
}

// 数组元素连接
func (d *DataPermissionSqlBuilder) joinArrayElementForSql(values []any, separator string) string {
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

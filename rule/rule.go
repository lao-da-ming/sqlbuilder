package rule

import (
	"context"
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"log"
	"strings"
)

// 维度
type DimensionType string

const (
	CreatedBy DimensionType = "created_by"
	Position  DimensionType = "position"
)

// 元素
type Element struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// 获取可以访问数据的sql条件，返回空sql表示无任何权限
func GetEmployeePermissionSql(ctx context.Context, loginEmployee int64, fieldAlias map[DimensionType]DimensionType, include, exclude [][]map[DimensionType][]Element) (sql string, err error) {
	//无权限返回的sql
	denySql := " 1==2 "
	includeValues, err := getFieldIds(ctx, loginEmployee, include)
	if err != nil {
		return "", err
	}
	log.Println("指定的:", includeValues)
	//包含的为空则无任何数据权限
	if includeValues == nil {
		return denySql, nil
	}
	//排除部分这层每个元素之间都是or关系
	excludeValues, err := getFieldIds(ctx, loginEmployee, exclude)
	if err != nil {
		return "", err
	}
	log.Println("排除的:", excludeValues)
	//包含跟排除之间对应字段取差集(map引用无需返回值)
	for field, _ := range includeValues {
		//不存在排除，继续下一个
		if _, ok := excludeValues[field]; !ok {
			continue
		}
		//存在则取差集
		includeValues[field] = slice.Difference(includeValues[field], excludeValues[field])
		//and条件如果include方差集为空，则全部不合适
		if len(includeValues[field]) == 0 {
			fmt.Println(fmt.Sprintf("字段:%s取差集后为空集合", field))
			return denySql, nil
		}
		//取了差集，删掉排除对应的字段
		delete(excludeValues, field)
	}
	log.Printf("最终差集为指定:%v  排除:%v \r\n", includeValues, excludeValues)
	//处理别名(map引用无需返回值)
	dealWithAlias(ctx, includeValues, excludeValues, fieldAlias)
	//拼接sql
	return buildSql(ctx, includeValues, excludeValues)
}

// 构建最终的sql
func buildSql(ctx context.Context, includeValues, excludeValues map[DimensionType][]int64) (string, error) {
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
		sqlBuilder.WriteString(slice.Join(value, ","))
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
	//初始化i
	index = 0
	for field, value := range excludeValues {
		index++
		sqlBuilder.WriteString(string(field))
		sqlBuilder.WriteString(" NOT IN (")
		sqlBuilder.WriteString(slice.Join(value, ","))
		sqlBuilder.WriteString(")")
		if index != lengthExclude {
			sqlBuilder.WriteString(" OR ")
		}
	}
	sqlBuilder.WriteString(") ")
	return sqlBuilder.String(), nil
}

// 处理别名
func dealWithAlias(ctx context.Context, includeValues, excludeValues map[DimensionType][]int64, fieldAlias map[DimensionType]DimensionType) {
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

// 获取对应字段的拥有可查看的id值
func getFieldIds(ctx context.Context, loginEmployee int64, exclude [][]map[DimensionType][]Element) (map[DimensionType][]int64, error) {
	if len(exclude) == 0 {
		return nil, nil
	}
	//每个元素之间是or关系，取并集
	mapFieldValues := make(map[DimensionType][]int64, len(exclude))
	for _, first := range exclude {
		secondLevelValues, err := secondLevel(ctx, loginEmployee, first)
		if err != nil {
			return nil, err
		}
		//这层是or，没有就跳过
		if secondLevelValues == nil {
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
func secondLevel(ctx context.Context, loginEmployee int64, first []map[DimensionType][]Element) (map[DimensionType][]int64, error) {
	//这层每个元素之间都是and关系,取交集
	mapFieldValues := make(map[DimensionType][]int64, len(first))
	for _, second := range first {
		thirdLevelValues, err := thirdLevel(ctx, loginEmployee, second)
		if err != nil {
			return nil, err
		}
		//and有不成立直接返回
		if thirdLevelValues == nil {
			return nil, nil
		}
		//取对应字段的交集(这里循环其实只有一个元素)
		for field, thirdLevelValue := range thirdLevelValues {
			//不存在，则直接赋值
			if _, ok := mapFieldValues[field]; !ok {
				mapFieldValues[field] = thirdLevelValue
			} else {
				//取交集
				mapFieldValues[field] = slice.Intersection(mapFieldValues[field], thirdLevelValue)
				//去重
				mapFieldValues[field] = slice.Unique(mapFieldValues[field])
			}
			//and关系判断交集,如果是空的则全部不成立
			if len(mapFieldValues[field]) == 0 {
				return nil, nil
			}
		}
	}
	return mapFieldValues, nil
}

// 第3层
func thirdLevel(ctx context.Context, loginEmployee int64, second map[DimensionType][]Element) (map[DimensionType][]int64, error) {
	//map只有一个元素
	var (
		err            error
		mapFieldValues = make(map[DimensionType][]int64, 1)
	)
	//其实只有一个元素
	for dimension, value := range second {
		var ids []int64
		switch dimension {
		case CreatedBy: //人员维度
			ids, err = createByDimensional(ctx, loginEmployee, value)
		case Position: //岗位维度
			ids, err = positionDimensional(ctx, loginEmployee, value)
		}
		if err != nil {
			return nil, err
		}
		if len(ids) == 0 {
			return nil, nil
		}
		//根据字段存值返回
		mapFieldValues[dimension] = ids
		break
	}
	return mapFieldValues, nil
}

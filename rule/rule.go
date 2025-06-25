package rule

import (
	"context"
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"log"
)

// 维度
type DbField string

const (
	CreatedBy DbField = "created_by"
	Position  DbField = "position"
)

// 元素
type Element struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// 获取可以访问数据的sql条件，返回空sql表示无任何权限
func GetEmployeePermissionSql(ctx context.Context, loginEmployee int64, fieldAlias map[DbField]DbField, include, exclude [][]map[DbField][]Element) (sql string, err error) {
	//无权限返回的sql
	denySql := " 1=2 "
	includeValues, err := getFieldValues(ctx, loginEmployee, include)
	if err != nil {
		return "", err
	}
	log.Println("指定的:", includeValues)
	//包含的为空则无任何数据权限
	if includeValues == nil {
		return denySql, nil
	}
	//排除部分这层每个元素之间都是or关系
	excludeValues, err := getFieldValues(ctx, loginEmployee, exclude)
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
	setAlias(ctx, includeValues, excludeValues, fieldAlias)
	//拼接sql
	return buildSql(ctx, includeValues, excludeValues)
}

// 获取对应字段的拥有可查看的id值
func getFieldValues(ctx context.Context, loginEmployee int64, exclude [][]map[DbField][]Element) (map[DbField][]any, error) {
	if len(exclude) == 0 {
		return nil, nil
	}
	//每个元素之间是or关系，取并集
	mapFieldValues := make(map[DbField][]any, len(exclude))
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
func secondLevel(ctx context.Context, loginEmployee int64, first []map[DbField][]Element) (map[DbField][]any, error) {
	//这层每个元素之间都是and关系,取交集
	mapFieldValues := make(map[DbField][]any, len(first))
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
func thirdLevel(ctx context.Context, loginEmployee int64, second map[DbField][]Element) (map[DbField][]any, error) {
	//map只有一个元素
	var (
		err            error
		mapFieldValues = make(map[DbField][]any, 1)
	)
	//其实只有一个元素
	for field, value := range second {
		var collections []any //集合
		//TODO添加不通的字段在这里维护数据源
		switch field {
		case CreatedBy: //人员id
			collections, err = createByDimensional(ctx, loginEmployee, value)
		case Position: //岗位id
			collections, err = positionDimensional(ctx, loginEmployee, value)
		}
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

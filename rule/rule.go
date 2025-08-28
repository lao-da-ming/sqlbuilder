package rule

import (
	"context"
	"errors"
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"log"
	"sqlbuilder/data_factory"
)

// 维度

// 元素

// 获取可以访问数据的sql条件，返回空sql表示无任何权限
func GetEmployeePermissionSql(ctx context.Context, loginEmployee int64, fieldAlias map[data_factory.DbField]data_factory.DbField, include, exclude [][]map[data_factory.DbField][]data_factory.Element) (sql string, err error) {
	//无权限返回的sql
	denySql := " 1=2 "
	includeValues, err := getFieldValues(ctx, loginEmployee, include)
	if err != nil {
		return "", err
	}
	log.Println("指定的:", includeValues)
	//包含的为空则无任何数据权限
	if len(includeValues) == 0 {
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
	setAlias(includeValues, excludeValues, fieldAlias)
	//拼接sql
	return buildSql(includeValues, excludeValues)
}

// 获取对应字段的拥有可查看的id值
func getFieldValues(ctx context.Context, loginEmployee int64, fieldValues [][]map[data_factory.DbField][]data_factory.Element) (map[data_factory.DbField][]any, error) {
	if len(fieldValues) == 0 {
		return nil, nil
	}
	//每个元素之间是or关系，取并集
	mapFieldValues := make(map[data_factory.DbField][]any, len(fieldValues))
	for _, firstLevelItem := range fieldValues {
		secondLevelValues, err := secondLevel(ctx, loginEmployee, firstLevelItem)
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
func secondLevel(ctx context.Context, loginEmployee int64, firstLevelItem []map[data_factory.DbField][]data_factory.Element) (map[data_factory.DbField][]any, error) {
	//这层每个元素之间都是and关系,取交集
	mapFieldValues := make(map[data_factory.DbField][]any, len(firstLevelItem))
	for _, secondLevelItem := range firstLevelItem {
		thirdLevelValues, err := thirdLevel(ctx, loginEmployee, secondLevelItem)
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
func thirdLevel(ctx context.Context, loginEmployee int64, secondLevelItem map[data_factory.DbField][]data_factory.Element) (map[data_factory.DbField][]any, error) {
	//map只有一个元素
	mapFieldValues := make(map[data_factory.DbField][]any, 1)
	//其实只有一个元素
	for field, value := range secondLevelItem {
		//根据field获取数据
		source := data_factory.NewDataSource(field)
		if source == nil {
			return nil, errors.New("invalid db field")
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

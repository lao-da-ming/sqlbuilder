package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sqlbuilder/rule"
	"time"
)

func main() {
	/*
			self:10
		    organization: all[1,2,3]  directly[1]
			position: true[11, 22, 33,44]

			include:
			exclude:

	*/
	includeStr := `[
			[
				{"created_by":[{"type":"self","value":true},{"type":"organization","value":"all"}]},
				{"created_by":[{"type":"self","value":true},{"type":"organization","value":"all"}]}
			],
			[
				{"position":[{"type":"self","value":false}]}
			]
    ]`
	excludeStr := `[
			[
				{"created_by":[{"type":"self","value":true},{"type":"organization","value":"directly"}]}
			],
			[
				{"created_by":[{"type":"self","value":true},{"type":"organization","value":"directly"}]}
			],
			[
				{"position":[{"type":"self","value":true}]}
			]
	]`
	//指定部分
	var include [][]map[rule.DimensionType][]rule.Element
	if err := json.Unmarshal([]byte(includeStr), &include); err != nil {
		panic(err)
	}
	//排除部分
	var exclude [][]map[rule.DimensionType][]rule.Element
	if err := json.Unmarshal([]byte(excludeStr), &exclude); err != nil {
		panic(err)
	}
	//字段别名
	fieldAlias := map[rule.DimensionType]rule.DimensionType{
		"created_by": "o.created_by",
	}
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()
	sql, err := rule.GetEmployeePermissionSql(ctx, 10, fieldAlias, include, exclude)
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)
}

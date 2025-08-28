package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sqlbuilder/data_permission"
	"sqlbuilder/data_permission/data_factory"
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
	var include [][]map[data_factory.TableField][]data_factory.Element
	if err := json.Unmarshal([]byte(includeStr), &include); err != nil {
		panic(err)
	}
	//排除部分
	var exclude [][]map[data_factory.TableField][]data_factory.Element
	if err := json.Unmarshal([]byte(excludeStr), &exclude); err != nil {
		panic(err)
	}
	//字段别名
	fieldAlias := map[data_factory.TableField]string{
		data_factory.CreatedBy: "o.created_by",
		data_factory.Position:  "a.position",
	}
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()
	builder := data_permission.NewDataPermissionSqlBuilder()
	sql, err := builder.GetDataPermissionSql(ctx, 10, fieldAlias, include, exclude)
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)
}

package snowflake

import (
	"fmt"
	"time"

	sf "github.com/bwmarrin/snowflake"
)

// node 是用于生成ID的Snowflake节点
var node *sf.Node

// Init 初始化Snowflake节点
// 参数:
//   - startTime: 起始时间，格式为"2006-01-02"
//   - machineID: 机器ID
//
// 返回:
//   - error: 可能发生的错误
func Init(startTime string, machineID int64) (err error) {
	var st time.Time
	// 解析起始时间
	st, err = time.Parse("2006-01-02", startTime)
	if err != nil {
		return
	}
	// 设置Snowflake的纪元（起始时间）
	sf.Epoch = st.UnixNano() / 1000000
	// 创建一个新的Snowflake节点
	node, err = sf.NewNode(machineID)
	fmt.Println("Snowflake node initialized success")
	return
}

// GetID 生成一个新的唯一ID
// 返回:
//   - int64: 生成的唯一ID
func GetID() uint64 {
	return uint64(node.Generate().Int64())
}

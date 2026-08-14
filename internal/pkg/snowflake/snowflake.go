package snowflake

import (
	"time"

	sf "github.com/bwmarrin/snowflake"
)

// 全局的雪花节点对象
var node *sf.Node

// Init 初始化雪花算法节点
// startTime: 项目启动基准时间（格式如："2026-01-01"）
// machineID: 当前机器/节点编号（0 ~ 1023 之间，支持 1024 台微服务机器并发）
func Init(startTime string, machineID int64) (err error) {
	var st time.Time
	// 1. 解析基准时间
	st, err = time.Parse("2006-01-02", startTime)
	if err != nil {
		return err
	}
	// 2. 设置雪花算法的时间纪元（Epoch）
	sf.Epoch = st.UnixNano() / 1000000

	// 3. 创建当前机器节点
	node, err = sf.NewNode(machineID)
	return err
}

// GenID 生成全局唯一的 64 位整数 ID
func GenID() int64 {
	return node.Generate().Int64()
}

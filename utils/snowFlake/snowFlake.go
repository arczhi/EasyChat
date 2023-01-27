package snowFlake

import (
	"log"
	"time"

	"github.com/bwmarrin/snowflake"
)

const (
	TIMEFORMAT = "2006-01-02"
)

var node *snowflake.Node

func init() {
	Initialization(time.Now().Format(TIMEFORMAT), 1)
}

func Initialization(startTime string, machineID int64) (err error) {
	var st time.Time
	// 格式化 1月2号下午3时4分5秒  2006年
	st, err = time.Parse(TIMEFORMAT, startTime)
	if err != nil {
		log.Println(err)
		return
	}

	snowflake.Epoch = st.UnixNano() / 1e6
	node, err = snowflake.NewNode(machineID)
	if err != nil {
		log.Println(err)
		return
	}

	return
}

// 生成 64 位的 雪花 ID
func GenID() int64 {
	return node.Generate().Int64()
}

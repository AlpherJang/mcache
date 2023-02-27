package oper

type OpType int32

const (
	OpType_AddTable OpType = 0
	OpType_DelTable OpType = 1
	OpType_Insert   OpType = 2
	OpType_Update   OpType = 3
	OpType_Delete   OpType = 4
)

type Operate struct {
	// Uid 操作id
	Uid string
	// Op 操作类型
	Op OpType
	// TableName 数据库的name
	TableName string
	// Key 操作数据的key
	Key string
	// Value 操作的数据结果
	Value interface{}
	// OpTime unix格式的时间
	OpTime int64
	// BeforeOp 前一个操作的id
	BeforeOp string
}

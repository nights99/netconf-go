package types

type RequestType int

const (
	Validate RequestType = iota
	Commit               = 1
	GetConf              = 2
	GetOper              = 3
	RpcOp                = 5
	EditConf             = 6
)

const (
	NewTokens   = 0
	ReplaceLast = 1
)

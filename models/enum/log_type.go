// 日志类型枚举

package enum

type LogType int8

const (
	LoginLogType   LogType = 1 // 登录日志
	ActionLogType  LogType = 2 // 操作日志
	RuntimeLogType LogType = 3 // 运行时日志
)

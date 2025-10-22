// core/init_logrus.go
package core

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"time"

	"myblogx/global"

	"github.com/sirupsen/logrus"
)

// 颜色映射
const (
	red    = 31
	yellow = 33
	blue   = 36
	gray   = 37
)

type LogFormatter struct{}

// 实现 Format 方法来实现 Formatter 接口
func (t *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 设置不同 level 的输出颜色
	var levelColor int
	switch entry.Level {
	case logrus.DebugLevel, logrus.TraceLevel:
		levelColor = gray
	case logrus.WarnLevel:
		levelColor = yellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = red
	default:
		levelColor = blue
	}

	// 使用 Buffer 初始化检测
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	// 自定义日期格式
	timestamp := entry.Time.Format("2006-01-02 15:04:05")

	if entry.HasCaller() {
		// 打印调用者路径
		funcVal := entry.Caller.Function
		fileVal := fmt.Sprintf("%s:%d", path.Base(entry.Caller.File), entry.Caller.Line)

		// 自定义输出格式
		fmt.Fprintf(b, "[%s] \x1b[%dm[%s]\x1b[0m %s %s %s\n", timestamp, levelColor, entry.Level, fileVal, funcVal, entry.Message)
	} else {
		fmt.Fprintf(b, "[%s] \x1b[%dm[%s]\x1b[0m %s\n", timestamp, levelColor, entry.Level, entry.Message)
	}

	return b.Bytes(), nil
}

// 使用 Hook 实现按日期切换日志文件 (实现Fire和Levels方法来实现Hook接口)
type FileDateHook struct {
	file     *os.File
	logPath  string
	fileDate string //判断日期切换目录
	appName  string
}

func createLogFile(logPath, timer, appName string) (*os.File, error) {
	dir := fmt.Sprintf("%s/%s", logPath, timer)

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}

	filename := fmt.Sprintf("%s/%s.log", dir, appName)

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败: %w", err)
	}

	return file, nil
}

func (hook FileDateHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *FileDateHook) Fire(entry *logrus.Entry) error {
	timer := entry.Time.Format("2006-01-02")
	line, _ := entry.String() // 将当前日志内容转换为字符串形式

	// 如果日期相等，直接写入文件
	if hook.fileDate == timer {
		if _, err := hook.file.Write([]byte(line)); err != nil {
			return fmt.Errorf("写入日志文件失败: %w", err)
		}
		return nil
	}

	// 时间不相等，关闭当前文件，创建新文件夹和新日志文件
	hook.file.Close()
	newFile, err := createLogFile(hook.logPath, timer, hook.appName)
	if err != nil {
		return fmt.Errorf("切换日志文件失败: %w", err)
	} else {
		hook.file = newFile
	}

	hook.fileDate = timer

	if _, err := hook.file.Write([]byte(line)); err != nil {
		hook.file.Close() // 写入失败时，关闭刚创建的文件
		return fmt.Errorf("写入新日志文件失败: %w", err)
	}

	return nil
}

func InitFile(logPath, appName string) {
	fileDate := time.Now().Format("2006-01-02")

	//创建目录和文件
	file, err := createLogFile(logPath, fileDate, appName)
	if err != nil {
		logrus.Error(err)
		return
	}

	fileHook := FileDateHook{file, logPath, fileDate, appName}
	logrus.AddHook(&fileHook)
}

func InitLogrus() {
	logrus.SetOutput(os.Stdout)          // 设置输出位置
	logrus.SetReportCaller(true)         // 日志显示函数名和行号
	logrus.SetFormatter(&LogFormatter{}) // 设置自己定义的Formatter
	logrus.SetLevel(logrus.DebugLevel)   // 设置最低的Level

	l := global.Config.Log
	InitFile(l.Dir, l.App)
}

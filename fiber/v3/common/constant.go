package common

// 项目
const (
	COMMON_PROJECT_NAME = "awesome-fiber-template" // 项目名
	COMMON_PROJECT_HELP = `
Awesome-Fiber-Template is a template for MVC backend service.
Usage:
  ./awesome-fiber-template [params]
    - install: install this backend to systemd.
    - uninstall: uninstall this backend from systemd.
    - version: show this backend version.
    - help: show this help message.
`
)

// 时间
const (
	TIME_FORMAT_DIGIT_DAY = "20060102"
	TIME_FORMAT_DIGIT     = "20060102150405"
	TIME_FORMAT_DATE      = "2006-01-02 15:04:05"
	TIME_FORMAT_DAY       = "2006-01-02"
	TIME_FORMAT_LOG       = "2006-01-02 15:04:05.000"
)

// 状态标识
const (
	RETURN_FAILED       = 0 //失败
	RETURN_SUCCESS      = 1 //成功
	RETURN_SUCCESS_CODE = 200
	RETURN_FAILED_CODE  = 500
)

// 常量
const (
	JWT_RELET_NUM     = 2 // JWT续租时间(小时)
	ONLINE_RELET_NUM  = 5 // 用户在线判定时间(分钟)
	EMAIL_CODE_LENGTH = 6 // 邮箱验证码长度
)

// 请求头
const (
	USER_AGENT      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	ACCEPT_LANGUAGE = "zh-CN,zh;q=0.9,en;q=0.8"
	APPLICATION     = "application/json"
)

// 事件
const (
	GLOBAL_MSG = "GLOBAL_MSG" // 全局事件
	COMMON_MSG = "COMMON_MSG" // 通用事件
)

package ecode

var (
	OkCode            int32 = 0
	ErrServerCode     int32 = -50000
	ErrSystemBusyCode int32 = -50001
	ErrDeviceTypeCode int32 = -50002
	ErrNoUserIDCode   int32 = -50003
	ErrParamCode      int32 = -50004
)

// common ecode
var (
	OK = add(OkCode) // 成功

	ErrServer     = add(ErrServerCode)
	ErrSystemBusy = add(ErrSystemBusyCode)
	ErrDeviceType = add(ErrDeviceTypeCode)
	ErrNoUserID   = add(ErrNoUserIDCode)
	ErrParam      = add(ErrParamCode)
)

var defaultCodeMap = map[int32]string{
	OkCode:            "成功",
	ErrServerCode:     "服务异常",
	ErrSystemBusyCode: "服务繁忙，请稍后重试",
	ErrDeviceTypeCode: "设备类型错误",
	ErrNoUserIDCode:   "用户ID缺失",
	ErrParamCode:      "输入参数错误",
}

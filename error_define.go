// 错误定义

package apitoken

import (
	"github.com/go-apibox/api"
)

// error type
const (
	errorMissingToken = iota
	errorInvalidToken
	errorParamMismatch
)

var ErrorDefines = map[api.ErrorType]*api.ErrorDefine{
	errorMissingToken: api.NewErrorDefine(
		"MissingToken",
		[]int{0},
		map[string]map[int]string{
			"en_us": {
				0: "Missing API token!",
			},
			"zh_cn": {
				0: "缺少 API Token！",
			},
		},
	),
	errorInvalidToken: api.NewErrorDefine(
		"InvalidToken",
		[]int{0},
		map[string]map[int]string{
			"en_us": {
				0: "API token is invalid!",
			},
			"zh_cn": {
				0: "API token不合法！",
			},
		},
	),
	errorParamMismatch: api.NewErrorDefine(
		"ParamMismatch",
		[]int{0},
		map[string]map[int]string{
			"en_us": {
				0: "Param mismatch with API token!",
			},
			"zh_cn": {
				0: "参数与 API Token 不匹配！",
			},
		},
	),
}

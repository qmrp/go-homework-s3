package request

import (
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type CreateTopicRequest struct {
	Topic string `json:"topic" binding:"required,name"`
}

// ^[a-zA-Z0-9_-]{4,30}$ 话题名称只能包含字母、数字、下划线和短横线，长度在4到30之间
// 注册自定义验证器
func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 注册密码验证器，只允许字母、数字、下划线和连字符
		_ = v.RegisterValidation("name", func(fl validator.FieldLevel) bool {
			topicName := fl.Field().String()
			// 使用正则表达式验证密码格式
			matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]{4,30}$`, topicName)
			return matched
		})
	}
}

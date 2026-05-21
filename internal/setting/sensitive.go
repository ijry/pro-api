package setting

import (
	"regexp"
	"strings"
)

// SensitiveKeys 是显式列出的敏感配置项白名单。
//
// 列出那些不被 sensitiveSuffixRE 兜底匹配到的"业务命名"项,
// 或语义上敏感但命名不带 _secret/_password/_key 后缀的 key。
var SensitiveKeys = map[string]struct{}{
	"auth.github_oauth.client_secret": {},
	"auth.smtp.password":              {},
	"payment.stripe.secret_key":       {},
	"payment.stripe.webhook_secret":   {},
}

// sensitiveSuffixRE 用正则兜底所有以 _secret / _password / _key (或裸 password/secret/key) 结尾的 key。
//
// 仅匹配 "." 之后的最后一段是 secret/password/key 或以 _secret/_password/_key 结尾的形态;
// 例如 "secret_path" 不算敏感,但 "foo.api_secret" 与 "smtp.password" 都算敏感。
var sensitiveSuffixRE = regexp.MustCompile(`\.([a-z0-9]+_)?(secret|password|key)$`)

// IsSensitive 判断一个 key 是否敏感。
//
// 规则:
//
//  1. 出现在显式 SensitiveKeys 白名单中
//  2. 形如 "<prefix>.<...>_secret" / "<prefix>.<...>_password" / "<prefix>.<...>_key" 后缀
func IsSensitive(key string) bool {
	if _, ok := SensitiveKeys[key]; ok {
		return true
	}
	return sensitiveSuffixRE.MatchString(key)
}

// EncryptedPlaceholder 是 GET 返回敏感字段时替换的占位字符串。
const EncryptedPlaceholder = "<encrypted>"

// IsEncryptedValue 判断 raw JSON value 是否为 ENC(v1,...) 形态的字符串。
//
// 仅识别 JSON 字符串形式;非字符串 value(数字 / bool / null / object) 不当作 ENC。
func IsEncryptedValue(raw []byte) bool {
	s := strings.TrimSpace(string(raw))
	return strings.HasPrefix(s, `"ENC(`) && strings.HasSuffix(s, `)"`)
}

// IsPlaceholderValue 判断 PATCH body 中是否为占位值 "<encrypted>"。
//
// 占位值表示用户在前端没有改动密文字段;handler 应原样跳过该 key,不更新。
func IsPlaceholderValue(raw []byte) bool {
	s := strings.TrimSpace(string(raw))
	return s == `"`+EncryptedPlaceholder+`"`
}

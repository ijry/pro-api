package setting

import "strings"

// Group 是前端渲染 tab 用的分组名。后端按 key prefix 分组。
type Group string

const (
	GroupAuth           Group = "auth"
	GroupSession        Group = "session"
	GroupToken          Group = "token"
	GroupPricing        Group = "pricing"
	GroupBilling        Group = "billing"
	GroupChannel        Group = "channel"
	GroupRatelimit      Group = "ratelimit"
	GroupLog            Group = "log"
	GroupNotice         Group = "notice"
	GroupManualRecharge Group = "manual_recharge"
	GroupRedeem         Group = "redeem"
	GroupOther          Group = "other"
)

// keyHead 返回 key 第一段(以 "." 分割)。
func keyHead(key string) string {
	if i := strings.IndexByte(key, '.'); i >= 0 {
		return key[:i]
	}
	return key
}

// GroupOf 根据 key 第一段返回分组。
//
//	"auth.allow_register"               -> GroupAuth
//	"billing.reserve_ttl_seconds"       -> GroupBilling
//	"auth.github_oauth.client_secret"   -> GroupAuth
func GroupOf(key string) Group {
	switch keyHead(key) {
	case "auth":
		return GroupAuth
	case "session":
		return GroupSession
	case "token":
		return GroupToken
	case "pricing":
		return GroupPricing
	case "billing":
		return GroupBilling
	case "channel":
		return GroupChannel
	case "ratelimit":
		return GroupRatelimit
	case "log":
		return GroupLog
	case "notice":
		return GroupNotice
	case "manual_recharge":
		return GroupManualRecharge
	case "redeem":
		return GroupRedeem
	}
	return GroupOther
}

// AllGroups 是稳定的分组渲染顺序(管理后台 tab 顺序)。
func AllGroups() []Group {
	return []Group{
		GroupAuth, GroupSession, GroupToken,
		GroupPricing, GroupBilling, GroupChannel,
		GroupRatelimit, GroupLog,
		GroupNotice, GroupManualRecharge, GroupRedeem,
		GroupOther,
	}
}

// groupLabels 是分组的中文显示名。
var groupLabels = map[Group]string{
	GroupAuth:           "认证",
	GroupSession:        "会话",
	GroupToken:          "令牌",
	GroupPricing:        "计费",
	GroupBilling:        "结算",
	GroupChannel:        "渠道",
	GroupRatelimit:      "限流",
	GroupLog:            "日志",
	GroupNotice:         "公告",
	GroupManualRecharge: "手动充值",
	GroupRedeem:         "兑换码",
	GroupOther:          "其他",
}

// GroupLabel 返回某分组的展示名。未知分组返回 "其他"。
func GroupLabel(g Group) string {
	if s, ok := groupLabels[g]; ok {
		return s
	}
	return groupLabels[GroupOther]
}

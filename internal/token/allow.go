package token

import (
	"net/netip"
	"strings"

	"github.com/ijry/pro-api/pkg/apierr"
)

// AssertIPAllowed 校验 ip 是否在 view.AllowedIPs 列表中。
//
//   - 空数组 = 不限,直接通过
//   - ip 形式接受 "1.2.3.4" / "1.2.3.4:1234" / "[::1]:8080" / "::1"
//   - 列表元素支持 CIDR("10.0.0.0/8")或单 IP("1.2.3.4");
//     单元素配置错误(无法解析)会跳过,不影响其他条目
//   - 全部不匹配返回 apierr.CodeIPNotAllowed
func AssertIPAllowed(v *View, ip string) error {
	if v == nil || len(v.AllowedIPs) == 0 {
		return nil
	}
	addr, ok := parseAddr(ip)
	if !ok {
		return apierr.New(apierr.CodeIPNotAllowed, "ip address invalid: "+ip)
	}
	for _, entry := range v.AllowedIPs {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if p, err := netip.ParsePrefix(entry); err == nil {
			if p.Contains(addr) {
				return nil
			}
			continue
		}
		// 单 IP 兼容
		if a, err := netip.ParseAddr(entry); err == nil {
			if a.Compare(addr) == 0 {
				return nil
			}
			continue
		}
		// 解析失败的条目跳过(配置错误不应整体放行)
	}
	return apierr.New(apierr.CodeIPNotAllowed, "ip not in allowlist")
}

// parseAddr 把客户端 IP 字符串(可能带端口)解析为 netip.Addr。
func parseAddr(s string) (netip.Addr, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return netip.Addr{}, false
	}
	if a, err := netip.ParseAddr(s); err == nil {
		return a, true
	}
	if ap, err := netip.ParseAddrPort(s); err == nil {
		return ap.Addr(), true
	}
	return netip.Addr{}, false
}

// AssertModelAllowed 校验 model 是否在 view.AllowedModels 中。
//
//   - 空数组 = 不限,直接通过
//   - 精确匹配优先于通配
//   - 通配仅支持末尾 "*"(例 "gpt-4*" 匹配 "gpt-4o" / "gpt-4-turbo");"*-turbo" 不被视为通配
//   - 大小写敏感
func AssertModelAllowed(v *View, model string) error {
	if v == nil || len(v.AllowedModels) == 0 {
		return nil
	}
	// 精确匹配先扫一遍
	for _, p := range v.AllowedModels {
		if !strings.HasSuffix(p, "*") && p == model {
			return nil
		}
	}
	// 末尾 * 通配
	for _, p := range v.AllowedModels {
		if !strings.HasSuffix(p, "*") {
			continue
		}
		prefix := strings.TrimSuffix(p, "*")
		if strings.HasPrefix(model, prefix) {
			return nil
		}
	}
	return apierr.New(apierr.CodeModelNotAllowed, "model "+model+" not allowed for this token")
}

// ModelInAllowList 是 AssertModelAllowed 的非 error 版本,供 /v1/models 这类需要批量过滤的路径用。
func ModelInAllowList(v *View, model string) bool {
	return AssertModelAllowed(v, model) == nil
}

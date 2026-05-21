// Package version 保存编译期注入的版本信息。
package version

var (
	// Version 由 ldflags 注入,默认 dev。
	Version = "dev"
	// Commit 是 git commit short hash,默认 unknown。
	Commit = "unknown"
	// BuildTime 是构建时间(RFC3339),默认 unknown。
	BuildTime = "unknown"
)

// String 返回完整版本字符串。
func String() string {
	return Version + " (" + Commit + " @ " + BuildTime + ")"
}

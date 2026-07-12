// Package config 加载并校验启动配置。
//
// 优先级:环境变量 > 配置文件 > 内置默认值。
// 启动后配置为只读;运行时可改的配置走 system_settings 表。
package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 是顶层启动配置。
type Config struct {
	NodeID    int            `mapstructure:"node_id"`
	MasterKey string         `mapstructure:"master_key"`
	Server    ServerConfig   `mapstructure:"server"`
	Log       LogConfig      `mapstructure:"log"`
	Database  DatabaseConfig `mapstructure:"database"`
	Redis     RedisConfig    `mapstructure:"redis"`
	SMTP      SMTPConfig     `mapstructure:"smtp"`
	Account   AccountConfig  `mapstructure:"account"`
}

// AccountConfig 是号池/OAuth/Probe 相关配置(M2b)。
// 默认值见 system_settings seed (000029_seed_account_settings)。
type AccountConfig struct {
	OAuthAnthropicTokenURL    string   `mapstructure:"oauth_anthropic_token_url"`
	OAuthAnthropicClientID    string   `mapstructure:"oauth_anthropic_client_id"`
	OAuthAnthropicAuthURL     string   `mapstructure:"oauth_anthropic_auth_url"`
	OAuthAnthropicRedirectURI string   `mapstructure:"oauth_anthropic_redirect_uri"`
	OAuthAnthropicScopes      []string `mapstructure:"oauth_anthropic_scopes"`
	OAuthOpenAITokenURL       string   `mapstructure:"oauth_openai_token_url"`
	OAuthOpenAIClientID       string   `mapstructure:"oauth_openai_client_id"`
	OAuthOpenAIAuthURL        string   `mapstructure:"oauth_openai_auth_url"`
	OAuthOpenAIRedirectURI    string   `mapstructure:"oauth_openai_redirect_uri"`
	OAuthOpenAIScopes         []string `mapstructure:"oauth_openai_scopes"`
	AnthropicProbeBase        string   `mapstructure:"anthropic_probe_base"`
	OpenAIProbeBase           string   `mapstructure:"openai_probe_base"`
	// ProbeLoop* 控制后台定时探测循环(周期扫描额度陈旧的 active 账号并回填额度、
	// 按失败类型标记)。0 值由 accountwire 兜底为内置默认。
	ProbeLoopTickSeconds  int `mapstructure:"probe_loop_tick_seconds"`  // 扫描间隔,默认 300
	ProbeLoopStaleSeconds int `mapstructure:"probe_loop_stale_seconds"` // 额度陈旧阈值,默认 600
	ProbeLoopBatchLimit   int `mapstructure:"probe_loop_batch_limit"`   // 单轮最多探测数,默认 100
	ProbeLoopConcurrency  int `mapstructure:"probe_loop_concurrency"`   // 单轮最大并发,默认 8
}

// ServerConfig 描述 HTTP server 启动参数。
type ServerConfig struct {
	Addr           string `mapstructure:"addr"`
	ReadTimeoutMS  int    `mapstructure:"read_timeout_ms"`
	WriteTimeoutMS int    `mapstructure:"write_timeout_ms"`
}

// LogConfig 描述日志输出形式。
type LogConfig struct {
	Level  string `mapstructure:"level"`  // debug | info | warn | error
	Format string `mapstructure:"format"` // json | console
}

// DatabaseConfig 描述数据库连接。
type DatabaseConfig struct {
	Driver         string `mapstructure:"driver"` // mysql | postgres
	DSN            string `mapstructure:"dsn"`
	MaxOpenConns   int    `mapstructure:"max_open_conns"`
	MaxIdleConns   int    `mapstructure:"max_idle_conns"`
	ConnMaxLifeSec int    `mapstructure:"conn_max_life_sec"`
}

// RedisConfig 描述 Redis 连接。
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// SMTPConfig 描述 SMTP 发件服务器。Host 为空时禁用 SMTP，回落到 stub。
type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"` // 25 / 465 / 587
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`     // "name <addr>" 或 "addr"
	TLS      bool   `mapstructure:"tls"`      // true: implicit TLS (port 465)
	Insecure bool   `mapstructure:"insecure"` // 跳过证书验证(开发环境)
}

// Load 加载配置。path 为空时仅使用 ENV 与默认值。
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("PROAPI")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 默认值
	v.SetDefault("node_id", 0)
	v.SetDefault("server.addr", ":8080")
	v.SetDefault("server.read_timeout_ms", 30000)
	v.SetDefault("server.write_timeout_ms", 60000)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("database.driver", "mysql")
	v.SetDefault("database.dsn", "root:proapi@tcp(127.0.0.1:3306)/proapi?charset=utf8mb4&parseTime=True&loc=UTC")
	v.SetDefault("database.max_open_conns", 50)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_life_sec", 1800)
	v.SetDefault("redis.addr", "127.0.0.1:6379")
	v.SetDefault("redis.db", 0)

	if path != "" {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("config: read file %s: %w", path, err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config: unmarshal: %w", err)
	}
	cfg.MasterKey = v.GetString("master_key")
	if cfg.MasterKey == "" {
		return nil, errors.New("config: PROAPI_MASTER_KEY (or master_key) is required")
	}
	if cfg.Database.Driver != "mysql" && cfg.Database.Driver != "postgres" {
		return nil, fmt.Errorf("config: unsupported database.driver %q (use mysql or postgres)", cfg.Database.Driver)
	}
	return &cfg, nil
}

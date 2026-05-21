package setting

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// Decryptor 是 store 解密的最小依赖(对应 crypto.AESGCM.Decrypt)。
//
// 抽出 interface 以便测试 GetSecret 无需启动真实 AES 链路。
type Decryptor interface {
	Decrypt(string) (string, error)
}

// ErrNotEncrypted 表示 GetSecret 时 value 不是 ENC(...) 形态(可能是明文 / 数字 / null)。
var ErrNotEncrypted = errors.New("setting: value is not encrypted")

// GetSecret 读取 key,若 value 是 JSON string + ENC(v1,...) 形态,
// 自动用 crypto 解密为明文。
//
// 用例:M1-02 LoginHandler 读 auth.smtp.password / auth.github_oauth.client_secret。
//
// 错误:
//   - key 不存在    -> ("", ErrNotFound)
//   - value 非字符串 -> ("", ErrNotEncrypted)
//   - 字符串非 ENC   -> ("", ErrNotEncrypted)
//   - 解密失败       -> ("", 解密错误)
func (s *store) GetSecret(ctx context.Context, key string, dec Decryptor) (string, error) {
	v, ok := s.Get(ctx, key)
	if !ok {
		return "", ErrNotFound
	}
	if !IsEncryptedValue(v) {
		return "", ErrNotEncrypted
	}
	var encoded string
	if err := json.Unmarshal(v, &encoded); err != nil {
		return "", fmt.Errorf("setting: secret unmarshal: %w", err)
	}
	plain, err := dec.Decrypt(encoded)
	if err != nil {
		return "", fmt.Errorf("setting: secret decrypt: %w", err)
	}
	return plain, nil
}

// ListAll 返回所有 setting 行(按 key 升序)。
//
// 主要用于管理后台 /api/admin/settings 列表;数量在 M1 范围内预计 < 200,可直接全表 SELECT。
// 不走缓存(管理后台调用频次低,且写后需立即看到最新)。
func (s *store) ListAll(ctx context.Context) ([]Setting, error) {
	if s.db == nil {
		return nil, fmt.Errorf("setting: DB not configured")
	}
	var rows []Setting
	if err := s.db.WithContext(ctx).Order("`key` ASC").Find(&rows).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("setting: list all: %w", err)
	}
	return rows, nil
}

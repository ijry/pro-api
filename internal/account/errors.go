package account

import "github.com/ijry/pro-api/pkg/apierr"

// ErrNotFound 由 Repo.Get 在记录缺失时返回。
var ErrNotFound = apierr.New(apierr.CodeAccountNotFound, "account not found")

// ErrPoolEmpty 由 Selector 在候选集为空时返回。P5 不直接用,但 P6 selector 与 fakeRepo 共享。
var ErrPoolEmpty = apierr.New(apierr.CodeAccountPoolEmpty, "account pool empty")

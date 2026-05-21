package token

import (
	"errors"

	"github.com/ijry/pro-api/pkg/apierr"
	"gorm.io/gorm"
)

// wrapDBErr 把 GORM 错误规整到 apierr。
//
//	gorm.ErrRecordNotFound → CodeNotFound("token not found")
//	其他                   → CodeDatabase(原因附 details)
func wrapDBErr(err error) error {
	if err == nil {
		return nil
	}
	var apiErr *apierr.Error
	if errors.As(err, &apiErr) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apierr.New(apierr.CodeNotFound, "token not found")
	}
	return apierr.Wrap(apierr.CodeDatabase, err.Error(), err)
}

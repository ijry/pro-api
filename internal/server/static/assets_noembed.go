//go:build !embed

package static

import "github.com/gin-gonic/gin"

// RegisterEmbedded is a no-op in local backend builds so developers can run the
// API server without first producing frontend dist directories.
func RegisterEmbedded(_ *gin.Engine) error {
	return nil
}

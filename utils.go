package cache

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"path/filepath"
)

func generateFileName(key string) string {
	name := hex.EncodeToString([]byte(fmt.Sprintf("%s", md5.Sum([]byte(key)))))

	level1 := name[31:32]
	level2 := name[29:31]

	return filepath.Join(level1, level2, name)
}

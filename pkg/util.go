package pkg

import (
	"fmt"
)

func genPath(name, version string) string {
	return fmt.Sprintf("%s-%s.tgz", name, version)
}

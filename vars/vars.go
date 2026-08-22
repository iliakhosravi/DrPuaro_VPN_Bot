package vars

import (
	"os"
)

func Get(name string) string {
	return os.Getenv(name)
}

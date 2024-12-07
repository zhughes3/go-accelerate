package url

import "strings"

func CreateFullPath(pathPrefix, path string) string {
	return strings.TrimRight(pathPrefix, "/") + "/" + strings.TrimLeft(path, "/")
}

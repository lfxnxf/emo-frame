package stringx

import (
	"fmt"
	"strings"
)

func RegistryKVPath(name, path string) (string, error) {
	namespace := strings.Split(name, ".")[0]
	if len(namespace) == 0 {
		return "", fmt.Errorf("wrong sdname %s", name)
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	path = fmt.Sprintf("%s%s/%s%s", "/service_config/", namespace, name, path)
	return path, nil
}

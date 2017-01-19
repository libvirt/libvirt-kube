package resource

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func GetResourcePartition(pid int) (string, error) {
	cgfile := fmt.Sprintf("/proc/%d/cgroup", pid)

	cgfh, err := os.Open(cgfile)
	if err != nil {
		return "", err
	}

	scan := bufio.NewScanner(cgfh)

	locations := make(map[string]string)

	for scan.Scan() {
		line := scan.Text()

		bits := strings.Split(line, ":")

		controllers := strings.Split(bits[1], ",")
		for _, controller := range controllers {
			locations[controller] = bits[2]
		}
	}

	val, ok := locations["memory"]
	if !ok {
		return "", fmt.Errorf("Memory controller not present for %s",
			cgfile)
	}

	return val, nil
}

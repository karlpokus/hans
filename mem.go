package hans

import (
	"fmt"
	"os/exec"
	"strconv"
  "strings"
)

// rss returns rss (in MB) by pid
func rss(pid int) (int, error) {
	Cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-O", "rss")
	res, err := Cmd.CombinedOutput()
	if err != nil {
		return 0, err
	}
	lines := strings.Split(string(res), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("bad ps output: %v", lines)
	}
	rssStr := strings.Fields(lines[1])[1]
	if rssStr == "" {
		return 0, fmt.Errorf("bad ps output: %s", rssStr)
	}
	rssInt, err := strconv.Atoi(rssStr)
	if err != nil {
		return 0, err
	}
	return rssInt / 1024, nil // rss is in 1024 byte units
}

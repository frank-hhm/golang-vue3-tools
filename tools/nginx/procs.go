package nginx

import (
	"strings"

	"github.com/shirou/gopsutil/v3/process"
)

// nginxPids 返回当前所有 nginx 进程 PID（master + worker，按镜像名匹配）
func nginxPids() []int {
	procs, err := process.Processes()
	if err != nil {
		return nil
	}
	var out []int
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}
		if strings.EqualFold(name, "nginx.exe") || strings.EqualFold(name, "nginx") {
			out = append(out, int(p.Pid))
		}
	}
	return out
}

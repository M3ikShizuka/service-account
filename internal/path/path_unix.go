//go:build linux

package path

var (
	logsDir string = "/var/log/service-account/logs"
)

func GetLogsDir() string {
	return logsDir
}

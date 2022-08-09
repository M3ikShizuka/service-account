//go:build unix
// +build unix

package path

var (
	logsDir string = "/var/log/service-account/logs"
)

func GetLogsDir() string {
	return logsDir
}

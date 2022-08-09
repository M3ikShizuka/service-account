//go:build windows
// +build windows

package path

var (
	logsDir string = "tmp\\logs"
)

func GetLogsDir() string {
	//if logsDir == "" {
	//	userCacheDir, _ := os.UserCacheDir()
	//	logsDir = userCacheDir + "\\service-account\\logs"
	//}

	return logsDir
}

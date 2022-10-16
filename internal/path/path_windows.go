//go:build windows

package path

var (
	logsDir = "tmp\\logs"
)

func GetLogsDir() string {
	//if logsDir == "" {
	//	userCacheDir, _ := os.UserCacheDir()
	//	logsDir = userCacheDir + "\\service-account\\logs"
	//}

	return logsDir
}

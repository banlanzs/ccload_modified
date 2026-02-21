// Package version 提供版本相关能力。
package version

// StartChecker 保留为兼容占位实现。
// 项目策略：禁用联网版本更新检测。
func StartChecker() {}

// GetUpdateInfo 返回空更新信息，供旧调用方兼容。
func GetUpdateInfo() (hasUpdate bool, latestVersion, releaseURL string) {
	return false, "", ""
}

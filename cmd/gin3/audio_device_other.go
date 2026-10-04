//go:build !linux

package main

// 非 Linux 維持原本行為，由平台後端判定裝置及回報初始化錯誤。
func hasAudioDevice() bool { return true }

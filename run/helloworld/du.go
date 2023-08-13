package main

import "syscall"

type diskUsage struct {
	// see: https://linuxjm.osdn.jp/html/LDP_man-pages/man2/statfs.2.html
	fs syscall.Statfs_t
}

func newDiskUsage(path string) (*diskUsage, error) {
	usage := &diskUsage{}
	err := syscall.Statfs(path, &usage.fs)
	if err != nil {
		return nil, err
	}
	return usage, nil
}

// ファイルシステムの総容量を返す
func (du *diskUsage) size() uint64 {
	return du.fs.Blocks * uint64(du.fs.Bsize)
}

// ファイルシステムの空き容量を返す
func (du *diskUsage) free() uint64 {
	return du.fs.Bfree * uint64(du.fs.Bsize)
}

// 非特権ユーザーが利用可能な空き容量を返す
func (du *diskUsage) avail() uint64 {
	return du.fs.Bavail * uint64(du.fs.Bsize)
}

// ファイルシステムの使用量を返す
func (du *diskUsage) used() uint64 {
	return du.size() - du.free()
}

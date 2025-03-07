// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package validator

import (
	"errors"
	"io/fs"
	"os"
)

// FileExists path 是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !errors.Is(err, fs.ErrNotExist)
}

// IsDir path 是否为一个目录
func IsDir(path string) bool {
	s, err := os.Stat(path)
	return err == nil && s.IsDir()
}

// FileExistsFS 判断文件是否存在于 fsys
func FileExistsFS(fsys fs.FS) func(string) bool {
	return func(path string) bool {
		_, err := fs.Stat(fsys, path)
		return err == nil || !errors.Is(err, fs.ErrNotExist)
	}
}

// IsDirFS 判断 fsys 中的 path 是否为目录
func IsDirFS(fsys fs.FS) func(string) bool {
	return func(path string) bool {
		s, err := os.Stat(path)
		return err == nil && s.IsDir()
	}
}

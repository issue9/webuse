// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package validator

import (
	"os"
	"testing"

	"github.com/issue9/assert/v4"
)

func TestFileExists(t *testing.T) {
	a := assert.New(t, false)

	a.True(FileExists("./file.go")).
		False(FileExists("./not-exists.go"))

	f := FileExistsFS(os.DirFS("./"))
	a.True(f("file.go")).
		False(f("not-exists.go"))
}

func TestIsDir(t *testing.T) {
	a := assert.New(t, false)

	a.True(IsDir(".")).
		False(IsDir("./file.go")).
		False(IsDir("not-exists.go"))

	f := IsDirFS(os.DirFS("./"))
	a.True(f(".")).
		False(f("file.go")).
		False(f("not-exists.go"))
}

# middleware

[![Build Status](https://github.com/issue9/middleware/workflows/Go/badge.svg)](https://github.com/issue9/middleware/actions?query=workflow%3AGo)
[![license](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat)](https://opensource.org/licenses/MIT)
[![codecov](https://codecov.io/gh/issue9/middleware/branch/master/graph/badge.svg)](https://codecov.io/gh/issue9/middleware)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/issue9/middleware/v5)](https://pkg.go.dev/github.com/issue9/middleware/v5)
[![Go version](https://img.shields.io/github/go-mod/go-version/issue9/middleware)](https://golang.org)

middleware Go HTTP 的中间件，提供了大部分实用的功能。

可以选择与 [mux](https://pkg.go.dev/github.com/issue9/mux/v5) 一起使用；
也可以采用 Middlewares 直接与标准库的 net/http 一起使用。

- auth 基本的验证处理；
- compress 对内容进行压缩；
- errorhandler 处理各类状态码下的输出；
- health 接口状态的监测；
- ratelimit x-rate-limit 的相关实现；
- recovery 对 Panic 的处理；
- debugger 用于输出测试的中间件；

## 安装

```shell
go get github.com/issue9/middleware/v5
```

## 版权

本项目采用 [MIT](https://opensource.org/licenses/MIT) 开源授权许可证，完整的授权说明可在 [LICENSE](LICENSE) 文件中找到。

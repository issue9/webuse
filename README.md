middleware
[![Build Status](https://github.com/issue9/middleware/workflows/Go/badge.svg)](https://github.com/issue9/middleware/actions?query=workflow%3AGo)
[![license](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat)](https://opensource.org/licenses/MIT)
[![codecov](https://codecov.io/gh/issue9/middleware/branch/master/graph/badge.svg)](https://codecov.io/gh/issue9/middleware)
======

middleware 是实现 http.Handler 接口的中间件，提供了大部分实用的功能。

- auth 基本的验证处理；
- compress 对内容进行压缩；
- errorhandler 处理各类状态码下的输出；
- header 输出指定的报头；
- health 接口状态的监测；
- host 匹配指定的域名；
- ratelimit x-rate-limit 的相关实现；
- recovery 对 Panic 的处理；
- version 匹配从 Accept 报头中的版本号信息；
- debugger 用于输出测试的中间件；

安装
---

```shell
go get github.com/issue9/middleware/v2
```

文档
---

[![Go Walker](https://gowalker.org/api/v1/badge)](https://gowalker.org/github.com/issue9/middleware)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/issue9/middleware/v2)](https://pkg.go.dev/github.com/issue9/middleware/v2)

版权
---

本项目采用 [MIT](https://opensource.org/licenses/MIT) 开源授权许可证，完整的授权说明可在 [LICENSE](LICENSE) 文件中找到。

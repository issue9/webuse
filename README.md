middleware
[![Build Status](https://github.com/issue9/middleware/workflows/Go/badge.svg)](https://github.com/issue9/middleware/actions?query=workflow%3AGo)
======

middleware 是实现 http.Handler 接口的中间件，提供了大部分实用的功能。

- version 匹配从 Accept 报头中的版本号信息；
- compress 对内容进行压缩；
- host 匹配指定的域名；
- recovery 对 Panic 的处理；
- errorhandler 处理各类状态码下的输出；
- header 输出指定的报头；
- auth 基本的验证处理；

安装
---

```shell
go get github.com/issue9/middleware
```

文档
---

[![Go Walker](https://gowalker.org/api/v1/badge)](https://gowalker.org/github.com/issue9/middleware)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/issue9/middleware)](https://pkg.go.dev/github.com/issue9/middleware)

版权
---

本项目采用 [MIT](https://opensource.org/licenses/MIT) 开源授权许可证，完整的授权说明可在 [LICENSE](LICENSE) 文件中找到。

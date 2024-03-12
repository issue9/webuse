# webuse

[![Go](https://github.com/issue9/webuse/actions/workflows/go.yml/badge.svg)](https://github.com/issue9/webuse/actions/workflows/go.yml)
[![license](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat)](https://opensource.org/licenses/MIT)
[![codecov](https://codecov.io/gh/issue9/webuse/branch/master/graph/badge.svg)](https://codecov.io/gh/issue9/webuse)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/issue9/webuse/v7)](https://pkg.go.dev/github.com/issue9/webuse/v7)
[![Go version](https://img.shields.io/github/go-mod/go-version/issue9/webuse)](https://golang.org)

 适用于 [web](https://pkg.go.dev/github.com/issue9/web) 的中间件和插件；

## 插件

插件位于 [plugins](plugins) 目录之下：

- access 客户端访问记录；
- health 接口状态的监测；
- compress 根据 CPU 使用率决定是否启用压缩功能；

## 中间件

中间件位于 [middlewares](middlewares) 目录之下：

- auth/basic 基本的验证处理；
- auth/jwt JSON Web Tokens 中间件；
- auth/session session 管理；
- acl/ratelimit x-rate-limit 的相关实现；
- acl/iplist 黑白名单；

## 安装

```shell
go get github.com/issue9/webuse/v7
```

## 版权

本项目采用 [MIT](https://opensource.org/licenses/MIT) 开源授权许可证，完整的授权说明可在 [LICENSE](LICENSE) 文件中找到。

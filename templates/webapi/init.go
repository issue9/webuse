package main

import (
	"flag"

	"github.com/issue9/web"
	"github.com/issue9/web/mimetype/cbor"
	"github.com/issue9/web/mimetype/json"
	"github.com/issue9/web/openapi"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/app"
	"github.com/issue9/webuse/v7/handlers/debug"
	"github.com/issue9/webuse/v7/middlewares/auth/token"
	"github.com/issue9/webuse/v7/openapis"
)

func Exec(id, version string) error {
	return app.NewCLI(&app.CLIOptions[struct{}]{
		ID:             id,
		Version:        version,
		NewServer:      initServer,
		ConfigDir:      "./",
		ConfigFilename: "web.yaml",
		ErrorHandling:  flag.ExitOnError,
		Daemon: &app.DaemonConfig{
			DisplayName: web.Phrase("TODO"),
			Description: web.Phrase("TODO"),
			Arguments:   []string{"-a=serve"},
		},
	}).Exec()
}

func initServer(id, ver string, o *server.Options, u struct{}) (web.Server, error) {
	s, err := server.NewHTTP(id, ver, o)
	if err != nil {
		return nil, err
	}

	router := s.Routers().New("main", nil,
		web.WithAnyInterceptor("any"),
		web.WithDigitInterceptor("digit"),
	)
	debug.RegisterDev(router, "/debug")

	doc := openapi.New(s, web.Phrase("The api doc of %s", s.ID()),
		openapi.WithMediaType(json.Mimetype, cbor.Mimetype),
		openapi.WithProblemResponse(),
		openapi.WithSecurityScheme(token.SecurityScheme("token", web.Phrase("token auth"))),
		openapis.WithCDNViewer(s, "scalar", ""),
	)
	router.Get("/openapi", doc.Handler())

	// TODO 初始化服务

	// 在所有模块加载完成之后调用，需要等待其它模块里的私有错误代码加载完成。
	doc.WithDescription(nil, web.Phrase(`problems response:

%s
`, openapi.MarkdownProblems(s, 4)))

	return s, nil
}

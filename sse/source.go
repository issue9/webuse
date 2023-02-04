// SPDX-License-Identifier: MIT

package sse

import (
	"net/http"
	"strconv"

	"github.com/issue9/errwrap"
	"github.com/issue9/web"
)

// WaitFunc 等待 [Source] 完成的阻塞函数
//
// 当 [Source] 因为某种原因退出时，WaitFunc 才会返回，
// 用于在 [web.Handler] 的处理函数阻止服务的退出。
type WaitFunc func()

type Source interface {
	// 关闭当前事件源
	//
	// 这将导致关联的 [WaitFunc] 返回。
	Close()

	// Sent 发送消息
	//
	// id、event 和  retry 都可以为空，表示不需要这些值；
	Sent(data []string, event, id string, retry uint)
}

type source struct {
	data chan []byte
	exit chan struct{}
	done chan struct{}
}

// Get 返回指定 ID 的事件源
func (srv *Server[T]) Get(id T) Source { return srv.sources[id] }

// NewSource 声明新的事件源
//
// NOTE: 只有采用此方法声明之后，才有可能通过 [Events.Get] 获取实例。
// id 表示是事件源的唯一 ID，如果事件是根据用户进行区分的，那么该值应该是表示用户的 ID 值。
func (srv *Server[T]) NewSource(id T, ctx *web.Context) (Source, WaitFunc) {
	if srv.sources[id] != nil {
		srv.sources[id].Close()
	}

	s := &source{
		data: make(chan []byte, 1),
		exit: make(chan struct{}, 1),
		done: make(chan struct{}, 1),
	}
	srv.sources[id] = s

	go func() {
		s.connect(ctx, srv.status) // 阻塞，出错退出
		close(s.data)              // 退出之前关闭，防止退出之后，依然有数据源源不断地从 Sent 输入。
		delete(srv.sources, id)    // 如果 connect 返回，说明断开了连接，删除 sources 中的记录。
	}()
	return s, s.wait
}

// 和客户端进行连接，如果返回，则表示连接被关闭。
func (s *source) connect(ctx *web.Context, status int) {
	ctx.Header().Set("content-type", "text/event-stream; charset=utf-8")
	ctx.Header().Set("Content-Length", "0")
	ctx.Header().Set("Cache-Control", "no-cache")
	ctx.Header().Set("Connection", "keep-alive")
	ctx.SetCharset("utf-8")
	ctx.SetEncoding("")

	var rw http.ResponseWriter = ctx
	f, ok := rw.(http.Flusher)
	for !ok {
		if rr, ok := rw.(interface{ Unwrap() http.ResponseWriter }); ok {
			rw = rr.Unwrap()
			f, ok = rw.(http.Flusher)
			continue
		}
		break
	}
	if f == nil {
		ctx.WriteHeader(http.StatusInternalServerError)
		ctx.Logs().ERROR().String("ctx 无法转换成 http.Flusher")
		return
	}

	ctx.WriteHeader(status)
	for {
		select {
		case <-s.exit:
			s.done <- struct{}{}
			return
		case data := <-s.data:
			if _, err := ctx.Write(data); err != nil { // 出错即退出，由客户端自行重连。
				ctx.Logs().ERROR().Error(err)
				s.done <- struct{}{}
				return
			}
			f.Flush()
		}
	}
}

func (s *source) Sent(d []string, event, id string, retry uint) {
	w := errwrap.Buffer{}
	for _, line := range d {
		w.WString("data: ").WString(line).WByte('\n')
	}
	if event != "" {
		w.WString("event: ").WString(event).WByte('\n')
	}
	if id != "" {
		w.WString("id: ").WString(id).WByte('\n')
	}
	if retry > 0 {
		w.WString("retry: ").WString(strconv.Itoa(int(retry))).WByte('\n')
	}
	w.WByte('\n')

	s.data <- w.Bytes()
}

func (s *source) Close() { s.exit <- struct{}{} }

func (s *source) wait() { <-s.done }

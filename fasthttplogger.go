package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

// https://github.com/AubSs/fasthttplogger

var (
	output = log.New(os.Stdout, "", 0)
)

var (
	green  = string([]byte{27, 91, 48, 48, 58, 51, 50, 109})
	yellow = string([]byte{27, 91, 48, 48, 59, 51, 51, 109})
	red    = string([]byte{27, 91, 48, 48, 59, 51, 49, 109})
	blue   = string([]byte{27, 91, 48, 48, 59, 51, 52, 109})
	white  = string([]byte{27, 91, 48, 109})
)

func getColorByStatus(code int) string {
	switch {
	case code >= 200 && code < 300:
		return green
	case code >= 300 && code < 400:
		return blue
	case code >= 400 && code < 500:
		return yellow
	default:
		return red
	}
}

func colorStatus(code int) string {
	return getColorByStatus(code) + strconv.Itoa(code) + white
}

func colorMethod(method []byte, code int) string {
	return getColorByStatus(code) + string(method) + white
}

func getHttp(ctx *fasthttp.RequestCtx) string {
	if ctx.Response.Header.IsHTTP11() {
		return "HTTP/1.1"
	}
	return "HTTP/1.0"
}

func getRealRemoteIP(ctx *fasthttp.RequestCtx) string {
	xForwardedFor := ctx.Request.Header.Peek("X-Forwarded-For")
	if len(xForwardedFor) != 0 {
		parts := strings.Split(string(xForwardedFor), ",")
		if len(parts) != 0 {
			ip := net.ParseIP(parts[0])
			if ip != nil && ip.IsGlobalUnicast() {
				return ip.String()
			}
		}
	}
	xRealIP := ctx.Request.Header.Peek("X-Real-IP")
	if len(xRealIP) > 0 {
		ip := net.ParseIP(string(xRealIP))
		if ip != nil && ip.IsGlobalUnicast() {
			return ip.String()
		}
	}
	return ctx.RemoteIP().String()
}

/* ========================== Predefined Formats =========================== */

// Tiny format:
// <method> <url> - <status> - <response-time us>
// GET / - 200 - 11.925 us
func Tiny(req fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		begin := time.Now()
		req(ctx)
		end := time.Now()
		output.Printf("%s %s - %v - %v",
			ctx.Method(),
			ctx.RequestURI(),
			ctx.Response.Header.StatusCode(),
			end.Sub(begin),
		)
	})
}

// TinyColored is same as Tiny but colored
func TinyColored(req fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		begin := time.Now()
		req(ctx)
		end := time.Now()
		output.Printf("%s %s - %v - %v",
			colorMethod(ctx.Method(), ctx.Response.Header.StatusCode()),
			ctx.RequestURI(),
			colorStatus(ctx.Response.Header.StatusCode()),
			end.Sub(begin),
		)
	})
}

// Short format:
// <remote-addr> | <real-remote-ip> | <HTTP/:http-version> | <method> <url> - <status> - <response-time us>
// 127.0.0.1:53324 | 1.1.1.1 | HTTP/1.1 | GET /hello - 200 - 44.8µs
func Short(req fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		begin := time.Now()
		req(ctx)
		end := time.Now()
		output.Printf("%v | %s | %s | %s %s - %v - %v",
			ctx.RemoteAddr(),
			getRealRemoteIP(ctx),
			getHttp(ctx),
			ctx.Method(),
			ctx.RequestURI(),
			ctx.Response.Header.StatusCode(),
			end.Sub(begin),
		)
	})
}

// ShortColored is same as Short but colored
func ShortColored(req fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		begin := time.Now()
		req(ctx)
		end := time.Now()
		output.Printf("%v | %s | %s | %s %s - %v - %v",
			ctx.RemoteAddr(),
			getRealRemoteIP(ctx),
			getHttp(ctx),
			colorMethod(ctx.Method(), ctx.Response.Header.StatusCode()),
			ctx.RequestURI(),
			colorStatus(ctx.Response.Header.StatusCode()),
			end.Sub(begin),
		)
	})
}

// Combined format:
// [<time>] <remote-addr> | <real-remote-ip> | <HTTP/http-version> | <method> <url> - <status> - <response-time us> | <user-agent>
// [2017/05/31 - 13:27:28] 127.0.0.1:54082 | HTTP/1.1 | GET /hello - 200 - 48.279µs | Paw/3.1.1 (Macintosh; OS X/10.12.5) GCDHTTPRequest
func Combined(req fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		begin := time.Now()
		req(ctx)
		end := time.Now()
		output.Printf("[%v] %v | %s | %s | %s %s - %v - %v | %s",
			end.Format("2006/01/02 - 15:04:05"),
			ctx.RemoteAddr(),
			getRealRemoteIP(ctx),
			getHttp(ctx),
			ctx.Method(),
			ctx.RequestURI(),
			ctx.Response.Header.StatusCode(),
			end.Sub(begin),
			ctx.UserAgent(),
		)
	})
}

// CombinedColored is same as Combined but colored
func CombinedColored(req fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		begin := time.Now()
		req(ctx)
		end := time.Now()
		output.Printf("[%v] %v | %s | %s | %s %s - %v - %v | %s",
			end.Format("2006/01/02 - 15:04:05"),
			ctx.RemoteAddr(),
			getRealRemoteIP(ctx),
			getHttp(ctx),
			colorMethod(ctx.Method(), ctx.Response.Header.StatusCode()),
			ctx.RequestURI(),
			colorStatus(ctx.Response.Header.StatusCode()),
			end.Sub(begin),
			ctx.UserAgent(),
		)
	})
}

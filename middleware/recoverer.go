package middleware

import (
	"bytes"
	"log"
	"runtime/debug"

	"github.com/emicklei/go-restful"

	"github.com/johnwilson/restapi/system"
)

// Recoverer is a middleware that recovers from panics, logs the panic (and a
// backtrace), and returns a HTTP 500 (Internal Server Error) status if
// possible.
//
// Recoverer prints a request ID if one is provided.
func Recoverer(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	reqID := GetReqID(req)

	defer func() {
		if err := recover(); err != nil {
			printPanic(reqID, err)
			debug.PrintStack()
			msg := "Application encountered and error. Contact admin."
			system.WriteError(msg, err, 500, resp)
		}
	}()

	chain.ProcessFilter(req, resp)
}

func printPanic(reqID string, err interface{}) {
	var buf bytes.Buffer

	if reqID != "" {
		cW(&buf, bBlack, "[%s] ", reqID)
	}
	cW(&buf, bRed, "panic: %+v", err)

	log.Print(buf.String())
}

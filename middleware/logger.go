package middleware

import (
	"bytes"
	"log"
	"time"

	"github.com/emicklei/go-restful"
)

// Logger is a middleware that logs the start and end of each request, along
// with some useful data about what was requested, what the response status was,
// and how long it took to return. When standard output is a TTY, Logger will
// print in color, otherwise it will print in black and white.
//
// Logger prints a request ID if one is provided.
//
// Logger has been designed explicitly to be Good Enough for use in small
// applications and for people just getting started with Goji. It is expected
// that applications will eventually outgrow this middleware and replace it with
// a custom request logger, such as one that produces machine-parseable output,
// outputs logs to a different service (e.g., syslog), or formats lines like
// those printed elsewhere in the application.
func Logger(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	reqID := GetReqID(req)
	printStart(reqID, req)

	t1 := time.Now()
	chain.ProcessFilter(req, resp)

	t2 := time.Now()
	printEnd(reqID, resp, t2.Sub(t1))
}

func printStart(reqID string, req *restful.Request) {
	var buf bytes.Buffer

	if reqID != "" {
		cW(&buf, bBlack, "[%s] ", reqID)
	}
	buf.WriteString("Started ")
	cW(&buf, bMagenta, "%s ", req.Request.Method)
	cW(&buf, nBlue, "%q ", req.Request.URL.String())
	buf.WriteString("from ")
	buf.WriteString(req.Request.RemoteAddr)

	log.Print(buf.String())
}

func printEnd(reqID string, resp *restful.Response, dt time.Duration) {
	var buf bytes.Buffer

	if reqID != "" {
		cW(&buf, bBlack, "[%s] ", reqID)
	}
	buf.WriteString("Returning ")
	status := resp.StatusCode()
	if status < 200 {
		cW(&buf, bBlue, "%03d", status)
	} else if status < 300 {
		cW(&buf, bGreen, "%03d", status)
	} else if status < 400 {
		cW(&buf, bCyan, "%03d", status)
	} else if status < 500 {
		cW(&buf, bYellow, "%03d", status)
	} else {
		cW(&buf, bRed, "%03d", status)
	}
	buf.WriteString(" in ")
	if dt < 500*time.Millisecond {
		cW(&buf, nGreen, "%s", dt)
	} else if dt < 5*time.Second {
		cW(&buf, nYellow, "%s", dt)
	} else {
		cW(&buf, nRed, "%s", dt)
	}

	log.Print(buf.String())
}

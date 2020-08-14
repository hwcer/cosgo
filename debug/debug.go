package debug

import (
	"context"
	"fmt"
	"log"
	"net/http"
	nhpprof "net/http/pprof"
	"net/url"
	"runtime/pprof"
	"strings"
	"text/template"
	"time"
)

var (
	srvPprof    *http.Server
	srvAddr     = ""
	TestHandler = func(w http.ResponseWriter, r *http.Request) {}
)

func StartPprofSrv(addr string) {
	if srvPprof != nil {
		return
	}
	if srvAddr == "" && addr != "" {
		srvAddr = addr
	}

	srvPprof = new(http.Server)
	r := http.NewServeMux()
	srvPprof.Addr = srvAddr
	srvPprof.Handler = r

	log.Printf("pprof server start, addr:%v\n", srvAddr)
	go func() {
		RegisterHandler("/debug/pprof/", r)
		if err := srvPprof.ListenAndServe(); err != nil {
			srvPprof = nil
			log.Printf("pprof server start result:%v", err)
		}
	}()
}

func StopPprofSrv() {
	if srvPprof == nil {
		return
	}

	//Wait no longer than 2 seconds before halting
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	err := srvPprof.Shutdown(ctx)
	log.Printf("pprof server gracefully stopped:%v\n", err)

	srvPprof = nil
}

func handler(token string) http.Handler {
	info := struct {
		Profiles []*pprof.Profile
		Token    string
		Gcstat   string
	}{
		Token: url.QueryEscape(token),
	}

	h := func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		switch name {
		case "":
			// Index page.

			info.Profiles = pprof.Profiles()
			info.Gcstat = GCSummary()
			if err := indexTmpl.Execute(w, info); err != nil {
				log.Println(err)
				return
			}
		case "index":
			nhpprof.Index(w, r)
		case "cmdline":
			nhpprof.Cmdline(w, r)
		case "profile":
			nhpprof.Profile(w, r)
		case "trace":
			nhpprof.Trace(w, r)
		case "symbol":
			nhpprof.Symbol(w, r)
		case "test":
			TestHandler(w, r)
		default:
			// Provides access to all profiles under runtime/pprof
			nhpprof.Handler(name).ServeHTTP(w, r)
		}
	}
	return http.HandlerFunc(h)
}

func Handler() http.Handler {
	return handler("")
}

func RegisterHandler(prefix string, mux *http.ServeMux) {
	mux.Handle(prefix, http.StripPrefix(prefix, Handler()))
}

func AuthHandler(token string) http.Handler {
	h := handler(token)
	ah := func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("token") == token {
			h.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintln(w, "Unauthorized.")
		}
	}
	return http.HandlerFunc(ah)
}

func RegisterAuthHandler(token, prefix string, mux *http.ServeMux) {
	mux.Handle(prefix, http.StripPrefix(prefix, AuthHandler(token)))
}

var indexTmpl = template.Must(template.New("index").Parse(`<html>
  <head>
    <title>Debug Information</title>
  </head>
  <br>
  <body>
    profiles:<br>
    <table>
    {{range .Profiles}}
      <tr><td align=right>{{.Count}}<td><a href="{{.Name}}?debug=1{{if $.Token}}&token={{$.Token}}{{end}}">{{.Name}}</a>
    {{end}}
    <tr><td align=right><td><a href="profile?seconds=5{{if .Token}}?token={{.Token}}{{end}}">5-second CPU</a>
    <tr><td align=right><td><a href="profile?seconds=30{{if .Token}}?token={{.Token}}{{end}}">30-second CPU</a>
    <tr><td align=right><td><a href="profile?seconds=60{{if .Token}}?token={{.Token}}{{end}}">60-secondCPU</a>
    <tr><td align=right><td><a href="trace?seconds=5{{if .Token}}&token={{.Token}}{{end}}">5-second trace</a>
    <tr><td align=right><td><a href="trace?seconds=30{{if .Token}}&token={{.Token}}{{end}}">30-second trace</a>
    <tr><td align=right><td><a href="trace?seconds=60{{if .Token}}&token={{.Token}}{{end}}">60-second trace</a>
    </table>
    <br>
    debug information:<br>
    <table>
      <tr><td align=right><td><a href="cmdline{{if .Token}}?token={{.Token}}{{end}}">cmdline</a>
      <tr><td align=right><td><a href="symbol{{if .Token}}?token={{.Token}}{{end}}">symbol</a>
    <tr><td align=right><td><a href="goroutine?debug=2{{if .Token}}&token={{.Token}}{{end}}">full goroutine stack dump</a><br>
    <table>
    <br>
    gcstat: {{.Gcstat}}
    <br>
    prof cmds:<br>
     go tool pprof http://HOST:PORT/debug/pprof/heap <br>
     go tool pprof http://HOST:PORT/debug/pprof/profile <br>
     go tool pprof http://HOST:PORT/debug/pprof/block <br>
  </body>
</html>`))

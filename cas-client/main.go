package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"

	"gopkg.in/cas.v2"
)

var casURL = "http://127.0.0.1:8000/cas"

var (
	port = flag.Int("port", 9999, "cas client port")
)

func main() {
	flag.Parse()

	m := http.NewServeMux()
	m.HandleFunc("/example/cas", handler)
	m.HandleFunc("/example/cas/logout", logoutHandler)
	url, _ := url.Parse(casURL)
	client := cas.NewClient(&cas.Options{
		URL:         url,
		SendService: true,
		// 默认的 TicketStore 是 MemoryStore，需要业务方自行修改为 RedisTicketStore
		// TicketStore: RedisTicketStore,
		// 默认的 SessionStore 是 MemorySessionStore，需要业务方自行修改为 RedisSessionStore
		// SessionStore: RedisSessionStore,
	})

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", *port),
		// 使用 cas client 包装后，会自动通过 ticket 获取用户信息，并且设置 cookie
		Handler: client.Handle(m),
	}
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if cas.IsAuthenticated(r) {
		cas.RedirectToLogout(w, r)
		return
	}

	fmt.Fprintf(w, logout_html)
}

func handler(w http.ResponseWriter, r *http.Request) {
	// 未登录时重定向到 SSO CAS 登陆页
	if !cas.IsAuthenticated(r) {
		log.Printf("not login, %s\n", r.RequestURI)
		cas.RedirectToLogin(w, r)
		return
	}

	// 已经登录
	// 打印从 SSO 获取到的用户信息

	w.Header().Add("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.New("index.html").Parse(index_html)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, error_500, err)
		return
	}

	var binding = struct {
		Username   string
		Attributes cas.UserAttributes
	}{
		Username:   cas.Username(r),
		Attributes: cas.Attributes(r),
	}

	html := new(bytes.Buffer)
	if err := tmpl.Execute(html, binding); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, error_500, err)
		return
	}

	html.WriteTo(w)
}

const index_html = `<!DOCTYPE html>
<html>
  <head>
    <title>Welcome {{.Username}}</title>
  </head>
  <body>
    <h1>Welcome {{.Username}} <a href="/example/cas/logout">Logout</a></h1>
  </body>
</html>
`

const error_500 = `<!DOCTYPE html>
<html>
  <head>
    <title>Error 500</title>
  </head>
  <body>
    <h1>Error 500</h1>
    <p>%v</p>
  </body>
</html>
`

const logout_html = `
<!DOCTYPE html>
<html>
  <head>
    <title>Logout Success</title>
  </head>
  <body>
	<p>Logout Success</p>
	<p><a href="/example/cas">Login</a></p>
  </body>
</html>`

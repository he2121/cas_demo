package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

// DEMO APP
var testAPPURL = map[string]string{
	"http://127.0.0.1:8001/example/cas": "app1",
	"http://127.0.0.1:8002/example/cas": "app2",
}

func main() {
	http.HandleFunc("/cas/login", HandLogin)
	http.HandleFunc("/cas/validate", HandValidate)

	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Println(err)
	}
}

func HandValidate(w http.ResponseWriter, r *http.Request) {
	st := r.URL.Query().Get("ticket")
	service := r.URL.Query().Get("service")
	// 校验 ST 通过
	if ticket, ok := memoryTicket[st]; ok && ticket.Service == service {
		delete(memoryTicket, st)
		fmt.Fprintf(w, `yes
heganghuan
`)
	} else {
		fmt.Fprintf(w, `no
`)
	}
}

// HandLogin http://cas-server.com/cas/login?service=http://app1.com
func HandLogin(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL, r.Host, r.Method, r.RemoteAddr)
	// 取需要单点登录服务的 URL
	service := r.URL.Query().Get("service")
	// 验证此 URL 是否已经在 CAS-Server 登记过
	if err := validateService(service); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	// 是否在 CAS-Sever 已经登录， Get 验证 TGT
	if r.Method == http.MethodGet {
		tgc, err := r.Cookie("CASTGC")
		if err != nil || memoryTicket[tgc.Value] == nil {
			// GET 方法没有登录过，返回登录页面
			w.Header().Add("Content-Type", "text/html; charset=utf-8")
			tmpl, err := template.New("login.html").Parse(loginHTML)
			if err != nil {
				log.Println(err)
			}
			var binding = struct {
				Service string
			}{
				Service: service,
			}
			html := new(bytes.Buffer)
			if err := tmpl.Execute(html, binding); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "%+v", err)
				return
			}
			html.WriteTo(w)
			return
		}
	}
	// post 需要用户密码认证
	if r.Method == http.MethodPost {
		userName, password := r.FormValue("username"), r.FormValue("password")
		// 假装验证 user_name 与 password
		if ok := validateUser(userName, password); !ok {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("user or passport wrong"))
			return
		}
		// 通过 CAS-Server 认证，利用 TGC 与 TGT 维护在 CAS-Server 的登录态
		tgt := NewTGT()
		http.SetCookie(w, &http.Cookie{
			Name:   "CASTGC",
			Value:  tgt.Ticket,
			MaxAge: 300,
		})
	}

	// 发放 ST ticket
	ticket := NewServiceTicket(service)
	http.Redirect(w, r, ticket.Service+"?ticket="+ticket.Ticket, http.StatusFound)
}

func validateUser(name string, passport string) bool {
	// 验证用户名密码, demo 判断
	if name == passport && name != "" {
		return true
	}
	return false
}

func validateService(service string) error {
	if _, ok := testAPPURL[service]; ok {
		return nil
	}
	return fmt.Errorf("invalid service")
}

const loginHTML = `<!DOCTYPE html>
<html>
  <head>
    <title> SSO CAS Login</title>
  </head>
  <body>
	<form action="/cas/login?service={{.Service}}" method="post">
		 <input type='text' name='username'/>
		 <input type='password' name='password'/>
		 <input type='submit' value='登录'/> 
	</form>
  </body>
</html>
`

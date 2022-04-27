// OpenConnect Auth Hack
// https://github.com/thomaswilbur/OpenConnect-Auth-Hack/
// From: https://gist.github.com/horsley/e286276f83cae9b60d98
package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
)

//const authFormOld = `<?xml version="1.0" encoding="UTF-8"?>
//<config-auth client="vpn" type="auth-request">
//<version who="sg">0.1(1)</version>
//<auth id="main">
//<message>Please enter your username</message>
//<form method="post" action="/auth">
//<input type="text" name="username" label="Username:" />
//</form></auth>
//</config-auth>`

const authFormOld = `<?xml version="1.0" encoding="UTF-8"?>
<config-auth client="vpn" type="auth-request">
<version who="sg">0.1(1)</version>
<auth id="main">
<message>Please enter your username.</message>
<form method="post" action="/auth">
<input type="text" name="username" label="Username:" />
</form></auth>
</config-auth>`

const authFormNew = `<?xml version="1.0" encoding="UTF-8"?>
<config-auth client="vpn" type="auth-request">
<version who="sg">0.1(1)</version>
<auth id="main">
<message>Authentication Required</message>
<form method="post" action="/auth">
<input type="text" name="username" label="Username:" />
<input type="password" name="password" label="Password:" />
</form></auth>
</config-auth>`

const loginFailOld = `<?xml version="1.0" encoding="UTF-8"?>
<config-auth client="vpn" type="auth-request">
<version who="sg">0.1(1)</version>
<auth id="main">
<message>Login failed.
Please enter your password.</message>
<form method="post" action="/auth">
<input type="password" name="password" label="Password:" />
</form></auth>
</config-auth>`

const loginFailNew = `<?xml version="1.0" encoding="UTF-8"?>
<config-auth client="vpn" type="auth-request">
<version who="sg">0.1(1)</version>
<auth id="main">
<message>Login failed.
Please enter your password.</message>
<form method="post" action="/auth">
<input type="text" name="username" label="Username:" />
<input type="password" name="password" label="Password:" />
</form></auth>
</config-auth>`

var authUserPassRE = regexp.MustCompile("<auth><username>(.*?)</username><password>(.*?)</password></auth>")
var vpnCookieRE = regexp.MustCompile("webvpncontext=(.*?);")

func main() {
	//@todo load config from file
	cert, _ := tls.LoadX509KeyPair("CERT.PEM", "KEY.PEM")
	var cfg tls.Config

	cfg.Certificates = []tls.Certificate{cert}
	l, err := tls.Listen("tcp", ":443", &cfg) //@todo load config from file
	if err != nil {
		log.Println("listen error:", err)
		os.Exit(1)
	}

	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue
		}

		go func(clientConn net.Conn) {
			log.Println("conn process start")
			defer clientConn.Close()

			//@todo load config from file
			config := &tls.Config{InsecureSkipVerify: true}
			remoteConn, err := tls.Dial("tcp", "MYSERVER.COM:4433", config) //remote can be unix cleartext socket of ocserv
			if err != nil {
				log.Println("connect upstream error:", err)
				return
			}

			defer remoteConn.Close()

			vpnCookie := make(chan string, 1)

			go func() { //remote => client
				hackRemoteResponse(clientConn, remoteConn, vpnCookie)

				io.Copy(clientConn, remoteConn)
			}()

			//client => remote
			hackClientRequest(clientConn, remoteConn, vpnCookie)
			io.Copy(remoteConn, clientConn)

			log.Println("conn process done")
		}(conn)
	}

}

//hackClientRequest
//hack auth by resend both user and pass form and cookie
func hackClientRequest(clientConn, remoteConn net.Conn, vpnCookie chan string) {
	reader := bufio.NewReader(clientConn)

	for { //client req hacking
		req, err := http.ReadRequest(reader)

		if err != nil {
			log.Println("ReadRequest from client conn err:", err)
			return
		}

		if req.URL.Path == "/auth" {
			var buf bytes.Buffer
			_, err = buf.ReadFrom(req.Body)
			req.Body.Close()

			if err != nil {
				log.Println("read request body from client conn err:", err)
				return
			}

			req.Body = ioutil.NopCloser(bytes.NewReader(buf.Bytes()))

			log.Println("3. second req, auth req, do first sending")
			req.Write(remoteConn)

			//waiting for cookie
			ck := <-vpnCookie

			log.Println("5. got cookie, resend auth req with cookie")

			req.Header.Set("Cookie", fmt.Sprintf("webvpncontext=%s;", ck))
			req.Body = ioutil.NopCloser(bytes.NewReader(buf.Bytes()))
			req.Write(remoteConn)

			log.Println("6. resend done, hack finish, going to do regular io.Copy")
			break
		} else {
			log.Println("1. first req, pass")
			req.Write(remoteConn)
		}
	}
}

//hackRemoteResponse
//hack auth by change auth form and drop the first cookie resp
func hackRemoteResponse(clientConn, remoteConn net.Conn, vpnCookie chan string) {
	reader := bufio.NewReader(remoteConn)

	for { //remote resp hacking
		resp, err := http.ReadResponse(reader, nil)
		if err != nil {
			log.Println("ReadResponse from remote conn err:", err)
			return
		}

		if resp.ContentLength == int64(len(authFormOld)) { //hack auth form
			log.Println("2. first resp, hack resp auth form")
			resp.Body.Close()
			resp.ContentLength = int64(len(authFormNew))
			resp.Body = ioutil.NopCloser(strings.NewReader(authFormNew))
		} else if resp.ContentLength == int64(len(loginFailOld)) {
			log.Println("2. first resp, hack resp auth form")
			resp.Body.Close()
			resp.ContentLength = int64(len(loginFailNew))
			resp.Body = ioutil.NopCloser(strings.NewReader(loginFailNew))
		} else if ck := resp.Header.Get("Set-Cookie"); ck != "" && vpnCookieRE.MatchString(ck) { //get context cookie
			log.Println("4. second resp, grep cookie, resp hacking finish")

			if m := vpnCookieRE.FindStringSubmatch(ck); m != nil && m[1] != "" {
				vpnCookie <- m[1]
				//log.Println("get vpn cookie:", m[1])
				break
			}
		}

		err = resp.Write(clientConn)
		if err != nil {
			log.Println("write to client conn err:", err)
			return
		}
	}
}

package smtp

import (
	"encoding/base64"
	"fmt"
	"net/smtp"
	"strings"
	"crypto/tls"
	"net"
)

type Smtp struct {
	Address  string
	Username string
	Password string
}

func New(address, username, password string) *Smtp {
	return &Smtp{
		Address:  address,
		Username: username,
		Password: password,
	}
}

func Dial(addr string) (*smtp.Client, error) {
    conn, err := tls.Dial("tcp", addr, nil)
    if err != nil {
        fmt.Println("Dialing Error:", err)
        return nil, err
    }
    host, _, _ := net.SplitHostPort(addr)
    return smtp.NewClient(conn, host)
}

func _SendMail4TLS(addr string, auth smtp.Auth, from string,
    to []string, msg []byte) (err error) {
    //create smtp client
    c, err := Dial(addr)
    if err != nil {
        fmt.Println("Create smpt client error:", err)
        return err
    }
    defer c.Close()
    if auth != nil {
        if ok, _ := c.Extension("AUTH"); ok {
            if err = c.Auth(auth); err != nil {
                fmt.Println("Error during AUTH", err)
                return err
            }
        }
    }
    if err = c.Mail(from); err != nil {
        fmt.Println("Error mail:",err)
        return err
    }
    for _, addr := range to {
        if err = c.Rcpt(addr); err != nil {
            fmt.Println("Error Rcpt:",err,addr)
            return err
        }
    }
    w, err := c.Data()
    if err != nil {
        return err
    }
    _, err = w.Write(msg)
    if err != nil {
        return err
    }
    err = w.Close()
    if err != nil {
        return err
    }
    return c.Quit()
}

func (this *Smtp) _SendMail(istsl bool,from, tos, subject, body string, contentType string) error {
	if this.Address == "" {
		return fmt.Errorf("address is necessary")
	}

	hp := strings.Split(this.Address, ":")
	if len(hp) != 2 {
		return fmt.Errorf("address format error")
	}

	arr := strings.Split(tos, ";")
	count := len(arr)
	safeArr := make([]string, 0, count)
	for i := 0; i < count; i++ {
		if arr[i] == "" {
			continue
		}
		safeArr = append(safeArr, arr[i])
	}

	if len(safeArr) == 0 {
		return fmt.Errorf("tos invalid")
	}

	tos = strings.Join(safeArr, ";")

	b64 := base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/")

	header := make(map[string]string)
	header["From"] = from
	header["To"] = tos
	header["Subject"] = fmt.Sprintf("=?UTF-8?B?%s?=", b64.EncodeToString([]byte(subject)))
	header["MIME-Version"] = "1.0"

	ct := "text/plain; charset=UTF-8"
	if len(contentType) > 0 && contentType == "html" {
		ct = "text/html; charset=UTF-8"
	}

	header["Content-Type"] = ct
	header["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + b64.EncodeToString([]byte(body))

	auth := smtp.PlainAuth("", this.Username, this.Password, hp[0])
	if(istsl){
		return _SendMail4TLS(this.Address, auth, from, strings.Split(tos, ";"), []byte(message))
	} else {
		return smtp.SendMail(this.Address, auth, from, strings.Split(tos, ";"), []byte(message))
	}
}

func (this *Smtp) SendMail(from, tos, subject, body string, contentType string) error {
	return this._SendMail(false,from, tos, subject, body, contentType)
}

func (this *Smtp) SendMail4TLS(from, tos, subject, body string, contentType string) error {
	return this._SendMail(true,from, tos, subject, body, contentType)
}

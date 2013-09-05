package models

import (
    "fmt"
    "Mango/management/utils"
)

type MangoMail struct {
	from      string
	to        []string
	cc        []string
	bcc       []string
	subject   string
	html      string
	text      string
	headers   map[string]string
	options   map[string]string
	variables map[string]string
}

func (m *MangoMail) From() string {
	return m.from
}

func (m *MangoMail) To() []string {
	return m.to
}

func (m *MangoMail) Cc() []string {
	return m.cc
}

func (m *MangoMail) Bcc() []string {
	return m.bcc
}

func (m *MangoMail) Subject() string {
	return m.subject
}

func (m *MangoMail) Html() string {
	return m.html
}

func (m *MangoMail) Text() string {
	return m.text
}

func (m *MangoMail) Headers() map[string]string {
	return m.headers
}

func (m *MangoMail) Options() map[string]string {
	return m.options
}

func (m *MangoMail) Variables() map[string]string {
	return m.variables
}

func NewRegisterMail(emailAddr, token string) *MangoMail {
    to := make([]string, 0)
    to = append(to, emailAddr)
    url := utils.GenerateRegisterUrl(token)
    m := &MangoMail {
        from : "Guoku <noreply@post.guoku.com>",
        to : to,
        subject : "Mango Registration URL",
        html : fmt.Sprintf("<a href='%s'>点此进入</a>", url),
    }
    return m
}

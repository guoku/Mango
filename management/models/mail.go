package models

type MangoMail struct {
	FromVar      string
	ToVar        []string
	CcVar        []string
	BccVar       []string
	SubjectVar   string
	HtmlVar      string
	TextVar      string
	HeadersVar   map[string]string
	OptionsVar   map[string]string
	VariablesVar map[string]string
}

func (m *MangoMail) From() string {
	return m.FromVar
}

func (m *MangoMail) To() []string {
	return m.ToVar
}

func (m *MangoMail) Cc() []string {
	return m.CcVar
}

func (m *MangoMail) Bcc() []string {
	return m.BccVar
}

func (m *MangoMail) Subject() string {
	return m.SubjectVar
}

func (m *MangoMail) Html() string {
	return m.HtmlVar
}

func (m *MangoMail) Text() string {
	return m.TextVar
}

func (m *MangoMail) Headers() map[string]string {
	return m.HeadersVar
}

func (m *MangoMail) Options() map[string]string {
	return m.OptionsVar
}

func (m *MangoMail) Variables() map[string]string {
	return m.VariablesVar
}

package auth

import (
	"fmt"
	"strings"
	"text/template"
)

var confirmInstruction = "Please wait for your admin to confirm."
var confirmQuestion = "Do you want to continue [Y/n]? : "

const (
	actionAccepted        = "Accepted"
	actionFailed          = "Failed"
	actionPartialAccepted = "Partial accepted"
)

func CreateChallengerInstruction(options []string) (string, string) {
	if len(options) == 1 {
		opt := options[0]
		return mfaOptionInstruction, fmt.Sprintf(mfaOptionQuestion, opt)
	}
	opts := make([]mfaOption, 0, len(options))
	for i := range options {
		opts = append(opts, mfaOption{
			Index: i + 1,
			Value: options[i],
		})
	}
	var out strings.Builder
	_ = mfaSelectTmpl.Execute(&out, opts)
	return mfaSelectInstruction, out.String()
}

var (
	mfaSelectTmpl        = template.Must(template.New("mfaOptions").Parse(mfaOptions))
	mfaSelectInstruction = "please select MFA choice"
)

type mfaOption struct {
	Index int
	Value string
}

var mfaOptions = `{{ range . }}{{ .Index }}: {{.Value}}
{{end}} num>: `

var (
	mfaOptionInstruction = "Please enter MFA code."
	mfaOptionQuestion    = "[%s auth]: "
)

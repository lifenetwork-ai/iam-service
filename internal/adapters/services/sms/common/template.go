package common

import "text/template"

var OTPTemplate = template.Must(template.New("otp").Parse(`
		Dear Valued Customer,

To authenticate, please use the following One Time Password (OTP) from {{ .TenantName }}:

*{{ .OTP }}*

Your OTP will be valid for *{{ .TTL }}* {{ if eq .TTL 1 }}minute{{ else }}minutes{{ end }}. Do not share this OTP with anyone. {{ .TenantName }} takes your account security very seriously.
	`))

// .{0,10}is your OTP number at GENETICA
var SMSOTPTemplate = template.Must(template.New("sms_otp").Parse(`
{{ .OTP }} is your OTP number at {{ .TenantName }}
	`))

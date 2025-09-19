package common

import "text/template"

var OTPTemplate = template.Must(template.New("otp").Parse(`
		Dear Valued Customer,

To authenticate, please use the following One Time Password (OTP) from {{ .TenantName }}:

*{{ .OTP }}*

Your OTP will be valid for *{{ .TTL }}* {{ if eq .TTL 1 }}minute{{ else }}minutes{{ end }}. Do not share this OTP with anyone. {{ .TenantName }} takes your account security very seriously.
	`))

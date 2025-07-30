package sms

import "text/template"

// "Dear Valued Customer,

// To authenticate, please use the following One Time Password (OTP) from LIFE AI:

// 234653

// Your OTP will be valid for 2 minutes. Do not share this OTP with anyone. LIFE AI takes your account security very seriously."
var (
	Template = template.Must(template.New("sms").Parse(`
		Dear Valued Customer,

To authenticate, please use the following One Time Password (OTP) from {{ .TenantName }}:

*{{ .OTP }}*

Your OTP will be valid for *2 minutes*. Do not share this OTP with anyone. {{ .TenantName }} takes your account security very seriously.
	`))
)

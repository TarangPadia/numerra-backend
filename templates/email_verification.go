package templates

import "fmt"

func EmailVerificationTemplate(otp string) string {
	header := EmailHeader()
	footer := EmailFooter()

	body := fmt.Sprintf(`
<p>Hello,</p>
<p>Your verification code is: <strong>%s</strong></p>
<p>If you did not request this code, please ignore this email.</p>
<p>Thanks,<br />SurrealXP</p>
`, otp)

	return header + body + footer
}

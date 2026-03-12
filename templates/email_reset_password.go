package templates

import "fmt"

func EmailResetPasswordTemplate(resetLink string) string {
	header := EmailHeader()
	footer := EmailFooter()

	body := fmt.Sprintf(`
<p>Hello,</p>
<p>We received a request from you for a password reset.</p>
<p>Please click the following button to reset your password:</p>
<p><a href="%s" style="display:inline-block;padding:10px 20px;background-color:#00ddfd;color:#1a1a1a;text-decoration:none;border-radius:4px;">Reset Password</a></p>
<p>Please ignore this email if this request was not initiated by you.</p>
`, resetLink)

	return header + body + footer
}

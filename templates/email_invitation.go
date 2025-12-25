package templates

import "fmt"

func EmailInvitationTemplate(orgName, inviterName, inviteLink string) string {
	header := EmailHeader()
	footer := EmailFooter()

	body := fmt.Sprintf(`
<p>Hello,</p>
<p>You have been invited to join the organization <strong>%s</strong> by <strong>%s</strong>.</p>
<p>Please click the following button to accept the invitation:</p>
<p><a href="%s" style="display:inline-block;padding:10px 20px;background-color:#EC1651;color:white;text-decoration:none;border-radius:5px;">Accept Invitation</a></p>
<p>This invitation will expire in 5 days.</p>
`, orgName, inviterName, inviteLink)

	return header + body + footer
}

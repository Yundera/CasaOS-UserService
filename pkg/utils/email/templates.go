package email

import "fmt"

// GenerateMagicLinkEmail generates the email content for magic link login
// Returns EmailContent with subject, HTML body, plain text body, and logo attachment
func GenerateMagicLinkEmail(code, link string) EmailContent {
	subject := "Your Magic Link to Sign In"

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0;">
    <div style="font-family: 'Trebuchet MS', sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; color: #3d617f;">
        <img src="%s" alt="Yundera" style="width: 150px; height: auto; margin-bottom: 20px; display: block; margin-left: auto; margin-right: auto;" />

        <h2 style="font-family: 'Comfortaa','Trebuchet MS', sans-serif; color: #27aae1; text-align: center;">🔐 Sign In to Your Personal Cloud</h2>

        <p>Hello!</p>
        <p>You requested a magic link to sign in to your Personal Cloud Server. Use one of the following methods:</p>

        <h3>Method 1: Click the Magic Link</h3>
        <div style="text-align: center; margin: 20px 0;">
            <a href="%s" style="display: inline-block; padding: 15px 30px; background-color: #27aae1; color: white; text-decoration: none; border-radius: 8px; font-weight: bold;">✨ Sign In with Magic Link</a>
        </div>

        <div style="margin: 30px 0; text-align: center; color: #999;">— OR —</div>

        <h3>Method 2: Enter the Code</h3>
        <p>Enter this 6-digit code on the login page:</p>
        <div style="background: white; border: 2px dashed #27aae1; padding: 20px; text-align: center; margin: 20px 0; border-radius: 8px;">
            <div style="font-size: 32px; font-weight: bold; letter-spacing: 8px; color: #27aae1; font-family: 'Courier New', monospace;">%s</div>
        </div>

        <div style="background-color: #dcebf9; padding: 15px; border-radius: 8px; margin: 20px 0;">
            <strong>⏰ This code expires in 10 minutes</strong><br>
            For your security, this code can only be used once.
        </div>

        <p>If you didn't request this magic link, you can safely ignore this email.</p>

        <p style="margin-top: 30px; font-family: 'Comfortaa','Trebuchet MS', sans-serif; font-weight: bold; color: #27aae1;">Powered by Yundera Personal Cloud</p>
        <p style="font-size: 12px; color: #999;">This is an automated email, please do not reply.</p>
    </div>
</body>
</html>
`, YunderaLogoContentID, link, code)

	text := fmt.Sprintf(`
Sign In to Your Personal Cloud

You requested a magic link to sign in to your Personal Cloud Server.

METHOD 1: Click the magic link below:
%s

METHOD 2: Enter this 6-digit code on the login page:
%s

⏰ This code expires in 10 minutes and can only be used once.

If you didn't request this magic link, you can safely ignore this email.

---
Powered by Yundera Personal Cloud
This is an automated email, please do not reply.
`, link, code)

	return EmailContent{
		Subject:     subject,
		HTMLBody:    html,
		TextBody:    text,
		Attachments: []InlineAttachment{YunderaLogo()},
	}
}

// GeneratePasswordResetEmail generates the email content for password reset
// Returns EmailContent with subject, HTML body, plain text body, and logo attachment
func GeneratePasswordResetEmail(code, link string) EmailContent {
	subject := "Reset Your Password"

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0;">
    <div style="font-family: 'Trebuchet MS', sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; color: #3d617f;">
        <img src="%s" alt="Yundera" style="width: 150px; height: auto; margin-bottom: 20px; display: block; margin-left: auto; margin-right: auto;" />

        <h2 style="font-family: 'Comfortaa','Trebuchet MS', sans-serif; color: #27aae1; text-align: center;">🔑 Reset Your Password</h2>

        <p>Hello!</p>
        <p>You requested to reset your password for your Personal Cloud Server. Use one of the following methods:</p>

        <h3>Method 1: Click the Reset Link</h3>
        <div style="text-align: center; margin: 20px 0;">
            <a href="%s" style="display: inline-block; padding: 15px 30px; background-color: #27aae1; color: white; text-decoration: none; border-radius: 8px; font-weight: bold;">🔓 Reset Password</a>
        </div>

        <div style="margin: 30px 0; text-align: center; color: #999;">— OR —</div>

        <h3>Method 2: Enter the Code</h3>
        <p>Enter this 6-digit code on the password reset page:</p>
        <div style="background: white; border: 2px dashed #27aae1; padding: 20px; text-align: center; margin: 20px 0; border-radius: 8px;">
            <div style="font-size: 32px; font-weight: bold; letter-spacing: 8px; color: #27aae1; font-family: 'Courier New', monospace;">%s</div>
        </div>

        <div style="background-color: #dcebf9; padding: 15px; border-radius: 8px; margin: 20px 0;">
            <strong>⏰ This code expires in 15 minutes</strong><br>
            For your security, this code can only be used once.
        </div>

        <div style="background-color: #dcebf9; padding: 15px; border-radius: 8px; margin: 20px 0;">
            <strong>🛡️ Security Notice</strong><br>
            If you didn't request a password reset, please ignore this email. Your password will remain unchanged.
        </div>

        <p style="margin-top: 30px; font-family: 'Comfortaa','Trebuchet MS', sans-serif; font-weight: bold; color: #27aae1;">Powered by Yundera Personal Cloud</p>
        <p style="font-size: 12px; color: #999;">This is an automated email, please do not reply.</p>
    </div>
</body>
</html>
`, YunderaLogoContentID, link, code)

	text := fmt.Sprintf(`
Reset Your Password

You requested to reset your password for your Personal Cloud Server.

METHOD 1: Click the reset link below:
%s

METHOD 2: Enter this 6-digit code on the password reset page:
%s

⏰ This code expires in 15 minutes and can only be used once.

🛡️ Security Notice:
If you didn't request a password reset, please ignore this email. Your password will remain unchanged.

---
Powered by Yundera Personal Cloud
This is an automated email, please do not reply.
`, link, code)

	return EmailContent{
		Subject:     subject,
		HTMLBody:    html,
		TextBody:    text,
		Attachments: []InlineAttachment{YunderaLogo()},
	}
}

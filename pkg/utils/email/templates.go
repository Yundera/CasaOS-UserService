package email

import "fmt"

// GenerateMagicLinkEmail generates the email content for magic link login
// Returns subject, HTML body, and plain text body
func GenerateMagicLinkEmail(code, link string) (subject, html, text string) {
	subject = "Your Magic Link to Sign In"

	html = fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .code-box { background: white; border: 2px dashed #667eea; padding: 20px; text-align: center; margin: 20px 0; border-radius: 8px; }
        .code { font-size: 32px; font-weight: bold; letter-spacing: 8px; color: #667eea; font-family: 'Courier New', monospace; }
        .button { display: inline-block; padding: 15px 30px; background: #667eea; color: white; text-decoration: none; border-radius: 8px; margin: 20px 0; font-weight: bold; }
        .button:hover { background: #764ba2; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 14px; }
        .divider { margin: 30px 0; text-align: center; color: #999; }
        .warning { background: #fff3cd; border-left: 4px solid #ffc107; padding: 15px; margin: 20px 0; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔐 Sign In to Your Personal Cloud</h1>
        </div>
        <div class="content">
            <p>Hello!</p>
            <p>You requested a magic link to sign in to your Personal Cloud Server. Use one of the following methods:</p>

            <h3>Method 1: Click the Magic Link</h3>
            <div style="text-align: center;">
                <a href="%s" class="button">✨ Sign In with Magic Link</a>
            </div>

            <div class="divider">— OR —</div>

            <h3>Method 2: Enter the Code</h3>
            <p>Enter this 6-digit code on the login page:</p>
            <div class="code-box">
                <div class="code">%s</div>
            </div>

            <div class="warning">
                <strong>⏰ This code expires in 10 minutes</strong><br>
                For your security, this code can only be used once.
            </div>

            <p>If you didn't request this magic link, you can safely ignore this email.</p>
        </div>
        <div class="footer">
            <p>Powered by Yundera Personal Cloud</p>
            <p style="font-size: 12px; color: #999;">This is an automated email, please do not reply.</p>
        </div>
    </div>
</body>
</html>
`, link, code)

	text = fmt.Sprintf(`
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

	return subject, html, text
}

// GeneratePasswordResetEmail generates the email content for password reset
// Returns subject, HTML body, and plain text body
func GeneratePasswordResetEmail(code, link string) (subject, html, text string) {
	subject = "Reset Your Password"

	html = fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #f093fb 0%%, #f5576c 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .code-box { background: white; border: 2px dashed #f5576c; padding: 20px; text-align: center; margin: 20px 0; border-radius: 8px; }
        .code { font-size: 32px; font-weight: bold; letter-spacing: 8px; color: #f5576c; font-family: 'Courier New', monospace; }
        .button { display: inline-block; padding: 15px 30px; background: #f5576c; color: white; text-decoration: none; border-radius: 8px; margin: 20px 0; font-weight: bold; }
        .button:hover { background: #f093fb; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 14px; }
        .divider { margin: 30px 0; text-align: center; color: #999; }
        .warning { background: #fff3cd; border-left: 4px solid #ffc107; padding: 15px; margin: 20px 0; border-radius: 4px; }
        .security { background: #e7f3ff; border-left: 4px solid #2196F3; padding: 15px; margin: 20px 0; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔑 Reset Your Password</h1>
        </div>
        <div class="content">
            <p>Hello!</p>
            <p>You requested to reset your password for your Personal Cloud Server. Use one of the following methods:</p>

            <h3>Method 1: Click the Reset Link</h3>
            <div style="text-align: center;">
                <a href="%s" class="button">🔓 Reset Password</a>
            </div>

            <div class="divider">— OR —</div>

            <h3>Method 2: Enter the Code</h3>
            <p>Enter this 6-digit code on the password reset page:</p>
            <div class="code-box">
                <div class="code">%s</div>
            </div>

            <div class="warning">
                <strong>⏰ This code expires in 15 minutes</strong><br>
                For your security, this code can only be used once.
            </div>

            <div class="security">
                <strong>🛡️ Security Notice</strong><br>
                If you didn't request a password reset, please ignore this email. Your password will remain unchanged.
            </div>
        </div>
        <div class="footer">
            <p>Powered by Yundera Personal Cloud</p>
            <p style="font-size: 12px; color: #999;">This is an automated email, please do not reply.</p>
        </div>
    </div>
</body>
</html>
`, link, code)

	text = fmt.Sprintf(`
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

	return subject, html, text
}

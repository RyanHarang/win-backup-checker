# Email Notifications Setup Guide

This guide covers email configuration for the most common email providers.

---

## Gmail

### Requirements

-   Gmail account with 2-Factor Authentication enabled
-   App Password (regular password will not work)

### Setup Steps

1. **Enable 2-Factor Authentication**

    - Go to [Google Account Security](https://myaccount.google.com/security)
    - Enable 2-Step Verification if not already enabled

2. **Generate App Password**

    - Visit [App Passwords](https://myaccount.google.com/apppasswords)
    - Select "Mail" and "Other (Custom name)"
    - Name it "Backup Checker" or similar
    - Google will generate a 16-character password
    - Copy this password (spaces are optional - they work either way)

3. **Configure**
    ```json
    {
        "email": {
            "enabled": true,
            "smtp_host": "smtp.gmail.com",
            "smtp_port": 587,
            "from": "your-email@gmail.com",
            "to": ["recipient@example.com"],
            "username": "your-email@gmail.com",
            "password": "abcd efgh ijkl mnop",
            "send_on_success": false,
            "send_on_warnings": false,
            "send_on_errors": true,
            "subject_prefix": "[Backup Alert]"
        }
    }
    ```

**Note:** If you don't see the App Passwords option, ensure 2-Factor Authentication is fully enabled and wait a few minutes.

---

## Outlook / Office 365 / Hotmail

### Requirements

-   Microsoft account
-   App Password (if 2FA is enabled)

### Setup Steps

1. **Check if 2FA is enabled**

    - Go to [Microsoft Account Security](https://account.microsoft.com/security)

2. **If 2FA is enabled:**

    - Visit [App Passwords](https://account.microsoft.com/security)
    - Click "Create a new app password"
    - Copy the generated password

3. **If 2FA is NOT enabled:**

    - You can use your regular password
    - However, Microsoft may still require an app password for security

4. **Configure**
    ```json
    {
        "email": {
            "enabled": true,
            "smtp_host": "smtp.office365.com",
            "smtp_port": 587,
            "from": "your-email@outlook.com",
            "to": ["recipient@example.com"],
            "username": "your-email@outlook.com",
            "password": "your-password-or-app-password",
            "send_on_success": false,
            "send_on_warnings": false,
            "send_on_errors": true,
            "subject_prefix": "[Backup Alert]"
        }
    }
    ```

**Note:** For custom domains using Office 365, use the same SMTP settings.

---

## Yahoo Mail

### Requirements

-   Yahoo account with 2-Factor Authentication enabled
-   App Password (regular password will not work)

### Setup Steps

1. **Enable 2-Factor Authentication**

    - Go to [Yahoo Account Security](https://login.yahoo.com/account/security)
    - Enable Two-Step Verification

2. **Generate App Password**

    - Visit [App Passwords](https://login.yahoo.com/account/security/app-passwords)
    - Select "Other App" and name it "Backup Checker"
    - Click "Generate"
    - Copy the generated password

3. **Configure**
    ```json
    {
        "email": {
            "enabled": true,
            "smtp_host": "smtp.mail.yahoo.com",
            "smtp_port": 587,
            "from": "your-email@yahoo.com",
            "to": ["recipient@example.com"],
            "username": "your-email@yahoo.com",
            "password": "your-app-password",
            "send_on_success": false,
            "send_on_warnings": false,
            "send_on_errors": true,
            "subject_prefix": "[Backup Alert]"
        }
    }
    ```

---

## ProtonMail

### Requirements

-   ProtonMail account (any plan)
-   ProtonMail Bridge application (free for all users)

### Setup Steps

1. **Install ProtonMail Bridge**

    - Download from [ProtonMail Bridge](https://proton.me/mail/bridge)
    - Install and launch the application
    - Sign in with your ProtonMail credentials

2. **Add Your Account to Bridge**

    - In Bridge, click "Add Account"
    - Sign in with your ProtonMail account
    - Bridge will generate SMTP credentials for you

3. **Get SMTP Credentials**

    - In Bridge, click on your account
    - Click "Mailbox configuration"
    - Note the SMTP username and password (these are NOT your ProtonMail login credentials)

4. **Configure**
    ```json
    {
        "email": {
            "enabled": true,
            "smtp_host": "127.0.0.1",
            "smtp_port": 1025,
            "from": "your-email@protonmail.com",
            "to": ["recipient@example.com"],
            "username": "your-bridge-username",
            "password": "your-bridge-password",
            "send_on_success": false,
            "send_on_warnings": false,
            "send_on_errors": true,
            "subject_prefix": "[Backup Alert]"
        }
    }
    ```

**Important:** ProtonMail Bridge must be running for emails to send.

---

## iCloud Mail

### Requirements

-   Apple ID with iCloud Mail
-   App-Specific Password

### Setup Steps

1. **Enable 2-Factor Authentication**

    - Go to [Apple ID Account](https://appleid.apple.com/)
    - Enable Two-Factor Authentication if not already enabled

2. **Generate App-Specific Password**

    - Visit [Apple ID Security](https://appleid.apple.com/account/manage)
    - Under "App-Specific Passwords", click "Generate Password"
    - Name it "Backup Checker"
    - Copy the generated password

3. **Configure**
    ```json
    {
        "email": {
            "enabled": true,
            "smtp_host": "smtp.mail.me.com",
            "smtp_port": 587,
            "from": "your-email@icloud.com",
            "to": ["recipient@example.com"],
            "username": "your-email@icloud.com",
            "password": "your-app-specific-password",
            "send_on_success": false,
            "send_on_warnings": false,
            "send_on_errors": true,
            "subject_prefix": "[Backup Alert]"
        }
    }
    ```

---

## Custom SMTP Server

If you're using a custom email provider, hosting provider, or self-hosted mail server:

### What You'll Need

-   SMTP server hostname (e.g., `mail.yourdomain.com`)
-   SMTP port (usually 587 for STARTTLS)
-   Your email credentials

### Configure

```json
{
    "email": {
        "enabled": true,
        "smtp_host": "mail.yourdomain.com",
        "smtp_port": 587,
        "from": "backups@yourdomain.com",
        "to": ["admin@yourdomain.com"],
        "username": "backups@yourdomain.com",
        "password": "your-password",
        "send_on_success": false,
        "send_on_warnings": false,
        "send_on_errors": true,
        "subject_prefix": "[Backup Alert]"
    }
}
```

**Common SMTP Ports:**

-   **Port 587** - STARTTLS (recommended, used by this tool) ✅
-   **Port 465** - Implicit SSL/TLS (not currently supported) ❌
-   **Port 25** - Unencrypted (not recommended, often blocked) ❌

---

## Testing Your Configuration

Before relying on email notifications, test your setup:

```bash
go run ./cmd/test-email/
```

This sends a mock error report to verify your SMTP configuration works correctly.

---

## Configuration Options Explained

| Option             | Description                           | Default            |
| ------------------ | ------------------------------------- | ------------------ |
| `enabled`          | Enable/disable email notifications    | `false`            |
| `smtp_host`        | SMTP server address                   | Required           |
| `smtp_port`        | SMTP server port                      | Required           |
| `from`             | Sender email address                  | Required           |
| `to`               | List of recipient email addresses     | Required           |
| `username`         | SMTP authentication username          | Required           |
| `password`         | SMTP authentication password          | Required           |
| `send_on_success`  | Send email when all backups are valid | `false`            |
| `send_on_warnings` | Send email when warnings are found    | `false`            |
| `send_on_errors`   | Send email when errors are found      | `true`             |
| `subject_prefix`   | Custom prefix for email subjects      | `"[Backup Alert]"` |

---

## Troubleshooting

### Authentication Failed

-   Verify you're using an **app password**, not your regular password (for Gmail, Yahoo, iCloud)
-   Ensure 2-Factor Authentication is enabled for providers that require it
-   Check that username is your full email address
-   For ProtonMail, ensure Bridge is running and using Bridge credentials

### Connection Refused / Timeout

-   Verify SMTP host and port are correct
-   Check that your firewall isn't blocking outgoing connections on port 587
-   Some ISPs block SMTP ports - try using a VPN
-   For ProtonMail, ensure Bridge application is running

### Emails Not Arriving

-   ✅ Check spam/junk folders
-   ✅ Verify recipient email addresses are correct in config
-   ✅ Ensure the sender email is valid for the SMTP server you're using
-   ✅ Check email provider's sending limits

### TLS Handshake Failed

-   ✅ This tool uses STARTTLS on port 587
-   ✅ Ensure your SMTP server supports STARTTLS
-   ✅ Port 465 (implicit TLS) is not currently supported

### "Failed to send email" Error

-   ✅ Run the test utility: `go run ./cmd/test-email/`
-   ✅ Check the error message for specific details
-   ✅ Verify all required config fields are filled
-   ✅ Ensure email config is valid JSON (no trailing commas, proper quotes)

---

## 💡 Best Practices

1. **Use App Passwords**: Always use app-specific passwords instead of your main password
2. **Secure Storage**: Keep your `config.json` file secure with appropriate file permissions
3. **Test First**: Always test email configuration before relying on automated alerts
4. **Monitor Logs**: Check `logs.json` to verify scans are running as expected
5. **Multiple Recipients**: Add multiple email addresses to the `to` array for redundancy
6. **Appropriate Triggers**: Consider enabling `send_on_warnings` if you want proactive notifications

---

## 🔒 Security Notes

-   **Never commit** your `config.json` with passwords to version control
-   Add `config.json` to your `.gitignore` file
-   Use environment variables for passwords in production environments
-   Regularly rotate app passwords for security
-   Keep ProtonMail Bridge updated if using ProtonMail

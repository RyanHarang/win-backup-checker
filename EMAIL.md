# Email Notifications Setup Guide

This guide covers how to configure email notifications for backup alerts.

---

## General Setup for Any Email Provider

Most modern email providers require **App Passwords** instead of your regular account password for security reasons.

### Common Steps:

1. **Enable 2-Factor Authentication** on your email account (usually required for app passwords)
2. **Generate an App Password** (sometimes called "App-Specific Password")
    - Look for this in your account security settings
    - Common names: "App Passwords", "App-Specific Passwords", "Security & Privacy"
3. **Use the app password** (not your regular password) in the config

---

## Example: Gmail Setup

### Requirements

-   Gmail account with 2-Factor Authentication enabled
-   App Password

### Steps

1. **Enable 2-Factor Authentication**

    - Go to [Google Account Security](https://myaccount.google.com/security)
    - Enable 2-Step Verification if not already enabled

2. **Generate App Password**

    - Visit [App Passwords](https://myaccount.google.com/apppasswords)
    - Name it "Backup Checker" or similar and create

---

## Common SMTP Settings

Here are SMTP settings for popular email providers:

| Provider           | SMTP Host             | Port | App Password Required? |
| ------------------ | --------------------- | ---- | ---------------------- |
| Gmail              | `smtp.gmail.com`      | 587  | Yes (with 2FA)         |
| Outlook/Office 365 | `smtp.office365.com`  | 587  | Yes (with 2FA)         |
| Yahoo Mail         | `smtp.mail.yahoo.com` | 587  | Yes (with 2FA)         |
| iCloud Mail        | `smtp.mail.me.com`    | 587  | Yes (with 2FA)         |
| Custom/Self-hosted | `mail.yourdomain.com` | 587  | Depends on server      |

**Note:** This tool uses port 587 with STARTTLS encryption. Port 465 (implicit SSL/TLS) is not currently supported.

---

### Configuration Options

| Option             | Description                                       | Default            |
| ------------------ | ------------------------------------------------- | ------------------ |
| `enabled`          | Enable/disable email notifications                | `false`            |
| `smtp_host`        | SMTP server address                               | Required           |
| `smtp_port`        | SMTP server port (use 587)                        | Required           |
| `from`             | Sender email address                              | Required           |
| `to`               | Array of recipient email addresses                | Required           |
| `username`         | SMTP authentication username (usually your email) | Required           |
| `password`         | SMTP authentication password (use app password)   | Required           |
| `send_on_success`  | Send email when all backups are valid             | `false`            |
| `send_on_warnings` | Send email when warnings are found                | `false`            |
| `send_on_errors`   | Send email when errors are found                  | `true`             |
| `subject_prefix`   | Custom prefix for email subjects                  | `"[Backup Alert]"` |

---

## Testing Your Configuration

Before relying on email notifications, test your setup:

```bash
go run ./cmd/test-email/
```

This sends a mock error report to verify your SMTP configuration works correctly.

**Expected output:**

```
Email Configuration:
  SMTP Host: smtp.gmail.com:587
  From: your-email@gmail.com
  To: [recipient@example.com]
  ...

Sending test email with mock error data...
Test email sent successfully!

Check your inbox at: recipient@example.com
```

---

## Disabling Email Notifications

You can disable emails in three ways:

1. **In config file:** Set `"enabled": false` in `email.config.json`
2. **Command line flag:** Use `--no-email` flag when running
3. **Remove config file:** Delete `email.config.json` (emails will be skipped)

---

## Troubleshooting

### Authentication Failed

-   Verify you're using an **app password**, not your regular password
-   Ensure 2-Factor Authentication is enabled (required by most providers)
-   Check that username is your full email address
-   Double-check you copied the app password correctly

### Connection Refused / Timeout

-   Verify SMTP host and port are correct
-   Check that your firewall isn't blocking outgoing connections on port 587
-   Some ISPs block SMTP ports - try using a VPN
-   Ensure you're using port 587 (not 465 or 25)

### Emails Not Arriving

-   Check spam/junk folders
-   Verify recipient email addresses are correct in config
-   Ensure the sender email is valid for the SMTP server you're using
-   Check email provider's sending limits (some have hourly/daily limits)

### "Failed to send email" Error

-   Run the test utility: `go run ./cmd/test-email/`
-   Check the error message for specific details
-   Verify all required config fields are filled
-   Ensure email.config.json is valid JSON (no trailing commas, proper quotes)
-   Check that the config file is in the correct location: `configs/email.config.json`

### "Email is not enabled" Error

-   Ensure `email.config.json` exists in the `configs/` directory
-   Verify `"enabled": true` is set in the config file
-   Check file permissions allow reading the config file

---

## Advanced Configuration

### Multiple Recipients

Add multiple email addresses to the `to` array:

```json
{
    "to": ["admin@example.com", "backup-team@example.com", "alerts@example.com"]
}
```

### Custom Subject Prefix

Change the email subject prefix to match your organization:

```json
{
    "subject_prefix": "[MyCompany Backups]"
}
```

### Notification Triggers

Control when emails are sent:

```json
{
    "send_on_success": true, // Get confirmation emails when everything is fine
    "send_on_warnings": true, // Be notified of potential issues
    "send_on_errors": true // Critical alerts for failures
}
```

---

## Getting Help

If you're having trouble with email configuration:

1. Run the test utility first: `go run ./cmd/test-email/`
2. Check the error message carefully
3. Consult your email provider's documentation for:
    - How to enable 2-Factor Authentication
    - How to generate app passwords
    - SMTP server settings
4. Verify your SMTP settings match the table above

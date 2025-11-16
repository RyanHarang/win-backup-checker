#!/bin/bash

# Windows Backup Checker - Setup Script
# This script creates the necessary configuration files

set -e

echo "========================================="
echo "Windows Backup Checker - Setup"
echo "========================================="
echo ""

# Create config.json
CONFIG_FILE="configs/config.json"
if [ -f "$CONFIG_FILE" ]; then
    echo ""
    read -p "config.json already exists. Overwrite? (y/N): " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "  Skipping config.json"
    else
        rm "$CONFIG_FILE"
        CREATE_CONFIG=true
    fi
else
    CREATE_CONFIG=true
fi

if [ "$CREATE_CONFIG" = true ]; then
    cat > "$CONFIG_FILE" << 'EOF'
{
  "backup_paths": ["/path/to/your/backup/directory"],
  "check_hash": false,
  "deep_validation": true,
  "max_zip_sample_size": 104857600,
  "required_catalog_extensions": [".wbcat", ".cat"],
  "min_backup_age": "1h",
  "max_backup_age": "90d"
}
EOF
    echo "✓ Created config.json"
fi

# Create email.config.json
EMAIL_CONFIG_FILE="configs/email.config.json"
if [ -f "$EMAIL_CONFIG_FILE" ]; then
    echo ""
    read -p "email.config.json already exists. Overwrite? (y/N): " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "  Skipping email.config.json"
    else
        rm "$EMAIL_CONFIG_FILE"
        CREATE_EMAIL=true
    fi
else
    CREATE_EMAIL=true
fi

if [ "$CREATE_EMAIL" = true ]; then
    cat > "$EMAIL_CONFIG_FILE" << 'EOF'
{
  "enabled": false,
  "smtp_host": "smtp.gmail.com",
  "smtp_port": 587,
  "from": "your-email@gmail.com",
  "to": ["recipient@example.com"],
  "username": "your-email@gmail.com",
  "password": "your-app-password",
  "send_on_success": false,
  "send_on_warnings": true,
  "send_on_errors": true,
  "subject_prefix": "[Backup Alert]"
}
EOF
    echo "Created email.config.json (disabled by default)"
fi

echo ""
echo "========================================="
echo "Setup Complete!"
echo "========================================="
echo ""
echo "Next steps:"
echo ""
echo "1. Edit configs/config.json:"
echo "   - Update 'backup_paths' with your actual backup directory paths"
echo ""
echo "2. (Optional) Configure email notifications:"
echo "   - Edit configs/email.config.json"
echo "   - Set 'enabled': true"
echo "   - Add your SMTP settings (see EMAIL.md for help)"
echo "   - Test with: go run ./cmd/test-email/"
echo ""
echo "3. Run the backup checker:"
echo "   go run ./cmd/checker/"
echo ""
echo "For more information, see SETUP.md"
echo ""
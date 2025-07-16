# PancakesGit 🥞

A privacy-focused, encrypted Git service that prioritizes user privacy and data security while providing a familiar Git hosting experience.

## Features

### 🔒 Privacy & Security
- **End-to-end encryption** for repositories, database, and logs
- **Zero-knowledge architecture** - we can't see your data
- **No tracking or analytics** by default
- **Tor-friendly** with .onion support
- **Strong password policies** and 2FA support
- **Minimal metadata collection**

### 🛡️ Data Protection
- **AES-256-GCM encryption** for all sensitive data
- **Argon2id password hashing**
- **Secure session management**
- **CSRF protection**
- **Rate limiting** and brute force protection

### 🚀 Git Features
- **Full Git protocol support** (HTTP/HTTPS and SSH)
- **Repository management** with web interface
- **Issue tracking** and pull requests
- **Webhooks** and API access
- **Organizations and teams**
- **Release management**

### 🌐 Deployment Options
- **Docker deployment** (recommended)
- **Host machine installation**
- **Cloudflare tunnel integration**
- **Automatic SSL/TLS** with Let's Encrypt

## Quick Start

### Prerequisites

- Linux or macOS system
- Docker and Docker Compose (for Docker deployment)
- Go 1.21+ (for host deployment)
- Git
- curl

### Installation

1. **Clone or download the setup script:**
   ```bash
   curl -O https://raw.githubusercontent.com/ilikepancakes-ink/pancakesgit/main/setup.sh
   chmod +x setup.sh
   ```

2. **Run the setup script:**
   ```bash
   ./setup.sh
   ```

3. **Follow the interactive prompts:**
   - Choose deployment method (Docker or host machine)
   - Create admin account
   - Configure Cloudflare tunnel (optional)

4. **Start your PancakesGit instance:**
   ```bash
   # For Docker deployment
   cd ~/pancakesgit && docker-compose up -d
   
   # For host deployment
   sudo systemctl start pancakesgit
   ```

5. **Access your instance:**
   - Local: http://localhost:3000
   - With Cloudflare tunnel: https://your-domain.com

## Configuration

### Environment Variables

Key environment variables for configuration:

```bash
# Server
PANCAKESGIT_DOMAIN=your-domain.com
PANCAKESGIT_ROOT_URL=https://your-domain.com
PANCAKESGIT_SECRET_KEY=your-secret-key
PANCAKESGIT_ENCRYPTION_KEY=your-encryption-key

# Database
PANCAKESGIT_DB_TYPE=postgres
PANCAKESGIT_DB_HOST=localhost
PANCAKESGIT_DB_PASSWORD=your-db-password

# Security
PANCAKESGIT_ENABLE_2FA=true
PANCAKESGIT_HTTPS_ONLY=true
```

### Configuration File

PancakesGit uses an INI-style configuration file (`app.ini`). Key sections:

- `[server]` - Server settings
- `[database]` - Database configuration
- `[security]` - Security policies
- `[privacy]` - Privacy settings
- `[encryption]` - Encryption configuration
- `[git]` - Git protocol settings

## Cloudflare Tunnel Setup

The setup script can automatically configure Cloudflare tunnels:

1. **Login to Cloudflare** (if not already logged in)
2. **Provide your domain** when prompted
3. **Script creates tunnel** named "pancakesgit"
4. **Custom config file** saved as `pancakes.yml`
5. **DNS record** automatically created

To start the tunnel:
```bash
cloudflared tunnel --config ~/.cloudflared/pancakes.yml run
```

## Security Best Practices

### 🔐 Encryption Keys
- **Generate strong keys** using the provided tools
- **Store keys securely** and create backups
- **Rotate keys regularly** (every 6-12 months)
- **Never share keys** in plain text

### 🛡️ Access Control
- **Use strong passwords** (minimum 12 characters)
- **Enable 2FA** for all accounts
- **Limit admin access** to necessary users only
- **Regular security audits**

### 🌐 Network Security
- **Use HTTPS only** in production
- **Configure firewall** properly
- **Regular security updates**
- **Monitor access logs**

## API Usage

PancakesGit provides a RESTful API for automation:

```bash
# Get user repositories
curl -H "Authorization: token YOUR_TOKEN" \
     https://your-domain.com/api/v1/user/repos

# Create repository
curl -X POST \
     -H "Authorization: token YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"name":"my-repo","private":true}' \
     https://your-domain.com/api/v1/user/repos
```

## Backup and Recovery

### Database Backup
```bash
# PostgreSQL
pg_dump pancakesgit > backup.sql

# SQLite
cp /data/pancakesgit.db backup.db
```

### Repository Backup
```bash
# Backup encrypted repositories
tar -czf repos-backup.tar.gz /data/repositories/
```

### Configuration Backup
```bash
# Backup configuration and keys
cp .env config-backup.env
cp config/app.ini app-backup.ini
```

## Troubleshooting

### Common Issues

**Service won't start:**
- Check logs: `docker-compose logs pancakesgit`
- Verify configuration file syntax
- Ensure database is accessible

**Can't login:**
- Verify admin account was created
- Check password requirements
- Review security logs

**Git operations fail:**
- Verify SSH keys are properly configured
- Check repository permissions
- Ensure Git service is running

### Getting Help

- **Documentation:** Check the `/docs` directory
- **Logs:** Review application logs for errors
- **Community:** Join our privacy-focused community
- **Issues:** Report bugs on our Git instance

## Contributing

We welcome contributions! Please:

1. **Fork the repository**
2. **Create a feature branch**
3. **Make your changes**
4. **Add tests** for new functionality
5. **Submit a pull request**

### Development Setup

```bash
# Clone repository
git clone https://github.com/ilikepancakes-ink/pancakesgit.git
cd pancakesgit

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o pancakesgit ./cmd/pancakesgit
```

## License

PancakesGit is released under the MIT License. See [LICENSE](LICENSE) for details.

## Privacy Policy

We are committed to protecting your privacy:

- **No data collection** beyond what's necessary for operation
- **No third-party tracking**
- **Encrypted storage** of all user data
- **Minimal logging** with automatic cleanup
- **User control** over all personal data

---

**Built with privacy in mind. Your code, your data, your control.** 🥞🔒

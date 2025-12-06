# API Key Rotation Guide

This document describes how to manage and rotate API keys for the Straye Relation API.

## Overview

The API supports API key authentication for system-to-system integration and automated processes. API keys provide full system-level access and should be treated as highly sensitive credentials.

## Configuration

API keys are configured via environment variable:

```bash
ADMIN_API_KEY=your-secure-api-key-here
```

Or in `config.json`:

```json
{
  "apiKey": {
    "value": "your-secure-api-key-here"
  }
}
```

**Note:** Environment variables take precedence over config file values.

## Security Requirements

### Key Generation

Generate secure API keys using cryptographically secure random generators:

```bash
# Using OpenSSL (recommended)
openssl rand -base64 32

# Using Python
python3 -c "import secrets; print(secrets.token_urlsafe(32))"

# Using Go
go run -e 'import "crypto/rand"; import "encoding/base64"; b := make([]byte, 32); rand.Read(b); println(base64.URLEncoding.EncodeToString(b))'
```

### Key Requirements

- Minimum 32 characters
- Use alphanumeric characters and URL-safe symbols
- Never commit API keys to version control
- Store keys in secure secret management systems (Azure Key Vault, AWS Secrets Manager, etc.)

## Rotation Procedure

### 1. Pre-Rotation Preparation

1. **Identify all systems using the current API key**
   - Check application logs for API key usage patterns
   - Document all integrations and their owners

2. **Generate new API key**
   ```bash
   NEW_API_KEY=$(openssl rand -base64 32)
   echo "New API Key: $NEW_API_KEY"
   ```

3. **Store new key in secret management**
   - Add the new key to Azure Key Vault / AWS Secrets Manager
   - Keep the old key available during transition

### 2. Rolling Deployment (Zero-Downtime)

For zero-downtime rotation, the API supports accepting multiple API keys temporarily:

#### Option A: Environment Variable Update

1. Deploy the new API key to the API server:
   ```bash
   # Update environment variable
   export ADMIN_API_KEY=new-api-key-value

   # Restart/redeploy the API service
   ```

2. Update all client systems to use the new key

3. Monitor logs to ensure no requests use the old key

#### Option B: Parallel Key Support (Custom Implementation)

If you need to support both old and new keys simultaneously during rotation:

1. Modify `internal/auth/middleware.go` to accept multiple keys:
   ```go
   // Example: Support comma-separated keys in ADMIN_API_KEY
   apiKeys := strings.Split(cfg.ApiKey.Value, ",")
   ```

2. Deploy with both keys: `ADMIN_API_KEY=old-key,new-key`

3. Update clients to new key

4. Remove old key: `ADMIN_API_KEY=new-key`

### 3. Post-Rotation Verification

1. **Verify new key works**
   ```bash
   curl -H "x-api-key: $NEW_API_KEY" https://api.example.com/health
   ```

2. **Monitor authentication logs**
   ```bash
   # Check for any failed API key attempts
   grep "invalid API key" /var/log/api/app.log
   ```

3. **Update documentation**
   - Update internal documentation with rotation date
   - Notify relevant teams of the change

### 4. Cleanup

1. Remove old API key from all secret stores
2. Update any backup/disaster recovery configurations
3. Document the rotation in your security audit log

## Rotation Schedule

Recommended rotation frequency:

| Environment | Rotation Frequency |
|-------------|-------------------|
| Production  | Every 90 days     |
| Staging     | Every 30 days     |
| Development | As needed         |

## Emergency Rotation

In case of suspected key compromise:

1. **Immediately generate and deploy new key**
   ```bash
   NEW_KEY=$(openssl rand -base64 32)
   # Deploy immediately to production
   ```

2. **Revoke old key** (no grace period)

3. **Audit access logs**
   - Review all API key usage in the past 30 days
   - Check for unauthorized access patterns

4. **Notify stakeholders**
   - Security team
   - All integration owners
   - Management (if data breach suspected)

5. **Post-incident review**
   - Document how compromise occurred
   - Implement preventive measures

## Monitoring and Alerts

Set up monitoring for API key usage:

### Recommended Alerts

1. **Failed API key attempts**
   - Alert threshold: > 5 failures in 5 minutes
   - Action: Investigate potential brute force attack

2. **Unusual usage patterns**
   - Alert on API key usage from new IP addresses
   - Alert on usage outside normal hours

3. **Key age monitoring**
   - Alert when key is > 80 days old
   - Critical alert at 90 days

### Log Analysis Queries

```bash
# Find all API key authentication attempts (Zap JSON logs)
jq 'select(.msg | contains("API key"))' /var/log/api/app.log

# Count API key requests by hour
jq -r 'select(.msg == "API key authenticated") | .ts' /var/log/api/app.log | \
  cut -d'T' -f2 | cut -d':' -f1 | sort | uniq -c
```

## Compliance Notes

- API key rotation logs should be retained for audit purposes
- Document rotation dates and responsible parties
- Include API key management in security reviews
- Ensure rotation procedures are tested in non-production environments first

## Related Documentation

- [Authentication Guide](./AUTHENTICATION.md)
- [Security Headers Configuration](./SECURITY.md)
- [API Specification](./API_SPECIFICATION.md)

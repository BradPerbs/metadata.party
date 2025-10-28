# Security Policy

## Supported Versions

We currently support the following versions with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |

## Reporting a Vulnerability

We take the security of metadata.party seriously. If you believe you have found a security vulnerability, please report it to us responsibly.

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please send an email to [bradperbs@gmail.com] with:

* A description of the vulnerability
* Steps to reproduce the issue
* Possible impact
* Suggested fix (if any)

You should receive a response within 48 hours. If the issue is confirmed, we will:

1. Work on a fix
2. Release a security update
3. Publicly disclose the vulnerability (with credit to you, if desired)

## Security Considerations

### Security Features

* **SSRF Protection**: Built-in protection blocks requests to:
  - Localhost and loopback addresses (127.0.0.0/8, ::1)
  - Private networks (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16)
  - Link-local addresses (169.254.0.0/16 - AWS metadata service)
  - Multicast and reserved IP ranges
  - The API resolves hostnames and checks IPs before making requests

### Known Limitations

* **Rate Limiting**: No built-in rate limiting - implement this at the infrastructure level (e.g., nginx, API gateway)
* **Memory**: Large HTML documents are limited to 10MB, but multiple concurrent requests could still cause memory issues
* **Timeout**: Default timeout is 30 seconds per request
* **Authentication**: No built-in authentication - add at reverse proxy or application level as needed

### Recommended Production Setup

1. **Use a reverse proxy** (nginx, Caddy) with rate limiting
2. **Set ALLOWED_ORIGIN** environment variable to restrict CORS
3. **Monitor resource usage** and set appropriate container limits
4. **Use HTTPS** for all communications
5. **Implement request logging** and monitoring
6. **Consider adding authentication** for your use case
7. **Keep dependencies updated** regularly

### Environment Variables

* `PORT` - Server port (default: 8080)
* `ALLOWED_ORIGIN` - CORS allowed origin (default: *)

### Docker Security

The Docker image:
* Runs as a non-root user (planned)
* Uses minimal Alpine Linux base
* Includes only necessary dependencies
* Has health checks enabled

## Best Practices for Users

1. Don't expose the API directly to the internet without authentication
2. Use environment variables for configuration, never hardcode
3. Keep dependencies updated
4. Monitor logs for suspicious activity
5. Implement rate limiting at the infrastructure level
6. Use container resource limits in production

## Acknowledgments

We appreciate the security research community's efforts in keeping open-source software secure.


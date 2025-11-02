# Metadata.party 🎉

A lightweight, production-ready Go API for extracting metadata from URLs. Perfect for link previews, social media cards, and content analysis.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/doc/devel/release.html)

## Features

- ✨ Extract page title, description, and images
- 🏷️ Parse Open Graph and Twitter Card metadata
- 🎨 Fetch favicons automatically
- ⚡ Fast extraction with duration metrics
- 🔢 **Batch processing: extract up to 5 URLs concurrently**
- 🔒 Production-ready with security best practices
- 🐳 Docker support with health checks
- 🌐 CORS support for browser requests
- 📊 Request logging middleware
- 🛡️ Graceful shutdown handling
- 🚦 Built-in rate limiting (15 req/min per IP, unlimited with API key from https://metadata.party/)

## Installation

1. Make sure you have Go 1.21+ installed
2. Clone this repository
3. Install dependencies:

```bash
go mod download
```

## Running the Server

```bash
go run main.go
```

The server will start on `http://localhost:8759`

## API Endpoints

### POST /extract

Extract metadata from 1-5 URLs. The endpoint automatically detects single vs. batch requests and returns the appropriate format.

#### Single URL Request

**Request:**
```bash
curl -X POST http://localhost:8759/extract \
  -H "Content-Type: application/json" \
  -d '{"url": "https://zapier.com/blog/best-crm-app/"}'
```

**Response:**
```json
{
  "title": "The 12 best CRM software in 2025",
  "description": "We put dozens of Salesforce alternatives through the wringer and came up with the 11 best CRM apps on the market.",
  "images": [
    "https://images.ctfassets.net/lzny33ho1g45/6HrRibvXMoNeGMPq3CIg8S/4ffcf4a0df0914f3dfc09a4914f89be7/best_apps_37.jpg"
  ],
  "sitename": ["Zapier"],
  "favicon": "https://cdn.zapier.com/zapier/images/favicon.ico",
  "duration": 746,
  "domain": "zapier.com",
  "url": "https://zapier.com/blog/best-crm-app/"
}
```

#### Batch Request (2-5 URLs)

**Request:**
```bash
curl -X POST http://localhost:8759/extract \
  -H "Content-Type: application/json" \
  -d '{
    "urls": [
      "https://github.com",
      "https://zapier.com/blog/best-crm-app/",
      "https://example.com"
    ]
  }'
```

**Response:**
```json
{
  "results": [
    {
      "title": "GitHub: Let's build from here",
      "description": "GitHub is where over 100 million developers shape the future of software...",
      "images": ["https://github.githubassets.com/images/modules/site/social-cards/github-social.png"],
      "sitename": ["GitHub"],
      "favicon": "https://github.com/favicon.ico",
      "duration": 523,
      "domain": "github.com",
      "url": "https://github.com"
    },
    {
      "title": "The 12 best CRM software in 2025",
      "description": "We put dozens of Salesforce alternatives through the wringer...",
      "images": ["https://images.ctfassets.net/..."],
      "sitename": ["Zapier"],
      "favicon": "https://cdn.zapier.com/zapier/images/favicon.ico",
      "duration": 612,
      "domain": "zapier.com",
      "url": "https://zapier.com/blog/best-crm-app/"
    },
    {
      "title": "Example Domain",
      "description": "",
      "images": [],
      "sitename": [],
      "favicon": "https://example.com/favicon.ico",
      "duration": 234,
      "domain": "example.com",
      "url": "https://example.com"
    }
  ],
  "total": 3
}
```

**Notes:**
- Use `"url"` for single URL, `"urls"` for multiple URLs
- Maximum 5 URLs per request
- Multiple URLs are processed concurrently for speed
- If a URL fails in batch mode, it returns with an `error` field
- Results are returned in the same order as input

#### Rate Limiting

By default, the API limits requests to **15 per minute per IP address**. To bypass this limit, provide an API key:

**Using Header:**
```bash
curl -X POST http://localhost:8759/extract \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secret-api-key" \
  -d '{"url": "https://example.com"}'
```

**Using Query Parameter:**
```bash
curl -X POST "http://localhost:8759/extract?api_key=your-secret-api-key" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'
```

**Rate Limit Response:**
When rate limit is exceeded, you'll receive a `429 Too Many Requests` response:
```json
{
  "error": "Rate limit exceeded. Maximum 15 requests per minute. Please try again later or use an API key for unlimited access."
}
```

### GET /health

Health check endpoint.

**Request:**
```bash
curl http://localhost:8759/health
```

**Response:**
```json
{
  "status": "ok"
}
```

## Building for Production

```bash
# Build binary
go build -o metadata-api

# Run the binary
./metadata-api
```

## Docker Deployment

### Using Docker

```bash
# Build the image
docker build -t metadata-api .

# Run the container
docker run -p 8759:8759 \
  -e PORT=8759 \
  -e ALLOWED_ORIGIN=https://yourdomain.com \
  -e API_KEY=your-secret-api-key \
  --name metadata-api \
  metadata-api
```

### Using Docker Compose

```bash
# Start the service
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the service
docker-compose down
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8759` |
| `ALLOWED_ORIGIN` | CORS allowed origin | `*` |
| `API_KEY` | API key for unlimited requests (bypasses rate limit) | _(none)_ |

## Production Considerations

### Security

- ✅ **SSRF Protection**: Built-in IP filtering to prevent SSRF attacks
- ✅ **Rate Limiting**: Built-in rate limiting (15 req/min per IP)
- 🔐 **API Key**: Set `API_KEY` environment variable for unlimited access
- 🌍 **CORS**: Set `ALLOWED_ORIGIN` to your domain in production
- 🔒 **API Key Security**: Keep your API key secret and use HTTPS in production

### Performance

- 📦 **Body Size Limit**: Responses are limited to 10MB
- ⏱️ **Timeout**: 30 second timeout for fetching URLs
- 🔄 **Redirects**: Maximum 10 redirects allowed
- 💾 **Memory**: Use container limits in production

### Recommended Setup

```nginx
# Example nginx configuration with API key forwarding
server {
    listen 80;
    server_name api.yourdomain.com;

    location / {
        proxy_pass http://localhost:8759;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-API-Key $http_x_api_key;
    }
}
```

## Metadata Extracted

The API extracts the following metadata:

- **title**: Page title (from `<title>`, `og:title`, or `twitter:title`)
- **description**: Page description (from meta description, `og:description`, or `twitter:description`)
- **images**: Array of images (from `og:image` and `twitter:image`)
- **sitename**: Site name (from `og:site_name`)
- **favicon**: Site favicon (from `<link rel="icon">` or default `/favicon.ico`)
- **duration**: Time taken to extract metadata (in milliseconds)
- **domain**: Domain name of the URL
- **url**: Original URL requested

## Error Handling

The API returns appropriate HTTP status codes:

- `200 OK`: Successful metadata extraction
- `400 Bad Request`: Invalid request (missing URL, invalid JSON)
- `405 Method Not Allowed`: Wrong HTTP method
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Failed to fetch or parse URL

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## Security

For security concerns, please see [SECURITY.md](SECURITY.md).

## Deployment Examples

### Deploy to Fly.io

```bash
# Install flyctl
curl -L https://fly.io/install.sh | sh

# Create and deploy
fly launch
fly deploy
```

### Deploy to Railway

```bash
# Install Railway CLI
npm i -g @railway/cli

# Deploy
railway login
railway init
railway up
```

### Deploy to Google Cloud Run

```bash
# Build and push to Container Registry
gcloud builds submit --tag gcr.io/PROJECT_ID/metadata-api

# Deploy to Cloud Run
gcloud run deploy metadata-api \
  --image gcr.io/PROJECT_ID/metadata-api \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated
```

## Roadmap

- [ ] Add caching layer (Redis)
- [x] Implement rate limiting middleware
- [x] Add API key authentication
- [ ] Support for more metadata types (JSON-LD, microdata)
- [ ] WebSocket support for real-time extraction
- [ ] Metrics and monitoring endpoints (Prometheus)

## License

MIT License - feel free to use this in your projects! See [LICENSE](LICENSE) for details.

## Acknowledgments

Built with:
- [Go](https://golang.org/) - The Go Programming Language
- [golang.org/x/net/html](https://pkg.go.dev/golang.org/x/net/html) - HTML parsing

## Support

- 📫 Open an issue for bug reports or feature requests
- ⭐ Star this repo if you find it useful
- 🔄 Fork and submit PRs for contributions

---

Made with ❤️ for the open-source community


#!/bin/bash

# metadata.party - Example API calls
# Make sure the server is running before executing these examples

echo "🎉 metadata.party - Example API Calls"
echo "======================================"
echo ""

# Check if server is running
echo "1️⃣  Health Check"
curl -s http://localhost:8759/health | jq .
echo ""
echo ""

# Root endpoint
echo "2️⃣  API Info"
curl -s http://localhost:8759/ | jq .
echo ""
echo ""

# Extract metadata from a blog post
echo "3️⃣  Extract metadata from Zapier blog"
curl -s -X POST http://localhost:8759/extract \
  -H "Content-Type: application/json" \
  -d '{"url": "https://zapier.com/blog/best-crm-app/"}' | jq .
echo ""
echo "" 

# Extract metadata from a news site
echo "4️⃣  Extract metadata from a news site"
curl -s -X POST http://localhost:8759/extract \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.theverge.com"}' | jq .
echo ""
echo ""

# Extract metadata from GitHub
echo "5️⃣  Extract metadata from GitHub"
curl -s -X POST http://localhost:8759/extract \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com"}' | jq .
echo ""
echo ""

# Batch extract - multiple URLs at once
echo "6️⃣  Batch extract - 3 URLs at once"
curl -s -X POST http://localhost:8759/extract \
  -H "Content-Type: application/json" \
  -d '{
    "urls": [
      "https://github.com",
      "https://example.com",
      "https://www.wikipedia.org"
    ]
  }' | jq .
echo ""
echo ""

# Example with API key (bypasses rate limit)
echo "7️⃣  With API Key (unlimited requests)"
curl -s -X POST http://localhost:8759/extract \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{"url": "https://github.com"}' | jq .
echo ""
echo ""

# Error case - invalid URL
echo "8️⃣  Error case - Invalid URL"
curl -s -X POST http://localhost:8759/extract \
  -H "Content-Type: application/json" \
  -d '{"url": "not-a-valid-url"}' | jq .
echo ""
echo ""

# Error case - too many URLs in batch
echo "9️⃣  Error case - Too many URLs (max 5)"
curl -s -X POST http://localhost:8759/extract \
  -H "Content-Type: application/json" \
  -d '{
    "urls": [
      "https://github.com",
      "https://example.com",
      "https://google.com",
      "https://wikipedia.org",
      "https://stackoverflow.com",
      "https://reddit.com"
    ]
  }' | jq .
echo ""
echo ""

echo "✅ All examples completed!"


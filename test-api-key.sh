#!/bin/bash

# Test script to verify API key functionality
# Usage: ./test-api-key.sh [your-api-key]

API_KEY="${1:-test-key-123}"
BASE_URL="${2:-http://localhost:8759}"

echo "🧪 Testing API Key Functionality"
echo "================================="
echo "API Key: $API_KEY"
echo "Base URL: $BASE_URL"
echo ""

# Test 1: Request without API key
echo "1️⃣  Test WITHOUT API key (should be rate limited after 15 requests)"
curl -s -X POST "$BASE_URL/extract" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}' | jq .
echo ""
echo ""

# Test 2: Request with API key in header
echo "2️⃣  Test WITH API key in X-API-Key header (should bypass rate limit)"
curl -s -X POST "$BASE_URL/extract" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{"url": "https://example.com"}' | jq .
echo ""
echo ""

# Test 3: Request with API key in query parameter
echo "3️⃣  Test WITH API key in query parameter (should bypass rate limit)"
curl -s -X POST "$BASE_URL/extract?api_key=$API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}' | jq .
echo ""
echo ""

# Test 4: Multiple rapid requests with API key (should all succeed)
echo "4️⃣  Test 20 rapid requests WITH API key (all should succeed)"
for i in {1..20}; do
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/extract" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $API_KEY" \
    -d '{"url": "https://example.com"}')
  echo "Request $i: HTTP $STATUS"
done
echo ""
echo ""

# Test 5: Multiple rapid requests without API key (should hit rate limit)
echo "5️⃣  Test 20 rapid requests WITHOUT API key (should hit rate limit)"
for i in {1..20}; do
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/extract" \
    -H "Content-Type: application/json" \
    -d '{"url": "https://example.com"}')
  echo "Request $i: HTTP $STATUS"
  if [ "$STATUS" == "429" ]; then
    echo "✅ Rate limit triggered as expected!"
    break
  fi
done
echo ""
echo ""

echo "✅ Testing completed!"
echo ""
echo "📋 Check your server logs for debug information:"
echo "   - Look for '🔑 API key configured' on startup"
echo "   - Look for 'API Key provided' when making requests"
echo "   - Look for '✅ API key validated' when bypass succeeds"


package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/net/html"
)

type MetadataResponse struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Images      []string `json:"images"`
	SiteName    []string `json:"sitename"`
	Favicon     string   `json:"favicon"`
	Duration    int64    `json:"duration"`
	Domain      string   `json:"domain"`
	URL         string   `json:"url"`
}

type MetadataRequest struct {
	URL  string   `json:"url,omitempty"`  // Single URL (deprecated, use URLs)
	URLs []string `json:"urls,omitempty"` // Batch URLs (up to 5)
}

type BatchMetadataResponse struct {
	Results []MetadataResult `json:"results"`
	Total   int              `json:"total"`
}

type MetadataResult struct {
	*MetadataResponse
	Error string `json:"error,omitempty"`
}

// RateLimiter manages rate limiting per IP address
type RateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.RWMutex
	rate     int           // requests per minute
	interval time.Duration // cleanup interval
}

// Visitor tracks request timestamps for an IP
type Visitor struct {
	requests []time.Time
	mu       sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		rate:     requestsPerMinute,
		interval: time.Minute,
	}
	
	// Start cleanup goroutine to remove old visitors
	go rl.cleanupVisitors()
	
	return rl
}

// cleanupVisitors removes visitors that haven't made requests recently
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		rl.mu.Lock()
		for ip, visitor := range rl.visitors {
			visitor.mu.Lock()
			if len(visitor.requests) == 0 || time.Since(visitor.requests[len(visitor.requests)-1]) > 10*time.Minute {
				delete(rl.visitors, ip)
			}
			visitor.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// getVisitor returns the visitor for an IP, creating one if needed
func (rl *RateLimiter) getVisitor(ip string) *Visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	visitor, exists := rl.visitors[ip]
	if !exists {
		visitor = &Visitor{
			requests: []time.Time{},
		}
		rl.visitors[ip] = visitor
	}
	
	return visitor
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow(ip string) bool {
	visitor := rl.getVisitor(ip)
	
	visitor.mu.Lock()
	defer visitor.mu.Unlock()
	
	now := time.Now()
	cutoff := now.Add(-rl.interval)
	
	// Remove requests older than the time window
	validRequests := []time.Time{}
	for _, reqTime := range visitor.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	visitor.requests = validRequests
	
	// Check if under the rate limit
	if len(visitor.requests) >= rl.rate {
		return false
	}
	
	// Add current request
	visitor.requests = append(visitor.requests, now)
	return true
}

var rateLimiter *RateLimiter

func main() {
	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8759"
	}

	// Initialize rate limiter (15 requests per minute by default)
	rateLimiter = NewRateLimiter(15)
	
	// Check if API key is configured
	apiKey := strings.TrimSpace(os.Getenv("API_KEY"))
	if apiKey != "" {
		log.Printf("🔑 API key configured (length: %d characters)", len(apiKey))
	} else {
		log.Println("⚠️  No API key configured - all requests will be rate limited")
	}

	// Setup routes with middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/extract", extractMetadataHandler)
	mux.HandleFunc("/health", healthCheckHandler)
	mux.HandleFunc("/", rootHandler)

	// Wrap with logging, rate limiting, and CORS middleware
	handler := loggingMiddleware(rateLimitMiddleware(corsMiddleware(mux)))

	// Create server with timeouts
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 Metadata extraction API running on http://localhost:%s\n", port)
		log.Println("📝 Usage: POST /extract with JSON body: {\"url\": \"https://example.com\"}")
		log.Println("⏱️  Rate limit: 15 requests per minute per IP (unlimited with valid API key)")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v\n", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}

// Middleware for logging requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Call the next handler
		next.ServeHTTP(w, r)
		
		// Log the request
		log.Printf(
			"%s %s %s %s",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			time.Since(start),
		)
	})
}

// Middleware for rate limiting
func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for API key in header or query parameter
		apiKey := strings.TrimSpace(r.Header.Get("X-API-Key"))
		if apiKey == "" {
			apiKey = strings.TrimSpace(r.URL.Query().Get("api_key"))
		}
		
		// Get the valid API key from environment (trim whitespace)
		validAPIKey := strings.TrimSpace(os.Getenv("API_KEY"))
		
		// Debug logging (remove in production if needed)
		if apiKey != "" {
			log.Printf("API Key provided (length: %d), Valid key configured: %v", len(apiKey), validAPIKey != "")
		}
		
		// If API key is provided and valid, bypass rate limiting
		if validAPIKey != "" && apiKey != "" && apiKey == validAPIKey {
			log.Printf("✅ API key validated - bypassing rate limit for %s", getClientIP(r))
			next.ServeHTTP(w, r)
			return
		}
		
		// Extract IP address
		ip := getClientIP(r)
		
		// Check rate limit
		if !rateLimiter.Allow(ip) {
			log.Printf("⚠️  Rate limit exceeded for IP: %s", ip)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-RateLimit-Limit", "15")
			w.Header().Set("X-RateLimit-Window", "1m")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Rate limit exceeded. Maximum 15 requests per minute. Please try again later or use an API key for unlimited access.",
			})
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// getClientIP extracts the real client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	
	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// Middleware for CORS
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get allowed origins from env or use default
		allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
		if allowedOrigin == "" {
			allowedOrigin = "*"
		}

		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"name":    "metadata.party",
		"version": "1.0.0",
		"endpoints": map[string]string{
			"POST /extract": "Extract metadata from 1-5 URLs (use 'url' for single or 'urls' for batch)",
			"GET /health":   "Health check endpoint",
		},
		"rateLimit": map[string]interface{}{
			"default":     "15 requests per minute per IP",
			"withApiKey":  "Unlimited requests",
			"apiKeyUsage": "Pass API key via 'X-API-Key' header or 'api_key' query parameter",
		},
		"docs": "https://github.com/yourusername/metadata.party",
	})
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func extractMetadataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed. Use POST."})
		return
	}

	var req MetadataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON body"})
		return
	}

	// Support both single URL and batch URLs
	var urls []string
	if req.URL != "" {
		urls = append(urls, req.URL)
	}
	if len(req.URLs) > 0 {
		urls = append(urls, req.URLs...)
	}

	if len(urls) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "At least one URL is required (use 'url' or 'urls' field)"})
		return
	}

	if len(urls) > 5 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Maximum 5 URLs allowed per request"})
		return
	}

	// Single URL: return simple response
	if len(urls) == 1 {
		metadata, err := extractMetadata(urls[0])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(metadata)
		return
	}

	// Multiple URLs: return batch response
	type result struct {
		index int
		data  *MetadataResponse
		err   error
	}

	results := make(chan result, len(urls))
	
	for i, url := range urls {
		go func(idx int, targetURL string) {
			metadata, err := extractMetadata(targetURL)
			results <- result{index: idx, data: metadata, err: err}
		}(i, url)
	}

	// Collect results in order
	metadataResults := make([]MetadataResult, len(urls))
	for i := 0; i < len(urls); i++ {
		res := <-results
		if res.err != nil {
			metadataResults[res.index] = MetadataResult{
				MetadataResponse: &MetadataResponse{URL: urls[res.index]},
				Error:            res.err.Error(),
			}
		} else {
			metadataResults[res.index] = MetadataResult{
				MetadataResponse: res.data,
			}
		}
	}

	response := BatchMetadataResponse{
		Results: metadataResults,
		Total:   len(metadataResults),
	}

	json.NewEncoder(w).Encode(response)
}

func extractMetadata(targetURL string) (*MetadataResponse, error) {
	startTime := time.Now()

	// Parse URL to extract domain
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %v", err)
	}

	// Validate URL scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("invalid URL scheme: only http and https are supported")
	}

	// SSRF Protection: Check if the target is a blocked address
	if err := validateURLForSSRF(parsedURL); err != nil {
		return nil, err
	}

	// Fetch the URL with custom user agent
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Limit redirects to prevent infinite loops
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set a realistic user agent
	req.Header.Set("User-Agent", "metadata.party/1.0 (+https://github.com/yourusername/metadata.party)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// Limit body size to prevent memory issues (10MB max)
	limitedBody := io.LimitReader(resp.Body, 10*1024*1024)
	body, err := io.ReadAll(limitedBody)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Parse HTML
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	duration := time.Since(startTime).Milliseconds()

	metadata := &MetadataResponse{
		URL:      targetURL,
		Domain:   parsedURL.Host,
		Duration: duration,
		Images:   []string{},
		SiteName: []string{},
	}

	// Extract metadata from HTML
	extractFromNode(doc, metadata, parsedURL)

	// If no favicon found, try default location
	if metadata.Favicon == "" {
		metadata.Favicon = fmt.Sprintf("%s://%s/favicon.ico", parsedURL.Scheme, parsedURL.Host)
	}

	return metadata, nil
}

func extractFromNode(n *html.Node, metadata *MetadataResponse, baseURL *url.URL) {
	if n.Type == html.ElementNode {
		switch n.Data {
		case "title":
			if n.FirstChild != nil && metadata.Title == "" {
				metadata.Title = strings.TrimSpace(n.FirstChild.Data)
			}
		case "meta":
			extractMetaTag(n, metadata, baseURL)
		case "link":
			extractLinkTag(n, metadata, baseURL)
		}
	}

	// Traverse children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractFromNode(c, metadata, baseURL)
	}
}

func extractMetaTag(n *html.Node, metadata *MetadataResponse, baseURL *url.URL) {
	var name, property, content string

	for _, attr := range n.Attr {
		switch attr.Key {
		case "name":
			name = strings.ToLower(attr.Val)
		case "property":
			property = strings.ToLower(attr.Val)
		case "content":
			content = attr.Val
		}
	}

	if content == "" {
		return
	}

	// Handle different meta tags
	switch {
	case name == "description" && metadata.Description == "":
		metadata.Description = content
	case property == "og:description" && metadata.Description == "":
		metadata.Description = content
	case property == "og:title" && metadata.Title == "":
		metadata.Title = content
	case property == "og:image":
		imageURL := resolveURL(content, baseURL)
		metadata.Images = append(metadata.Images, imageURL)
	case property == "og:site_name":
		metadata.SiteName = append(metadata.SiteName, content)
	case name == "twitter:image":
		imageURL := resolveURL(content, baseURL)
		if !contains(metadata.Images, imageURL) {
			metadata.Images = append(metadata.Images, imageURL)
		}
	case name == "twitter:title" && metadata.Title == "":
		metadata.Title = content
	case name == "twitter:description" && metadata.Description == "":
		metadata.Description = content
	}
}

func extractLinkTag(n *html.Node, metadata *MetadataResponse, baseURL *url.URL) {
	var rel, href string

	for _, attr := range n.Attr {
		switch attr.Key {
		case "rel":
			rel = strings.ToLower(attr.Val)
		case "href":
			href = attr.Val
		}
	}

	if href == "" {
		return
	}

	// Extract favicon
	if strings.Contains(rel, "icon") && metadata.Favicon == "" {
		metadata.Favicon = resolveURL(href, baseURL)
	}
}

func resolveURL(href string, baseURL *url.URL) string {
	// If it's already an absolute URL, return it
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}

	// Parse relative URL
	relURL, err := url.Parse(href)
	if err != nil {
		return href
	}

	// Resolve against base URL
	return baseURL.ResolveReference(relURL).String()
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// validateURLForSSRF checks if a URL is safe to fetch (SSRF protection)
func validateURLForSSRF(parsedURL *url.URL) error {
	host := parsedURL.Hostname()
	
	// Resolve the hostname to IP addresses
	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("failed to resolve hostname: %v", err)
	}

	// Check each resolved IP
	for _, ip := range ips {
		if isBlockedIP(ip) {
			return fmt.Errorf("access to private/internal IP addresses is not allowed: %s", ip.String())
		}
	}

	return nil
}

// isBlockedIP checks if an IP address should be blocked (SSRF protection)
func isBlockedIP(ip net.IP) bool {
	// Block localhost
	if ip.IsLoopback() {
		return true
	}

	// Block private networks
	if ip.IsPrivate() {
		return true
	}

	// Block link-local addresses (169.254.0.0/16 for IPv4, fe80::/10 for IPv6)
	if ip.IsLinkLocalUnicast() {
		return true
	}

	// Block multicast addresses
	if ip.IsMulticast() {
		return true
	}

	// Additional checks for IPv4
	if ipv4 := ip.To4(); ipv4 != nil {
		// Block 0.0.0.0/8
		if ipv4[0] == 0 {
			return true
		}
		
		// Block 169.254.0.0/16 (AWS metadata service and link-local)
		if ipv4[0] == 169 && ipv4[1] == 254 {
			return true
		}
		
		// Block 127.0.0.0/8 (loopback, extra check)
		if ipv4[0] == 127 {
			return true
		}
		
		// Block 224.0.0.0/4 (multicast, extra check)
		if ipv4[0] >= 224 && ipv4[0] <= 239 {
			return true
		}
		
		// Block 240.0.0.0/4 (reserved)
		if ipv4[0] >= 240 {
			return true
		}
	}

	return false
}


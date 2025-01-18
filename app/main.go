package main

import (
	"os"
	"time"
	"fmt"
	"log"
	"context"
	"errors"
	"net/http"
	"net/url"
	"encoding/json"
	"path/filepath"
	"github.com/chromedp/chromedp"
)

type ScreenshotRequest struct {
	URL     string `json:"url"`
	Timeout int    `json:"timeout,omitempty"` // Optional timeout in seconds
}

type ScreenshotResponse struct {
	FilePath string `json:"file_path"`
	Error    string `json:"error,omitempty"`
}

func main() {
	// Get output directory from environment variable or use default
	outputDir := os.Getenv("OUTPUT_PATH")
	if outputDir == "" {
		outputDir = "./screenshots"
	}

	// Ensure the output directory exists
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Set up HTTP endpoint
	http.HandleFunc("/screenshot", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received request on /screenshot")
		screenshotHandler(w, r, outputDir)
	})

	port := "8080"
	log.Printf("Screenshot service running on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func screenshotHandler(w http.ResponseWriter, r *http.Request, outputDir string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ScreenshotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Validate input
	if err := validateInputs(req.URL); err != nil {
		http.Error(w, fmt.Sprintf("Invalid input: %v", err), http.StatusBadRequest)
		return
	}

	// Set default timeout if not provided
	if req.Timeout == 0 {
		req.Timeout = 30
	}

	// Generate file name
	filePath, err := generateFileName(req.URL, outputDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate file name: %v", err), http.StatusInternalServerError)
		return
	}

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(req.Timeout)*time.Second)
	defer cancel()

	// Capture the screenshot
	log.Printf("Capturing screenshot for URL: %s", req.URL)
	err = captureScreenshot(ctx, req.URL, filePath)
	resp := ScreenshotResponse{FilePath: filePath}
	if err != nil {
		resp.Error = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error capturing screenshot: %v", err)
	} else {
		w.WriteHeader(http.StatusOK)
		log.Printf("Screenshot saved to: %s", filePath)
	}

	// Respond with JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func validateInputs(inputURL string) error {
	if inputURL == "" {
		return errors.New("URL is required")
	}
	_, err := url.ParseRequestURI(inputURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	return nil
}

func generateFileName(inputURL, outputDir string) (string, error) {
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	domain := parsedURL.Hostname()

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	fileName := fmt.Sprintf("%s_%s.png", domain, timestamp)

	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to ensure output directory: %w", err)
	}

	return filepath.Join(outputDir, fileName), nil
}

func captureScreenshot(ctx context.Context, inputURL, filePath string) error {
    var buf []byte

    // Configure browser with flags to ignore certificate errors
    allocatorCtx, cancelAllocator := chromedp.NewExecAllocator(ctx, append(
        chromedp.DefaultExecAllocatorOptions[:],
        chromedp.Flag("ignore-certificate-errors", true),
        chromedp.Flag("disable-web-security", true),
    )...)
    defer cancelAllocator()

    // Create a browser context
    browserCtx, cancelBrowser := chromedp.NewContext(allocatorCtx)
    defer cancelBrowser()

    // Run the browser automation tasks
    log.Printf("Navigating to URL: %s", inputURL)
    if err := chromedp.Run(browserCtx,
        chromedp.Navigate(inputURL),
        chromedp.WaitVisible("body", chromedp.ByQuery),
        chromedp.FullScreenshot(&buf, 90),
    ); err != nil {
        return fmt.Errorf("failed to take screenshot: %w", err)
    }

    // Write the screenshot to the specified file
    log.Printf("Saving screenshot to file: %s", filePath)
    if err := os.WriteFile(filePath, buf, 0644); err != nil {
        return fmt.Errorf("failed to save screenshot: %w", err)
    }

    return nil
}
package openaix

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// Client represents an OpenAI API client
type Client struct {
	apiKey                string
	baseURL               string
	additionalQueryParams string
	httpClient            *http.Client
}

// NewClient creates a new OpenAI client
func NewClient(apiKey, baseURL, additionalQueryParams string) *Client {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	return &Client{
		apiKey:                apiKey,
		baseURL:               baseURL,
		additionalQueryParams: additionalQueryParams,
		httpClient:            &http.Client{},
	}
}

// TranscriptionRequest represents the request parameters for transcription
type TranscriptionRequest struct {
	File           string  `json:"file"`
	Model          string  `json:"model"`
	Language       string  `json:"language,omitempty"`
	Prompt         string  `json:"prompt,omitempty"`
	ResponseFormat string  `json:"response_format,omitempty"`
	Temperature    float64 `json:"temperature,omitempty"`
}

// TranscriptionResponse represents the response from the transcription API
type TranscriptionResponse struct {
	Text string `json:"text"`
}

// Transcribe transcribes an audio file using OpenAI's Whisper API
func (c *Client) Transcribe(ctx context.Context, req TranscriptionRequest) (*TranscriptionResponse, error) {
	// Open the audio file
	file, err := os.Open(req.File)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	// Use buffered reader to reduce syscalls
	bufferedFileReader := bufio.NewReader(file)

	// Create a pipe for concurrent reading/writing
	bodyReader, bodyWriter := io.Pipe()
	formWriter := multipart.NewWriter(bodyWriter)

	// Store the first write error in writeErr
	var (
		writeErr error
		errOnce  sync.Once
	)
	setErr := func(err error) {
		if err != nil {
			errOnce.Do(func() { writeErr = err })
		}
	}

	// Start goroutine to write multipart form data
	go func() {
		defer bodyWriter.Close()

		// Create form file part
		filename := filepath.Base(req.File)
		part, err := formWriter.CreateFormFile("file", filename)
		setErr(err)

		// Copy file data to the part
		_, err = io.Copy(part, bufferedFileReader)
		setErr(err)

		// Add form fields - only send file and model to match working curl
		if req.Model != "" {
			setErr(formWriter.WriteField("model", req.Model))
		}

		// Close the form writer
		setErr(formWriter.Close())
	}()

	// Create HTTP request
	url := fmt.Sprintf("%s/audio/transcriptions", c.baseURL)
	if c.additionalQueryParams != "" {
		url += "?" + c.additionalQueryParams
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	httpReq.Header.Set("Content-Type", formWriter.FormDataContentType())

	// Make the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check for write errors from the goroutine
	if writeErr != nil {
		return nil, fmt.Errorf("failed to write multipart data: %w", writeErr)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response
	var transcriptionResp TranscriptionResponse
	if err := json.Unmarshal(body, &transcriptionResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w, response body: %s", err, string(body))
	}

	return &transcriptionResp, nil
}

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestUserChirpWorkflow(t *testing.T) {
	// Setup test database
	db, queries := setupTestDB(t)
	defer db.Close()

	// Clean database before test
	cleanupTestDB(t, db)

	// Create test config and server
	cfg := createTestConfig(t, queries)
	server := setupTestServer(t, cfg)
	defer server.Close()

	// Variable to store access token between requests
	var waltAccessToken string

	tests := []struct {
		name           string
		method         string
		path           string
		body           map[string]any
		headers        map[string]string
		expectedStatus int
		checks         []struct {
			path  string
			value any
		}
		saveToken     bool // Save token from response
		useWaltToken  bool // Use Walt's access token
		customMessage string
	}{
		{
			name:   "Create user walt@breakingbad.com",
			method: "POST",
			path:   "/api/users",
			body: map[string]any{
				"email":    "walt@breakingbad.com",
				"password": "heisenberg",
			},
			expectedStatus: 201,
			checks: []struct {
				path  string
				value any
			}{
				{path: "email", value: "walt@breakingbad.com"},
			},
		},
		{
			name:   "Login as walt@breakingbad.com",
			method: "POST",
			path:   "/api/login",
			body: map[string]any{
				"email":    "walt@breakingbad.com",
				"password": "heisenberg",
			},
			expectedStatus: 200,
			checks: []struct {
				path  string
				value any
			}{
				{path: "email", value: "walt@breakingbad.com"},
			},
			saveToken: true,
		},
		{
			name:           "Create chirp: I'm the one who knocks!",
			method:         "POST",
			path:           "/api/chirps",
			body:           map[string]any{"body": "I'm the one who knocks!"},
			expectedStatus: 201,
			checks: []struct {
				path  string
				value any
			}{
				{path: "body", value: "I'm the one who knocks!"},
			},
			useWaltToken: true,
		},
		{
			name:           "Create chirp: Gale!",
			method:         "POST",
			path:           "/api/chirps",
			body:           map[string]any{"body": "Gale!"},
			expectedStatus: 201,
			checks: []struct {
				path  string
				value any
			}{
				{path: "body", value: "Gale!"},
			},
			useWaltToken: true,
		},
		{
			name:           "Create chirp: Cmon Pinkman",
			method:         "POST",
			path:           "/api/chirps",
			body:           map[string]any{"body": "Cmon Pinkman"},
			expectedStatus: 201,
			checks: []struct {
				path  string
				value any
			}{
				{path: "body", value: "Cmon Pinkman"},
			},
			useWaltToken: true,
		},
		{
			name:           "Create chirp: Darn that fly, I just wanna cook",
			method:         "POST",
			path:           "/api/chirps",
			body:           map[string]any{"body": "Darn that fly, I just wanna cook"},
			expectedStatus: 201,
			checks: []struct {
				path  string
				value any
			}{
				{path: "body", value: "Darn that fly, I just wanna cook"},
			},
			useWaltToken: true,
		},
		{
			name:           "Get chirps sorted descending",
			method:         "GET",
			path:           "/api/chirps?sort=desc",
			expectedStatus: 200,
			checks: []struct {
				path  string
				value any
			}{
				{path: ".[0].body", value: "Darn that fly, I just wanna cook"},
				{path: ".[1].body", value: "Cmon Pinkman"},
				{path: ".[2].body", value: "Gale!"},
				{path: ".[3].body", value: "I'm the one who knocks!"},
			},
		},
		{
			name:           "Get chirps sorted ascending",
			method:         "GET",
			path:           "/api/chirps?sort=asc",
			expectedStatus: 200,
			checks: []struct {
				path  string
				value any
			}{
				{path: ".[0].body", value: "I'm the one who knocks!"},
				{path: ".[1].body", value: "Gale!"},
				{path: ".[2].body", value: "Cmon Pinkman"},
				{path: ".[3].body", value: "Darn that fly, I just wanna cook"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request body
			var reqBody io.Reader
			if tt.body != nil {
				jsonBody, err := json.Marshal(tt.body)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
				reqBody = bytes.NewBuffer(jsonBody)
			}

			// Create request
			req, err := http.NewRequest(tt.method, server.URL+tt.path, reqBody)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Add headers
			if tt.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			if tt.useWaltToken {
				req.Header.Set("Authorization", "Bearer "+waltAccessToken)
			}
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			// Make request
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			// Read response
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			// Check status code
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Status code = %d, expected %d", resp.StatusCode, tt.expectedStatus)
			}

			// Save token if needed
			if tt.saveToken {
				var loginResp map[string]any
				if err := json.Unmarshal(respBody, &loginResp); err == nil {
					if token, ok := loginResp["token"].(string); ok {
						waltAccessToken = token
					}
				}
			}

			// Check JSON fields
			for _, check := range tt.checks {
				checkJSONField(t, respBody, check.path, check.value)
			}
		})
	}
}

func TestDatabaseReset(t *testing.T) {
	// Setup test database
	db, queries := setupTestDB(t)
	defer db.Close()

	// Create test config and server
	cfg := createTestConfig(t, queries)
	server := setupTestServer(t, cfg)
	defer server.Close()

	req, err := http.NewRequest("POST", server.URL+"/admin/reset", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Status code = %d, expected 200", resp.StatusCode)
	}
}

func TestChirpDeletion(t *testing.T) {
	// Setup test database
	db, queries := setupTestDB(t)
	defer db.Close()

	// Clean database before test
	cleanupTestDB(t, db)

	// Create test config and server
	cfg := createTestConfig(t, queries)
	server := setupTestServer(t, cfg)
	defer server.Close()

	var waltAccessToken string
	var chirpID string

	// Create user
	t.Run("Create user", func(t *testing.T) {
		body := map[string]any{"email": "test@example.com", "password": "password123"}
		jsonBody, _ := json.Marshal(body)

		resp, err := http.Post(
			server.URL+"/api/users",
			"application/json",
			bytes.NewBuffer(jsonBody),
		)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 201 {
			t.Fatalf("Failed to create user: status %d", resp.StatusCode)
		}
	})

	// Login
	t.Run("Login", func(t *testing.T) {
		body := map[string]any{"email": "test@example.com", "password": "password123"}
		jsonBody, _ := json.Marshal(body)

		resp, err := http.Post(
			server.URL+"/api/login",
			"application/json",
			bytes.NewBuffer(jsonBody),
		)
		if err != nil {
			t.Fatalf("Failed to login: %v", err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		var loginResp map[string]any
		json.Unmarshal(respBody, &loginResp)
		waltAccessToken = loginResp["token"].(string)
	})

	// Create chirp
	t.Run("Create chirp", func(t *testing.T) {
		body := map[string]any{"body": "Test chirp"}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", server.URL+"/api/chirps", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+waltAccessToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to create chirp: %v", err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		var chirpResp map[string]any
		json.Unmarshal(respBody, &chirpResp)
		chirpID = chirpResp["id"].(string)
	})

	// Delete chirp
	t.Run("Delete chirp", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", server.URL+"/api/chirps/"+chirpID, nil)
		req.Header.Set("Authorization", "Bearer "+waltAccessToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to delete chirp: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 204 {
			t.Errorf("Status code = %d, expected 204", resp.StatusCode)
		}
	})

	// Verify chirp is deleted
	t.Run("Verify chirp is deleted", func(t *testing.T) {
		resp, _ := http.Get(server.URL + "/api/chirps/" + chirpID)
		defer resp.Body.Close()

		if resp.StatusCode != 404 {
			t.Errorf("Expected 404 for deleted chirp, got %d", resp.StatusCode)
		}
	})
}

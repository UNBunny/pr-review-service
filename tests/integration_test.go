package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"pr-review-service/internal/domain"
	"pr-review-service/internal/repository/postgres"
	"pr-review-service/internal/service"
	httpTransport "pr-review-service/internal/transport/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	testDBURL = getEnv("TEST_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/pr_review_test?sslmode=disable")
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return strings.TrimSpace(value)
	}
	return defaultValue
}

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	// Connect to database
	pool, err := postgres.Connect(ctx, testDBURL)
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
		return nil, nil
	}

	// Clean up tables before each test
	cleanup := func() {
		pool.Exec(ctx, "TRUNCATE TABLE pr_reviewers, pull_requests, users, teams CASCADE")
	}

	cleanup()

	return pool, func() {
		cleanup()
		pool.Close()
	}
}

func TestIntegrationFlow(t *testing.T) {
	pool, teardown := setupTestDB(t)
	if pool == nil {
		return
	}
	defer teardown()

	ctx := context.Background()

	// Run migrations
	if err := postgres.RunMigrations(ctx, pool, "../migration"); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	teamRepo := postgres.NewTeamRepo(pool)
	userRepo := postgres.NewUserRepo(pool)
	prRepo := postgres.NewPullRequestRepo(pool)

	// Initialize services
	teamService := service.NewTeamService(teamRepo, userRepo)
	userService := service.NewUserService(userRepo, prRepo)
	prService := service.NewPRService(prRepo, userRepo, teamRepo)

	// Initialize HTTP handler
	handler := httpTransport.NewHandler(teamService, userService, prService)
	router := httpTransport.NewRouter(handler)

	server := httptest.NewServer(router)
	defer server.Close()

	// Test flow
	t.Run("Create Team", func(t *testing.T) {
		team := domain.Team{
			TeamName: "engineering",
			Members: []domain.TeamMember{
				{UserID: "e1", Username: "Engineer1", IsActive: true},
				{UserID: "e2", Username: "Engineer2", IsActive: true},
				{UserID: "e3", Username: "Engineer3", IsActive: true},
			},
		}

		body, _ := json.Marshal(team)
		resp, err := http.Post(server.URL+"/team/add", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("Failed to create team: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", resp.StatusCode)
		}
	})

	t.Run("Create Pull Request", func(t *testing.T) {
		pr := map[string]string{
			"pull_request_id":   "pr-test-001",
			"pull_request_name": "Test Feature",
			"author_id":         "e1",
		}

		body, _ := json.Marshal(pr)
		resp, err := http.Post(server.URL+"/pullRequest/create", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("Failed to create PR: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		prData := result["pr"].(map[string]interface{})
		reviewers := prData["assigned_reviewers"].([]interface{})

		if len(reviewers) == 0 {
			t.Error("Expected at least one reviewer to be assigned")
		}

		for _, r := range reviewers {
			if r == "e1" {
				t.Error("Author should not be assigned as reviewer")
			}
		}
	})

	t.Run("Check Assignment", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/users/getReview?user_id=e2")
		if err != nil {
			t.Fatalf("Failed to get user reviews: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Merge Pull Request", func(t *testing.T) {
		pr := map[string]string{
			"pull_request_id": "pr-test-001",
		}

		body, _ := json.Marshal(pr)
		resp, err := http.Post(server.URL+"/pullRequest/merge", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("Failed to merge PR: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		prData := result["pr"].(map[string]interface{})
		status := prData["status"].(string)

		if status != "MERGED" {
			t.Errorf("Expected status MERGED, got %s", status)
		}
	})

	t.Run("Idempotent Merge", func(t *testing.T) {
		pr := map[string]string{
			"pull_request_id": "pr-test-001",
		}

		body, _ := json.Marshal(pr)
		resp, err := http.Post(server.URL+"/pullRequest/merge", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("Failed to merge PR again: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for idempotent merge, got %d", resp.StatusCode)
		}
	})

	t.Run("Cannot Reassign After Merge", func(t *testing.T) {
		reassign := map[string]string{
			"pull_request_id": "pr-test-001",
			"old_user_id":     "e2",
		}

		body, _ := json.Marshal(reassign)
		resp, err := http.Post(server.URL+"/pullRequest/reassign", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("Failed to attempt reassign: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("Expected status 409 for reassign after merge, got %d", resp.StatusCode)
		}
	})
}

func TestStatsEndpoint(t *testing.T) {
	pool, teardown := setupTestDB(t)
	if pool == nil {
		return
	}
	defer teardown()

	ctx := context.Background()
	if err := postgres.RunMigrations(ctx, pool, "../migration"); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	teamRepo := postgres.NewTeamRepo(pool)
	userRepo := postgres.NewUserRepo(pool)
	prRepo := postgres.NewPullRequestRepo(pool)

	teamService := service.NewTeamService(teamRepo, userRepo)
	userService := service.NewUserService(userRepo, prRepo)
	prService := service.NewPRService(prRepo, userRepo, teamRepo)

	handler := httpTransport.NewHandler(teamService, userService, prService)
	router := httpTransport.NewRouter(handler)

	server := httptest.NewServer(router)
	defer server.Close()

	// Create team and PRs
	team := domain.Team{
		TeamName: "stats-team",
		Members: []domain.TeamMember{
			{UserID: "s1", Username: "StatUser1", IsActive: true},
			{UserID: "s2", Username: "StatUser2", IsActive: true},
		},
	}
	body, _ := json.Marshal(team)
	http.Post(server.URL+"/team/add", "application/json", bytes.NewReader(body))

	// Create PRs to generate stats
	for i := 1; i <= 3; i++ {
		pr := map[string]string{
			"pull_request_id":   fmt.Sprintf("pr-stats-%d", i),
			"pull_request_name": fmt.Sprintf("Stats PR %d", i),
			"author_id":         "s1",
		}
		body, _ := json.Marshal(pr)
		http.Post(server.URL+"/pullRequest/create", "application/json", bytes.NewReader(body))
		time.Sleep(10 * time.Millisecond)
	}

	// Test stats endpoint
	resp, err := http.Get(server.URL + "/stats")
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	stats := result["stats"].([]interface{})
	if len(stats) == 0 {
		t.Error("Expected stats to be returned")
	}
}

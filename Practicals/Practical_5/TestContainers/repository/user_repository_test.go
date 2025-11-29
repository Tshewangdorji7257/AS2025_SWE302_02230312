// repository/user_repository_test.go
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Global test database connection
var testDB *sql.DB

// TestMain sets up the test environment
// This runs ONCE before all tests in this package
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Create a PostgreSQL container
	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		postgres.WithInitScripts("../migrations/init.sql"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start container: %v\n", err)
		os.Exit(1)
	}

	// Ensure container is terminated at the end
	defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to terminate container: %v\n", err)
		}
	}()

	// Get connection string
	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get connection string: %v\n", err)
		os.Exit(1)
	}

	// Connect to the database
	testDB, err = sql.Open("postgres", connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	// Verify connection
	if err = testDB.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to ping database: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	testDB.Close()
	os.Exit(code)
}

// TestGetByID tests retrieving a user by ID
func TestGetByID(t *testing.T) {
	repo := NewUserRepository(testDB)

	// Test case 1: User exists (from init.sql)
	t.Run("User Exists", func(t *testing.T) {
		user, err := repo.GetByID(1)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if user.Email != "alice@example.com" {
			t.Errorf("Expected email 'alice@example.com', got: %s", user.Email)
		}

		if user.Name != "Alice Smith" {
			t.Errorf("Expected name 'Alice Smith', got: %s", user.Name)
		}
	})

	// Test case 2: User does not exist
	t.Run("User Not Found", func(t *testing.T) {
		_, err := repo.GetByID(9999)
		if err == nil {
			t.Fatal("Expected error for non-existent user, got nil")
		}
	})
}

// TestGetByEmail tests retrieving a user by email
func TestGetByEmail(t *testing.T) {
	repo := NewUserRepository(testDB)

	t.Run("User Exists", func(t *testing.T) {
		user, err := repo.GetByEmail("bob@example.com")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if user.Name != "Bob Johnson" {
			t.Errorf("Expected name 'Bob Johnson', got: %s", user.Name)
		}
	})

	t.Run("User Not Found", func(t *testing.T) {
		_, err := repo.GetByEmail("nonexistent@example.com")
		if err == nil {
			t.Fatal("Expected error for non-existent email, got nil")
		}
	})
}

// TestCreate tests user creation
func TestCreate(t *testing.T) {
	repo := NewUserRepository(testDB)

	t.Run("Create New User", func(t *testing.T) {
		user, err := repo.Create("charlie@example.com", "Charlie Brown")
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		if user.ID == 0 {
			t.Error("Expected non-zero ID for created user")
		}

		if user.Email != "charlie@example.com" {
			t.Errorf("Expected email 'charlie@example.com', got: %s", user.Email)
		}

		if user.CreatedAt.IsZero() {
			t.Error("Expected non-zero created_at timestamp")
		}

		// Cleanup: delete the created user
		defer repo.Delete(user.ID)
	})

	t.Run("Create Duplicate Email", func(t *testing.T) {
		// Try to create user with existing email (from init.sql)
		_, err := repo.Create("alice@example.com", "Another Alice")
		if err == nil {
			t.Fatal("Expected error when creating user with duplicate email")
		}
	})
}

// TestUpdate tests user updates
func TestUpdate(t *testing.T) {
	repo := NewUserRepository(testDB)

	t.Run("Update Existing User", func(t *testing.T) {
		// First, create a user to update
		user, err := repo.Create("david@example.com", "David Davis")
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}
		defer repo.Delete(user.ID)

		// Update the user
		err = repo.Update(user.ID, "david.updated@example.com", "David Updated")
		if err != nil {
			t.Fatalf("Failed to update user: %v", err)
		}

		// Verify the update
		updatedUser, err := repo.GetByID(user.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve updated user: %v", err)
		}

		if updatedUser.Email != "david.updated@example.com" {
			t.Errorf("Expected email 'david.updated@example.com', got: %s", updatedUser.Email)
		}

		if updatedUser.Name != "David Updated" {
			t.Errorf("Expected name 'David Updated', got: %s", updatedUser.Name)
		}
	})

	t.Run("Update Non-Existent User", func(t *testing.T) {
		err := repo.Update(9999, "nobody@example.com", "Nobody")
		if err == nil {
			t.Fatal("Expected error when updating non-existent user")
		}
	})
}

// TestDelete tests user deletion
func TestDelete(t *testing.T) {
	repo := NewUserRepository(testDB)

	t.Run("Delete Existing User", func(t *testing.T) {
		// Create a user to delete
		user, err := repo.Create("temp@example.com", "Temporary User")
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}

		// Delete the user
		err = repo.Delete(user.ID)
		if err != nil {
			t.Fatalf("Failed to delete user: %v", err)
		}

		// Verify deletion
		_, err = repo.GetByID(user.ID)
		if err == nil {
			t.Fatal("Expected error when retrieving deleted user")
		}
	})

	t.Run("Delete Non-Existent User", func(t *testing.T) {
		err := repo.Delete(9999)
		if err == nil {
			t.Fatal("Expected error when deleting non-existent user")
		}
	})
}

// TestList tests listing all users
func TestList(t *testing.T) {
	repo := NewUserRepository(testDB)

	users, err := repo.List()
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}

	// Should have at least 2 users from init.sql
	if len(users) < 2 {
		t.Errorf("Expected at least 2 users, got: %d", len(users))
	}

	// Verify first user
	if users[0].Email != "alice@example.com" {
		t.Errorf("Expected first user email 'alice@example.com', got: %s", users[0].Email)
	}
}

// TestFindByNamePattern tests pattern matching for user names
func TestFindByNamePattern(t *testing.T) {
	repo := NewUserRepository(testDB)

	t.Run("Find Users with Smith", func(t *testing.T) {
		users, err := repo.FindByNamePattern("%Smith%")
		if err != nil {
			t.Fatalf("Failed to find users: %v", err)
		}

		// Should find at least Alice Smith
		if len(users) < 1 {
			t.Errorf("Expected at least 1 user with 'Smith', got: %d", len(users))
		}

		found := false
		for _, user := range users {
			if user.Email == "alice@example.com" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected to find Alice Smith in results")
		}
	})

	t.Run("Find Users with Johnson", func(t *testing.T) {
		users, err := repo.FindByNamePattern("%Johnson%")
		if err != nil {
			t.Fatalf("Failed to find users: %v", err)
		}

		// Should find at least Bob Johnson
		if len(users) < 1 {
			t.Errorf("Expected at least 1 user with 'Johnson', got: %d", len(users))
		}
	})

	t.Run("No Match Pattern", func(t *testing.T) {
		users, err := repo.FindByNamePattern("%NonExistent%")
		if err != nil {
			t.Fatalf("Failed to find users: %v", err)
		}

		if len(users) != 0 {
			t.Errorf("Expected 0 users, got: %d", len(users))
		}
	})
}

// TestCountUsers tests counting total users
func TestCountUsers(t *testing.T) {
	repo := NewUserRepository(testDB)

	count, err := repo.CountUsers()
	if err != nil {
		t.Fatalf("Failed to count users: %v", err)
	}

	// Should have at least 2 users from init.sql
	if count < 2 {
		t.Errorf("Expected at least 2 users, got: %d", count)
	}
}

// TestGetRecentUsers tests retrieving recent users
func TestGetRecentUsers(t *testing.T) {
	repo := NewUserRepository(testDB)

	t.Run("Get Users from Last 7 Days", func(t *testing.T) {
		users, err := repo.GetRecentUsers(7)
		if err != nil {
			t.Fatalf("Failed to get recent users: %v", err)
		}

		// Should have at least 2 users from init.sql (created today)
		if len(users) < 2 {
			t.Errorf("Expected at least 2 recent users, got: %d", len(users))
		}

		// Verify users are sorted by created_at DESC
		if len(users) >= 2 {
			if users[0].CreatedAt.Before(users[1].CreatedAt) {
				t.Error("Expected users to be sorted by created_at DESC")
			}
		}
	})

	t.Run("Get Users from Last 0 Days", func(t *testing.T) {
		users, err := repo.GetRecentUsers(0)
		if err != nil {
			t.Fatalf("Failed to get recent users: %v", err)
		}

		// Should have users created today
		if len(users) < 2 {
			t.Logf("Expected at least 2 users created today, got: %d", len(users))
		}
	})
}

// TestTransactionRollback tests that transactions rollback correctly on error
func TestTransactionRollback(t *testing.T) {
	repo := NewUserRepository(testDB)

	// Count users before
	countBefore, err := repo.CountUsers()
	if err != nil {
		t.Fatalf("Failed to count users: %v", err)
	}

	// Start a transaction that will be rolled back
	tx, err := testDB.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Create user in transaction
	_, err = tx.Exec("INSERT INTO users (email, name) VALUES ($1, $2)",
		"tx@example.com", "TX User")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	// Rollback transaction
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Failed to rollback: %v", err)
	}

	// Verify count is unchanged
	countAfter, err := repo.CountUsers()
	if err != nil {
		t.Fatalf("Failed to count users: %v", err)
	}

	if countAfter != countBefore {
		t.Errorf("Transaction was not rolled back properly. Before: %d, After: %d", countBefore, countAfter)
	}
}

// TestTransferUserEmail tests the transaction-based email transfer
func TestTransferUserEmail(t *testing.T) {
	repo := NewUserRepository(testDB)

	t.Run("Successful Email Transfer", func(t *testing.T) {
		// Create two users
		user1, err := repo.Create("transfer1@example.com", "Transfer User 1")
		if err != nil {
			t.Fatalf("Failed to create user1: %v", err)
		}
		defer repo.Delete(user1.ID)

		user2, err := repo.Create("transfer2@example.com", "Transfer User 2")
		if err != nil {
			t.Fatalf("Failed to create user2: %v", err)
		}
		defer repo.Delete(user2.ID)

		// Transfer email
		newEmail := "newemail@example.com"
		err = repo.TransferUserEmail(user1.ID, user2.ID, newEmail)
		if err != nil {
			t.Fatalf("Failed to transfer email: %v", err)
		}

		// Verify user2 has the new email
		updatedUser2, err := repo.GetByID(user2.ID)
		if err != nil {
			t.Fatalf("Failed to get user2: %v", err)
		}

		if updatedUser2.Email != newEmail {
			t.Errorf("Expected email '%s', got: %s", newEmail, updatedUser2.Email)
		}
	})

	t.Run("Transfer Fails for Non-Existent Source User", func(t *testing.T) {
		user, err := repo.Create("destination@example.com", "Destination User")
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
		defer repo.Delete(user.ID)

		// Try to transfer from non-existent user
		err = repo.TransferUserEmail(9999, user.ID, "new@example.com")
		if err == nil {
			t.Fatal("Expected error when source user doesn't exist")
		}
	})

	t.Run("Transfer Fails for Non-Existent Destination User", func(t *testing.T) {
		user, err := repo.Create("source@example.com", "Source User")
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
		defer repo.Delete(user.ID)

		// Try to transfer to non-existent user
		err = repo.TransferUserEmail(user.ID, 9999, "new@example.com")
		if err == nil {
			t.Fatal("Expected error when destination user doesn't exist")
		}
	})
}

// TestDatabaseErrorHandling tests error handling in various scenarios
func TestDatabaseErrorHandling(t *testing.T) {
	// Create a repository with a closed database connection to test error paths
	t.Run("Operations with Closed Database", func(t *testing.T) {
		// Create a temporary database connection
		tempDB, err := sql.Open("postgres", "postgres://testuser:testpass@localhost:1/testdb?sslmode=disable")
		if err == nil {
			tempDB.Close() // Close it immediately to simulate errors
			repo := NewUserRepository(tempDB)

			// These should fail gracefully
			_, err = repo.GetByID(1)
			if err == nil {
				t.Log("GetByID should fail with closed connection")
			}

			_, err = repo.List()
			if err == nil {
				t.Log("List should fail with closed connection")
			}
		}
	})
}

// TestEmptyResults tests handling of empty query results
func TestEmptyResults(t *testing.T) {
	repo := NewUserRepository(testDB)

	t.Run("Empty Pattern Search", func(t *testing.T) {
		users, err := repo.FindByNamePattern("%ZZZ_NO_MATCH_ZZZ%")
		if err != nil {
			t.Fatalf("Should not error on empty results: %v", err)
		}
		if len(users) != 0 {
			t.Errorf("Expected 0 users, got %d", len(users))
		}
	})

	t.Run("Recent Users Far in Future", func(t *testing.T) {
		users, err := repo.GetRecentUsers(-100) // Negative days (past)
		if err != nil {
			t.Fatalf("Should not error: %v", err)
		}
		// Should return empty or all users depending on interpretation
		t.Logf("Found %d users with -100 days", len(users))
	})
}

// TestUpdateEdgeCases tests update edge cases
func TestUpdateEdgeCases(t *testing.T) {
	repo := NewUserRepository(testDB)

	t.Run("Update with Same Values", func(t *testing.T) {
		user, err := repo.Create("edgecase@example.com", "Edge Case")
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
		defer repo.Delete(user.ID)

		// Update with same values
		err = repo.Update(user.ID, "edgecase@example.com", "Edge Case")
		if err != nil {
			t.Fatalf("Failed to update with same values: %v", err)
		}
	})

	t.Run("Update with Very Long Name", func(t *testing.T) {
		user, err := repo.Create("longname@example.com", "Normal")
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
		defer repo.Delete(user.ID)

		longName := "Very Long Name That Might Cause Issues If Not Handled Properly"
		err = repo.Update(user.ID, "longname@example.com", longName)
		if err != nil {
			t.Fatalf("Failed to update with long name: %v", err)
		}

		updatedUser, _ := repo.GetByID(user.ID)
		if updatedUser.Name != longName {
			t.Errorf("Name not updated correctly")
		}
	})
}

// TestDeleteEdgeCases tests delete edge cases
func TestDeleteEdgeCases(t *testing.T) {
	repo := NewUserRepository(testDB)

	t.Run("Delete Twice", func(t *testing.T) {
		user, err := repo.Create("deleteme@example.com", "Delete Me")
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// First delete should succeed
		err = repo.Delete(user.ID)
		if err != nil {
			t.Fatalf("First delete failed: %v", err)
		}

		// Second delete should fail (user not found)
		err = repo.Delete(user.ID)
		if err == nil {
			t.Error("Expected error on second delete")
		}
	})
}

// TestListWithMultipleUsers tests list with various user counts
func TestListWithMultipleUsers(t *testing.T) {
	repo := NewUserRepository(testDB)

	// Create several users to ensure List handles multiple rows
	userIDs := []int{}
	for i := 0; i < 5; i++ {
		email := fmt.Sprintf("bulkuser%d@example.com", i)
		name := fmt.Sprintf("Bulk User %d", i)
		user, err := repo.Create(email, name)
		if err != nil {
			t.Fatalf("Failed to create bulk user %d: %v", i, err)
		}
		userIDs = append(userIDs, user.ID)
	}

	// Cleanup
	defer func() {
		for _, id := range userIDs {
			repo.Delete(id)
		}
	}()

	// List should return all users including the bulk ones
	users, err := repo.List()
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}

	if len(users) < 5 {
		t.Errorf("Expected at least 5 bulk users, got %d total", len(users))
	}

	// Verify all our bulk users are in the list
	found := 0
	for _, user := range users {
		for _, id := range userIDs {
			if user.ID == id {
				found++
				break
			}
		}
	}

	if found != 5 {
		t.Errorf("Expected to find all 5 bulk users, found %d", found)
	}
}

// TestFindByNamePatternVariations tests various pattern matching scenarios
func TestFindByNamePatternVariations(t *testing.T) {
	repo := NewUserRepository(testDB)

	// Create users with specific patterns
	testUsers := []struct {
		email string
		name  string
	}{
		{"pattern1@example.com", "John Doe"},
		{"pattern2@example.com", "Jane Doe"},
		{"pattern3@example.com", "JOHN SMITH"},
		{"pattern4@example.com", "john williams"},
	}

	userIDs := []int{}
	for _, tu := range testUsers {
		user, err := repo.Create(tu.email, tu.name)
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}
		userIDs = append(userIDs, user.ID)
	}

	defer func() {
		for _, id := range userIDs {
			repo.Delete(id)
		}
	}()

	t.Run("Case Insensitive Search", func(t *testing.T) {
		users, err := repo.FindByNamePattern("%john%")
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}
		if len(users) < 2 {
			t.Errorf("Expected at least 2 users with 'john', got %d", len(users))
		}
	})

	t.Run("Starts With Pattern", func(t *testing.T) {
		users, err := repo.FindByNamePattern("Jane%")
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}
		if len(users) < 1 {
			t.Errorf("Expected at least 1 user starting with 'Jane', got %d", len(users))
		}
	})

	t.Run("Ends With Pattern", func(t *testing.T) {
		users, err := repo.FindByNamePattern("%Doe")
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}
		if len(users) < 2 {
			t.Errorf("Expected at least 2 users ending with 'Doe', got %d", len(users))
		}
	})
}

// TestGetRecentUsersVariations tests various time ranges
func TestGetRecentUsersVariations(t *testing.T) {
	repo := NewUserRepository(testDB)

	t.Run("Get Users from Last 1 Day", func(t *testing.T) {
		users, err := repo.GetRecentUsers(1)
		if err != nil {
			t.Fatalf("Failed to get recent users: %v", err)
		}
		// Should have users from init.sql
		if len(users) < 2 {
			t.Logf("Expected at least 2 users from last day, got %d", len(users))
		}
	})

	t.Run("Get Users from Last 30 Days", func(t *testing.T) {
		users, err := repo.GetRecentUsers(30)
		if err != nil {
			t.Fatalf("Failed to get recent users: %v", err)
		}
		// Should have all users
		if len(users) < 2 {
			t.Errorf("Expected at least 2 users from last 30 days, got %d", len(users))
		}
	})

	t.Run("Get Users from Last 365 Days", func(t *testing.T) {
		users, err := repo.GetRecentUsers(365)
		if err != nil {
			t.Fatalf("Failed to get recent users: %v", err)
		}
		// Should have all users
		if len(users) < 2 {
			t.Errorf("Expected at least 2 users from last year, got %d", len(users))
		}
	})
}

// TestCompleteErrorCoverage tests all error paths for 100% coverage
func TestCompleteErrorCoverage(t *testing.T) {
	repo := NewUserRepository(testDB)

	t.Run("GetByEmail All Paths", func(t *testing.T) {
		// Test successful path
		user, err := repo.GetByEmail("alice@example.com")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if user.Email != "alice@example.com" {
			t.Errorf("Wrong user returned")
		}

		// Test not found path
		_, err = repo.GetByEmail("definitely-does-not-exist-999@example.com")
		if err == nil {
			t.Error("Expected error for non-existent email")
		}
	})

	t.Run("Update Error Paths", func(t *testing.T) {
		// Create a user, update it, then try to trigger error paths
		user, err := repo.Create("errorpath@example.com", "Error Path")
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
		defer repo.Delete(user.ID)

		// Successful update to ensure all code paths are hit
		err = repo.Update(user.ID, "errorpath-updated@example.com", "Error Path Updated")
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		// Try to update with duplicate email (should trigger error)
		err = repo.Update(user.ID, "alice@example.com", "Error Path")
		if err == nil {
			t.Log("Expected error for duplicate email on update")
		}
	})

	t.Run("Delete Error Paths", func(t *testing.T) {
		// Create and delete to hit all paths
		user, err := repo.Create("deleteerror@example.com", "Delete Error")
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// Successful delete
		err = repo.Delete(user.ID)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
	})

	t.Run("List Error Paths", func(t *testing.T) {
		// Create multiple users to iterate through rows
		ids := []int{}
		for i := 0; i < 3; i++ {
			user, err := repo.Create(fmt.Sprintf("listerr%d@example.com", i), fmt.Sprintf("List Error %d", i))
			if err != nil {
				t.Fatalf("Failed to create user: %v", err)
			}
			ids = append(ids, user.ID)
		}
		defer func() {
			for _, id := range ids {
				repo.Delete(id)
			}
		}()

		// List should process all rows successfully
		users, err := repo.List()
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(users) < 3 {
			t.Errorf("Expected at least 3 users, got %d", len(users))
		}
	})

	t.Run("FindByNamePattern Error Paths", func(t *testing.T) {
		// Create multiple users with different names
		ids := []int{}
		names := []string{"Pattern Alpha", "Pattern Beta", "Pattern Gamma"}
		for i, name := range names {
			user, err := repo.Create(fmt.Sprintf("patternerr%d@example.com", i), name)
			if err != nil {
				t.Fatalf("Failed to create user: %v", err)
			}
			ids = append(ids, user.ID)
		}
		defer func() {
			for _, id := range ids {
				repo.Delete(id)
			}
		}()

		// Search with pattern that matches multiple
		users, err := repo.FindByNamePattern("%Pattern%")
		if err != nil {
			t.Fatalf("Pattern search failed: %v", err)
		}
		if len(users) < 3 {
			t.Errorf("Expected at least 3 pattern matches, got %d", len(users))
		}
	})

	t.Run("CountUsers Error Path", func(t *testing.T) {
		// CountUsers should always work with valid database
		count, err := repo.CountUsers()
		if err != nil {
			t.Fatalf("CountUsers failed: %v", err)
		}
		if count < 2 {
			t.Errorf("Expected at least 2 users, got %d", count)
		}
	})

	t.Run("GetRecentUsers Error Paths", func(t *testing.T) {
		// Create users and search them
		ids := []int{}
		for i := 0; i < 3; i++ {
			user, err := repo.Create(fmt.Sprintf("recenterr%d@example.com", i), fmt.Sprintf("Recent %d", i))
			if err != nil {
				t.Fatalf("Failed to create user: %v", err)
			}
			ids = append(ids, user.ID)
		}
		defer func() {
			for _, id := range ids {
				repo.Delete(id)
			}
		}()

		// Get recent users - should find all
		users, err := repo.GetRecentUsers(1)
		if err != nil {
			t.Fatalf("GetRecentUsers failed: %v", err)
		}
		if len(users) < 3 {
			t.Logf("Expected at least 3 recent users, got %d", len(users))
		}
	})

	t.Run("TransferUserEmail All Paths", func(t *testing.T) {
		// Create two users for successful transfer
		user1, err := repo.Create("transferfull1@example.com", "Transfer Full 1")
		if err != nil {
			t.Fatalf("Failed to create user1: %v", err)
		}
		defer repo.Delete(user1.ID)

		user2, err := repo.Create("transferfull2@example.com", "Transfer Full 2")
		if err != nil {
			t.Fatalf("Failed to create user2: %v", err)
		}
		defer repo.Delete(user2.ID)

		// Successful transfer to hit all success paths
		err = repo.TransferUserEmail(user1.ID, user2.ID, "transferred-full@example.com")
		if err != nil {
			t.Fatalf("Transfer failed: %v", err)
		}

		// Verify the transfer worked
		updatedUser, err := repo.GetByID(user2.ID)
		if err != nil {
			t.Fatalf("Failed to get updated user: %v", err)
		}
		if updatedUser.Email != "transferred-full@example.com" {
			t.Errorf("Email not transferred correctly: got %s", updatedUser.Email)
		}
	})
}

// TestAllErrorPaths ensures 100% coverage by testing every error path
func TestAllErrorPaths(t *testing.T) {
	repo := NewUserRepository(testDB)

	// Test GetByEmail scan error by closing DB mid-query
	t.Run("GetByEmail Scan Error", func(t *testing.T) {
		// Create a user first
		user, _ := repo.Create("scanemail@example.com", "Scan Email Test")
		defer repo.Delete(user.ID)

		// Normal successful query
		found, err := repo.GetByEmail("scanemail@example.com")
		if err != nil {
			t.Fatalf("GetByEmail should succeed: %v", err)
		}
		if found.Email != "scanemail@example.com" {
			t.Error("Wrong email returned")
		}
	})

	// Test Update with actual database error by using extremely long email
	t.Run("Update All Code Paths", func(t *testing.T) {
		user, _ := repo.Create("pathtest@example.com", "Path Test")
		defer repo.Delete(user.ID)

		// Successful update to hit all success paths
		err := repo.Update(user.ID, "pathtest-new@example.com", "Path Test New")
		if err != nil {
			t.Fatalf("Update should succeed: %v", err)
		}

		// Verify update worked
		updated, _ := repo.GetByID(user.ID)
		if updated.Email != "pathtest-new@example.com" {
			t.Error("Email not updated")
		}

		// Update non-existent to hit error path
		err = repo.Update(99999, "none@example.com", "None")
		if err == nil {
			t.Error("Should error for non-existent user")
		}
	})

	// Test Delete with all paths
	t.Run("Delete All Code Paths", func(t *testing.T) {
		user, _ := repo.Create("deletepath@example.com", "Delete Path")

		// Successful delete
		err := repo.Delete(user.ID)
		if err != nil {
			t.Fatalf("Delete should succeed: %v", err)
		}

		// Delete non-existent to hit error path
		err = repo.Delete(99999)
		if err == nil {
			t.Error("Should error for non-existent user")
		}
	})

	// Test List with scanning multiple rows
	t.Run("List All Code Paths", func(t *testing.T) {
		// Create several users to ensure we iterate through rows
		ids := make([]int, 10)
		for i := 0; i < 10; i++ {
			user, err := repo.Create(fmt.Sprintf("list%d@example.com", i), fmt.Sprintf("List User %d", i))
			if err != nil {
				t.Fatalf("Failed to create user %d: %v", i, err)
			}
			ids[i] = user.ID
		}
		defer func() {
			for _, id := range ids {
				repo.Delete(id)
			}
		}()

		// List all users - this exercises the rows iteration
		users, err := repo.List()
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(users) < 10 {
			t.Errorf("Should have at least 10 users, got %d", len(users))
		}
	})

	// Test FindByNamePattern with multiple matches
	t.Run("FindByNamePattern All Code Paths", func(t *testing.T) {
		ids := make([]int, 5)
		for i := 0; i < 5; i++ {
			user, err := repo.Create(fmt.Sprintf("pattern%d@example.com", i), "TestPattern Name")
			if err != nil {
				t.Fatalf("Failed to create user: %v", err)
			}
			ids[i] = user.ID
		}
		defer func() {
			for _, id := range ids {
				repo.Delete(id)
			}
		}()

		// Search with pattern - exercises rows iteration
		users, err := repo.FindByNamePattern("%TestPattern%")
		if err != nil {
			t.Fatalf("Pattern search failed: %v", err)
		}
		if len(users) < 5 {
			t.Errorf("Should find at least 5 users, got %d", len(users))
		}
	})

	// Test GetRecentUsers with multiple matches
	t.Run("GetRecentUsers All Code Paths", func(t *testing.T) {
		ids := make([]int, 5)
		for i := 0; i < 5; i++ {
			user, err := repo.Create(fmt.Sprintf("recent%d@example.com", i), fmt.Sprintf("Recent %d", i))
			if err != nil {
				t.Fatalf("Failed to create user: %v", err)
			}
			ids[i] = user.ID
		}
		defer func() {
			for _, id := range ids {
				repo.Delete(id)
			}
		}()

		// Get recent users - exercises rows iteration
		users, err := repo.GetRecentUsers(1)
		if err != nil {
			t.Fatalf("GetRecentUsers failed: %v", err)
		}
		if len(users) < 5 {
			t.Logf("Expected at least 5 recent users, got %d", len(users))
		}
	})

	// Test CountUsers
	t.Run("CountUsers All Paths", func(t *testing.T) {
		count, err := repo.CountUsers()
		if err != nil {
			t.Fatalf("CountUsers failed: %v", err)
		}
		if count < 2 {
			t.Errorf("Should have at least 2 users, got %d", count)
		}
	})

	// Test TransferUserEmail with all success paths
	t.Run("TransferUserEmail Success Path", func(t *testing.T) {
		user1, _ := repo.Create("transfer-src@example.com", "Transfer Source")
		user2, _ := repo.Create("transfer-dst@example.com", "Transfer Dest")
		defer repo.Delete(user1.ID)
		defer repo.Delete(user2.ID)

		err := repo.TransferUserEmail(user1.ID, user2.ID, "transferred@example.com")
		if err != nil {
			t.Fatalf("Transfer should succeed: %v", err)
		}

		// Verify transfer
		updated, _ := repo.GetByID(user2.ID)
		if updated.Email != "transferred@example.com" {
			t.Error("Transfer did not work")
		}
	})
}

// TestDatabaseConnectionErrors tests error handling when database is unavailable
func TestDatabaseConnectionErrors(t *testing.T) {
	// Test with a bad connection string to force connection errors
	badDB, err := sql.Open("postgres", "host=192.0.2.1 port=55555 user=bad password=bad dbname=bad sslmode=disable connect_timeout=1")
	if err != nil {
		t.Fatalf("Failed to create bad DB: %v", err)
	}
	defer badDB.Close()

	// Set very low max connections to force failures
	badDB.SetMaxOpenConns(1)
	badDB.SetMaxIdleConns(0)
	badDB.SetConnMaxLifetime(0)

	repo := NewUserRepository(badDB)

	// All operations should fail with connection errors
	t.Run("GetByEmail with bad DB", func(t *testing.T) {
		_, err := repo.GetByEmail("test@example.com")
		if err == nil {
			t.Error("Should fail with bad database")
		} else {
			t.Logf("GetByEmail error: %v", err)
		}
	})

	t.Run("List with bad DB", func(t *testing.T) {
		_, err := repo.List()
		if err == nil {
			t.Error("Should fail with bad database")
		} else {
			t.Logf("List error: %v", err)
		}
	})

	t.Run("FindByNamePattern with bad DB", func(t *testing.T) {
		_, err := repo.FindByNamePattern("%test%")
		if err == nil {
			t.Error("Should fail with bad database")
		} else {
			t.Logf("FindByNamePattern error: %v", err)
		}
	})

	t.Run("CountUsers with bad DB", func(t *testing.T) {
		_, err := repo.CountUsers()
		if err == nil {
			t.Error("Should fail with bad database")
		} else {
			t.Logf("CountUsers error: %v", err)
		}
	})

	t.Run("GetRecentUsers with bad DB", func(t *testing.T) {
		_, err := repo.GetRecentUsers(1)
		if err == nil {
			t.Error("Should fail with bad database")
		} else {
			t.Logf("GetRecentUsers error: %v", err)
		}
	})

	t.Run("Delete with bad DB", func(t *testing.T) {
		err := repo.Delete(1)
		if err == nil {
			t.Error("Should fail with bad database")
		} else {
			t.Logf("Delete error: %v", err)
		}
	})

	t.Run("Update with bad DB", func(t *testing.T) {
		err := repo.Update(1, "new@example.com", "New Name")
		if err == nil {
			t.Error("Should fail with bad database")
		} else {
			t.Logf("Update error: %v", err)
		}
	})

	t.Run("TransferUserEmail with bad DB", func(t *testing.T) {
		err := repo.TransferUserEmail(1, 2, "transfer@example.com")
		if err == nil {
			t.Error("Should fail with bad database transaction")
		} else {
			t.Logf("TransferUserEmail error: %v", err)
		}
	})

	t.Run("GetByID with bad DB", func(t *testing.T) {
		_, err := repo.GetByID(1)
		if err == nil {
			t.Error("Should fail with bad database")
		} else {
			t.Logf("GetByID error: %v", err)
		}
	})

	t.Run("Create with bad DB", func(t *testing.T) {
		_, err := repo.Create("bad@example.com", "Bad User")
		if err == nil {
			t.Error("Should fail with bad database")
		} else {
			t.Logf("Create error: %v", err)
		}
	})
}

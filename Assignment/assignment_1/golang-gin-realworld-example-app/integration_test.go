package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"realworld-backend/articles"
	"realworld-backend/common"
	"realworld-backend/users"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

var test_db *gorm.DB

// Setup test database and routes
func setupIntegrationTest() (*gin.Engine, *gorm.DB) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Initialize test database
	db := common.TestDBInit()

	// Auto-migrate all models
	db.AutoMigrate(&users.UserModel{})
	db.AutoMigrate(&users.FollowModel{})
	db.AutoMigrate(&articles.ArticleModel{})
	db.AutoMigrate(&articles.ArticleUserModel{})
	db.AutoMigrate(&articles.FavoriteModel{})
	db.AutoMigrate(&articles.TagModel{})
	db.AutoMigrate(&articles.CommentModel{})

	// Setup routes
	r := gin.New()

	// API v1 routes
	v1 := r.Group("/api")

	// Public routes (no auth required, but context needs to be set)
	users.UsersRegister(v1.Group("/users"))

	v1Anon := v1.Group("")
	v1Anon.Use(users.AuthMiddleware(false))
	articles.ArticlesAnonymousRegister(v1Anon.Group("/articles"))
	articles.TagsAnonymousRegister(v1Anon.Group("/tags"))

	// Protected routes (auth required)
	v1Auth := v1.Group("")
	v1Auth.Use(users.AuthMiddleware(true))
	users.UserRegister(v1Auth.Group("/user"))
	users.ProfileRegister(v1Auth.Group("/profiles"))
	articles.ArticlesRegister(v1Auth.Group("/articles"))

	return r, db
}

// Helper function to make authenticated requests
func makeAuthRequest(t *testing.T, r *gin.Engine, method, url, body string, token string) *httptest.ResponseRecorder {
	var req *http.Request
	if body != "" {
		req, _ = http.NewRequest(method, url, bytes.NewBufferString(body))
	} else {
		req, _ = http.NewRequest(method, url, nil)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Token %s", token))
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// =============================================================================
// Authentication Integration Tests
// =============================================================================

func TestUserRegistrationIntegration(t *testing.T) {
	asserts := assert.New(t)
	r, db := setupIntegrationTest()
	defer common.TestDBFree(db)

	// Test user registration with valid data
	registrationData := `{"user":{"username":"testuser","email":"test@example.com","password":"password123"}}`
	w := makeAuthRequest(t, r, "POST", "/api/users/", registrationData, "")

	asserts.Equal(http.StatusCreated, w.Code, "Should return 201 Created")

	// Parse response
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	user := response["user"].(map[string]interface{})
	asserts.Equal("testuser", user["username"], "Username should match")
	asserts.Equal("test@example.com", user["email"], "Email should match")
	asserts.NotEmpty(user["token"], "Token should be present")

	// Verify user is saved in database
	var userModel users.UserModel
	db.Where("email = ?", "test@example.com").First(&userModel)
	asserts.NotZero(userModel.ID, "User should exist in database")
}

func TestUserRegistrationDuplicateEmail(t *testing.T) {
	asserts := assert.New(t)
	r, db := setupIntegrationTest()
	defer common.TestDBFree(db)

	// Register first user
	registrationData := `{"user":{"username":"user1","email":"duplicate@example.com","password":"password123"}}`
	makeAuthRequest(t, r, "POST", "/api/users/", registrationData, "")

	// Try to register with same email
	duplicateData := `{"user":{"username":"user2","email":"duplicate@example.com","password":"password456"}}`
	w := makeAuthRequest(t, r, "POST", "/api/users/", duplicateData, "")

	asserts.Equal(http.StatusUnprocessableEntity, w.Code, "Should return 422 for duplicate email")
}

func TestUserLoginIntegration(t *testing.T) {
	asserts := assert.New(t)
	r, db := setupIntegrationTest()
	defer common.TestDBFree(db)

	// Register user first
	registrationData := `{"user":{"username":"loginuser","email":"login@example.com","password":"password123"}}`
	makeAuthRequest(t, r, "POST", "/api/users/", registrationData, "")

	// Test login with valid credentials
	loginData := `{"user":{"email":"login@example.com","password":"password123"}}`
	w := makeAuthRequest(t, r, "POST", "/api/users/login", loginData, "")

	asserts.Equal(http.StatusOK, w.Code, "Login should return 200 OK")

	// Parse response
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	user := response["user"].(map[string]interface{})
	asserts.Equal("loginuser", user["username"], "Username should match")
	asserts.NotEmpty(user["token"], "JWT token should be returned")
}

func TestUserLoginInvalidCredentials(t *testing.T) {
	asserts := assert.New(t)
	r, db := setupIntegrationTest()
	defer common.TestDBFree(db)

	// Register user
	registrationData := `{"user":{"username":"testuser2","email":"test2@example.com","password":"password123"}}`
	makeAuthRequest(t, r, "POST", "/api/users/", registrationData, "")

	// Test login with wrong password
	loginData := `{"user":{"email":"test2@example.com","password":"wrongpassword"}}`
	w := makeAuthRequest(t, r, "POST", "/api/users/login", loginData, "")

	asserts.Equal(http.StatusForbidden, w.Code, "Should return 403 Forbidden for invalid credentials")
}

func TestGetCurrentUserAuthenticated(t *testing.T) {
	asserts := assert.New(t)
	r, db := setupIntegrationTest()
	defer common.TestDBFree(db)

	// Register and get token
	registrationData := `{"user":{"username":"currentuser","email":"current@example.com","password":"password123"}}`
	regResp := makeAuthRequest(t, r, "POST", "/api/users/", registrationData, "")

	var regResponse map[string]interface{}
	json.Unmarshal(regResp.Body.Bytes(), &regResponse)
	token := regResponse["user"].(map[string]interface{})["token"].(string)

	// Get current user with token
	w := makeAuthRequest(t, r, "GET", "/api/user/", "", token)

	asserts.Equal(http.StatusOK, w.Code, "Should return 200 OK")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	user := response["user"].(map[string]interface{})
	asserts.Equal("currentuser", user["username"], "Should return authenticated user data")
}

func TestGetCurrentUserWithoutToken(t *testing.T) {
	asserts := assert.New(t)
	r, db := setupIntegrationTest()
	defer common.TestDBFree(db)

	// Try to get current user without token
	w := makeAuthRequest(t, r, "GET", "/api/user/", "", "")

	asserts.Equal(http.StatusUnauthorized, w.Code, "Should return 401 Unauthorized")
}

func TestGetCurrentUserWithInvalidToken(t *testing.T) {
	asserts := assert.New(t)
	r, db := setupIntegrationTest()
	defer common.TestDBFree(db)

	// Try with invalid token
	w := makeAuthRequest(t, r, "GET", "/api/user/", "", "invalidtoken123")

	asserts.Equal(http.StatusUnauthorized, w.Code, "Should return 401 for invalid token")
}

// =============================================================================
// Article CRUD Integration Tests
// =============================================================================

func TestCreateArticleAuthenticated(t *testing.T) {
	asserts := assert.New(t)
	r, db := setupIntegrationTest()
	defer common.TestDBFree(db)

	// Register user and get token
	registrationData := `{"user":{"username":"articleauthor","email":"author@example.com","password":"password123"}}`
	regResp := makeAuthRequest(t, r, "POST", "/api/users/", registrationData, "")

	var regResponse map[string]interface{}
	json.Unmarshal(regResp.Body.Bytes(), &regResponse)
	token := regResponse["user"].(map[string]interface{})["token"].(string)

	// Create article
	articleData := `{"article":{"title":"Test Article","description":"Test Description","body":"Test Body","tagList":["test","golang"]}}`
	w := makeAuthRequest(t, r, "POST", "/api/articles/", articleData, token)

	asserts.Equal(http.StatusCreated, w.Code, "Should return 201 Created")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	article := response["article"].(map[string]interface{})
	asserts.Equal("Test Article", article["title"], "Article title should match")
	asserts.Equal("Test Description", article["description"], "Article description should match")
	asserts.NotEmpty(article["slug"], "Slug should be generated")
}

func TestCreateArticleUnauthenticated(t *testing.T) {
	asserts := assert.New(t)
	r, db := setupIntegrationTest()
	defer common.TestDBFree(db)

	// Try to create article without authentication
	articleData := `{"article":{"title":"Test","description":"Test","body":"Test"}}`
	w := makeAuthRequest(t, r, "POST", "/api/articles/", articleData, "")

	asserts.Equal(http.StatusUnauthorized, w.Code, "Should return 401 Unauthorized")
}

func TestListArticles(t *testing.T) {
	asserts := assert.New(t)
	r, db := setupIntegrationTest()
	defer common.TestDBFree(db)

	// Register user and create articles
	registrationData := `{"user":{"username":"listauthor","email":"listauthor@example.com","password":"password123"}}`
	regResp := makeAuthRequest(t, r, "POST", "/api/users/", registrationData, "")

	var regResponse map[string]interface{}
	json.Unmarshal(regResp.Body.Bytes(), &regResponse)
	token := regResponse["user"].(map[string]interface{})["token"].(string)

	// Create multiple articles
	for i := 1; i <= 3; i++ {
		articleData := fmt.Sprintf(`{"article":{"title":"Article %d","description":"Description %d","body":"Body %d"}}`, i, i, i)
		makeAuthRequest(t, r, "POST", "/api/articles/", articleData, token)
	}

	// List articles
	w := makeAuthRequest(t, r, "GET", "/api/articles/", "", "")

	asserts.Equal(http.StatusOK, w.Code, "Should return 200 OK")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	articles := response["articles"].([]interface{})
	asserts.GreaterOrEqual(len(articles), 3, "Should return at least 3 articles")
}

func TestGetSingleArticle(t *testing.T) {
	asserts := assert.New(t)
	r, db := setupIntegrationTest()
	defer common.TestDBFree(db)

	// Register user and create article
	registrationData := `{"user":{"username":"singleauthor","email":"singleauthor@example.com","password":"password123"}}`
	regResp := makeAuthRequest(t, r, "POST", "/api/users/", registrationData, "")

	var regResponse map[string]interface{}
	json.Unmarshal(regResp.Body.Bytes(), &regResponse)
	token := regResponse["user"].(map[string]interface{})["token"].(string)

	// Create article
	articleData := `{"article":{"title":"Single Article Test","description":"Description","body":"Body"}}`
	createResp := makeAuthRequest(t, r, "POST", "/api/articles/", articleData, token)

	var createResponse map[string]interface{}
	json.Unmarshal(createResp.Body.Bytes(), &createResponse)
	slug := createResponse["article"].(map[string]interface{})["slug"].(string)

	// Get single article
	w := makeAuthRequest(t, r, "GET", fmt.Sprintf("/api/articles/%s", slug), "", "")

	asserts.Equal(http.StatusOK, w.Code, "Should return 200 OK")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	article := response["article"].(map[string]interface{})
	asserts.Equal("Single Article Test", article["title"], "Article title should match")
}

func TestUpdateArticleAsAuthor(t *testing.T) {
	asserts := assert.New(t)
	r, db := setupIntegrationTest()
	defer common.TestDBFree(db)

	// Register user and create article
	registrationData := `{"user":{"username":"updateauthor","email":"updateauthor@example.com","password":"password123"}}`
	regResp := makeAuthRequest(t, r, "POST", "/api/users/", registrationData, "")

	var regResponse map[string]interface{}
	json.Unmarshal(regResp.Body.Bytes(), &regResponse)
	token := regResponse["user"].(map[string]interface{})["token"].(string)

	// Create article
	articleData := `{"article":{"title":"Original Title","description":"Original Description","body":"Original Body"}}`
	createResp := makeAuthRequest(t, r, "POST", "/api/articles/", articleData, token)

	var createResponse map[string]interface{}
	json.Unmarshal(createResp.Body.Bytes(), &createResponse)
	slug := createResponse["article"].(map[string]interface{})["slug"].(string)

	// Update article
	updateData := `{"article":{"title":"Updated Title","description":"Updated Description","body":"Updated Body"}}`
	w := makeAuthRequest(t, r, "PUT", fmt.Sprintf("/api/articles/%s", slug), updateData, token)

	asserts.Equal(http.StatusOK, w.Code, "Should return 200 OK")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	article := response["article"].(map[string]interface{})
	asserts.Equal("Updated Title", article["title"], "Title should be updated")
}

func TestDeleteArticleAsAuthor(t *testing.T) {
	asserts := assert.New(t)
	r, db := setupIntegrationTest()
	defer common.TestDBFree(db)

	// Register user and create article
	registrationData := `{"user":{"username":"deleteauthor","email":"deleteauthor@example.com","password":"password123"}}`
	regResp := makeAuthRequest(t, r, "POST", "/api/users/", registrationData, "")

	var regResponse map[string]interface{}
	json.Unmarshal(regResp.Body.Bytes(), &regResponse)
	token := regResponse["user"].(map[string]interface{})["token"].(string)

	// Create article
	articleData := `{"article":{"title":"Delete Me","description":"Description","body":"Body"}}`
	createResp := makeAuthRequest(t, r, "POST", "/api/articles/", articleData, token)

	var createResponse map[string]interface{}
	json.Unmarshal(createResp.Body.Bytes(), &createResponse)
	slug := createResponse["article"].(map[string]interface{})["slug"].(string)

	// Delete article
	w := makeAuthRequest(t, r, "DELETE", fmt.Sprintf("/api/articles/%s", slug), "", token)

	asserts.Equal(http.StatusOK, w.Code, "Should return 200 OK")

	// Verify article is deleted
	getResp := makeAuthRequest(t, r, "GET", fmt.Sprintf("/api/articles/%s", slug), "", "")
	asserts.Equal(http.StatusNotFound, getResp.Code, "Article should not be found after deletion")
}

// =============================================================================
// Article Interaction Integration Tests
// =============================================================================

func TestFavoriteArticle(t *testing.T) {
	asserts := assert.New(t)
	r, db := setupIntegrationTest()
	defer common.TestDBFree(db)

	// Register author and create article
	authorData := `{"user":{"username":"favauthor","email":"favauthor@example.com","password":"password123"}}`
	authorResp := makeAuthRequest(t, r, "POST", "/api/users/", authorData, "")
	var authorResponse map[string]interface{}
	json.Unmarshal(authorResp.Body.Bytes(), &authorResponse)
	authorToken := authorResponse["user"].(map[string]interface{})["token"].(string)

	articleData := `{"article":{"title":"Favorite Test","description":"Description","body":"Body"}}`
	createResp := makeAuthRequest(t, r, "POST", "/api/articles/", articleData, authorToken)
	var createResponse map[string]interface{}
	json.Unmarshal(createResp.Body.Bytes(), &createResponse)
	slug := createResponse["article"].(map[string]interface{})["slug"].(string)

	// Register another user to favorite
	userData := `{"user":{"username":"favoriter","email":"favoriter@example.com","password":"password123"}}`
	userResp := makeAuthRequest(t, r, "POST", "/api/users/", userData, "")
	var userResponse map[string]interface{}
	json.Unmarshal(userResp.Body.Bytes(), &userResponse)
	userToken := userResponse["user"].(map[string]interface{})["token"].(string)

	// Favorite article
	w := makeAuthRequest(t, r, "POST", fmt.Sprintf("/api/articles/%s/favorite", slug), "", userToken)

	asserts.Equal(http.StatusOK, w.Code, "Should return 200 OK")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	article := response["article"].(map[string]interface{})
	asserts.True(article["favorited"].(bool), "Article should be favorited")
	asserts.Equal(float64(1), article["favoritesCount"].(float64), "Favorites count should be 1")
}

func TestUnfavoriteArticle(t *testing.T) {
	asserts := assert.New(t)
	r, db := setupIntegrationTest()
	defer common.TestDBFree(db)

	// Register author and create article
	authorData := `{"user":{"username":"unfavauthor","email":"unfavauthor@example.com","password":"password123"}}`
	authorResp := makeAuthRequest(t, r, "POST", "/api/users/", authorData, "")
	var authorResponse map[string]interface{}
	json.Unmarshal(authorResp.Body.Bytes(), &authorResponse)
	authorToken := authorResponse["user"].(map[string]interface{})["token"].(string)

	articleData := `{"article":{"title":"Unfavorite Test","description":"Description","body":"Body"}}`
	createResp := makeAuthRequest(t, r, "POST", "/api/articles/", articleData, authorToken)
	var createResponse map[string]interface{}
	json.Unmarshal(createResp.Body.Bytes(), &createResponse)
	slug := createResponse["article"].(map[string]interface{})["slug"].(string)

	// Register another user
	userData := `{"user":{"username":"unfavoriter","email":"unfavoriter@example.com","password":"password123"}}`
	userResp := makeAuthRequest(t, r, "POST", "/api/users/", userData, "")
	var userResponse map[string]interface{}
	json.Unmarshal(userResp.Body.Bytes(), &userResponse)
	userToken := userResponse["user"].(map[string]interface{})["token"].(string)

	// Favorite first
	makeAuthRequest(t, r, "POST", fmt.Sprintf("/api/articles/%s/favorite", slug), "", userToken)

	// Then unfavorite
	w := makeAuthRequest(t, r, "DELETE", fmt.Sprintf("/api/articles/%s/favorite", slug), "", userToken)

	asserts.Equal(http.StatusOK, w.Code, "Should return 200 OK")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	article := response["article"].(map[string]interface{})
	asserts.False(article["favorited"].(bool), "Article should not be favorited")
	asserts.Equal(float64(0), article["favoritesCount"].(float64), "Favorites count should be 0")
}

func TestCreateComment(t *testing.T) {
	asserts := assert.New(t)
	r, db := setupIntegrationTest()
	defer common.TestDBFree(db)

	// Register author and create article
	authorData := `{"user":{"username":"commentauthor","email":"commentauthor@example.com","password":"password123"}}`
	authorResp := makeAuthRequest(t, r, "POST", "/api/users/", authorData, "")
	var authorResponse map[string]interface{}
	json.Unmarshal(authorResp.Body.Bytes(), &authorResponse)
	authorToken := authorResponse["user"].(map[string]interface{})["token"].(string)

	articleData := `{"article":{"title":"Comment Test","description":"Description","body":"Body"}}`
	createResp := makeAuthRequest(t, r, "POST", "/api/articles/", articleData, authorToken)
	var createResponse map[string]interface{}
	json.Unmarshal(createResp.Body.Bytes(), &createResponse)
	slug := createResponse["article"].(map[string]interface{})["slug"].(string)

	// Create comment
	commentData := `{"comment":{"body":"This is a test comment"}}`
	w := makeAuthRequest(t, r, "POST", fmt.Sprintf("/api/articles/%s/comments", slug), commentData, authorToken)

	asserts.Equal(http.StatusCreated, w.Code, "Should return 201 Created")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	comment := response["comment"].(map[string]interface{})
	asserts.Equal("This is a test comment", comment["body"], "Comment body should match")
	asserts.NotZero(comment["id"], "Comment should have an ID")
}

func TestListComments(t *testing.T) {
	asserts := assert.New(t)
	r, db := setupIntegrationTest()
	defer common.TestDBFree(db)

	// Register user and create article
	userData := `{"user":{"username":"listcomauthor","email":"listcomauthor@example.com","password":"password123"}}`
	userResp := makeAuthRequest(t, r, "POST", "/api/users/", userData, "")
	var userResponse map[string]interface{}
	json.Unmarshal(userResp.Body.Bytes(), &userResponse)
	token := userResponse["user"].(map[string]interface{})["token"].(string)

	articleData := `{"article":{"title":"List Comments","description":"Description","body":"Body"}}`
	createResp := makeAuthRequest(t, r, "POST", "/api/articles/", articleData, token)
	var createResponse map[string]interface{}
	json.Unmarshal(createResp.Body.Bytes(), &createResponse)
	slug := createResponse["article"].(map[string]interface{})["slug"].(string)

	// Create multiple comments
	for i := 1; i <= 3; i++ {
		commentData := fmt.Sprintf(`{"comment":{"body":"Comment %d"}}`, i)
		makeAuthRequest(t, r, "POST", fmt.Sprintf("/api/articles/%s/comments", slug), commentData, token)
	}

	// List comments
	w := makeAuthRequest(t, r, "GET", fmt.Sprintf("/api/articles/%s/comments", slug), "", "")

	asserts.Equal(http.StatusOK, w.Code, "Should return 200 OK")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	comments := response["comments"].([]interface{})
	asserts.GreaterOrEqual(len(comments), 3, "Should have at least 3 comments")
}

func TestDeleteComment(t *testing.T) {
	asserts := assert.New(t)
	r, db := setupIntegrationTest()
	defer common.TestDBFree(db)

	// Register user and create article
	userData := `{"user":{"username":"delcomauthor","email":"delcomauthor@example.com","password":"password123"}}`
	userResp := makeAuthRequest(t, r, "POST", "/api/users/", userData, "")
	var userResponse map[string]interface{}
	json.Unmarshal(userResp.Body.Bytes(), &userResponse)
	token := userResponse["user"].(map[string]interface{})["token"].(string)

	articleData := `{"article":{"title":"Delete Comment","description":"Description","body":"Body"}}`
	createResp := makeAuthRequest(t, r, "POST", "/api/articles/", articleData, token)
	var createResponse map[string]interface{}
	json.Unmarshal(createResp.Body.Bytes(), &createResponse)
	slug := createResponse["article"].(map[string]interface{})["slug"].(string)

	// Create comment
	commentData := `{"comment":{"body":"Delete this comment"}}`
	commentResp := makeAuthRequest(t, r, "POST", fmt.Sprintf("/api/articles/%s/comments", slug), commentData, token)
	var commentResponse map[string]interface{}
	json.Unmarshal(commentResp.Body.Bytes(), &commentResponse)
	commentID := int(commentResponse["comment"].(map[string]interface{})["id"].(float64))

	// Delete comment
	w := makeAuthRequest(t, r, "DELETE", fmt.Sprintf("/api/articles/%s/comments/%d", slug, commentID), "", token)

	asserts.Equal(http.StatusOK, w.Code, "Should return 200 OK")
}

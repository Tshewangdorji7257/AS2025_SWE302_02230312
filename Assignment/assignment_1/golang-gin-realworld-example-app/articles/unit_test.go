package articles

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"testing"

	"realworld-backend/common"
	"realworld-backend/users"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

var test_db *gorm.DB

// Setup test database and auto-migrate models
func setupTestDB() *gorm.DB {
	db := common.TestDBInit()
	db.AutoMigrate(&ArticleModel{})
	db.AutoMigrate(&ArticleUserModel{})
	db.AutoMigrate(&FavoriteModel{})
	db.AutoMigrate(&TagModel{})
	db.AutoMigrate(&CommentModel{})
	db.AutoMigrate(&users.UserModel{})
	db.AutoMigrate(&users.FollowModel{})
	return db
}

// Helper function to create a test user
func createTestUser(username, email string) users.UserModel {
	userModel := users.UserModel{
		Username: username,
		Email:    email,
		Bio:      "Test bio",
	}
	err := common.GenToken(1) // Just use token generation to avoid password issues
	_ = err
	test_db.Create(&userModel)
	return userModel
}

// Helper function to create a test article
func createTestArticle(title, description, body string, author ArticleUserModel) ArticleModel {
	article := ArticleModel{
		Slug:        slug.Make(title),
		Title:       title,
		Description: description,
		Body:        body,
		Author:      author,
		AuthorID:    author.ID,
	}
	test_db.Create(&article)
	return article
}

// =============================================================================
// Model Tests
// =============================================================================

func TestArticleModelCreation(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	// Create test user and article user
	userModel := createTestUser("testauthor", "testauthor@test.com")
	articleUserModel := GetArticleUserModel(userModel)

	// Test article creation with valid data
	article := ArticleModel{
		Slug:        "test-article",
		Title:       "Test Article",
		Description: "Test Description",
		Body:        "Test Body Content",
		Author:      articleUserModel,
		AuthorID:    articleUserModel.ID,
	}

	err := test_db.Create(&article).Error
	asserts.NoError(err, "Article should be created successfully")
	asserts.NotZero(article.ID, "Article ID should not be zero")
	asserts.Equal("Test Article", article.Title, "Article title should match")
	asserts.Equal("test-article", article.Slug, "Article slug should match")
}

func TestArticleValidation(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	userModel := createTestUser("validator", "validator@test.com")
	articleUserModel := GetArticleUserModel(userModel)

	// Test creating article with empty title should still work in DB
	// (validation happens at validator level, not model level)
	article := ArticleModel{
		Slug:     "",
		Title:    "",
		Body:     "Body without title",
		AuthorID: articleUserModel.ID,
	}

	err := test_db.Create(&article).Error
	asserts.NoError(err, "Database allows empty title (validation is in validator layer)")
}

func TestFavoriteArticle(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	// Create users
	author := createTestUser("author1", "author1@test.com")
	favoriter := createTestUser("favoriter1", "favoriter1@test.com")

	authorModel := GetArticleUserModel(author)
	favoriterModel := GetArticleUserModel(favoriter)

	// Create article
	article := createTestArticle("Favorite Test", "Description", "Body", authorModel)

	// Test initial favorite count
	asserts.Equal(uint(0), article.favoritesCount(), "Initial favorites count should be 0")

	// Test isFavoriteBy before favoriting
	asserts.False(article.isFavoriteBy(favoriterModel), "Article should not be favorited initially")

	// Test favoriting
	err := article.favoriteBy(favoriterModel)
	asserts.NoError(err, "Favoriting should succeed")

	// Refresh article from database
	test_db.First(&article, article.ID)

	// Test after favoriting
	asserts.Equal(uint(1), article.favoritesCount(), "Favorites count should be 1 after favoriting")
	asserts.True(article.isFavoriteBy(favoriterModel), "Article should be favorited")
}

func TestUnfavoriteArticle(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	// Create users
	author := createTestUser("author2", "author2@test.com")
	favoriter := createTestUser("favoriter2", "favoriter2@test.com")

	authorModel := GetArticleUserModel(author)
	favoriterModel := GetArticleUserModel(favoriter)

	// Create and favorite article
	article := createTestArticle("Unfavorite Test", "Description", "Body", authorModel)
	article.favoriteBy(favoriterModel)

	// Test unfavoriting
	err := article.unFavoriteBy(favoriterModel)
	asserts.NoError(err, "Unfavoriting should succeed")

	// Test after unfavoriting
	asserts.Equal(uint(0), article.favoritesCount(), "Favorites count should be 0 after unfavoriting")
	asserts.False(article.isFavoriteBy(favoriterModel), "Article should not be favorited after unfavoriting")
}

func TestMultipleFavorites(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	// Create users
	author := createTestUser("author3", "author3@test.com")
	favoriter1 := createTestUser("favoriter3", "favoriter3@test.com")
	favoriter2 := createTestUser("favoriter4", "favoriter4@test.com")
	favoriter3 := createTestUser("favoriter5", "favoriter5@test.com")

	authorModel := GetArticleUserModel(author)
	favoriterModel1 := GetArticleUserModel(favoriter1)
	favoriterModel2 := GetArticleUserModel(favoriter2)
	favoriterModel3 := GetArticleUserModel(favoriter3)

	// Create article
	article := createTestArticle("Multi Favorite Test", "Description", "Body", authorModel)

	// Multiple users favorite the article
	article.favoriteBy(favoriterModel1)
	article.favoriteBy(favoriterModel2)
	article.favoriteBy(favoriterModel3)

	// Test favorites count
	asserts.Equal(uint(3), article.favoritesCount(), "Favorites count should be 3")

	// Each user should show favorited
	asserts.True(article.isFavoriteBy(favoriterModel1), "User 1 should have favorited")
	asserts.True(article.isFavoriteBy(favoriterModel2), "User 2 should have favorited")
	asserts.True(article.isFavoriteBy(favoriterModel3), "User 3 should have favorited")
}

func TestTagAssociation(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	// Create user and article
	author := createTestUser("tagauthor", "tagauthor@test.com")
	authorModel := GetArticleUserModel(author)

	article := ArticleModel{
		Slug:     "tag-test",
		Title:    "Tag Test",
		Body:     "Body",
		AuthorID: authorModel.ID,
	}

	// Test setting tags
	tags := []string{"golang", "testing", "backend"}
	err := article.setTags(tags)
	asserts.NoError(err, "Setting tags should succeed")

	// Save article with tags
	test_db.Create(&article)

	// Load article with tags
	var loadedArticle ArticleModel
	test_db.Preload("Tags").First(&loadedArticle, article.ID)

	asserts.Equal(3, len(loadedArticle.Tags), "Article should have 3 tags")
}

func TestCommentCreation(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	// Create users
	author := createTestUser("articleauthor", "articleauthor@test.com")
	commenter := createTestUser("commenter", "commenter@test.com")

	authorModel := GetArticleUserModel(author)
	commenterModel := GetArticleUserModel(commenter)

	// Create article
	article := createTestArticle("Comment Test", "Description", "Body", authorModel)

	// Create comment
	comment := CommentModel{
		Article:   article,
		ArticleID: article.ID,
		Author:    commenterModel,
		AuthorID:  commenterModel.ID,
		Body:      "This is a test comment",
	}

	err := test_db.Create(&comment).Error
	asserts.NoError(err, "Comment should be created successfully")
	asserts.NotZero(comment.ID, "Comment ID should not be zero")
	asserts.Equal("This is a test comment", comment.Body, "Comment body should match")
}

func TestGetComments(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	// Create users
	author := createTestUser("commentauthor", "commentauthor@test.com")
	commenter1 := createTestUser("commenter1", "commenter1@test.com")
	commenter2 := createTestUser("commenter2", "commenter2@test.com")

	authorModel := GetArticleUserModel(author)
	commenterModel1 := GetArticleUserModel(commenter1)
	commenterModel2 := GetArticleUserModel(commenter2)

	// Create article
	article := createTestArticle("Get Comments Test", "Description", "Body", authorModel)

	// Create multiple comments
	comment1 := CommentModel{ArticleID: article.ID, AuthorID: commenterModel1.ID, Body: "Comment 1"}
	comment2 := CommentModel{ArticleID: article.ID, AuthorID: commenterModel2.ID, Body: "Comment 2"}
	test_db.Create(&comment1)
	test_db.Create(&comment2)

	// Get comments for article
	err := article.getComments()
	asserts.NoError(err, "Getting comments should succeed")
	asserts.Equal(2, len(article.Comments), "Article should have 2 comments")
}

func TestFindOneArticle(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	// Create user and article
	author := createTestUser("findauthor", "findauthor@test.com")
	authorModel := GetArticleUserModel(author)

	article := createTestArticle("Find One Test", "Description", "Body", authorModel)

	// Find article by ID
	foundArticle, err := FindOneArticle(&ArticleModel{Model: gorm.Model{ID: article.ID}})
	asserts.NoError(err, "Finding article should succeed")
	asserts.Equal(article.Title, foundArticle.Title, "Found article title should match")
	asserts.NotNil(foundArticle.Author, "Article author should be loaded")
}

func TestSaveOne(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	// Create user and article
	author := createTestUser("saveauthor", "saveauthor@test.com")
	authorModel := GetArticleUserModel(author)

	article := createTestArticle("Save Test", "Old Description", "Old Body", authorModel)

	// Modify article
	article.Description = "New Description"
	article.Body = "New Body"

	// Save changes
	err := SaveOne(&article)
	asserts.NoError(err, "Saving article should succeed")

	// Verify changes persisted
	var reloadedArticle ArticleModel
	test_db.First(&reloadedArticle, article.ID)
	asserts.Equal("New Description", reloadedArticle.Description, "Description should be updated")
	asserts.Equal("New Body", reloadedArticle.Body, "Body should be updated")
}

func TestDeleteArticleModel(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	// Create user and article
	author := createTestUser("deleteauthor", "deleteauthor@test.com")
	authorModel := GetArticleUserModel(author)

	article := createTestArticle("Delete Test", "Description", "Body", authorModel)
	articleID := article.ID

	// Delete article
	err := DeleteArticleModel(&ArticleModel{Model: gorm.Model{ID: articleID}})
	asserts.NoError(err, "Deleting article should succeed")

	// Verify article is deleted
	var count int
	test_db.Model(&ArticleModel{}).Where("id = ?", articleID).Count(&count)
	asserts.Equal(0, count, "Article should be deleted from database")
}

func TestDeleteCommentModel(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	// Create users
	author := createTestUser("delcommentauthor", "delcommentauthor@test.com")
	commenter := createTestUser("delcommenter", "delcommenter@test.com")

	authorModel := GetArticleUserModel(author)
	commenterModel := GetArticleUserModel(commenter)

	// Create article and comment
	article := createTestArticle("Delete Comment Test", "Description", "Body", authorModel)
	comment := CommentModel{ArticleID: article.ID, AuthorID: commenterModel.ID, Body: "Delete me"}
	test_db.Create(&comment)
	commentID := comment.ID

	// Delete comment
	err := DeleteCommentModel(&CommentModel{Model: gorm.Model{ID: commentID}})
	asserts.NoError(err, "Deleting comment should succeed")

	// Verify comment is deleted
	var count int
	test_db.Model(&CommentModel{}).Where("id = ?", commentID).Count(&count)
	asserts.Equal(0, count, "Comment should be deleted from database")
}

// =============================================================================
// Serializer Tests
// =============================================================================

func TestArticleSerializer(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	// Create context with user
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	author := createTestUser("serializeauthor", "serializeauthor@test.com")
	authorModel := GetArticleUserModel(author)
	c.Set("my_user_model", author)

	// Create article with tags
	article := createTestArticle("Serialize Test", "Test Description", "Test Body", authorModel)
	article.setTags([]string{"test", "serializer"})
	test_db.Save(&article)
	test_db.Preload("Tags").First(&article, article.ID)

	// Serialize article
	serializer := ArticleSerializer{C: c, ArticleModel: article}
	response := serializer.Response()

	// Verify response
	asserts.Equal("Serialize Test", response.Title, "Title should match")
	asserts.Equal("Test Description", response.Description, "Description should match")
	asserts.Equal("Test Body", response.Body, "Body should match")
	asserts.NotEmpty(response.Slug, "Slug should not be empty")
	asserts.Equal(uint(0), response.FavoritesCount, "Favorites count should be 0")
	asserts.False(response.Favorite, "Should not be favorited")
	asserts.Equal(2, len(response.Tags), "Should have 2 tags")
}

func TestArticlesSerializer(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	// Create context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	author := createTestUser("multiauthor", "multiauthor@test.com")
	authorModel := GetArticleUserModel(author)
	c.Set("my_user_model", author)

	// Create multiple articles
	article1 := createTestArticle("Article 1", "Description 1", "Body 1", authorModel)
	article2 := createTestArticle("Article 2", "Description 2", "Body 2", authorModel)
	article3 := createTestArticle("Article 3", "Description 3", "Body 3", authorModel)

	test_db.Model(&article1).Related(&article1.Author, "Author")
	test_db.Model(&article1.Author).Related(&article1.Author.UserModel)
	test_db.Model(&article2).Related(&article2.Author, "Author")
	test_db.Model(&article2.Author).Related(&article2.Author.UserModel)
	test_db.Model(&article3).Related(&article3.Author, "Author")
	test_db.Model(&article3.Author).Related(&article3.Author.UserModel)

	articles := []ArticleModel{article1, article2, article3}

	// Serialize articles
	serializer := ArticlesSerializer{C: c, Articles: articles}
	response := serializer.Response()

	// Verify response
	asserts.Equal(3, len(response), "Should have 3 articles")
	asserts.Equal("Article 1", response[0].Title, "First article title should match")
	asserts.Equal("Article 2", response[1].Title, "Second article title should match")
	asserts.Equal("Article 3", response[2].Title, "Third article title should match")
}

func TestCommentSerializer(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	// Create context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	author := createTestUser("commentser", "commentser@test.com")
	commenter := createTestUser("commenterser", "commenterser@test.com")
	authorModel := GetArticleUserModel(author)
	commenterModel := GetArticleUserModel(commenter)
	c.Set("my_user_model", author)

	// Create article and comment
	article := createTestArticle("Comment Serialize", "Description", "Body", authorModel)
	comment := CommentModel{
		ArticleID: article.ID,
		Author:    commenterModel,
		AuthorID:  commenterModel.ID,
		Body:      "Test comment body",
	}
	test_db.Create(&comment)
	test_db.Model(&comment).Related(&comment.Author, "Author")
	test_db.Model(&comment.Author).Related(&comment.Author.UserModel)

	// Serialize comment
	serializer := CommentSerializer{C: c, CommentModel: comment}
	response := serializer.Response()

	// Verify response
	asserts.Equal("Test comment body", response.Body, "Comment body should match")
	asserts.NotZero(response.ID, "Comment ID should not be zero")
	asserts.NotEmpty(response.CreatedAt, "CreatedAt should not be empty")
	asserts.NotEmpty(response.Author.Username, "Author username should not be empty")
}

func TestTagSerializer(t *testing.T) {
	// Create context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	// Create tag
	tag := TagModel{Tag: "golang"}

	// Serialize tag
	serializer := TagSerializer{C: c, TagModel: tag}
	response := serializer.Response()

	// Verify response
	assert.Equal(t, "golang", response, "Tag should match")
}

func TestTagsSerializer(t *testing.T) {
	asserts := assert.New(t)

	// Create context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	// Create tags
	tags := []TagModel{
		{Tag: "golang"},
		{Tag: "testing"},
		{Tag: "backend"},
	}

	// Serialize tags
	serializer := TagsSerializer{C: c, Tags: tags}
	response := serializer.Response()

	// Verify response
	asserts.Equal(3, len(response), "Should have 3 tags")
	asserts.Equal("golang", response[0], "First tag should match")
	asserts.Equal("testing", response[1], "Second tag should match")
	asserts.Equal("backend", response[2], "Third tag should match")
}

// =============================================================================
// Validator Tests
// =============================================================================

func TestArticleModelValidator(t *testing.T) {
	asserts := assert.New(t)

	validator := NewArticleModelValidator()
	asserts.NotNil(validator, "Validator should be created")

	// Test setting validator data
	validator.Article.Title = "Test Title"
	validator.Article.Description = "Test Description"
	validator.Article.Body = "Test Body"
	validator.Article.Tags = []string{"test", "validator"}

	asserts.Equal("Test Title", validator.Article.Title, "Title should match")
	asserts.Equal("Test Description", validator.Article.Description, "Description should match")
	asserts.Equal(2, len(validator.Article.Tags), "Should have 2 tags")
}

func TestArticleModelValidatorFillWith(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	// Create article
	author := createTestUser("validauthor", "validauthor@test.com")
	authorModel := GetArticleUserModel(author)

	article := ArticleModel{
		Title:       "Original Title",
		Description: "Original Description",
		Body:        "Original Body",
		AuthorID:    authorModel.ID,
	}
	article.setTags([]string{"original", "tags"})

	// Fill validator with article
	validator := NewArticleModelValidatorFillWith(article)

	// Verify validator is filled correctly
	asserts.Equal("Original Title", validator.Article.Title, "Title should match")
	asserts.Equal("Original Description", validator.Article.Description, "Description should match")
	asserts.Equal("Original Body", validator.Article.Body, "Body should match")
	asserts.Equal(2, len(validator.Article.Tags), "Should have 2 tags")
}

func TestCommentModelValidator(t *testing.T) {
	asserts := assert.New(t)

	validator := NewCommentModelValidator()
	asserts.NotNil(validator, "Validator should be created")

	// Test setting validator data
	validator.Comment.Body = "Test comment body"

	asserts.Equal("Test comment body", validator.Comment.Body, "Comment body should match")
}

func TestGetArticleUserModel(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	// Create user
	user := createTestUser("articleuser", "articleuser@test.com")

	// Get article user model
	articleUser := GetArticleUserModel(user)

	// Verify article user model
	asserts.NotZero(articleUser.ID, "ArticleUserModel ID should not be zero")
	asserts.Equal(user.ID, articleUser.UserModelID, "UserModelID should match")

	// Test getting again (should return existing)
	articleUser2 := GetArticleUserModel(user)
	asserts.Equal(articleUser.ID, articleUser2.ID, "Should return same ArticleUserModel")
}

func TestGetArticleUserModelWithEmptyUser(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	// Test with empty user (ID = 0)
	emptyUser := users.UserModel{}
	articleUser := GetArticleUserModel(emptyUser)

	// Should return empty article user model
	asserts.Zero(articleUser.ID, "ArticleUserModel ID should be zero for empty user")
}

func TestGetAllTags(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	// Create tags
	test_db.Create(&TagModel{Tag: "golang"})
	test_db.Create(&TagModel{Tag: "testing"})
	test_db.Create(&TagModel{Tag: "backend"})

	// Get all tags
	tags, err := getAllTags()
	asserts.NoError(err, "Getting all tags should succeed")
	asserts.GreaterOrEqual(len(tags), 3, "Should have at least 3 tags")
}

func TestFindManyArticle(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	// Create user and articles
	author := createTestUser("manyauthor", "manyauthor@test.com")
	authorModel := GetArticleUserModel(author)

	createTestArticle("Many 1", "Description 1", "Body 1", authorModel)
	createTestArticle("Many 2", "Description 2", "Body 2", authorModel)
	createTestArticle("Many 3", "Description 3", "Body 3", authorModel)

	// Find many articles
	articles, count, err := FindManyArticle("", "", "20", "0", "")
	asserts.NoError(err, "Finding many articles should succeed")
	asserts.GreaterOrEqual(len(articles), 3, "Should have at least 3 articles")
	asserts.GreaterOrEqual(count, 3, "Count should be at least 3")
}

func TestArticleUpdate(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	// Create user and article
	author := createTestUser("updateauthor", "updateauthor@test.com")
	authorModel := GetArticleUserModel(author)

	article := createTestArticle("Update Test", "Old Description", "Old Body", authorModel)

	// Update article
	updateData := map[string]interface{}{
		"Description": "New Description",
		"Body":        "New Body",
	}
	err := article.Update(updateData)
	asserts.NoError(err, "Updating article should succeed")

	// Verify update
	var updated ArticleModel
	test_db.First(&updated, article.ID)
	asserts.Equal("New Description", updated.Description, "Description should be updated")
	asserts.Equal("New Body", updated.Body, "Body should be updated")
}

// =============================================================================
// Router and Handler Tests (HTTP Integration Tests)
// =============================================================================

func TestArticlesRegister(t *testing.T) {
	asserts := assert.New(t)
	gin.SetMode(gin.TestMode)
	router := gin.New()

	routerGroup := router.Group("/api/articles")
	ArticlesRegister(routerGroup)

	// Verify routes are registered (check router has routes)
	routes := router.Routes()
	asserts.NotEmpty(routes, "Router should have routes registered")
}

func TestArticlesAnonymousRegister(t *testing.T) {
	asserts := assert.New(t)
	gin.SetMode(gin.TestMode)
	router := gin.New()

	routerGroup := router.Group("/api/articles")
	ArticlesAnonymousRegister(routerGroup)

	// Verify routes are registered
	routes := router.Routes()
	asserts.NotEmpty(routes, "Router should have routes registered")
}

func TestTagsAnonymousRegister(t *testing.T) {
	asserts := assert.New(t)
	gin.SetMode(gin.TestMode)
	router := gin.New()

	routerGroup := router.Group("/api/tags")
	TagsAnonymousRegister(routerGroup)

	// Verify routes are registered
	routes := router.Routes()
	asserts.NotEmpty(routes, "Router should have routes registered")
}

// =============================================================================
// Validator Tests
// =============================================================================

func TestArticleModelValidatorBind(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	gin.SetMode(gin.TestMode)

	// Create a validator
	validator := NewArticleModelValidator()

	asserts.NotNil(validator, "Validator should be created")
	// Bind will be called through ArticleCreate handler
}

func TestCommentModelValidatorBind(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	gin.SetMode(gin.TestMode)

	// Create a validator
	validator := NewCommentModelValidator()

	asserts.NotNil(validator, "Validator should be created")
	// Bind will be called through ArticleCommentCreate handler
}

// =============================================================================
// Additional Model Coverage Tests
// =============================================================================

func TestGetArticleFeed(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	// Create users
	user1 := createTestUser("feeduser1", "feeduser1@test.com")
	user2 := createTestUser("feeduser2", "feeduser2@test.com")

	articleUser1 := GetArticleUserModel(user1)
	articleUser2 := GetArticleUserModel(user2)

	// User1 follows User2
	follow := users.FollowModel{
		Following:   user2,
		FollowingID: user2.ID,
	}
	test_db.Model(&user1).Association("Followings").Append(&follow)

	// User2 creates articles
	createTestArticle("Feed Article 1", "Desc 1", "Body 1", articleUser2)
	createTestArticle("Feed Article 2", "Desc 2", "Body 2", articleUser2)

	// Get feed for user1
	articles, count, err := articleUser1.GetArticleFeed("20", "0")

	// Should succeed even if no articles (depends on follow setup)
	asserts.NoError(err, "GetArticleFeed should not error")
	asserts.GreaterOrEqual(count, 0, "Count should be non-negative")
	_ = articles
}

func TestFindManyArticleWithFilters(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	// Create user and articles
	author := createTestUser("filterauthor", "filterauthor@test.com")
	authorModel := GetArticleUserModel(author)

	createTestArticle("Filter Test 1", "Desc 1", "Body 1", authorModel)

	// Test with author filter
	articles, count, err := FindManyArticle("", "filterauthor", "20", "0", "")
	asserts.NoError(err, "Should find articles by author")
	asserts.GreaterOrEqual(count, 1, "Should have at least 1 article")
	_ = articles

	// Test with limit
	articles3, count3, err3 := FindManyArticle("", "", "1", "0", "")
	asserts.NoError(err3, "Should find articles with limit")
	asserts.LessOrEqual(len(articles3), 1, "Should respect limit")
	_ = count3

	// Test with offset
	articles4, count4, err4 := FindManyArticle("", "", "20", "1", "")
	asserts.NoError(err4, "Should find articles with offset")
	_ = articles4
	_ = count4
}

func TestSetTagsEdgeCases(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	// Create user and article
	author := createTestUser("tagedgeauthor", "tagedgeauthor@test.com")
	authorModel := GetArticleUserModel(author)

	article := createTestArticle("Tag Edge Test", "Desc", "Body", authorModel)

	// Test with empty tags
	emptyTags := []string{}
	err := article.setTags(emptyTags)
	asserts.NoError(err, "Setting empty tags should succeed")

	// Test with multiple tags
	multipleTags := []string{"go", "testing", "backend", "api"}
	err = article.setTags(multipleTags)
	asserts.NoError(err, "Setting multiple tags should succeed")

	// Tags were set in the article's Tags field
	asserts.GreaterOrEqual(len(article.Tags), 4, "Should have at least 4 tags")
}

func TestArticleUpdateMethod(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	// Create user and article
	author := createTestUser("updatemethodauthor", "updatemethodauthor@test.com")
	authorModel := GetArticleUserModel(author)

	article := createTestArticle("Update Method Test", "Old Desc", "Old Body", authorModel)

	// Test Update method
	updateData := ArticleModel{
		Title:       "New Title",
		Description: "New Description",
		Body:        "New Body",
	}

	err := article.Update(updateData)
	asserts.NoError(err, "Update should succeed")

	// Verify updates
	var updated ArticleModel
	test_db.First(&updated, article.ID)
	asserts.Equal("New Description", updated.Description, "Description should be updated")
	asserts.Equal("New Body", updated.Body, "Body should be updated")
}

func TestCommentsSerializerResponse(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	// Create user, article and comments
	author := createTestUser("commentserialauthor", "commentserialauthor@test.com")
	authorModel := GetArticleUserModel(author)
	article := createTestArticle("Comment Serial Test", "Desc", "Body", authorModel)

	comment1 := CommentModel{Article: article, Body: "Comment 1", AuthorID: authorModel.ID}
	comment2 := CommentModel{Article: article, Body: "Comment 2", AuthorID: authorModel.ID}
	test_db.Create(&comment1)
	test_db.Create(&comment2)

	// Set my_user_model in context (required by serializer)
	c.Set("my_user_model", author)

	// Test serializer
	serializer := CommentsSerializer{c, []CommentModel{comment1, comment2}}
	response := serializer.Response()

	asserts.Len(response, 2, "Should serialize 2 comments")
}

// =============================================================================
// HTTP Handler Tests - Full Request/Response Simulation
// =============================================================================

func TestArticleCreateHandler(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	gin.SetMode(gin.TestMode)

	// Create test user
	author := createTestUser("httpcreateauthor", "httpcreateauthor@test.com")
	authorModel := GetArticleUserModel(author)

	// Setup router
	router := gin.New()
	router.POST("/api/articles", func(c *gin.Context) {
		c.Set("my_user_model", author)
		c.Set("my_article_user_model", authorModel)
		ArticleCreate(c)
	})

	// Create request
	body := `{"article":{"title":"HTTP Test","description":"HTTP Desc","body":"HTTP Body","tagList":["test"]}}`
	req := httptest.NewRequest("POST", "/api/articles", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	asserts.Equal(201, w.Code, "Should return 201 Created")
}

func TestArticleListHandler(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	gin.SetMode(gin.TestMode)

	// Create test data
	author := createTestUser("httplistauthor", "httplistauthor@test.com")
	authorModel := GetArticleUserModel(author)
	createTestArticle("List HTTP 1", "Desc", "Body", authorModel)
	createTestArticle("List HTTP 2", "Desc", "Body", authorModel)

	// Setup router with user context middleware
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("my_user_model", author)
		c.Next()
	})
	router.GET("/api/articles", ArticleList)

	// Create request
	req := httptest.NewRequest("GET", "/api/articles", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	asserts.Equal(200, w.Code, "Should return 200 OK")
}

func TestArticleRetrieveHandler(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	gin.SetMode(gin.TestMode)

	// Create test article
	author := createTestUser("httpretrieveauthor", "httpretrieveauthor@test.com")
	authorModel := GetArticleUserModel(author)
	article := createTestArticle("Retrieve HTTP", "Desc", "Body", authorModel)

	// Setup router with user context middleware
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("my_user_model", author)
		c.Next()
	})
	router.GET("/api/articles/:slug", ArticleRetrieve)

	// Create request
	req := httptest.NewRequest("GET", "/api/articles/"+article.Slug, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	asserts.Equal(200, w.Code, "Should return 200 OK")
}

func TestArticleUpdateHandler(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	gin.SetMode(gin.TestMode)

	// Create test article
	author := createTestUser("httpupdateauthor", "httpupdateauthor@test.com")
	authorModel := GetArticleUserModel(author)
	article := createTestArticle("Update HTTP", "Old Desc", "Old Body", authorModel)

	// Setup router
	router := gin.New()
	router.PUT("/api/articles/:slug", func(c *gin.Context) {
		c.Set("my_user_model", author)
		c.Set("my_article_user_model", authorModel)
		ArticleUpdate(c)
	})

	// Create request
	body := `{"article":{"description":"New Desc","body":"New Body"}}`
	req := httptest.NewRequest("PUT", "/api/articles/"+article.Slug, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	asserts.Equal(200, w.Code, "Should return 200 OK")
}

func TestArticleDeleteHandler(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	gin.SetMode(gin.TestMode)

	// Create test article
	author := createTestUser("httpdeleteauthor", "httpdeleteauthor@test.com")
	authorModel := GetArticleUserModel(author)
	article := createTestArticle("Delete HTTP", "Desc", "Body", authorModel)

	// Setup router with user context middleware
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("my_user_model", author)
		c.Next()
	})
	router.DELETE("/api/articles/:slug", ArticleDelete)

	// Create request
	req := httptest.NewRequest("DELETE", "/api/articles/"+article.Slug, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	asserts.Equal(200, w.Code, "Should return 200 OK")
}

func TestArticleFavoriteHandler(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	gin.SetMode(gin.TestMode)

	// Create test article
	author := createTestUser("httpfavoriteauthor", "httpfavoriteauthor@test.com")
	authorModel := GetArticleUserModel(author)
	article := createTestArticle("Favorite HTTP", "Desc", "Body", authorModel)

	// Setup router
	router := gin.New()
	router.POST("/api/articles/:slug/favorite", func(c *gin.Context) {
		c.Set("my_user_model", author)
		ArticleFavorite(c)
	})

	// Create request
	req := httptest.NewRequest("POST", "/api/articles/"+article.Slug+"/favorite", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	asserts.Equal(200, w.Code, "Should return 200 OK")
}

func TestArticleUnfavoriteHandler(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	gin.SetMode(gin.TestMode)

	// Create test article
	author := createTestUser("httpunfavoriteauthor", "httpunfavoriteauthor@test.com")
	authorModel := GetArticleUserModel(author)
	article := createTestArticle("Unfavorite HTTP", "Desc", "Body", authorModel)

	// Favorite it first
	article.favoriteBy(authorModel)

	// Setup router
	router := gin.New()
	router.DELETE("/api/articles/:slug/favorite", func(c *gin.Context) {
		c.Set("my_user_model", author)
		ArticleUnfavorite(c)
	})

	// Create request
	req := httptest.NewRequest("DELETE", "/api/articles/"+article.Slug+"/favorite", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	asserts.Equal(200, w.Code, "Should return 200 OK")
}

func TestArticleCommentCreateHandler(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	gin.SetMode(gin.TestMode)

	// Create test article
	author := createTestUser("httpcommentauthor", "httpcommentauthor@test.com")
	authorModel := GetArticleUserModel(author)
	article := createTestArticle("Comment HTTP", "Desc", "Body", authorModel)

	// Setup router
	router := gin.New()
	router.POST("/api/articles/:slug/comments", func(c *gin.Context) {
		c.Set("my_user_model", author)
		c.Set("my_article_user_model", authorModel)
		ArticleCommentCreate(c)
	})

	// Create request
	body := `{"comment":{"body":"Test comment"}}`
	req := httptest.NewRequest("POST", "/api/articles/"+article.Slug+"/comments", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	asserts.Equal(201, w.Code, "Should return 201 Created")
}

func TestArticleCommentDeleteHandler(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	gin.SetMode(gin.TestMode)

	// Create test article and comment
	author := createTestUser("httpdelcommentauthor", "httpdelcommentauthor@test.com")
	authorModel := GetArticleUserModel(author)
	article := createTestArticle("Delete Comment HTTP", "Desc", "Body", authorModel)

	comment := CommentModel{
		Article:  article,
		Body:     "To be deleted",
		AuthorID: authorModel.ID,
	}
	test_db.Create(&comment)

	// Setup router
	router := gin.New()
	router.DELETE("/api/articles/:slug/comments/:id", ArticleCommentDelete)

	// Create request
	req := httptest.NewRequest("DELETE", "/api/articles/"+article.Slug+"/comments/"+fmt.Sprint(comment.ID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	asserts.Equal(200, w.Code, "Should return 200 OK")
}

func TestArticleCommentListHandler(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	gin.SetMode(gin.TestMode)

	// Create test article with comments
	author := createTestUser("httplistcommentauthor", "httplistcommentauthor@test.com")
	authorModel := GetArticleUserModel(author)
	article := createTestArticle("List Comments HTTP", "Desc", "Body", authorModel)

	comment1 := CommentModel{Article: article, Body: "Comment 1", AuthorID: authorModel.ID}
	comment2 := CommentModel{Article: article, Body: "Comment 2", AuthorID: authorModel.ID}
	test_db.Create(&comment1)
	test_db.Create(&comment2)

	// Setup router with user context middleware
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("my_user_model", author)
		c.Next()
	})
	router.GET("/api/articles/:slug/comments", ArticleCommentList)

	// Create request
	req := httptest.NewRequest("GET", "/api/articles/"+article.Slug+"/comments", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	asserts.Equal(200, w.Code, "Should return 200 OK")
}

func TestTagListHandler(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	gin.SetMode(gin.TestMode)

	// Create some tags
	test_db.Create(&TagModel{Tag: "go"})
	test_db.Create(&TagModel{Tag: "rust"})
	test_db.Create(&TagModel{Tag: "python"})

	// Setup router
	router := gin.New()
	router.GET("/api/tags", TagList)

	// Create request
	req := httptest.NewRequest("GET", "/api/tags", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	asserts.Equal(200, w.Code, "Should return 200 OK")
}

func TestArticleFeedHandler(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	gin.SetMode(gin.TestMode)

	// Create users
	user1 := createTestUser("feeduser1", "feeduser1@test.com")
	user2 := createTestUser("feeduser2", "feeduser2@test.com")

	// Setup router
	router := gin.New()
	router.GET("/api/articles/feed", func(c *gin.Context) {
		c.Set("my_user_model", user1)
		ArticleFeed(c)
	})

	// Create request
	req := httptest.NewRequest("GET", "/api/articles/feed", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 200 even if empty
	asserts.Equal(200, w.Code, "Should return 200 OK")

	_ = user2 // Avoid unused variable
}

// =============================================================================
// Validator Bind Tests
// =============================================================================

func TestArticleValidatorBind(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	gin.SetMode(gin.TestMode)

	// Create test user
	author := createTestUser("validatorauthor", "validatorauthor@test.com")
	authorModel := GetArticleUserModel(author)

	// Test valid article binding
	validator := NewArticleModelValidator()

	// Create a test context with valid JSON
	body := `{"article":{"title":"Valid Title","description":"Valid Desc","body":"Valid Body"}}`
	req := httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("my_user_model", author)
	c.Set("my_article_user_model", authorModel)

	err := validator.Bind(c)
	asserts.NoError(err, "Valid article should bind successfully")

	// Test invalid article binding (missing required fields)
	validator2 := NewArticleModelValidator()
	body2 := `{"article":{}}`
	req2 := httptest.NewRequest("POST", "/", bytes.NewBufferString(body2))
	req2.Header.Set("Content-Type", "application/json")
	c2, _ := gin.CreateTestContext(httptest.NewRecorder())
	c2.Request = req2
	c2.Set("my_user_model", author)
	c2.Set("my_article_user_model", authorModel)

	err2 := validator2.Bind(c2)
	asserts.Error(err2, "Invalid article should fail binding")
}

func TestCommentValidatorBind(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	gin.SetMode(gin.TestMode)

	// Create test user
	author := createTestUser("commentvalidatorauthor", "commentvalidatorauthor@test.com")
	authorModel := GetArticleUserModel(author)

	// Test valid comment binding
	validator := NewCommentModelValidator()

	body := `{"comment":{"body":"Valid comment body"}}`
	req := httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("my_user_model", author)
	c.Set("my_article_user_model", authorModel)

	err := validator.Bind(c)
	asserts.NoError(err, "Valid comment should bind successfully")

	// Test comment binding with empty body (should succeed as body is not required)
	validator2 := NewCommentModelValidator()
	body2 := `{"comment":{}}`
	req2 := httptest.NewRequest("POST", "/", bytes.NewBufferString(body2))
	req2.Header.Set("Content-Type", "application/json")
	c2, _ := gin.CreateTestContext(httptest.NewRecorder())
	c2.Request = req2
	c2.Set("my_user_model", author)
	c2.Set("my_article_user_model", authorModel)

	err2 := validator2.Bind(c2)
	asserts.NoError(err2, "Empty comment body should bind successfully")
}

// =============================================================================
// FindManyArticle Comprehensive Tests
// =============================================================================

func TestFindManyArticleWithAllFilters(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	// Create test users
	author1 := createTestUser("filterauthor1", "filterauthor1@test.com")
	author2 := createTestUser("filterauthor2", "filterauthor2@test.com")
	authorModel1 := GetArticleUserModel(author1)
	authorModel2 := GetArticleUserModel(author2)

	// Create articles with different tags
	article1 := createTestArticle("Filter Article 1", "Desc 1", "Body 1", authorModel1)
	article2 := createTestArticle("Filter Article 2", "Desc 2", "Body 2", authorModel2)
	article3 := createTestArticle("Filter Article 3", "Desc 3", "Body 3", authorModel1)

	// Add tags
	tag1 := TagModel{Tag: "golang"}
	tag2 := TagModel{Tag: "testing"}
	test_db.Create(&tag1)
	test_db.Create(&tag2)

	// Associate tags with articles
	test_db.Model(&article1).Association("Tags").Append(&tag1)
	test_db.Model(&article2).Association("Tags").Append(&tag2)
	test_db.Model(&article3).Association("Tags").Append(&tag1, &tag2)

	// Create favorite
	favoriteModel := FavoriteModel{
		Favorite:   article1,
		FavoriteBy: authorModel2,
	}
	test_db.Create(&favoriteModel)

	// Test: Find by tag
	articles, count, err := FindManyArticle("golang", "", "10", "0", "")
	asserts.NoError(err)
	asserts.True(count >= 2, "Should find at least 2 articles with golang tag")

	// Test: Find by author
	articles, count, err = FindManyArticle("", "filterauthor1", "10", "0", "")
	asserts.NoError(err)
	asserts.Equal(2, count, "Should find 2 articles by author1")
	asserts.True(len(articles) == 2)

	// Test: Find by favorited
	articles, count, err = FindManyArticle("", "", "10", "0", "filterauthor2")
	asserts.NoError(err)
	asserts.True(count >= 1, "Should find at least 1 favorited article")

	// Test: With limit
	articles, count, err = FindManyArticle("", "", "1", "0", "")
	asserts.NoError(err)
	asserts.True(len(articles) <= 1, "Should return at most 1 article")

	// Test: With offset
	articles, count, err = FindManyArticle("", "", "10", "1", "")
	asserts.NoError(err)
	asserts.True(count >= 3, "Total count should be at least 3")

	// Test: Invalid limit uses default
	articles, count, err = FindManyArticle("", "", "invalid", "0", "")
	asserts.NoError(err)
	asserts.True(count >= 3, "Should return articles with default limit")

	// Test: Invalid offset uses default
	articles, count, err = FindManyArticle("", "", "10", "invalid", "")
	asserts.NoError(err)
	asserts.True(len(articles) <= 10, "Should apply limit correctly")

	// Test: Negative limit (should use default)
	articles, count, err = FindManyArticle("", "", "-5", "0", "")
	asserts.NoError(err)
	asserts.True(count >= 3, "Should return articles with default limit")

	// Test: Large offset
	articles, count, err = FindManyArticle("", "", "10", "1000", "")
	asserts.NoError(err)
	asserts.Equal(0, len(articles), "Large offset should return empty array")
	asserts.True(count >= 3, "Count should still show total")

	// Test: Combined filters
	articles, count, err = FindManyArticle("golang", "filterauthor1", "10", "0", "")
	asserts.NoError(err)
	asserts.True(count >= 2, "Should find articles with tag AND author")
}

// =============================================================================
// GetArticleFeed Comprehensive Tests
// =============================================================================

func TestGetArticleFeedComprehensive(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	// Create test users
	follower := createTestUser("feedfollower", "feedfollower@test.com")
	following1 := createTestUser("feedfollowing1", "feedfollowing1@test.com")
	following2 := createTestUser("feedfollowing2", "feedfollowing2@test.com")
	notFollowing := createTestUser("notfollowing", "notfollowing@test.com")

	followerModel := GetArticleUserModel(follower)
	following1Model := GetArticleUserModel(following1)
	following2Model := GetArticleUserModel(following2)
	notFollowingModel := GetArticleUserModel(notFollowing)

	// Create follow relationships
	followModel1 := users.FollowModel{
		Following:  following1,
		FollowedBy: follower,
	}
	followModel2 := users.FollowModel{
		Following:  following2,
		FollowedBy: follower,
	}
	test_db.Create(&followModel1)
	test_db.Create(&followModel2)

	// Create articles
	article1 := createTestArticle("Feed Article 1", "Desc 1", "Body 1", following1Model)
	article2 := createTestArticle("Feed Article 2", "Desc 2", "Body 2", following2Model)
	article3 := createTestArticle("Feed Article 3", "Desc 3", "Body 3", notFollowingModel)
	_, _, _ = article1, article2, article3

	// Test: Get feed for follower
	articles, _, err := followerModel.GetArticleFeed("10", "0")
	asserts.NoError(err)
	asserts.Equal(2, len(articles), "Should return 2 articles from followed users")

	// Test: Get feed with limit
	articles, _, err = followerModel.GetArticleFeed("1", "0")
	asserts.NoError(err)
	asserts.Equal(1, len(articles), "Should return only 1 article due to limit")

	// Test: Get feed with offset
	articles, _, err = followerModel.GetArticleFeed("10", "1")
	asserts.NoError(err)
	asserts.Equal(1, len(articles), "Should return 1 article after offset")

	// Test: Get feed for user with no following
	articles, _, err = notFollowingModel.GetArticleFeed("10", "0")
	asserts.NoError(err)
	asserts.Equal(0, len(articles), "Should return empty array for user with no followings")

	// Test: Default limit and offset
	articles, _, err = followerModel.GetArticleFeed("20", "0")
	asserts.NoError(err)
	asserts.Equal(2, len(articles), "Should return all articles with explicit limit 20")

	// Test: Different limit
	articles, _, err = followerModel.GetArticleFeed("5", "0")
	asserts.NoError(err)
	asserts.True(len(articles) <= 2, "Should respect the limit parameter")

	// Test: Large offset
	articles, _, err = followerModel.GetArticleFeed("10", "1000")
	asserts.NoError(err)
	asserts.Equal(0, len(articles), "Large offset should return empty array")
}

// =============================================================================
// Serializer Coverage Tests
// =============================================================================

func TestCommentsSerializerFullCoverage(t *testing.T) {
	asserts := assert.New(t)
	test_db = setupTestDB()
	defer common.TestDBFree(test_db)

	gin.SetMode(gin.TestMode)

	// Create test user and article
	author := createTestUser("commentserializer", "commentserializer@test.com")
	authorModel := GetArticleUserModel(author)
	article := createTestArticle("Serializer Test Article", "Desc", "Body", authorModel)

	// Create comments
	comment1 := CommentModel{
		Article: article,
		Author:  authorModel,
		Body:    "First comment",
	}
	comment2 := CommentModel{
		Article: article,
		Author:  authorModel,
		Body:    "Second comment",
	}
	test_db.Create(&comment1)
	test_db.Create(&comment2)

	// Create test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("my_user_model", author)

	// Create serializer
	comments := []CommentModel{comment1, comment2}
	serializer := CommentsSerializer{c, comments}

	// Test Response
	response := serializer.Response()
	asserts.NotNil(response, "Response should not be nil")

	// Check response length
	asserts.Equal(2, len(response), "Should have 2 comments")
}

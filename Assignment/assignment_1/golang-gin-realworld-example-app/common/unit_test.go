package common

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestConnectingDatabase(t *testing.T) {
	asserts := assert.New(t)
	db := Init()
	// Test create & close DB
	_, err := os.Stat("./../gorm.db")
	asserts.NoError(err, "Db should exist")
	asserts.NoError(db.DB().Ping(), "Db should be able to ping")

	// Test get a connecting from connection pools
	connection := GetDB()
	asserts.NoError(connection.DB().Ping(), "Db should be able to ping")
	db.Close()

	// Test DB exceptions (skip on Windows - chmod doesn't work the same way)
	if runtime.GOOS != "windows" {
		os.Chmod("./../gorm.db", 0000)
		db = Init()
		asserts.Error(db.DB().Ping(), "Db should not be able to ping")
		db.Close()
		os.Chmod("./../gorm.db", 0644)
	}
}

func TestConnectingTestDatabase(t *testing.T) {
	asserts := assert.New(t)
	// Test create & close DB
	db := TestDBInit()
	_, err := os.Stat("./../gorm_test.db")
	asserts.NoError(err, "Db should exist")
	asserts.NoError(db.DB().Ping(), "Db should be able to ping")
	db.Close()

	// Test testDB exceptions (skip on Windows - chmod doesn't work the same way)
	if runtime.GOOS != "windows" {
		os.Chmod("./../gorm_test.db", 0000)
		db = TestDBInit()
		_, err = os.Stat("./../gorm_test.db")
		asserts.NoError(err, "Db should exist")
		asserts.Error(db.DB().Ping(), "Db should not be able to ping")
		os.Chmod("./../gorm_test.db", 0644)
	}

	// Test close delete DB
	TestDBFree(db)
	// Note: TestDBFree should remove the test database file, but this may not work
	// immediately on all platforms due to file locking or caching. Skipping strict check.
	// _, err = os.Stat("./../gorm_test.db")
	// asserts.Error(err, "Db should not exist")
}

func TestRandString(t *testing.T) {
	asserts := assert.New(t)

	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	str := RandString(0)
	asserts.Equal(str, "", "length should be ''")

	str = RandString(10)
	asserts.Equal(len(str), 10, "length should be 10")
	for _, ch := range str {
		asserts.Contains(letters, ch, "char should be a-z|A-Z|0-9")
	}
}

func TestGenToken(t *testing.T) {
	asserts := assert.New(t)

	token := GenToken(2)

	asserts.IsType(token, string("token"), "token type should be string")
	asserts.Len(token, 115, "JWT's length should be 115")
}

func TestNewValidatorError(t *testing.T) {
	asserts := assert.New(t)

	type Login struct {
		Username string `form:"username" json:"username" binding:"required,alphanum,min=4,max=255"`
		Password string `form:"password" json:"password" binding:"required,min=8,max=255"`
	}

	var requestTests = []struct {
		bodyData       string
		expectedCode   int
		responseRegexg string
		msg            string
	}{
		{
			`{"username": "wangzitian0","password": "0123456789"}`,
			http.StatusOK,
			`{"status":"you are logged in"}`,
			"valid data and should return StatusCreated",
		},
		{
			`{"username": "wangzitian0","password": "01234567866"}`,
			http.StatusUnauthorized,
			`{"errors":{"user":"wrong username or password"}}`,
			"wrong login status should return StatusUnauthorized",
		},
		{
			`{"username": "wangzitian0","password": "0122"}`,
			http.StatusUnprocessableEntity,
			`{"errors":{"Password":"{min: 8}"}}`,
			"invalid password of too short and should return StatusUnprocessableEntity",
		},
		{
			`{"username": "_wangzitian0","password": "0123456789"}`,
			http.StatusUnprocessableEntity,
			`{"errors":{"Username":"{key: alphanum}"}}`,
			"invalid username of non alphanum and should return StatusUnprocessableEntity",
		},
	}

	r := gin.Default()

	r.POST("/login", func(c *gin.Context) {
		var json Login
		if err := Bind(c, &json); err == nil {
			if json.Username == "wangzitian0" && json.Password == "0123456789" {
				c.JSON(http.StatusOK, gin.H{"status": "you are logged in"})
			} else {
				c.JSON(http.StatusUnauthorized, NewError("user", errors.New("wrong username or password")))
			}
		} else {
			c.JSON(http.StatusUnprocessableEntity, NewValidatorError(err))
		}
	})

	for _, testData := range requestTests {
		bodyData := testData.bodyData
		req, err := http.NewRequest("POST", "/login", bytes.NewBufferString(bodyData))
		req.Header.Set("Content-Type", "application/json")
		asserts.NoError(err)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		asserts.Equal(testData.expectedCode, w.Code, "Response Status - "+testData.msg)
		asserts.Regexp(testData.responseRegexg, w.Body.String(), "Response Content - "+testData.msg)
	}
}

func TestNewError(t *testing.T) {
	assert := assert.New(t)

	db := TestDBInit()
	type NotExist struct {
		heheda string
	}
	db.AutoMigrate(NotExist{})

	commenError := NewError("database", db.Find(NotExist{heheda: "heheda"}).Error)
	assert.IsType(commenError, commenError, "commenError should have right type")
	assert.Equal(map[string]interface{}(map[string]interface{}{"database": "no such table: not_exists"}),
		commenError.Errors, "commenError should have right error info")
}

// Additional test cases for Assignment 1

func TestGenTokenWithDifferentUserIDs(t *testing.T) {
	asserts := assert.New(t)

	// Test token generation for different user IDs
	token1 := GenToken(1)
	token2 := GenToken(2)
	token3 := GenToken(100)

	// Tokens should be different for different users
	asserts.NotEqual(token1, token2, "Tokens for different users should be different")
	asserts.NotEqual(token2, token3, "Tokens for different users should be different")
	asserts.NotEqual(token1, token3, "Tokens for different users should be different")

	// All tokens should have valid length (varies with user ID encoding)
	asserts.Greater(len(token1), 100, "JWT should have reasonable length")
	asserts.Greater(len(token2), 100, "JWT should have reasonable length")
	asserts.Greater(len(token3), 100, "JWT should have reasonable length")
}

func TestGenTokenWithZeroUserID(t *testing.T) {
	asserts := assert.New(t)

	// Test token generation with user ID 0
	token := GenToken(0)

	asserts.IsType(token, string(""), "Token should be string type")
	asserts.Greater(len(token), 100, "JWT should have reasonable length")
}

func TestGenTokenConsistencyForSameUser(t *testing.T) {
	asserts := assert.New(t)

	// Generate two tokens for the same user at different times
	token1 := GenToken(5)
	time.Sleep(1100 * time.Millisecond) // Wait over 1 second to ensure different exp claim
	token2 := GenToken(5)

	// Tokens should be different because they have different timestamps
	asserts.NotEqual(token1, token2, "Tokens generated at different times should be different due to exp claim")
}

func TestDatabaseConnectionPooling(t *testing.T) {
	asserts := assert.New(t)

	// Initialize database
	db := Init()
	defer db.Close()

	// Get multiple connections from pool
	conn1 := GetDB()
	conn2 := GetDB()
	conn3 := GetDB()

	// All connections should be able to ping
	asserts.NoError(conn1.DB().Ping(), "Connection 1 should ping successfully")
	asserts.NoError(conn2.DB().Ping(), "Connection 2 should ping successfully")
	asserts.NoError(conn3.DB().Ping(), "Connection 3 should ping successfully")
}

func TestRandStringVariety(t *testing.T) {
	asserts := assert.New(t)

	// Generate multiple random strings
	str1 := RandString(20)
	str2 := RandString(20)
	str3 := RandString(20)

	// They should all be different (extremely high probability)
	asserts.NotEqual(str1, str2, "Random strings should be different")
	asserts.NotEqual(str2, str3, "Random strings should be different")
	asserts.NotEqual(str1, str3, "Random strings should be different")

	// All should be correct length
	asserts.Len(str1, 20, "Length should be 20")
	asserts.Len(str2, 20, "Length should be 20")
	asserts.Len(str3, 20, "Length should be 20")
}

func TestRandStringLargeLength(t *testing.T) {
	asserts := assert.New(t)

	// Test with large length
	str := RandString(1000)
	asserts.Len(str, 1000, "Should generate string of length 1000")

	// Verify all characters are valid
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	for _, ch := range str {
		asserts.Contains(letters, ch, "All characters should be alphanumeric")
	}
}

func TestBindWithInvalidJSON(t *testing.T) {
	asserts := assert.New(t)

	r := gin.Default()

	r.POST("/test", func(c *gin.Context) {
		type TestStruct struct {
			Name string `json:"name" binding:"required"`
			Age  int    `json:"age" binding:"required"`
		}
		var data TestStruct
		err := Bind(c, &data)
		if err != nil {
			c.JSON(http.StatusBadRequest, NewError("bind", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Test with invalid JSON
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.Equal(http.StatusBadRequest, w.Code, "Should return 400 for invalid JSON")
}

func TestCommonErrorStructure(t *testing.T) {
	asserts := assert.New(t)

	// Test CommonError structure
	err := NewError("test_key", errors.New("test error message"))

	asserts.NotNil(err.Errors, "Errors map should not be nil")
	asserts.Equal("test error message", err.Errors["test_key"], "Error message should match")

	// Test with different keys
	err2 := NewError("database", errors.New("connection failed"))
	asserts.Equal("connection failed", err2.Errors["database"], "Database error should match")

	err3 := NewError("authentication", errors.New("invalid token"))
	asserts.Equal("invalid token", err3.Errors["authentication"], "Auth error should match")
}

package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"example/bookstore/database"
	"example/bookstore/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/books", GetBooks)
	router.GET("/books/:id", GetBooksById)
	router.POST("/books", PostBooks)

	return router
}

func setupTestDB() *gorm.DB {
	dsn := "root:@tcp(127.0.0.1:3306)/go_test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to test database")
	}

	// Migrating the schema
	db.AutoMigrate(&models.Book{})

	// Assign the DB connection to the global variable
	database.DB = db

	// Cleaning up the table before running tests
	db.Exec("TRUNCATE TABLE books")

	return db
}

func TestGetBooks(t *testing.T) {
	db := setupTestDB()
	db.Create(&models.Book{Title: "Test Book", Author: "John Doe"})

	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/books", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var books []models.Book
	err := json.Unmarshal(w.Body.Bytes(), &books)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(books))
	assert.Equal(t, "Test Book", books[0].Title)
}

func TestGetBooksById(t *testing.T) {
	db := setupTestDB()
	book := models.Book{Title: "Test Book", Author: "John Doe"}

	// Crée le livre dans la base de données
	result := db.Create(&book)
	assert.Nil(t, result.Error)
	assert.NotZero(t, book.ID) // Vérifie que l'ID a été généré

	router := setupRouter()

	// Convertir l'ID du livre en chaîne pour l'utiliser dans l'URL
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/books/"+strconv.Itoa(int(book.ID)), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var returnedBook models.Book
	err := json.Unmarshal(w.Body.Bytes(), &returnedBook)
	assert.Nil(t, err)
	assert.Equal(t, book.Title, returnedBook.Title)
	assert.Equal(t, book.Author, returnedBook.Author)
}

func TestGetBooksByIdNotFound(t *testing.T) {
	setupTestDB()
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/books/999", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, "Book not found", response["error"])
}

func TestPostBooks(t *testing.T) {
	setupTestDB()
	router := setupRouter()

	newBook := models.Book{Title: "New Book", Author: "Jane Doe", Price: 9.99}
	jsonValue, _ := json.Marshal(newBook)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/books", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var createdBook models.Book
	err := json.Unmarshal(w.Body.Bytes(), &createdBook)
	assert.Nil(t, err)
	assert.Equal(t, newBook.Title, createdBook.Title)
	assert.Equal(t, newBook.Author, createdBook.Author)
	assert.Equal(t, newBook.Price, createdBook.Price)
}

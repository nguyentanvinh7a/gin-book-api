package main

import (
    "context"
    "log"
    "net/http"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v4"
    "github.com/jackc/pgx/v4/pgxpool"
)

var db *pgxpool.Pool

type Book struct {
    ID     string `json:"id"`
    Title  string `json:"title"`
    Author string `json:"author"`
    Year   int    `json:"year"`
}

func getBooks(c *gin.Context) {
    rows, err := db.Query(context.Background(), "SELECT id, title, author, year FROM books")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    var books []Book
    for rows.Next() {
        var book Book
        if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Year); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        books = append(books, book)
    }

    c.JSON(http.StatusOK, books)
}

func getBook(c *gin.Context) {
    id := c.Param("id")
    var book Book
    err := db.QueryRow(context.Background(), "SELECT id, title, author, year FROM books WHERE id=$1", id).Scan(&book.ID, &book.Title, &book.Author, &book.Year)
    if err != nil {
        if err == pgx.ErrNoRows {
            c.JSON(http.StatusNotFound, gin.H{"message": "book not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        }
        return
    }

    c.JSON(http.StatusOK, book)
}

func createBook(c *gin.Context) {
    var newBook Book
    if err := c.ShouldBindJSON(&newBook); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    _, err := db.Exec(context.Background(), "INSERT INTO books (id, title, author, year) VALUES ($1, $2, $3, $4)", newBook.ID, newBook.Title, newBook.Author, newBook.Year)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, newBook)
}

func updateBook(c *gin.Context) {
    id := c.Param("id")
    var updatedBook Book
    if err := c.ShouldBindJSON(&updatedBook); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    _, err := db.Exec(context.Background(), "UPDATE books SET title=$1, author=$2, year=$3 WHERE id=$4", updatedBook.Title, updatedBook.Author, updatedBook.Year, id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, updatedBook)
}

func deleteBook(c *gin.Context) {
    id := c.Param("id")
    _, err := db.Exec(context.Background(), "DELETE FROM books WHERE id=$1", id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "book deleted"})
}

func main() {
    var err error
    db, err = pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatalf("Unable to connect to database: %v\n", err)
    }
    defer db.Close()

    r := gin.Default()
    r.GET("/books", getBooks)
    r.GET("/books/:id", getBook)
    r.POST("/books", createBook)
    r.PUT("/books/:id", updateBook)
    r.DELETE("/books/:id", deleteBook)
    r.Run(":8080")
}

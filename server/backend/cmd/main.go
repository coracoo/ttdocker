package main

import (
    "log"
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"
    "gorm.io/gorm"
    "gorm.io/driver/sqlite"
    "server/handlers"  // 添加这行
)

type Template struct {
    gorm.Model
    Name        string   `json:"name"`
    Category    string   `json:"category"`
    Description string   `json:"description"`
    Version     string   `json:"version"`
    Website     string   `json:"website"`
    Logo        string   `json:"logo"`
    Tutorial    string   `json:"tutorial"`
    Compose     string   `json:"compose"`
    Screenshots []string `json:"screenshots" gorm:"type:json"`
}

func main() {
    db, err := gorm.Open(sqlite.Open("templates.db"), &gorm.Config{})
    if err != nil {
        log.Fatal("failed to connect database")
    }
    
    db.AutoMigrate(&Template{})

    r := gin.Default()

    r.Use(cors.New(cors.Config{
        AllowOrigins: []string{"*"},
        AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders: []string{"Origin", "Content-Type"},
    }))

    r.Static("/uploads", "./uploads")

    api := r.Group("/api")
    {
        api.GET("/templates", handlers.ListTemplates(db))
        api.GET("/templates/:id", handlers.GetTemplate(db))
        api.POST("/templates", handlers.CreateTemplate(db))
        api.PUT("/templates/:id", handlers.UpdateTemplate(db))
        api.DELETE("/templates/:id", handlers.DeleteTemplate(db))
        api.POST("/upload", handlers.UploadFile)
    }

    r.Run(":3002")
}
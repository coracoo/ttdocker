package handlers

import (
    "path/filepath"
    "github.com/gin-gonic/gin"
)

func UploadFile(c *gin.Context) {
    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(400, gin.H{"error": "文件上传失败"})
        return
    }

    filename := filepath.Join("uploads", file.Filename)
    if err := c.SaveUploadedFile(file, filename); err != nil {
        c.JSON(500, gin.H{"error": "文件保存失败"})
        return
    }

    c.JSON(200, gin.H{
        "url": "/uploads/" + file.Filename,
    })
}
package handlers

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
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

func ListTemplates(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var templates []Template
        if err := db.Find(&templates).Error; err != nil {
            c.JSON(500, gin.H{"error": "获取模板列表失败"})
            return
        }
        c.JSON(200, templates)
    }
}

func GetTemplate(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var template Template
        if err := db.First(&template, c.Param("id")).Error; err != nil {
            c.JSON(404, gin.H{"error": "模板不存在"})
            return
        }
        c.JSON(200, template)
    }
}

func CreateTemplate(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var template Template
        if err := c.ShouldBindJSON(&template); err != nil {
            c.JSON(400, gin.H{"error": "无效的请求数据"})
            return
        }
        
        if err := db.Create(&template).Error; err != nil {
            c.JSON(500, gin.H{"error": "创建模板失败"})
            return
        }
        c.JSON(201, template)
    }
}

func UpdateTemplate(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var template Template
        if err := db.First(&template, c.Param("id")).Error; err != nil {
            c.JSON(404, gin.H{"error": "模板不存在"})
            return
        }
        
        if err := c.ShouldBindJSON(&template); err != nil {
            c.JSON(400, gin.H{"error": "无效的请求数据"})
            return
        }
        
        if err := db.Save(&template).Error; err != nil {
            c.JSON(500, gin.H{"error": "更新模板失败"})
            return
        }
        c.JSON(200, template)
    }
}

func DeleteTemplate(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        if err := db.Delete(&Template{}, c.Param("id")).Error; err != nil {
            c.JSON(500, gin.H{"error": "删除模板失败"})
            return
        }
        c.JSON(200, gin.H{"message": "删除成功"})
    }
}
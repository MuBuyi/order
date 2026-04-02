package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    "ordercount/internal/models"
)

// ListNavLinks 列出当前登录用户的导航链接
func ListNavLinks(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        uidVal, _ := c.Get("userID")
        userID, _ := uidVal.(uint)
        if userID == 0 {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
            return
        }

        var links []models.NavLink
        if err := db.Where("user_id = ?", userID).Order("created_at desc").Find(&links).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, links)
    }
}

// SaveNavLink 新增或更新导航链接
func SaveNavLink(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        uidVal, _ := c.Get("userID")
        userID, _ := uidVal.(uint)
        if userID == 0 {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
            return
        }

        var body struct {
            ID      uint   `json:"id"`
            Title   string `json:"title"`
            URL     string `json:"url"`
            Account string `json:"account"`
            Remark  string `json:"remark"`
        }
        if err := c.ShouldBindJSON(&body); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        if body.Title == "" || body.URL == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "title and url are required"})
            return
        }

        var link models.NavLink
        if body.ID != 0 {
            if err := db.Where("id = ? AND user_id = ?", body.ID, userID).First(&link).Error; err != nil {
                if err == gorm.ErrRecordNotFound {
                    // 记录不存在则当作新建
                    link = models.NavLink{UserID: userID}
                } else {
                    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                    return
                }
            }
        } else {
            link.UserID = userID
        }

        link.Title = body.Title
        link.URL = body.URL
        link.Account = body.Account
        link.Remark = body.Remark

        if link.ID == 0 {
            if err := db.Create(&link).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }
        } else {
            if err := db.Save(&link).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }
        }

        c.JSON(http.StatusOK, link)
    }
}

// DeleteNavLink 删除导航链接
func DeleteNavLink(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        uidVal, _ := c.Get("userID")
        userID, _ := uidVal.(uint)
        if userID == 0 {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
            return
        }

        id := c.Param("id")
        if id == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
            return
        }

        if err := db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.NavLink{}).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"success": true})
    }
}

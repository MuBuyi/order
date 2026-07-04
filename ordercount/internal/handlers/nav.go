package handlers

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    "ordercount/internal/models"
)

type navLinkItem struct {
	models.NavLink
	CreatorUsername string `json:"creator_username"`
	CreatorRole     string `json:"creator_role"`
}

// ListNavLinks 列出当前登录用户的导航链接
func ListNavLinks(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        uidVal, _ := c.Get("userID")
        userID, _ := uidVal.(uint)
        roleVal, _ := c.Get("role")
        roleStr, _ := roleVal.(string)
        if userID == 0 {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
            return
        }

        query := db.Model(&models.NavLink{}).
			Select("nav_links.*, users.username as creator_username, users.role as creator_role").
			Joins("LEFT JOIN users ON users.id = nav_links.user_id")
        if strings.TrimSpace(strings.ToLower(roleStr)) != "superadmin" {
            query = query.Where("nav_links.user_id = ?", userID)
        }

        var links []navLinkItem
        if err := query.Order("nav_links.created_at desc").Scan(&links).Error; err != nil {
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
        roleVal, _ := c.Get("role")
        roleStr, _ := roleVal.(string)
        if userID == 0 {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
            return
        }

        var body struct {
            Category string `json:"category"`
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

        category := strings.TrimSpace(body.Category)

        var link models.NavLink
        if body.ID != 0 {
            q := db
            if strings.TrimSpace(strings.ToLower(roleStr)) != "superadmin" {
                q = q.Where("user_id = ?", userID)
            }
            if err := q.First(&link, body.ID).Error; err != nil {
                if err == gorm.ErrRecordNotFound {
                    c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
                    return
                }
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }
        } else {
            link.UserID = userID
        }

        link.Category = category
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
        roleVal, _ := c.Get("role")
        roleStr, _ := roleVal.(string)
        if userID == 0 {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
            return
        }

        id := c.Param("id")
        if id == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
            return
        }

        q := db.Where("id = ?", id)
        if strings.TrimSpace(strings.ToLower(roleStr)) != "superadmin" {
            q = q.Where("user_id = ?", userID)
        }

        if err := q.Delete(&models.NavLink{}).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"success": true})
    }
}

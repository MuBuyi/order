package handlers

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ordercount/internal/models"
)

var orderIDAllowedPattern = regexp.MustCompile(`^[A-Za-z0-9]+$`)
var orderIDHasDigitPattern = regexp.MustCompile(`[0-9]`)

type returnRecordItem struct {
	models.ReturnRecord
	CreatorUsername string `json:"creator_username"`
	CreatorRole     string `json:"creator_role"`
}

// ListReturns 列出当前登录用户的退货记录
func ListReturns(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidVal, _ := c.Get("userID")
		userID, _ := uidVal.(uint)
		roleVal, _ := c.Get("role")
		roleStr, _ := roleVal.(string)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		startDate := strings.TrimSpace(c.Query("start_date"))
		endDate := strings.TrimSpace(c.Query("end_date"))

		var items []returnRecordItem
		q := db.Model(&models.ReturnRecord{}).
			Select("return_records.*, users.username as creator_username, users.role as creator_role").
			Joins("LEFT JOIN users ON users.id = return_records.user_id")
		if strings.ToLower(strings.TrimSpace(roleStr)) != "superadmin" {
			q = q.Where("return_records.user_id = ?", userID)
		}
		if startDate != "" {
			q = q.Where("return_records.return_date >= ?", startDate)
		}
		if endDate != "" {
			q = q.Where("return_records.return_date <= ?", endDate)
		}

		if err := q.Order("return_records.id desc").Scan(&items).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, items)
	}
}

// SaveReturn 新增或更新退货记录
func SaveReturn(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidVal, _ := c.Get("userID")
		userID, _ := uidVal.(uint)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var body models.ReturnRecord
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		body.OrderID = strings.TrimSpace(body.OrderID)
		if body.OrderID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "不能为空", "field": "order_id"})
			return
		}
		if !orderIDAllowedPattern.MatchString(body.OrderID) || !orderIDHasDigitPattern.MatchString(body.OrderID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "需为纯数字或数字+字母组合（仅限大小写字母和数字）", "field": "order_id"})
			return
		}

		var rec models.ReturnRecord
		if body.ID != 0 {
			if err := db.Where("id = ? AND user_id = ?", body.ID, userID).First(&rec).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					c.JSON(http.StatusNotFound, gin.H{"error": "record not found"})
					return
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}
		} else {
			rec.UserID = userID
		}

		// 复制字段
		rec.OrderID = body.OrderID
		rec.Country = body.Country
		rec.StoreName = body.StoreName
		rec.ProductName = body.ProductName
		rec.SKU = body.SKU
		rec.Quantity = body.Quantity
		rec.RefundAmount = body.RefundAmount
		rec.LossAmount = body.LossAmount
		rec.ReturnDate = body.ReturnDate
		rec.Handler = body.Handler
		rec.Remark = body.Remark

		if rec.ID == 0 {
			if err := db.Create(&rec).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			if err := db.Save(&rec).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, rec)
	}
}

// DeleteReturn 删除退货记录
func DeleteReturn(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, _ := c.Get("role")
		roleStr, _ := roleVal.(string)
		if strings.ToLower(strings.TrimSpace(roleStr)) != "superadmin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "仅超级管理员可删除退货记录"})
			return
		}

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

		if err := db.Where("id = ?", id).Delete(&models.ReturnRecord{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

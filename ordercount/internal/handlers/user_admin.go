package handlers

import (
    "net/http"
    "strconv"
    "strings"

    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"

    "ordercount/internal/models"
)

var allowedRoles = map[string]bool{
    "superadmin": true,
    "admin":      true,
    "staff":      true,
}

var allowedPerms = map[string]bool{
    "stats":      true,
    "settlement": true,
    "returns":    true,
    "product":    true,
    "shop":       true,
    "exchange":   true,
    "charts":     true,
    "nav-helper": true,
    "users":      true,
}

// ListUserOptions 返回可用于前端下拉的人名选项（登录即可访问）
func ListUserOptions(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var users []models.User
        if err := db.Order("created_at asc").Find(&users).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        type item struct {
            ID       uint   `json:"id"`
            Username string `json:"username"`
        }

        res := make([]item, 0, len(users))
        for _, u := range users {
            if strings.TrimSpace(u.Username) == "" {
                continue
            }
            res = append(res, item{
                ID:       u.ID,
                Username: u.Username,
            })
        }

        c.JSON(http.StatusOK, res)
    }
}

// ListUsers 列出所有用户（仅超级管理员）
func ListUsers(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var users []models.User
        if err := db.Order("created_at asc").Find(&users).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        // 不返回密码哈希
        type item struct {
            ID          uint     `json:"id"`
            Username    string   `json:"username"`
            Role        string   `json:"role"`
            IsActive    bool     `json:"is_active"`
            Permissions []string `json:"permissions"`
            CreatedAt   string   `json:"created_at"`
        }
        res := make([]item, 0, len(users))
        for _, u := range users {
            var perms []string
            if u.Permissions != "" {
                for _, p := range strings.Split(u.Permissions, ",") {
                    p = strings.TrimSpace(p)
                    if p != "" {
                        perms = append(perms, p)
                    }
                }
            }
            res = append(res, item{
                ID:          u.ID,
                Username:    u.Username,
                Role:        u.Role,
                IsActive:    u.IsActive,
                Permissions: perms,
                CreatedAt:   u.CreatedAt.Format("2006-01-02 15:04"),
            })
        }
        c.JSON(http.StatusOK, res)
    }
}

// CreateUser 创建新用户（仅超级管理员）
func CreateUser(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var body struct {
            Username    string   `json:"username"`
            Password    string   `json:"password"`
            Role        string   `json:"role"`
            Permissions []string `json:"permissions"`
        }
        if err := c.ShouldBindJSON(&body); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
            return
        }
        if body.Username == "" || body.Password == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "用户名和密码不能为空"})
            return
        }
        if !allowedRoles[body.Role] {
            c.JSON(http.StatusBadRequest, gin.H{"error": "角色不合法"})
            return
        }

        var cnt int64
        if err := db.Model(&models.User{}).Where("username = ?", body.Username).Count(&cnt).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        if cnt > 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
            return
        }

        hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
            return
        }

        // 过滤非法权限
        var perms []string
        for _, p := range body.Permissions {
            if allowedPerms[p] {
                perms = append(perms, p)
            }
        }
        u := models.User{
            Username:     body.Username,
            PasswordHash: string(hash),
            Role:         body.Role,
            IsActive:     true,
            Permissions:  strings.Join(perms, ","),
        }
        if err := db.Create(&u).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"id": u.ID})
    }
}

// UpdateUserPermissions 更新用户页面权限（仅超级管理员）
func UpdateUserPermissions(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        idStr := c.Param("id")
        id, err := strconv.Atoi(idStr)
        if err != nil || id <= 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
            return
        }

        var body struct {
            Permissions []string `json:"permissions"`
        }
        if err := c.ShouldBindJSON(&body); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
            return
        }

        var user models.User
        if err := db.First(&user, id).Error; err != nil {
            if err == gorm.ErrRecordNotFound {
                c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            }
            return
        }

        // 防止超级管理员修改自己的权限，避免把自己锁死
        if curIDVal, ok := c.Get("userID"); ok {
            if curID, ok2 := curIDVal.(uint); ok2 && curID == user.ID {
                c.JSON(http.StatusBadRequest, gin.H{"error": "不能修改自己的权限"})
                return
            }
        }

        var perms []string
        for _, p := range body.Permissions {
            if allowedPerms[p] {
                perms = append(perms, p)
            }
        }
        user.Permissions = strings.Join(perms, ",")
        if err := db.Save(&user).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"ok": true})
    }
}

// UpdateUserRole 更新用户角色（仅超级管理员）
func UpdateUserRole(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        idStr := c.Param("id")
        id, err := strconv.Atoi(idStr)
        if err != nil || id <= 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
            return
        }

        var body struct {
            Role string `json:"role"`
        }
        if err := c.ShouldBindJSON(&body); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
            return
        }
        if !allowedRoles[body.Role] {
            c.JSON(http.StatusBadRequest, gin.H{"error": "角色不合法"})
            return
        }

        var user models.User
        if err := db.First(&user, id).Error; err != nil {
            if err == gorm.ErrRecordNotFound {
                c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            }
            return
        }

        // 防止超级管理员把自己降级，避免误锁死
        if curIDVal, ok := c.Get("userID"); ok {
            if curID, ok2 := curIDVal.(uint); ok2 && curID == user.ID && body.Role != "superadmin" {
                c.JSON(http.StatusBadRequest, gin.H{"error": "不能修改自己的超级管理员角色"})
                return
            }
        }

        user.Role = body.Role
        if err := db.Save(&user).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"ok": true})
    }
}

// UpdateUserPassword 更新用户密码（仅超级管理员）
func UpdateUserPassword(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        idStr := c.Param("id")
        id, err := strconv.Atoi(idStr)
        if err != nil || id <= 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
            return
        }

        var body struct {
            Password string `json:"password"`
        }
        if err := c.ShouldBindJSON(&body); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
            return
        }
        if strings.TrimSpace(body.Password) == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "新密码不能为空"})
            return
        }

        var user models.User
        if err := db.First(&user, id).Error; err != nil {
            if err == gorm.ErrRecordNotFound {
                c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            }
            return
        }

        hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
            return
        }

        user.PasswordHash = string(hash)
        if err := db.Save(&user).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"ok": true})
    }
}

// UpdateUserStatus 启用或停用用户（仅超级管理员）
func UpdateUserStatus(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        roleVal, _ := c.Get("role")
        roleStr, _ := roleVal.(string)
        if strings.ToLower(strings.TrimSpace(roleStr)) != "superadmin" {
            c.JSON(http.StatusForbidden, gin.H{"error": "仅超级管理员可以修改用户状态"})
            return
        }

        idStr := c.Param("id")
        id, err := strconv.Atoi(idStr)
        if err != nil || id <= 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
            return
        }

        var body struct {
            IsActive bool `json:"is_active"`
        }
        if err := c.ShouldBindJSON(&body); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
            return
        }

        var user models.User
        if err := db.First(&user, id).Error; err != nil {
            if err == gorm.ErrRecordNotFound {
                c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            }
            return
        }

        if curIDVal, ok := c.Get("userID"); ok {
            if curID, ok2 := curIDVal.(uint); ok2 && curID == user.ID {
                c.JSON(http.StatusBadRequest, gin.H{"error": "不能修改当前登录用户状态"})
                return
            }
        }

        if !body.IsActive && strings.ToLower(strings.TrimSpace(user.Role)) == "superadmin" {
            var cnt int64
            if err := db.Model(&models.User{}).Where("role = ? AND is_active = ?", "superadmin", true).Count(&cnt).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }
            if cnt <= 1 {
                c.JSON(http.StatusBadRequest, gin.H{"error": "至少保留一个启用中的超级管理员"})
                return
            }
        }

        user.IsActive = body.IsActive
        if err := db.Save(&user).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"ok": true})
    }
}

// DeleteUser 删除用户（仅超级管理员）
func DeleteUser(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        roleVal, _ := c.Get("role")
        roleStr, _ := roleVal.(string)
        if strings.ToLower(strings.TrimSpace(roleStr)) != "superadmin" {
            c.JSON(http.StatusForbidden, gin.H{"error": "仅超级管理员可以删除用户"})
            return
        }

        idStr := c.Param("id")
        id, err := strconv.Atoi(idStr)
        if err != nil || id <= 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
            return
        }

        var user models.User
        if err := db.First(&user, id).Error; err != nil {
            if err == gorm.ErrRecordNotFound {
                c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            }
            return
        }

        // 防止删除当前登录用户，避免误锁死
        if curIDVal, ok := c.Get("userID"); ok {
            if curID, ok2 := curIDVal.(uint); ok2 && curID == user.ID {
                c.JSON(http.StatusBadRequest, gin.H{"error": "不能删除当前登录用户"})
                return
            }
        }

        // 防止删除最后一个超级管理员
        if strings.ToLower(strings.TrimSpace(user.Role)) == "superadmin" {
            var cnt int64
            if err := db.Model(&models.User{}).Where("role = ?", "superadmin").Count(&cnt).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }
            if cnt <= 1 {
                c.JSON(http.StatusBadRequest, gin.H{"error": "至少保留一个超级管理员"})
                return
            }
        }

        // 先删除店铺授权关系，再删除用户
        if err := db.Where("user_id = ?", user.ID).Delete(&models.StoreUser{}).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        if err := db.Delete(&user).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"ok": true})
    }
}

package handlers

import (
    "log"
    "net/http"
    "os"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"

    "ordercount/internal/models"
)

var jwtSecret []byte

func init() {
    s := os.Getenv("JWT_SECRET")
    env := os.Getenv("APP_ENV")
    if s == "" {
        // 生产环境必须显式配置 JWT_SECRET，避免使用弱默认值
        if env == "prod" || env == "production" {
            log.Fatal("JWT_SECRET must be set in production environment")
        }
        // 非生产环境仍允许使用默认值，便于本地开发
        s = "ordercount-secret"
    }
    jwtSecret = []byte(s)
}

type UserClaims struct {
    UserID   uint   `json:"uid"`
    Username string `json:"username"`
    Role     string `json:"role"`
    Permissions string `json:"permissions"`
    jwt.RegisteredClaims
}

// RequireRole 要求当前登录用户具备指定角色之一
func RequireRole(roles ...string) gin.HandlerFunc {
    allowed := make(map[string]struct{}, len(roles))
    for _, r := range roles {
        allowed[r] = struct{}{}
    }
    return func(c *gin.Context) {
        val, ok := c.Get("role")
        if !ok {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "无权限"})
            return
        }
        role, _ := val.(string)
        if _, ok := allowed[role]; !ok {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "无权限"})
            return
        }
        c.Next()
    }
}

// RequirePermission 要求当前登录用户具备指定页面权限之一（超级管理员默认放行）
func RequirePermission(perms ...string) gin.HandlerFunc {
    allowed := make(map[string]struct{}, len(perms))
    for _, p := range perms {
        p = strings.ToLower(strings.TrimSpace(p))
        if p != "" {
            allowed[p] = struct{}{}
        }
    }

    return func(c *gin.Context) {
        roleVal, _ := c.Get("role")
        roleStr, _ := roleVal.(string)
        roleStr = strings.ToLower(strings.TrimSpace(roleStr))
        if roleStr == "superadmin" {
            c.Next()
            return
        }

        rawPerms, _ := c.Get("permissions")
        permStr, _ := rawPerms.(string)
        for _, p := range strings.Split(permStr, ",") {
            p = strings.ToLower(strings.TrimSpace(p))
            if p == "" {
                continue
            }
            if _, ok := allowed[p]; ok {
                c.Next()
                return
            }
        }

        c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "无权限"})
    }
}

// Login 用户登录，返回 JWT 和基础用户信息
func Login(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var body struct {
            Username string `json:"username"`
            Password string `json:"password"`
        }
        if err := c.ShouldBindJSON(&body); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
            return
        }
        if body.Username == "" || body.Password == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "用户名和密码不能为空"})
            return
        }

        var u models.User
        if err := db.Where("username = ?", body.Username).First(&u).Error; err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
            return
        }

        if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(body.Password)); err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
            return
        }
        if !u.IsActive {
            c.JSON(http.StatusForbidden, gin.H{"error": "账号已停用，请联系超级管理员"})
            return
        }

        // 生成 token
        now := time.Now()
        claims := UserClaims{
            UserID:   u.ID,
            Username: u.Username,
            Role:     u.Role,
            Permissions: u.Permissions,
            RegisteredClaims: jwt.RegisteredClaims{
                ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
                IssuedAt:  jwt.NewNumericDate(now),
            },
        }
        token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
        signed, err := token.SignedString(jwtSecret)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "token": signed,
            "user": gin.H{
                "id":          u.ID,
                "username":    u.Username,
                "role":        u.Role,
                "is_active":   u.IsActive,
                "permissions": u.Permissions,
            },
        })
    }
}

// AuthMiddleware 解析 JWT，将用户信息放到上下文中
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        auth := c.GetHeader("Authorization")
        if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
            return
        }
        tokenStr := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
        token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
            return jwtSecret, nil
        })
        if err != nil || !token.Valid {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "登录已失效，请重新登录"})
            return
        }
        claims, ok := token.Claims.(*UserClaims)
        if !ok {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "无效令牌"})
            return
        }

        // 写入上下文，后续 handler 可读取
        c.Set("userID", claims.UserID)
        c.Set("username", claims.Username)
        c.Set("role", claims.Role)
        c.Set("permissions", claims.Permissions)

        c.Next()
    }
}

// Me 返回当前登录用户的基础信息
func Me() gin.HandlerFunc {
    return func(c *gin.Context) {
        id, _ := c.Get("userID")
        username, _ := c.Get("username")
        role, _ := c.Get("role")
        c.JSON(http.StatusOK, gin.H{
            "id":       id,
            "username": username,
            "role":     role,
        })
    }
}

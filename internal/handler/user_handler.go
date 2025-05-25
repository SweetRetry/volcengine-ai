package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"jimeng-go-server/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

type CreateUserRequest struct {
	Email string `json:"email" binding:"required,email,max=100" validate:"email"`
	Name  string `json:"name" binding:"required,min=2,max=50" validate:"required,min=2,max=50"`
}

type UpdateUserRequest struct {
	Email string `json:"email" binding:"omitempty,email,max=100" validate:"omitempty,email"`
	Name  string `json:"name" binding:"omitempty,min=2,max=50" validate:"omitempty,min=2,max=50"`
}

// 创建用户
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest

	// 使用新的校验机制
	if errors := ValidateRequest(c, &req); len(errors) > 0 {
		ResponseValidationError(c, errors)
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), req.Email, req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "创建用户失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    user,
		"message": "用户创建成功",
	})
}

// 获取用户信息
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "用户ID不能为空",
		})
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "用户不存在",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
	})
}

// 通过邮箱获取用户
func (h *UserHandler) GetUserByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "邮箱不能为空",
		})
		return
	}

	user, err := h.userService.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "用户不存在",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
	})
}

// 更新用户信息
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "用户ID不能为空",
		})
		return
	}

	var req UpdateUserRequest

	// 使用新的校验机制
	if errors := ValidateRequest(c, &req); len(errors) > 0 {
		ResponseValidationError(c, errors)
		return
	}

	// 获取现有用户信息
	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "用户不存在",
			"message": err.Error(),
		})
		return
	}

	// 更新字段
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Name != "" {
		user.Name = req.Name
	}

	err = h.userService.UpdateUser(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "更新用户失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
		"message": "用户更新成功",
	})
}

// 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "用户ID不能为空",
		})
		return
	}

	err := h.userService.DeleteUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "删除用户失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户删除成功",
	})
}

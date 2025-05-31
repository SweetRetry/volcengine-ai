package handlers

import (
	"github.com/gin-gonic/gin"

	"volcengine-go-server/internal/models"
	"volcengine-go-server/internal/service"
	"volcengine-go-server/internal/util"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

type CreateUserRequest struct {
	Email string `json:"email" binding:"required,email,max=100"`
	Name  string `json:"name" binding:"required,min=2,max=50"`
}

type UpdateUserRequest struct {
	Email string `json:"email" binding:"omitempty,email,max=100"`
	Name  string `json:"name" binding:"omitempty,min=2,max=50"`
}

// 创建用户
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest

	// 使用新的校验机制
	if errors := util.ValidateRequest(c, &req); len(errors) > 0 {
		util.ValidationErrorResponse(c, errors)
		return
	}

	// 检查用户是否已存在
	existingUser, err := h.userService.GetUserByEmail(c.Request.Context(), req.Email)
	if err == nil && existingUser != nil {
		util.BadRequestResponse(c, "创建用户失败", "用户已存在: "+req.Email)
		return
	}

	// 创建用户对象
	user := &models.User{
		Email: req.Email,
		Name:  req.Name,
	}

	err = h.userService.CreateUser(c.Request.Context(), user)
	if err != nil {
		util.BadRequestResponse(c, "创建用户失败", err.Error())
		return
	}

	util.CreatedResponse(c, user, "用户创建成功")
}

// 获取用户信息
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		util.BadRequestResponse(c, "用户ID不能为空", "")
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		util.NotFoundResponse(c, "用户不存在", err.Error())
		return
	}

	util.SuccessResponse(c, user, "")
}

// 通过邮箱获取用户
func (h *UserHandler) GetUserByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		util.BadRequestResponse(c, "邮箱不能为空", "")
		return
	}

	user, err := h.userService.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		util.NotFoundResponse(c, "用户不存在", err.Error())
		return
	}

	util.SuccessResponse(c, user, "")
}

// 更新用户信息
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		util.BadRequestResponse(c, "用户ID不能为空", "")
		return
	}

	var req UpdateUserRequest

	// 使用新的校验机制
	if errors := util.ValidateRequest(c, &req); len(errors) > 0 {
		util.ValidationErrorResponse(c, errors)
		return
	}

	// 获取现有用户信息
	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		util.NotFoundResponse(c, "用户不存在", err.Error())
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
		util.BadRequestResponse(c, "更新用户失败", err.Error())
		return
	}

	util.SuccessResponse(c, user, "用户更新成功")
}

// 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		util.BadRequestResponse(c, "用户ID不能为空", "")
		return
	}

	err := h.userService.DeleteUser(c.Request.Context(), userID)
	if err != nil {
		util.BadRequestResponse(c, "删除用户失败", err.Error())
		return
	}

	util.SuccessResponse(c, nil, "用户删除成功")
}

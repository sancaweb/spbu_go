package handler

import (
	"net/http"
	"spbu_go/internal/entity"
	"spbu_go/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService       service.UserService
	roleService       service.RoleService
	permissionService service.PermissionService
}

func NewUserHandler(userService service.UserService, roleService service.RoleService, permissionService service.PermissionService) *UserHandler {
	return &UserHandler{userService, roleService, permissionService}
}

func (h *UserHandler) Index(c *gin.Context) {
	users, err := h.userService.GetAll()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"Error": err.Error()})
		return
	}

	roles, _ := h.roleService.GetAll()
	permissions, _ := h.permissionService.GetAll() // Fetch all permissions
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")

	c.HTML(http.StatusOK, "users/index.html", gin.H{
		"Users":       users,
		"Roles":       roles,
		"Permissions": permissions,
		"User":        user,
		"Favicon":     favicon,
		"Title":       "User Management",
		"ActiveMenu":  "settings_users",
	})
}

func (h *UserHandler) CreateView(c *gin.Context) {
	roles, _ := h.roleService.GetAll()
	c.HTML(http.StatusOK, "users/create.html", gin.H{
		"Roles": roles,
		"Title": "Create User",
	})
}

func (h *UserHandler) Create(c *gin.Context) {
	roleID, _ := strconv.Atoi(c.PostForm("role_id"))
	user := &entity.User{
		FirstName: c.PostForm("first_name"),
		LastName:  c.PostForm("last_name"),
		Username:  c.PostForm("username"),
		Password:  c.PostForm("password"),
		Email:     c.PostForm("email"),
		Phone:     c.PostForm("phone"),
		IsActive:  c.PostForm("is_active") == "on" || c.PostForm("is_active") == "true",
		RoleID:    uint(roleID),
	}

	if err := h.userService.Create(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "User created successfully"})
}

func (h *UserHandler) EditView(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	user, err := h.userService.GetByID(uint(id))
	if err != nil {
		c.Redirect(http.StatusFound, "/users")
		return
	}

	roles, _ := h.roleService.GetAll()
	c.HTML(http.StatusOK, "users/edit.html", gin.H{
		"User":  user,
		"Roles": roles,
		"Title": "Edit User",
	})
}

func (h *UserHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	roleID, _ := strconv.Atoi(c.PostForm("role_id"))

	user := &entity.User{
		FirstName: c.PostForm("first_name"),
		LastName:  c.PostForm("last_name"),
		Username:  c.PostForm("username"),
		Email:     c.PostForm("email"),
		Phone:     c.PostForm("phone"),
		IsActive:  c.PostForm("is_active") == "on" || c.PostForm("is_active") == "true",
		RoleID:    uint(roleID),
	}

	if password := c.PostForm("password"); password != "" {
		user.Password = password
	}

	if err := h.userService.Update(uint(id), user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "User updated successfully"})
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.userService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "User deleted successfully"})
}

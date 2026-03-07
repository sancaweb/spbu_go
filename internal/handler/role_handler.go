package handler

import (
	"net/http"
	"spbu_go/internal/entity"
	"spbu_go/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	roleService service.RoleService
}

func NewRoleHandler(roleService service.RoleService) *RoleHandler {
	return &RoleHandler{roleService}
}

func (h *RoleHandler) Index(c *gin.Context) {
	roles, err := h.roleService.GetAll()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"Error": err.Error()})
		return
	}

	user, _ := c.Get("user")

	c.HTML(http.StatusOK, "roles/index.html", gin.H{
		"Roles": roles,
		"User":  user,
		"Title": "Role Management",
	})
}

func (h *RoleHandler) CreateView(c *gin.Context) {
	c.HTML(http.StatusOK, "roles/create.html", gin.H{
		"Title": "Create Role",
	})
}

func (h *RoleHandler) Create(c *gin.Context) {
	role := &entity.Role{
		Name: c.PostForm("name"),
		Code: c.PostForm("code"),
	}

	if err := h.roleService.Create(role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Role created successfully"})
}

func (h *RoleHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	role := &entity.Role{
		Name: c.PostForm("name"),
		Code: c.PostForm("code"),
	}

	if err := h.roleService.Update(uint(id), role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Role updated successfully"})
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.roleService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Role deleted successfully"})
}

func (h *RoleHandler) UpdatePermissions(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	// Parse permission IDs from form data (e.g., permissions[]=1&permissions[]=2)
	// Or simplified: comma separated string? standard approach is array in form.
	// Since we use x-www-form-urlencoded in alpine:
	// formData.append('permissions', JSON.stringify(ids)) ? or key[]=val

	// Let's assume sending JSON in body for array data is cleaner, but if we stick to form:
	// c.PostFormArray("permissions")

	permsStr := c.PostFormArray("permissions[]")
	if len(permsStr) == 0 {
		// Fallback for some clients sending without brackets
		permsStr = c.PostFormArray("permissions")
	}

	var permIDs []uint
	for _, p := range permsStr {
		if pid, err := strconv.Atoi(p); err == nil {
			permIDs = append(permIDs, uint(pid))
		}
	}

	if err := h.roleService.UpdatePermissions(uint(id), permIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Permissions updated successfully"})
}

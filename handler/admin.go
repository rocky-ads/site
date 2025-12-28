package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/logger"
	"github.com/rocky-ads/site/ui"
	"github.com/rocky-ads/site/user"
)

func AdminDashboardHandler(c *fiber.Ctx) error {
	// Load users for the default users tab
	sortBy := c.Query("sort", "id")
	sortOrder := c.Query("order", "ASC")
	currentUserID := local.GetUserID(c)

	users, err := user.GetAllUsers(sortBy, sortOrder)
	if err != nil {
		logger.Error("Failed to get users", "error", err)
		return showError(c, "Failed to load users")
	}

	// Render page with users tab active and users table loaded
	page := ui.AdminDashboardPage(users, sortBy, sortOrder, currentUserID)
	return renderPage(c, "Admin Dashboard", page)
}

func AdminTabHandler(c *fiber.Ctx) error {
	tabID := c.Params("tab")
	if tabID != "users" && tabID != "settings" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid tab")
	}

	if tabID == "users" {
		sortBy := c.Query("sort", "id")
		sortOrder := c.Query("order", "ASC")
		currentUserID := local.GetUserID(c)

		users, err := user.GetAllUsers(sortBy, sortOrder)
		if err != nil {
			logger.Error("Failed to get users", "error", err)
			return showError(c, "Failed to load users")
		}

		return render(c, ui.AdminDashboardContainer("users", users, sortBy, sortOrder, currentUserID))
	}

	return render(c, ui.AdminDashboardContainer("settings", nil, "", "", 0))
}

func AdminUsersHandler(c *fiber.Ctx) error {
	sortBy := c.Query("sort", "id")
	sortOrder := c.Query("order", "ASC")
	currentUserID := local.GetUserID(c)

	users, err := user.GetAllUsers(sortBy, sortOrder)
	if err != nil {
		logger.Error("Failed to get users", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to load users")
	}

	return render(c, ui.UsersTable(users, sortBy, sortOrder, currentUserID))
}

func AdminUserDeleteHandler(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
	}

	if err := user.DeleteUser(userID); err != nil {
		logger.Error("Failed to delete user", "error", err, "userID", userID)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete user")
	}

	return refreshUserRow(c)
}

func AdminUserRestoreHandler(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
	}

	if err := user.RestoreUser(userID); err != nil {
		logger.Error("Failed to restore user", "error", err, "userID", userID)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to restore user")
	}

	return refreshUserRow(c)
}

func AdminUserPromoteHandler(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
	}

	if err := user.PromoteToAdmin(userID); err != nil {
		logger.Error("Failed to promote user", "error", err, "userID", userID)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to promote user")
	}

	return refreshUserRow(c)
}

func AdminUserDemoteHandler(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
	}

	currentUserID := local.GetUserID(c)
	if userID == currentUserID {
		return fiber.NewError(fiber.StatusBadRequest, "Cannot demote yourself")
	}

	if err := user.DemoteFromAdmin(userID); err != nil {
		logger.Error("Failed to demote user", "error", err, "userID", userID)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to demote user")
	}

	return refreshUserRow(c)
}

func refreshUserRow(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
	}

	currentUserID := local.GetUserID(c)

	// Load the updated user (including deleted users)
	u, err := user.GetByIDIncludingDeleted(userID)
	if err != nil {
		logger.Error("Failed to get user", "error", err, "userID", userID)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to load user")
	}

	return render(c, ui.UserRow(u, currentUserID))
}

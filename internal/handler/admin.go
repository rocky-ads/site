package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/local"
	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/message"
	"github.com/rocky-ads/site/internal/service/sms"
	"github.com/rocky-ads/site/internal/ui"
	"github.com/rocky-ads/site/internal/user"
)

func userRowData(u user.User) ui.UserRowData {
	return ui.UserRowData{
		ID:        u.ID,
		Name:      u.Name,
		PhoneE64:  u.PhoneE64,
		IsAdmin:   u.IsAdmin,
		CreatedAt: u.CreatedAt,
		DeletedAt: u.DeletedAt,
	}
}

func userRowsData(users []user.User) []ui.UserRowData {
	rows := make([]ui.UserRowData, len(users))
	for i, u := range users {
		rows[i] = userRowData(u)
	}
	return rows
}

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
	page := ui.AdminDashboardPage(userRowsData(users), sortBy, sortOrder, currentUserID)
	return renderPage(c, "Admin Dashboard", page)
}

func AdminTabHandler(c *fiber.Ctx) error {
	tabID := c.Params("tab")
	if tabID != "users" && tabID != "settings" && tabID != "sms-queue" &&
		tabID != "embeddings" && tabID != "clicks" {
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

		return render(c, ui.AdminDashboardContainer("users", userRowsData(users), sortBy, sortOrder, currentUserID))
	}

	if tabID == "sms-queue" {
		return AdminSMSQueueHandler(c)
	}

	if tabID == "embeddings" {
		return AdminEmbeddingsHandler(c)
	}

	if tabID == "clicks" {
		return AdminClicksHandler(c)
	}

	return render(c, ui.AdminDashboardContainer("settings", nil, "", "", 0))
}

func AdminSMSQueueHandler(c *fiber.Ctx) error {
	currentUserID := local.GetUserID(c)
	status := c.Query("status", "all")
	limit := 50
	offset := 0

	// Get queue stats
	stats, err := sms.GetQueueStats()
	if err != nil {
		logger.Error("Failed to get queue stats", "error", err)
		return showError(c, "Failed to load queue statistics")
	}

	// Get queue entries
	queueEntries, err := sms.GetQueueEntries(status, limit, offset)
	if err != nil {
		logger.Error("Failed to get queue entries", "error", err)
		return showError(c, "Failed to load queue entries")
	}

	// Convert queue entries to UI format
	inputs := make([]ui.SMSQueueEntryInput, len(queueEntries))
	for i, entry := range queueEntries {
		inputs[i] = ui.SMSQueueEntryInput{
			ID:            entry.ID,
			RecipientName: entry.RecipientName,
			AdTitle:       entry.AdTitle,
			Status:        entry.Status,
			CreatedAt:     entry.CreatedAt,
			ProcessedAt:   entry.ProcessedAt,
		}
	}
	uiEntries := ui.SMSQueueEntriesFrom(inputs)

	uiStats := ui.QueueStats{
		Pending:    stats.Pending,
		Processed:  stats.Processed,
		Suppressed: stats.Suppressed,
	}

	return render(c, ui.AdminDashboardContainerWithQueue("sms-queue", nil, "", "", currentUserID, uiStats, uiEntries))
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

	return render(c, ui.UsersTable(userRowsData(users), sortBy, sortOrder, currentUserID))
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
	deleteUserAccountPicture(userID)

	convs, err := message.CloseConversationsForDeletedAccount(userID)
	if err != nil {
		logger.Error("Failed to close conversations for deleted account",
			"error", err, "userID", userID)
	} else {
		NotifyConversationsClosed(convs, userID)
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

	return render(c, ui.UserRow(userRowData(u), currentUserID))
}

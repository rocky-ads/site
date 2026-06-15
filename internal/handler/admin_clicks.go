package handler

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/local"
	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/ui"
	"github.com/rocky-ads/site/internal/user"
)

func AdminClicksHandler(c *fiber.Ctx) error {
	currentUserID := local.GetUserID(c)
	data, err := clickAdminData()
	if err != nil {
		logger.Error("Failed to load click admin data", "error", err)
		return showError(c, "Failed to load click data")
	}
	return render(c, ui.AdminDashboardContainerWithClicks(
		"clicks", nil, "", "", currentUserID, data,
	))
}

func clickAdminData() (ui.ClickAdminData, error) {
	summary, err := ad.GetClickSummary()
	if err != nil {
		return ui.ClickAdminData{}, err
	}
	topAds, err := ad.GetTopAdsByClicks(20)
	if err != nil {
		return ui.ClickAdminData{}, err
	}
	topImages, err := ad.GetTopImagesByClicks(20)
	if err != nil {
		return ui.ClickAdminData{}, err
	}
	recent, err := ad.GetRecentClickActivity(25)
	if err != nil {
		return ui.ClickAdminData{}, err
	}
	topUsers, err := ad.GetTopUsersByClicks(15)
	if err != nil {
		return ui.ClickAdminData{}, err
	}
	names, err := userNameMap()
	if err != nil {
		return ui.ClickAdminData{}, err
	}
	return ui.ClickAdminData{
		UsersWithClicks: summary.UsersWithClicks,
		AdsClicked:      summary.AdsClicked,
		AdDetailViews:   summary.AdDetailViews,
		ImageNavClicks:  summary.ImageNavClicks,
		ActiveLast7Days: summary.ActiveLast7Days,
		TopAds:          clickAdRows(topAds),
		TopImages:       clickImageRows(topImages),
		RecentActivity:  clickActivityRows(recent, names),
		TopUsers:        clickUserRows(topUsers, names),
	}, nil
}

func userNameMap() (map[int]string, error) {
	users, err := user.GetAllUsers("id", "ASC")
	if err != nil {
		return nil, err
	}
	names := make(map[int]string, len(users))
	for _, u := range users {
		names[u.ID] = u.Name
	}
	return names, nil
}

func clickAdRows(rows []ad.TopAdClick) []ui.ClickAdRow {
	out := make([]ui.ClickAdRow, len(rows))
	for i, r := range rows {
		out[i] = ui.ClickAdRow{
			AdID:         r.AdID,
			Title:        r.Title,
			CategoryName: r.CategoryName,
			UserCount:    r.UserCount,
			AdViews:      r.AdViews,
			ImageClicks:  r.ImageClicks,
			LastActivity: formatClickTime(r.LastActivity),
		}
	}
	return out
}

func clickImageRows(rows []ad.TopImageClick) []ui.ClickImageRow {
	out := make([]ui.ClickImageRow, len(rows))
	for i, r := range rows {
		out[i] = ui.ClickImageRow{
			AdID:       r.AdID,
			Title:      r.Title,
			ImageIndex: r.ImageIndex,
			UserCount:  r.UserCount,
			Clicks:     r.Clicks,
			LastClick:  formatClickTime(r.LastClickedAt),
		}
	}
	return out
}

func clickActivityRows(
	rows []ad.RecentClick, names map[int]string,
) []ui.ClickActivityRow {
	out := make([]ui.ClickActivityRow, len(rows))
	for i, r := range rows {
		label := "Ad view"
		if r.ClickType == "image" && r.ImageIndex != nil {
			label = fmt.Sprintf("Image %d", *r.ImageIndex)
		}
		name := names[r.UserID]
		if name == "" {
			name = fmt.Sprintf("user #%d", r.UserID)
		}
		out[i] = ui.ClickActivityRow{
			When:       formatClickTime(&r.LastClickedAt),
			UserName:   name,
			UserID:     r.UserID,
			AdID:       r.AdID,
			AdTitle:    r.Title,
			ClickLabel: label,
			ClickCount: r.ClickCount,
		}
	}
	return out
}

func clickUserRows(
	rows []ad.TopUserClick, names map[int]string,
) []ui.ClickUserRow {
	out := make([]ui.ClickUserRow, len(rows))
	for i, r := range rows {
		name := names[r.UserID]
		if name == "" {
			name = fmt.Sprintf("user #%d", r.UserID)
		}
		out[i] = ui.ClickUserRow{
			UserID:      r.UserID,
			UserName:    name,
			AdClicks:    r.AdClicks,
			ImageClicks: r.ImageClicks,
			LastActive:  formatClickTime(r.LastActive),
		}
	}
	return out
}

func formatClickTime(t *time.Time) string {
	if t == nil {
		return "—"
	}
	now := time.Now()
	d := now.Sub(t.In(now.Location()))
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("Jan 2, 2006")
	}
}

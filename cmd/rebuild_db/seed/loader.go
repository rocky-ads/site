package seed

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/nyaruka/phonenumbers"
	adp "github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/currency"
	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/encryption"
	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/password"
)

// categoryJSON represents a category from ad-category.json
type categoryJSON struct {
	Name       string   `json:"name"`
	ImageFile  string   `json:"image_file"`
	SeedAdFile string   `json:"seed_ad_file"`
	Facets     []string `json:"facets"`
}

// CategoryFiles stores file information for a category
type CategoryFiles struct {
	ID         int
	SeedAdFile string
}

var categoryFiles = make(map[string]CategoryFiles)

// LoadAll loads all seed data into the database
func LoadAll(includeTestAds bool) error {
	startTime := time.Now()
	if err := LoadUsers(); err != nil {
		return fmt.Errorf("loading users: %w", err)
	}
	logger.Info("LoadUsers completed", "duration", time.Since(startTime))

	startTime = time.Now()
	if err := LoadCategories(); err != nil {
		return fmt.Errorf("loading categories: %w", err)
	}
	logger.Info("LoadCategories completed", "duration", time.Since(startTime))

	if includeTestAds {
		startTime = time.Now()
		if err := LoadAds(includeTestAds); err != nil {
			return fmt.Errorf("loading ads: %w", err)
		}
		logger.Info("LoadAds completed", "duration", time.Since(startTime))
	}
	return nil
}

// userJSON represents a user from user.json
type userJSON struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
	IsAdmin  bool   `json:"is_admin,omitempty"`
}

// LocationData represents location data from ad JSON files
type LocationData struct {
	City      string  `json:"city"`
	AdminArea string  `json:"admin_area"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Ad represents an ad for loading purposes
type Ad struct {
	ID          int
	CategoryID  int
	Title       string
	Description string
	Price       float64
	CreatedAt   string
	UserID      int
	ImageCount  int
	Location    LocationData
}

// LoadUsers loads users from user.json into the users table
func LoadUsers() error {
	data, err := os.ReadFile("cmd/rebuild_db/seed/user.json")
	if err != nil {
		return err
	}

	var users []userJSON
	if err := json.Unmarshal(data, &users); err != nil {
		return err
	}

	for _, u := range users {
		passwordHash, passwordSalt, err := password.HashPassword(u.Password)
		if err != nil {
			return fmt.Errorf("hashing password for user %s: %w", u.Name, err)
		}

		var phoneE64 string
		num, err := phonenumbers.Parse(u.Phone, "")
		if err != nil {
			if !strings.HasPrefix(u.Phone, "+") {
				num, err = phonenumbers.Parse(u.Phone, "US")
				if err != nil {
					return fmt.Errorf("parsing phone number for user %s: %w", u.Name, err)
				}
			} else {
				return fmt.Errorf("parsing phone number for user %s: %w", u.Name, err)
			}
		}

		if !phonenumbers.IsPossibleNumber(num) {
			return fmt.Errorf("invalid phone number format for user %s: %s", u.Name, u.Phone)
		}

		phoneE64 = phonenumbers.Format(num, phonenumbers.E164)

		result, err := db.Exec(
			"INSERT INTO users (encrypted_name, name_nonce, name_hash, password_hash, password_salt, password_algo, encrypted_phone, phone_nonce, phone_hash, encrypted_email, email_nonce, email_hash, is_admin) VALUES (?, ?, ?, ?, ?, 'argon2id', ?, ?, ?, ?, ?, ?, ?)",
			[]byte{}, []byte{}, db.HashString(u.Name), passwordHash, passwordSalt,
			[]byte{}, []byte{}, db.HashString(phoneE64),
			[]byte{}, []byte{}, nil, u.IsAdmin,
		)
		if err != nil {
			return fmt.Errorf("inserting user %s: %w", u.Name, err)
		}
		userID, _ := result.LastInsertId()

		encryptedName, nameNonce, err := encryption.Encrypt(int(userID), u.Name, config.UserEncryptionKey)
		if err != nil {
			return fmt.Errorf("encrypting name for user %s: %w", u.Name, err)
		}
		nameHash := db.HashString(u.Name)

		encryptedPhone, phoneNonce, err := encryption.Encrypt(int(userID), phoneE64, config.UserEncryptionKey)
		if err != nil {
			return fmt.Errorf("encrypting phone for user %s: %w", u.Name, err)
		}
		phoneHash := db.HashString(phoneE64)

		encryptedEmail, emailNonce, err := encryption.Encrypt(int(userID), "", config.UserEncryptionKey)
		if err != nil {
			return fmt.Errorf("encrypting email for user %s: %w", u.Name, err)
		}
		var emailHash any = nil

		encryptedNameBytes, _ := base64.StdEncoding.DecodeString(encryptedName)
		nameNonceBytes, _ := base64.StdEncoding.DecodeString(nameNonce)
		encryptedPhoneBytes, _ := base64.StdEncoding.DecodeString(encryptedPhone)
		phoneNonceBytes, _ := base64.StdEncoding.DecodeString(phoneNonce)
		encryptedEmailBytes, _ := base64.StdEncoding.DecodeString(encryptedEmail)
		emailNonceBytes, _ := base64.StdEncoding.DecodeString(emailNonce)

		_, err = db.Exec(
			`UPDATE users SET 
				encrypted_name = ?, name_nonce = ?, name_hash = ?,
				encrypted_phone = ?, phone_nonce = ?, phone_hash = ?,
				encrypted_email = ?, email_nonce = ?, email_hash = ?,
				phone_verified = 1
			WHERE id = ?`,
			encryptedNameBytes, nameNonceBytes, nameHash,
			encryptedPhoneBytes, phoneNonceBytes, phoneHash,
			encryptedEmailBytes, emailNonceBytes, emailHash,
			userID,
		)
		if err != nil {
			return fmt.Errorf("updating encrypted fields for user %s: %w", u.Name, err)
		}
	}

	return nil
}

// LoadCategories loads ad-category.json into categories
func LoadCategories() error {
	data, err := os.ReadFile("cmd/rebuild_db/seed/ad-category.json")
	if err != nil {
		return err
	}

	var categories []categoryJSON
	if err := json.Unmarshal(data, &categories); err != nil {
		return err
	}

	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Name < categories[j].Name
	})

	for _, cat := range categories {
		facetKeys := cat.Facets
		if facetKeys == nil {
			facetKeys = []string{}
		}
		facetsJSON, err := json.Marshal(facetKeys)
		if err != nil {
			return fmt.Errorf("marshaling facets for category %s: %w", cat.Name, err)
		}

		result, err := db.Exec(
			"INSERT INTO categories (name, seed_ad_file, image_file, facets) VALUES (?, ?, ?, ?)",
			cat.Name, cat.SeedAdFile, cat.ImageFile, string(facetsJSON),
		)
		if err != nil {
			return fmt.Errorf("inserting category %s: %w", cat.Name, err)
		}

		categoryID, _ := result.LastInsertId()
		categoryFiles[cat.Name] = CategoryFiles{
			ID:         int(categoryID),
			SeedAdFile: cat.SeedAdFile,
		}
	}

	return nil
}

// LoadAds loads ad files into ads table
func LoadAds(includeTestAds bool) error {
	if !includeTestAds {
		return nil
	}

	usedIDs := make(map[int]string)

	categoryNames := make([]string, 0, len(categoryFiles))
	for name := range categoryFiles {
		categoryNames = append(categoryNames, name)
	}
	sort.Strings(categoryNames)

	for _, categoryName := range categoryNames {
		files := categoryFiles[categoryName]
		if err := loadAdsFromFile(files.ID, files.SeedAdFile, usedIDs); err != nil {
			return fmt.Errorf("loading ads from %s for category %s: %w", files.SeedAdFile, categoryName, err)
		}
	}

	return nil
}

// HistoryEntryJSON is one description edit-history block in seed ad files.
type HistoryEntryJSON struct {
	Label string `json:"label"`
	Body  string `json:"body"`
	At    string `json:"at"`
}

// SuggestionJSON is one description tag in seed ad files.
type SuggestionJSON struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// adJSON represents the flat JSON structure from ad-*.json files
type adJSON struct {
	ID                 int                  `json:"id"`
	Title              string               `json:"title"`
	Description        string               `json:"description,omitempty"`
	DescriptionHistory []HistoryEntryJSON   `json:"description_history,omitempty"`
	Suggestions        []SuggestionJSON     `json:"suggestions,omitempty"`
	Price              float64              `json:"price"`
	CreatedAt          string               `json:"created_at"`
	UserID             int                  `json:"user_id"`
	ImageCount         int                  `json:"image_count"`
	Location           LocationData         `json:"location"`
	Facets             map[string]facetJSON `json:"facets,omitempty"`
}

// facetJSON is a single facet value in a seed ad file (maps to ad_facets).
type facetJSON struct {
	Num  *int    `json:"num,omitempty"`
	Text *string `json:"text,omitempty"`
}

var seedHistoryLocation *time.Location

func init() {
	loc, err := time.LoadLocation("America/Detroit")
	if err != nil {
		seedHistoryLocation = time.UTC
		return
	}
	seedHistoryLocation = loc
}

// AssembleDescription applies seed history entries to the original body.
func AssembleDescription(
	original string,
	history []HistoryEntryJSON,
) (string, error) {
	desc := original
	for _, h := range history {
		at, err := time.Parse(time.RFC3339, h.At)
		if err != nil {
			return "", fmt.Errorf("parsing history at %q: %w", h.At, err)
		}
		desc = adp.AppendHistoryEntry(
			desc, h.Label, h.Body, at, seedHistoryLocation,
		)
	}
	return desc, nil
}

func seedSuggestions(aj adJSON) []adp.Suggestion {
	if len(aj.Suggestions) == 0 {
		return nil
	}
	out := make([]adp.Suggestion, 0, len(aj.Suggestions))
	for _, s := range aj.Suggestions {
		out = append(out, adp.Suggestion{Label: s.Label, Value: s.Value})
	}
	return out
}

func convertAdJSON(aj adJSON) Ad {
	return Ad{
		ID:          aj.ID,
		Title:       aj.Title,
		Description: aj.Description,
		Price:       aj.Price,
		CreatedAt:   aj.CreatedAt,
		UserID:      aj.UserID,
		ImageCount:  aj.ImageCount,
		Location:    aj.Location,
	}
}

func loadAdsFromFile(categoryID int, filename string, usedIDs map[int]string) error {
	data, err := os.ReadFile("cmd/rebuild_db/seed/" + filename)
	if err != nil {
		return err
	}

	var adsJSON []adJSON
	if err := json.Unmarshal(data, &adsJSON); err != nil {
		return err
	}

	var testUserID int
	err = db.QueryRow("SELECT id FROM users WHERE name_hash = ?", db.HashString("test")).Scan(&testUserID)
	if err != nil {
		return fmt.Errorf("finding test user: %w", err)
	}

	for i, aj := range adsJSON {
		if aj.ID == 0 {
			return fmt.Errorf("ad at index %d in %s is missing required 'id' field", i, filename)
		}

		adID := aj.ID
		if existingFile, exists := usedIDs[adID]; exists {
			return fmt.Errorf("duplicate ad ID %d: found in both %s and %s", adID, filename, existingFile)
		}
		usedIDs[adID] = filename

		ad := convertAdJSON(aj)

		locationID, err := getOrCreateLocation(aj.Location)
		if err != nil {
			return fmt.Errorf("getting/creating location: %w", err)
		}

		price := int(ad.Price + 0.5)
		priceCurrency := currency.DefaultFromRegion(aj.Location.Country)

		createdAt, err := time.Parse(time.RFC3339, ad.CreatedAt)
		if err != nil {
			return fmt.Errorf("parsing created_at: %w", err)
		}

		description, err := AssembleDescription(
			ad.Description, aj.DescriptionHistory,
		)
		if err != nil {
			return fmt.Errorf("assembling description for ad %d: %w", adID, err)
		}
		suggestionsJSON := adp.SuggestionsJSON(seedSuggestions(aj))

		_, err = db.Exec(
			`INSERT INTO ads (id, category_id, title, description, created_at,
			 user_id, image_count, location_id, suggestions)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			adID, categoryID, ad.Title, description, createdAt,
			testUserID, ad.ImageCount, locationID, suggestionsJSON,
		)
		if err != nil {
			return fmt.Errorf("inserting ad with ID %d: %w", adID, err)
		}

		if err := insertAdFacets(adID, price, priceCurrency, aj.Facets); err != nil {
			return fmt.Errorf("inserting facets for ad %d: %w", adID, err)
		}
	}

	return nil
}

// insertAdFacets writes the price facet (from the ad's top-level price) plus any
// generic facets declared in the seed file into ad_facets.
func insertAdFacets(adID, price int, priceCurrency string, facets map[string]facetJSON) error {
	if _, err := db.Exec(
		`INSERT INTO ad_facets (ad_id, "key", num, "text") VALUES (?, 'price', ?, ?)`,
		adID, price, priceCurrency,
	); err != nil {
		return err
	}
	for key, v := range facets {
		if key == "price" {
			continue
		}
		if _, err := db.Exec(
			`INSERT INTO ad_facets (ad_id, "key", num, "text") VALUES (?, ?, ?, ?)`,
			adID, key, v.Num, v.Text,
		); err != nil {
			return err
		}
	}
	return nil
}

func getOrCreateLocation(loc LocationData) (int, error) {
	rawText := fmt.Sprintf("%s, %s, %s", loc.City, loc.AdminArea, loc.Country)

	var locationID int
	err := db.QueryRow("SELECT id FROM locations WHERE raw_text = ?", rawText).Scan(&locationID)
	if err == nil {
		return locationID, nil
	}

	result, err := db.Exec(
		"INSERT INTO locations (raw_text, city, admin_area, country, latitude, longitude) VALUES (?, ?, ?, ?, ?, ?)",
		rawText, loc.City, loc.AdminArea, loc.Country, loc.Latitude, loc.Longitude,
	)
	if err != nil {
		return 0, fmt.Errorf("inserting location: %w", err)
	}

	locationIDInt64, _ := result.LastInsertId()
	return int(locationIDInt64), nil
}

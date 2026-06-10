package eggopinion

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/message"
	"github.com/rocky-ads/site/internal/service/grok"
	"github.com/rocky-ads/site/internal/user"
)

var ErrUnavailable = errors.New("egg opinion unavailable")

// Opinion is a cached or freshly generated dispute assessment.
type Opinion struct {
	ConversationID   int       `db:"conversation_id"`
	GeneratedAt      time.Time `db:"generated_at"`
	Summary          string    `db:"summary"`
	Assessment       int       `db:"assessment"`
	AssessmentDetail string    `db:"assessment_detail"`
	Resolution       string    `db:"resolution"`
	Reasoning        string    `db:"reasoning"`
}

// GetOrGenerate returns a cached opinion or generates and stores one.
func GetOrGenerate(
	conv message.Conversation,
	loc *time.Location,
) (Opinion, error) {
	if conv.EggThrowerID == nil {
		return Opinion{}, fmt.Errorf("conversation has no egg")
	}

	cached, err := loadOpinion(conv.ID)
	if err == nil {
		return cached, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return Opinion{}, err
	}

	op, err := generate(conv, loc)
	if err != nil {
		return Opinion{}, err
	}
	if err := storeOpinion(op); err != nil {
		return Opinion{}, err
	}
	return loadOpinion(conv.ID)
}

func loadOpinion(conversationID int) (Opinion, error) {
	var op Opinion
	err := db.QueryRow(`
		SELECT conversation_id, generated_at, summary, assessment,
			assessment_detail, resolution, reasoning
		FROM egg_opinions
		WHERE conversation_id = ?
	`, conversationID).Scan(
		&op.ConversationID,
		&op.GeneratedAt,
		&op.Summary,
		&op.Assessment,
		&op.AssessmentDetail,
		&op.Resolution,
		&op.Reasoning,
	)
	if err != nil {
		return Opinion{}, err
	}
	return op, nil
}

func storeOpinion(op Opinion) error {
	_, err := db.Exec(`
		INSERT INTO egg_opinions (
			conversation_id, generated_at, summary, assessment,
			assessment_detail, resolution, reasoning
		) VALUES (?, CURRENT_TIMESTAMP, ?, ?, ?, ?, ?)
	`, op.ConversationID, op.Summary, op.Assessment,
		op.AssessmentDetail, op.Resolution, op.Reasoning)
	if err != nil {
		return fmt.Errorf("store egg opinion: %w", err)
	}
	return nil
}

// Invalidate removes the cached opinion for one conversation.
func Invalidate(conversationID int) error {
	_, err := db.Exec(
		`DELETE FROM egg_opinions WHERE conversation_id = ?`,
		conversationID,
	)
	if err != nil {
		return fmt.Errorf("invalidate egg opinion: %w", err)
	}
	return nil
}

// InvalidateForAd removes cached opinions for all conversations on an ad.
func InvalidateForAd(adID int) error {
	_, err := db.Exec(`
		DELETE FROM egg_opinions
		WHERE conversation_id IN (
			SELECT id FROM conversations WHERE ad_id = ?
		)
	`, adID)
	if err != nil {
		return fmt.Errorf("invalidate egg opinions for ad: %w", err)
	}
	return nil
}

func generate(conv message.Conversation, loc *time.Location) (Opinion, error) {
	a, err := ad.GetAd(0, conv.AdID, loc)
	if err != nil {
		return Opinion{}, fmt.Errorf("load ad: %w", err)
	}
	if err := ad.LoadTags(&a); err != nil {
		return Opinion{}, fmt.Errorf("load ad tags: %w", err)
	}

	category, err := ad.GetCategory(a.CategoryID)
	if err != nil {
		return Opinion{}, fmt.Errorf("load category: %w", err)
	}

	owner, err := user.GetByID(conv.OwnerID)
	if err != nil {
		return Opinion{}, fmt.Errorf("load owner: %w", err)
	}
	enquirer, err := user.GetByID(conv.EnquirerID)
	if err != nil {
		return Opinion{}, fmt.Errorf("load enquirer: %w", err)
	}

	messages, err := listMessages(conv.ID, loc)
	if err != nil {
		return Opinion{}, err
	}

	redacted := make([]message.Message, len(messages))
	for i, m := range messages {
		redacted[i] = m
		content := RedactText(m.Content)
		content = RedactNames(content, owner.Name, enquirer.Name)
		redacted[i].Content = content
	}

	desc := ad.ParseDescriptionForDisplay(a.Description)
	tags := make([]string, len(a.Tags))
	for i, t := range a.Tags {
		tags[i] = t.PromptDisplay()
	}

	thrownAt := time.Now()
	if conv.EggThrownAt != nil {
		thrownAt = *conv.EggThrownAt
	}

	userPrompt := buildUserPrompt(promptInput{
		AdTitle:      a.Title,
		AdOriginal:   desc.Original,
		AdHistory:    desc.History,
		FormalFacets: ad.FormalFacetLines(category, a.Facets),
		Tags:         tags,
		Messages:     redacted,
		OwnerID:      conv.OwnerID,
		EnquirerID:   conv.EnquirerID,
		EggThrowerID: *conv.EggThrowerID,
		EggThrownAt:  thrownAt,
		Loc:          loc,
	})

	resp, err := grok.CallGrokConv(
		opinionSystemPrompt, userPrompt, eggOpinionConvID,
	)
	if err != nil {
		logger.Warn("egg opinion: grok call failed",
			"error", err, "conversationID", conv.ID)
		return Opinion{}, ErrUnavailable
	}

	op, err := parseOpinionResponse(resp)
	if err != nil {
		logger.Warn("egg opinion: parse failed",
			"error", err, "conversationID", conv.ID)
		return Opinion{}, ErrUnavailable
	}

	op.ConversationID = conv.ID
	return op, nil
}

func listMessages(conversationID int, loc *time.Location) ([]message.Message, error) {
	var messages []message.Message
	err := db.Select(&messages, `
		SELECT id, conversation_id, sender_id, content, created_at
		FROM messages
		WHERE conversation_id = ?
		ORDER BY created_at ASC
	`, conversationID)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	for i := range messages {
		if loc != nil {
			messages[i].CreatedAt = messages[i].CreatedAt.In(loc)
		}
	}
	return messages, nil
}

// AdFactLines returns description history lines for UI display.
func AdFactLines(a ad.Ad, loc *time.Location) []string {
	desc := ad.ParseDescriptionForDisplay(a.Description)
	var lines []string
	if desc.Original != "" {
		lines = append(lines, "Original: "+ad.DisplayDescription(desc.Original))
	}
	for _, e := range desc.History {
		line := e.Header
		if e.Body != "" {
			line += " — " + ad.DisplayDescription(e.Body)
		}
		lines = append(lines, line)
	}
	return lines
}

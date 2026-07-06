package testagent

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/rocky-ads/site/internal/browserclient"
)

// Config holds runtime settings for agents.
type Config struct {
	SiteURL  string
	StateDir string
	MinDelay time.Duration
	MaxDelay time.Duration
}

// Agent is one simulated user with a headless browser.
type Agent struct {
	Index    int
	Persona  Persona
	Username string
	Phone    string

	status    Status
	journal   *Journal
	client    *browserclient.Client
	statePath string
	password  string

	mu          sync.Mutex
	cancel      context.CancelFunc
	done        chan struct{}
	currentPath string
	loggedIn    bool
	inbox       *inbox
}

// Snapshot is a read-only view for the control UI.
type Snapshot struct {
	Index       int    `json:"index"`
	Username    string `json:"username"`
	Persona     string `json:"persona"`
	Status      Status `json:"status"`
	LastAction  string `json:"last_action,omitempty"`
	ErrorCount  int    `json:"error_count"`
	CurrentPath string `json:"current_path,omitempty"`
}

// NewAgent creates a stopped agent with identity for index (1-based).
func NewAgent(index int, persona Persona, cfg Config) (*Agent, error) {
	client, err := browserclient.New(cfg.SiteURL)
	if err != nil {
		return nil, err
	}
	statePath := StatePath(cfg.StateDir, index)
	st, err := LoadState(statePath)
	password := ""
	if st != nil {
		password = st.Password
	}
	username := AgentUsername(index)
	phone := AgentPhone(index)
	if st != nil && st.Username != "" {
		username = st.Username
	}
	if st != nil && st.Phone != "" {
		phone = st.Phone
	}
	return &Agent{
		Index:       index,
		Persona:     persona,
		Username:    username,
		Phone:       phone,
		status:      StatusStopped,
		journal:     NewJournal(),
		client:      client,
		statePath:   statePath,
		password:    password,
		currentPath: "/",
		inbox:       newInbox(),
	}, nil
}

// Snapshot returns current agent state for the UI.
func (a *Agent) Snapshot() Snapshot {
	a.mu.Lock()
	defer a.mu.Unlock()
	last := ""
	if e := a.journal.Last(); e != nil {
		last = e.Action
		if e.Error != "" {
			last = e.Error
		}
	}
	return Snapshot{
		Index:       a.Index,
		Username:    a.Username,
		Persona:     a.Persona.Name,
		Status:      a.status,
		LastAction:  last,
		ErrorCount:  a.journal.ErrorCount(),
		CurrentPath: a.currentPath,
	}
}

// Journal returns the agent journal.
func (a *Agent) Journal() *Journal {
	return a.journal
}

// Status returns current status.
func (a *Agent) Status() Status {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.status
}

// Start begins the agent loop.
func (a *Agent) Start(cfg Config) error {
	a.mu.Lock()
	if a.status == StatusRunning {
		a.mu.Unlock()
		return fmt.Errorf("agent %d already running", a.Index)
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	a.done = make(chan struct{})
	a.status = StatusRunning
	a.mu.Unlock()

	go func() {
		defer close(a.done)
		a.run(ctx, cfg)
	}()
	return nil
}

// Stop cancels the agent loop.
func (a *Agent) Stop() {
	a.mu.Lock()
	cancel := a.cancel
	a.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// Wait blocks until the agent goroutine exits.
func (a *Agent) Wait() {
	a.mu.Lock()
	done := a.done
	a.mu.Unlock()
	if done != nil {
		<-done
	}
}

func (a *Agent) run(ctx context.Context, cfg Config) {
	defer func() {
		a.client.Close()
		a.mu.Lock()
		if a.status == StatusRunning {
			a.status = StatusStopped
		}
		a.mu.Unlock()
	}()

	if err := a.client.Start(); err != nil {
		a.journal.Append(JournalEntry{
			Phase: PhaseBoot, Action: "browser start", Error: err.Error(),
		})
		a.mu.Lock()
		a.status = StatusStalled
		a.mu.Unlock()
		return
	}

	if err := a.bootstrap(); err != nil {
		a.journal.Append(JournalEntry{
			Phase: PhaseBoot, Action: "bootstrap", Error: err.Error(),
		})
		a.mu.Lock()
		a.status = StatusStalled
		a.mu.Unlock()
		return
	}
	a.loggedIn = true

	for {
		if ctx.Err() != nil {
			return
		}
		if a.Status() == StatusStalled {
			return
		}
		if err := a.iteration(ctx, cfg); err != nil {
			if ctx.Err() != nil {
				return
			}
			a.journal.Append(JournalEntry{
				Phase: PhaseAct, Action: "iteration", Error: err.Error(),
			})
		}
		delay := randomDelay(cfg.MinDelay, cfg.MaxDelay)
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
	}
}

func (a *Agent) bootstrap() error {
	a.journal.Append(JournalEntry{Phase: PhaseBoot, Action: "bootstrap start"})

	if a.password == "" {
		pw, err := NewPassword()
		if err != nil {
			return err
		}
		a.password = pw
	}

	if err := a.client.Login(a.Username, a.password); err == nil {
		a.journal.Append(JournalEntry{Phase: PhaseBoot, Action: "login"})
		return a.saveState()
	}

	if err := a.client.RegisterTestUser(a.Username, a.Phone, a.password); err != nil {
		if err2 := a.client.Login(a.Username, a.password); err2 != nil {
			return fmt.Errorf("register: %v; login retry: %v", err, err2)
		}
		a.journal.Append(JournalEntry{
			Phase: PhaseBoot, Action: "login after register conflict",
		})
		return a.saveState()
	}
	a.journal.Append(JournalEntry{Phase: PhaseBoot, Action: "register"})
	return a.saveState()
}

func (a *Agent) saveState() error {
	return SaveState(a.statePath, AgentState{
		Username: a.Username,
		Phone:    a.Phone,
		Password: a.password,
	})
}

func (a *Agent) iteration(ctx context.Context, cfg Config) error {
	if err := a.ensureSession(); err != nil {
		a.journal.Append(JournalEntry{
			Phase: PhaseBoot, Action: "ensure session", Error: err.Error(),
		})
		a.mu.Lock()
		a.status = StatusStalled
		a.mu.Unlock()
		return nil
	}

	obs, _ := a.client.Observe()
	path := obs.Path
	page := obs.Page
	a.setPath(path)

	a.syncInbox(obs)

	a.journal.Append(JournalEntry{
		Phase: PhaseObserve, URL: path, Action: "observe " + path,
	})

	a.syncLoggedIn(page)

	omitModals := modalLoopDetected(a.journal.Entries())
	planPage := browserclient.FilterAffordancesForPlanner(page, omitModals)
	if a.loggedIn {
		planPage = browserclient.FilterForLoggedInUser(planPage)
		planPage = browserclient.AppendKnownAuthPaths(planPage)
		planPage = browserclient.FilterSensitiveForms(planPage)
		planPage = a.inbox.enrich(planPage)
		planPage = browserclient.FilterMessageSends(planPage)
		if myAdsTabLoopDetected(a.journal.Entries()) {
			planPage = browserclient.FilterHTMXPrefix(planPage, "/auth/user/myads/tab/")
		}
		if settingsTabLoopDetected(a.journal.Entries()) {
			planPage = browserclient.FilterHTMXPrefix(planPage, "/auth/user/settings/")
		}
	}

	var act PlannedAction
	entries := a.journal.Entries()
	if replyAct, ok := replyReadyAction(page, a.inbox, a.Persona); ok {
		act = replyAct
	} else if replyAct, ok := pendingReplyAction(path, page, a.inbox); ok {
		act = replyAct
	} else if loopPath := repeatedConversationClickPath(entries); loopPath != "" {
		if recoverAct, ok := a.recoverConversationSend(loopPath); ok {
			act = recoverAct
		} else {
			act = PlannedAction{
				Action: "get", Path: "/",
				Reason: "escape conversation click loop",
			}
		}
	} else if unreadAct, ok := unreadMessageAction(path, page, entries); ok {
		act = unreadAct
	} else {
		var err error
		act, err = Plan(a.Persona, planPage, a.journal.Entries(),
			a.loggedIn, a.Username, a.Phone)
		if err != nil {
			a.journal.Append(JournalEntry{
				Phase: PhasePlan, URL: path, Action: "plan", Error: err.Error(),
			})
			return nil
		}
	}

	if browserclient.IsModalPath(act.Path) && modalLoopDetected(a.journal.Entries()) {
		act = escapeModalLoop(page)
		a.journal.Append(JournalEntry{
			Phase:     PhasePlan,
			URL:       path,
			Action:    "loop break",
			Reasoning: act.Reason,
		})
	}

	if a.loggedIn && browserclient.IsAuthEntryPath(act.Path) {
		act = sellerFallbackAction(a.Persona, page)
		a.journal.Append(JournalEntry{
			Phase:     PhasePlan,
			URL:       path,
			Action:    "skip auth page",
			Reasoning: act.Reason,
		})
	}

	if act.Action == "wait" && browserclient.IsStuckPage(page) {
		act = escapeStuckPage(page, a.Persona)
		a.journal.Append(JournalEntry{
			Phase:     PhasePlan,
			URL:       path,
			Action:    "escape stuck page",
			Reasoning: act.Reason,
		})
	}

	if act.Action == "noop" && noopLoopDetected(a.journal.Entries(), path) {
		act = escapeNoopLoop(path, page, a.Persona)
		a.journal.Append(JournalEntry{
			Phase:     PhasePlan,
			URL:       path,
			Action:    "escape noop loop",
			Reasoning: act.Reason,
		})
	}

	if act.Action == "click" && strings.HasPrefix(act.Path, "/auth/user/myads/tab/") &&
		myAdsTabLoopDetected(a.journal.Entries()) {
		act = escapeMyAdsTabLoop(a.Persona, page)
		a.journal.Append(JournalEntry{
			Phase:     PhasePlan,
			URL:       path,
			Action:    "escape myads tab loop",
			Reasoning: act.Reason,
		})
	}

	if act.Action == "click" && strings.HasPrefix(act.Path, "/auth/user/settings/") &&
		settingsTabLoopDetected(a.journal.Entries()) {
		act = escapeSettingsTabLoop(page, a.Persona)
		a.journal.Append(JournalEntry{
			Phase:     PhasePlan,
			URL:       path,
			Action:    "escape settings tab loop",
			Reasoning: act.Reason,
		})
	}

	a.journal.Append(JournalEntry{
		Phase: PhasePlan, URL: path, Action: act.Action + " " + act.Path,
		Reasoning: act.Reason,
	})

	validateErr := ValidateAction(act, planPage, a.loggedIn)
	if validateErr != nil && act.Action == "get" && browserclient.HTMXPaths(planPage)[act.Path] {
		act.Action = "click"
		validateErr = ValidateAction(act, planPage, a.loggedIn)
	}
	if validateErr != nil {
		if act.Reason != "escape modal loop" &&
			act.Reason != "already logged in; create an ad" &&
			act.Reason != "skip auth page" &&
			act.Reason != "escape stuck page" &&
			act.Reason != "escape noop loop" &&
			act.Reason != "escape myads tab loop" &&
			act.Reason != "escape settings tab loop" &&
			act.Reason != "reply to message" &&
			act.Reason != "reply after conversation open loop" &&
			act.Reason != "escape conversation click loop" {
			a.journal.Append(JournalEntry{
				Phase: PhasePlan, URL: path, Action: "validate", Error: validateErr.Error(),
			})
			return nil
		}
	}

	act = ensureMessageFields(act, a.Persona)

	switch act.Action {
	case "wait":
		sec := act.Wait
		if sec <= 0 {
			sec = 30
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(sec) * time.Second):
		}
		return nil
	case "noop":
		return nil
	case "get":
		err := a.client.ActGet(act.Path)
		a.journal.Append(JournalEntry{
			Phase: PhaseAct, URL: act.Path, Action: "GET " + act.Path,
			Error: errString(err),
		})
		if err == nil {
			a.snapshotAfterAct(act.Path)
		}
		return nil
	case "click":
		err := a.client.ActClick(act.Path)
		a.journal.Append(JournalEntry{
			Phase: PhaseAct, URL: act.Path, Action: "CLICK " + act.Path,
			Error: errString(err),
		})
		if err == nil {
			a.snapshotAfterAct(act.Path)
		}
		return nil
	case "post_form":
		if browserclient.IsBlockedAgentForm(act.Path) {
			a.journal.Append(JournalEntry{
				Phase: PhaseAct, URL: act.Path, Action: "blocked form",
				Error: "form not allowed for test agents",
			})
			return nil
		}
		active := a.inbox.activeSnapshot()
		var activePtr *browserclient.ConversationSnapshot
		if active.ID != 0 {
			activePtr = &active
		}
		if browserclient.IsMessageSendPath(act.Path) &&
			!browserclient.AllowsMessageSend(act.Path, activePtr) {
			a.journal.Append(JournalEntry{
				Phase: PhaseAct, URL: act.Path, Action: "blocked message send",
				Error: "wait for the other person to reply",
			})
			return nil
		}
		if act.Path == "/auth/ad/new" {
			if err := a.ensureAdCategory(page, path); err != nil {
				a.journal.Append(JournalEntry{
					Phase: PhaseAct, URL: path, Action: "ensure category",
					Error: err.Error(),
				})
				return nil
			}
		}
		if browserclient.IsMessageSendPath(act.Path) {
			if err := a.ensureMessageForm(act.Path); err != nil {
				a.journal.Append(JournalEntry{
					Phase: PhaseAct, URL: act.Path, Action: "open message form",
					Error: err.Error(),
				})
				return nil
			}
		}
		err := a.client.ActPostForm(act.Path, act.Fields)
		a.journal.Append(JournalEntry{
			Phase: PhaseAct, URL: act.Path, Action: "POST " + act.Path,
			Error: errString(err),
		})
		if err == nil {
			a.snapshotAfterAct(act.Path)
		}
		return nil
	}
	return nil
}

func (a *Agent) ensureAdCategory(page browserclient.PageAffordances, returnPath string) error {
	want := a.Persona.PreferredAdCategory()
	if want == "" || page.CurrentCategory == want {
		return nil
	}
	if switchPath := browserclient.CategorySwitchPath(page, want); switchPath != "" {
		if err := a.client.ActClick(switchPath); err != nil {
			return err
		}
		a.snapshotAfterAct(returnPath)
		return nil
	}
	modalPath := "/api/category-select?return=" + url.QueryEscape(returnPath)
	if err := a.client.ActClick(modalPath); err != nil {
		return err
	}
	obs, _ := a.client.Observe()
	if switchPath := browserclient.CategorySwitchPath(obs.Page, want); switchPath == "" {
		return fmt.Errorf("category %q not in picker", want)
	} else if err := a.client.ActClick(switchPath); err != nil {
		return err
	}
	a.snapshotAfterAct(returnPath)
	a.journal.Append(JournalEntry{
		Phase: PhaseAct, URL: returnPath, Action: "category " + want,
	})
	return nil
}

func (a *Agent) snapshotAfterAct(fallbackPath string) {
	obs, _ := a.client.Observe()
	path := obs.Path
	if path == "" || path == "/" {
		path = fallbackPath
	}
	a.setPath(path)
	a.syncInbox(obs)
	a.journal.Append(JournalEntry{
		Phase: PhaseObserve, URL: path, Action: "observe " + path,
	})
	a.syncLoggedIn(obs.Page)
	if obs.Page.Conversation != nil {
		conv := *obs.Page.Conversation
		a.inbox.setActive(conv)
	}
}

func (a *Agent) setPath(p string) {
	a.mu.Lock()
	a.currentPath = p
	a.mu.Unlock()
}

func (a *Agent) getPath() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.currentPath
}

func (a *Agent) syncInbox(obs browserclient.Observation) {
	for _, conv := range obs.Page.Conversations {
		if a.inbox.update(conv) {
			a.journalAppendConversation(conv, "new message")
		}
	}
	if obs.Page.Conversation != nil {
		conv := *obs.Page.Conversation
		if a.inbox.update(conv) {
			a.journalAppendConversation(conv, "new message")
		}
	}
}

func unreadMessageAction(path string, page browserclient.PageAffordances,
	entries []JournalEntry) (PlannedAction, bool) {
	if !page.HasUnreadMessages {
		return PlannedAction{}, false
	}
	if len(page.OpenConversationForms) > 0 {
		return PlannedAction{}, false
	}
	if path != "/auth/user/messages" {
		return PlannedAction{
			Action: "get",
			Path:   "/auth/user/messages",
			Reason: "unread messages indicator",
		}, true
	}
	if len(page.UnreadConversationPaths) == 0 {
		return PlannedAction{}, false
	}
	openPath := page.UnreadConversationPaths[0]
	if conversationClickLoopDetected(entries, openPath) {
		return PlannedAction{}, false
	}
	return PlannedAction{
		Action: "click",
		Path:   openPath,
		Reason: "open unread conversation",
	}, true
}

func pendingReplyAction(path string, page browserclient.PageAffordances,
	inbox *inbox) (PlannedAction, bool) {
	conv := inbox.awaitingReplySnapshot()
	if conv.ID == 0 {
		return PlannedAction{}, false
	}
	if browserclient.PageHasOpenConversationForm(page, conv.ID) {
		return PlannedAction{}, false
	}
	openPath := fmt.Sprintf("/auth/conversation/%d", conv.ID)
	if path == "/auth/user/messages" {
		if browserclient.HTMXPaths(page)[openPath] {
			return PlannedAction{
				Action: "click", Path: openPath,
				Reason: "open conversation to reply",
			}, true
		}
	}
	return PlannedAction{
		Action: "get",
		Path:   "/auth/user/messages",
		Reason: "check messages to reply",
	}, true
}

func (a *Agent) ensureConversationForm(convID int) error {
	obs, _ := a.client.Observe()
	page := obs.Page
	if browserclient.PageHasOpenConversationForm(page, convID) {
		return a.client.WaitForConversationForm(convID)
	}
	openPath := fmt.Sprintf("/auth/conversation/%d", convID)
	if browserclient.HTMXPaths(page)[openPath] {
		if err := a.client.ActClick(openPath); err != nil {
			return err
		}
	} else if err := a.client.ActGet("/auth/user/messages"); err != nil {
		return err
	} else {
		obs, _ = a.client.Observe()
		if !browserclient.HTMXPaths(obs.Page)[openPath] {
			return fmt.Errorf("conversation %d not in messages list", convID)
		}
		if err := a.client.ActClick(openPath); err != nil {
			return err
		}
	}
	return a.client.WaitForConversationForm(convID)
}

func (a *Agent) ensureAdMessageForm(adID int) error {
	obs, _ := a.client.Observe()
	page := obs.Page
	if browserclient.PageHasAdMessageForm(page, adID) {
		return a.client.WaitForAdMessageForm(adID)
	}
	for _, convID := range page.OpenConversationForms {
		if convID != 0 {
			return a.client.WaitForAdMessageForm(adID)
		}
	}
	openPath := browserclient.AdNewConversationPath(adID)
	if browserclient.HTMXPaths(page)[openPath] {
		if err := a.client.ActClick(openPath); err != nil {
			return err
		}
		return a.client.WaitForAdMessageForm(adID)
	}
	return fmt.Errorf("message button not on page for ad %d", adID)
}

func (a *Agent) ensureMessageForm(postPath string) error {
	if convID, ok := browserclient.ConversationIDFromSendPath(postPath); ok {
		return a.ensureConversationForm(convID)
	}
	if adID, ok := browserclient.AdIDFromSendPath(postPath); ok {
		return a.ensureAdMessageForm(adID)
	}
	return nil
}

func (a *Agent) ensureSession() error {
	if a.client.SessionActive() {
		a.loggedIn = true
		return nil
	}
	a.loggedIn = false
	if err := a.client.Login(a.Username, a.password); err != nil {
		if a.client.SessionActive() {
			a.loggedIn = true
			return nil
		}
		return err
	}
	if !a.client.SessionActive() {
		return fmt.Errorf("re-login failed")
	}
	a.loggedIn = true
	a.journal.Append(JournalEntry{Phase: PhaseBoot, Action: "re-login"})
	return nil
}

func (a *Agent) syncLoggedIn(page browserclient.PageAffordances) {
	if a.client.SessionActive() {
		a.loggedIn = true
		return
	}
	if browserclient.PageLooksLoggedOut(page) {
		a.loggedIn = false
	}
}

func errString(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

func randomDelay(min, max time.Duration) time.Duration {
	if max <= min {
		return min
	}
	diff := max - min
	n, err := rand.Int(rand.Reader, big.NewInt(int64(diff/time.Second)+1))
	if err != nil {
		return min
	}
	return min + time.Duration(n.Int64())*time.Second
}

func (a *Agent) journalAppendConversation(conv browserclient.ConversationSnapshot, action string) {
	latest := ""
	fromOther := false
	if n := len(conv.Messages); n > 0 {
		latest = conv.Messages[n-1].Text
		fromOther = !conv.Messages[n-1].FromSelf
	}
	a.journal.Append(JournalEntry{
		Phase:     PhaseObserve,
		URL:       fmt.Sprintf("/auth/conversation/%d", conv.ID),
		Action:    action,
		Reasoning: fmt.Sprintf("%d messages; latest from other=%v: %q", len(conv.Messages), fromOther, latest),
	})
}

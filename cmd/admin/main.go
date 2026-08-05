package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rocky-ads/site/internal/backup"
	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/dbinit"
	"github.com/rocky-ads/site/internal/imagestore"
	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/user"
	"golang.org/x/term"
)

type screen int

const (
	screenMenu screen = iota
	screenBackupConfirm
	screenBackupRunning
	screenRestorePick
	screenRestoreKey
	screenRestoreConfirm
	screenRestoreRunning
	screenInitConfirm
	screenInitRunning
	screenAdminName
	screenAdminRunning
	screenResult
)

type menuItem int

const (
	menuBackup menuItem = iota
	menuRestore
	menuInit
	menuInitSeed
	menuPromote
	menuDemote
	menuQuit
)

type model struct {
	screen        screen
	menuCursor    int
	archives      []string
	archiveCursor int
	nameInput     textinput.Model
	keyInput      textinput.Model
	adminMode     string // promote | demote
	initSeed      bool
	selected      string
	status        string
	err           string
	store         imagestore.Store
	dbTarget      string
}

type doneMsg struct {
	ok  string
	err error
}

func initialModel(store imagestore.Store, dbTarget string) model {
	ni := textinput.New()
	ni.Placeholder = "username"
	ni.CharLimit = 64
	ni.Width = 40

	ki := textinput.New()
	ki.Placeholder = "BACKUP_DB_ENCRYPTION_KEY (empty = DB_ENCRYPTION_KEY)"
	ki.EchoMode = textinput.EchoPassword
	ki.EchoCharacter = '•'
	ki.Width = 60

	return model{
		screen:    screenMenu,
		nameInput: ni,
		keyInput:  ki,
		store:     store,
		dbTarget:  dbTarget,
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case doneMsg:
		if msg.err != nil {
			m.err = msg.err.Error()
			m.status = ""
		} else {
			m.status = msg.ok
			m.err = ""
		}
		m.screen = screenResult
		return m, nil
	case tea.KeyMsg:
		switch m.screen {
		case screenMenu:
			return m.updateMenu(msg)
		case screenBackupConfirm:
			return m.updateBackupConfirm(msg)
		case screenRestorePick:
			return m.updateRestorePick(msg)
		case screenRestoreKey:
			return m.updateRestoreKey(msg)
		case screenRestoreConfirm:
			return m.updateRestoreConfirm(msg)
		case screenInitConfirm:
			return m.updateInitConfirm(msg)
		case screenAdminName:
			return m.updateAdminName(msg)
		case screenResult:
			if msg.String() == "enter" || msg.String() == "esc" {
				m.screen = screenMenu
				m.status = ""
				m.err = ""
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	switch m.screen {
	case screenAdminName:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case screenRestoreKey:
		m.keyInput, cmd = m.keyInput.Update(msg)
	}
	return m, cmd
}

func (m model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.menuCursor > 0 {
			m.menuCursor--
		}
	case "down", "j":
		if m.menuCursor < int(menuQuit) {
			m.menuCursor++
		}
	case "enter":
		switch menuItem(m.menuCursor) {
		case menuBackup:
			m.screen = screenBackupConfirm
		case menuRestore:
			names, err := backup.ListArchives()
			if err != nil {
				m.err = err.Error()
				m.screen = screenResult
				return m, nil
			}
			if len(names) == 0 {
				m.err = "no backups in " + backup.DefaultDir + "/"
				m.screen = screenResult
				return m, nil
			}
			m.archives = names
			m.archiveCursor = 0
			m.screen = screenRestorePick
		case menuInit:
			m.initSeed = false
			m.screen = screenInitConfirm
		case menuInitSeed:
			m.initSeed = true
			m.screen = screenInitConfirm
		case menuPromote:
			m.adminMode = "promote"
			m.nameInput.SetValue("")
			m.nameInput.Focus()
			m.screen = screenAdminName
			return m, textinput.Blink
		case menuDemote:
			m.adminMode = "demote"
			m.nameInput.SetValue("")
			m.nameInput.Focus()
			m.screen = screenAdminName
			return m, textinput.Blink
		case menuQuit:
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) updateBackupConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = screenMenu
	case "y", "enter":
		m.screen = screenBackupRunning
		store := m.store
		return m, func() tea.Msg {
			path, err := backup.BackupToArchive("", store, false, false)
			if err != nil {
				return doneMsg{err: err}
			}
			return doneMsg{ok: "Backup written: " + path}
		}
	}
	return m, nil
}

func (m model) updateInitConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n":
		m.screen = screenMenu
	case "y":
		m.screen = screenInitRunning
		loadSeed := m.initSeed
		return m, func() tea.Msg {
			if err := dbinit.Rebuild(loadSeed); err != nil {
				return doneMsg{err: err}
			}
			if loadSeed {
				return doneMsg{ok: "Database rebuilt with seed data"}
			}
			return doneMsg{ok: "Database rebuilt (categories only)"}
		}
	}
	return m, nil
}

func (m model) updateRestorePick(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = screenMenu
		return m, nil
	case "up", "k":
		if m.archiveCursor > 0 {
			m.archiveCursor--
		}
	case "down", "j":
		if m.archiveCursor < len(m.archives)-1 {
			m.archiveCursor++
		}
	case "enter":
		if len(m.archives) == 0 {
			return m, nil
		}
		m.selected = m.archives[m.archiveCursor]
		m.keyInput.SetValue("")
		m.keyInput.Focus()
		m.screen = screenRestoreKey
		return m, textinput.Blink
	}
	return m, nil
}

func (m model) updateRestoreKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = screenRestorePick
		return m, nil
	case "enter":
		m.screen = screenRestoreConfirm
		return m, nil
	}
	var cmd tea.Cmd
	m.keyInput, cmd = m.keyInput.Update(msg)
	return m, cmd
}

func (m model) updateRestoreConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n":
		m.screen = screenRestoreKey
	case "y":
		m.screen = screenRestoreRunning
		store := m.store
		selected := m.selected
		keyStr := strings.TrimSpace(m.keyInput.Value())
		return m, func() tea.Msg {
			var backupKey []byte
			if keyStr != "" {
				k, err := base64.StdEncoding.DecodeString(keyStr)
				if err != nil || len(k) != 32 {
					return doneMsg{err: fmt.Errorf(
						"BACKUP_DB_ENCRYPTION_KEY must be 32-byte base64")}
				}
				backupKey = k
			}
			err := backup.RestoreFromArchive(
				selected, store, backupKey, false, false,
			)
			if err != nil {
				return doneMsg{err: err}
			}
			return doneMsg{ok: "Restored from " + selected}
		}
	}
	return m, nil
}

func (m model) updateAdminName(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = screenMenu
		return m, nil
	case "enter":
		name := strings.TrimSpace(m.nameInput.Value())
		if name == "" {
			return m, nil
		}
		m.screen = screenAdminRunning
		mode := m.adminMode
		return m, func() tea.Msg {
			u, err := user.GetByName(name)
			if err != nil {
				return doneMsg{err: fmt.Errorf("user %q: %w", name, err)}
			}
			switch mode {
			case "promote":
				if u.IsAdmin {
					return doneMsg{ok: name + " is already an admin"}
				}
				if err := user.PromoteToAdmin(u.ID); err != nil {
					return doneMsg{err: err}
				}
				return doneMsg{ok: "Promoted " + name + " to admin"}
			case "demote":
				if !u.IsAdmin {
					return doneMsg{ok: name + " is not an admin"}
				}
				if err := user.DemoteFromAdmin(u.ID); err != nil {
					return doneMsg{err: err}
				}
				return doneMsg{ok: "Demoted " + name + " from admin"}
			}
			return doneMsg{err: fmt.Errorf("unknown mode")}
		}
	}
	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	title := lipgloss.NewStyle().Bold(true).Render("Rocky Ads Admin")
	if m.dbTarget != "" {
		muted := lipgloss.NewStyle().Faint(true).Render(m.dbTarget)
		title = title + "\n" + muted
	}
	var body string
	switch m.screen {
	case screenMenu:
		items := []string{
			"Backup database",
			"Restore database",
			"Init database (categories only)",
			"Init database with seed",
			"Promote admin",
			"Demote admin",
			"Quit",
		}
		var b strings.Builder
		b.WriteString("Choose an action:\n\n")
		for i, item := range items {
			cursor := "  "
			if i == m.menuCursor {
				cursor = "> "
			}
			b.WriteString(cursor + item + "\n")
		}
		b.WriteString("\n↑/↓ move · enter select · q quit")
		body = b.String()
	case screenBackupConfirm:
		body = "Create backup in " + backup.DefaultDir + "/ ?\n\n" +
			"y/enter confirm · esc cancel"
	case screenInitConfirm:
		if m.initSeed {
			body = "Wipe DB and reload schema + seed users/ads?\n\n" +
				"y confirm · n/esc cancel"
		} else {
			body = "Wipe DB and reload schema + categories only?\n\n" +
				"y confirm · n/esc cancel"
		}
	case screenBackupRunning, screenRestoreRunning, screenAdminRunning,
		screenInitRunning:
		body = "Working…"
	case screenRestorePick:
		var b strings.Builder
		b.WriteString("Select backup:\n\n")
		for i, name := range m.archives {
			cursor := "  "
			if i == m.archiveCursor {
				cursor = "> "
			}
			b.WriteString(cursor + name + "\n")
		}
		b.WriteString("\n↑/↓ move · enter select · esc cancel")
		body = b.String()
	case screenRestoreKey:
		body = "Archive: " + m.selected + "\n\n" +
			"Source DB encryption key (base64):\n" + m.keyInput.View() +
			"\n\nenter continue · esc back"
	case screenRestoreConfirm:
		body = fmt.Sprintf(
			"Restore %s?\nThis WIPES the current database.\n\ny confirm · n/esc cancel",
			m.selected,
		)
	case screenAdminName:
		body = m.adminMode + " user:\n\n" + m.nameInput.View() +
			"\n\nenter submit · esc cancel"
	case screenResult:
		if m.err != "" {
			body = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).
				Render("Error: "+m.err) + "\n\nenter/esc back"
		} else {
			body = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).
				Render(m.status) + "\n\nenter/esc back"
		}
	}
	return title + "\n\n" + body + "\n"
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h", "--help", "help":
			fmt.Print(`Usage:
  admin    Interactive TUI (backup, restore, init DB, promote/demote)
`)
			return
		default:
			fmt.Fprintf(os.Stderr, "unknown argument %q (try -h)\n", os.Args[1])
			os.Exit(1)
		}
	}
	runTUI()
}

func runTUI() {
	if !term.IsTerminal(int(os.Stdin.Fd())) ||
		!term.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Fprintln(os.Stderr,
			"admin requires a TTY (use ssh -t, or docker compose exec -it)")
		os.Exit(1)
	}

	mustInit(true)
	defer db.Close()

	host, database := db.ConnectionTarget(config.DatabaseURL)
	dbTarget := database
	if host != "" {
		dbTarget = database + " @ " + host
	}
	fmt.Fprintf(os.Stderr, "Rocky Ads Admin — connected to %s\n", dbTarget)

	store, err := imagestore.NewDefault()
	if err != nil {
		fmt.Fprintf(os.Stderr, "image store: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(
		initialModel(store, dbTarget), tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "admin: %v\n", err)
		os.Exit(1)
	}
}

func mustInit(quiet bool) {
	logOut := ""
	if quiet {
		logOut = os.DevNull
	}
	if err := logger.Init("info", "text", logOut); err != nil {
		fmt.Fprintf(os.Stderr, "logger: %v\n", err)
		os.Exit(1)
	}
	if config.DatabaseURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL must be set")
		os.Exit(1)
	}
	if len(config.DBEncryptionKey) != 32 {
		fmt.Fprintln(os.Stderr,
			"DB_ENCRYPTION_KEY must be set (32-byte base64)")
		os.Exit(1)
	}
	if len(config.DBHashPepper) != 32 {
		fmt.Fprintln(os.Stderr,
			"DB_HASH_PEPPER must be set (32-byte base64, "+
				"distinct from DB_ENCRYPTION_KEY)")
		os.Exit(1)
	}
	db.SetHashPepper(config.DBHashPepper)
	if err := db.Init(config.DatabaseURL); err != nil {
		fmt.Fprintf(os.Stderr, "database: %v\n", err)
		os.Exit(1)
	}
}

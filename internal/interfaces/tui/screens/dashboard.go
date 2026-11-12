package screens

import (
	"context"
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/forgedash/forgedash/internal/application/ports"
	"github.com/forgedash/forgedash/internal/application/usecase"
	"github.com/forgedash/forgedash/internal/domain/account"
	"github.com/forgedash/forgedash/internal/domain/change"
	"github.com/forgedash/forgedash/internal/domain/issue"
	"github.com/forgedash/forgedash/internal/domain/notification"
	"github.com/forgedash/forgedash/internal/domain/shared"
	"github.com/forgedash/forgedash/internal/infrastructure/config"
	"github.com/forgedash/forgedash/internal/interfaces/tui/components"
	"github.com/forgedash/forgedash/internal/interfaces/tui/components/modals"
	"github.com/forgedash/forgedash/internal/interfaces/tui/keymap"
	"github.com/forgedash/forgedash/internal/interfaces/tui/safebrowser"
	"github.com/forgedash/forgedash/internal/interfaces/tui/styles"
)

type FocusedPane int

const (
	PaneTable FocusedPane = iota
	PanePreview
)

type ModalState int

const (
	ModalNone ModalState = iota
	ModalHelp
	ModalSearch
	ModalAccountSwitch
	ModalConfirm
	ModalComment
)

type dataLoadedMsg struct {
	result *usecase.DashboardResult
	err    error
}

type previewDetailLoadedMsg struct {
	diffs    []shared.DiffFile
	commits  []shared.Commit
	checks   []shared.CheckRun
	comments []shared.Comment
}

type statusFlashMsg string

type tickMsg time.Time

type DashboardModel struct {
	cfg         *config.Config
	dashboardUC *usecase.DashboardUseCase
	changeUC    *usecase.ChangeActionUseCase
	issueUC     *usecase.IssueActionUseCase
	theme       *styles.Theme
	keys        keymap.KeyMap

	width  int
	height int

	activeView    usecase.ViewMode
	activeScope   usecase.AccountScope
	activeSection int
	sections      []usecase.SectionFilter
	previewOpen   bool
	previewPos    string
	focusedPane   FocusedPane
	changes       []change.ChangeRequest
	issues        []issue.Issue
	notifications []notification.Notification
	errors        map[account.AccountID]error
	loading       bool

	curDiffs    []shared.DiffFile
	curCommits  []shared.Commit
	curChecks   []shared.CheckRun
	curComments []shared.Comment

	activeModal         ModalState
	searchInput         textinput.Model
	commentInput        textinput.Model
	accountModalIndex   int
	accountModalClients []ports.ForgeClient
	helpPage            int
	confirmAction       func() tea.Cmd
	confirmPrompt       string
	statusFlash         string
	statusFlashExpiry   time.Time

	header      *components.Header
	sectionTabs *components.SectionTabs
	itemTable   *components.ItemTable
	previewPane *components.PreviewPane
	statusBar   *components.StatusBar
}

func NewDashboardModel(
	cfg *config.Config,
	dashboardUC *usecase.DashboardUseCase,
	changeUC *usecase.ChangeActionUseCase,
	issueUC *usecase.IssueActionUseCase,
) *DashboardModel {
	theme := styles.NewTheme(styles.DefaultPalette())
	if cfg.Theme == "nord" {
		theme = styles.NewTheme(styles.NordPalette())
	}

	activeView := usecase.ViewModePRs
	if cfg.DefaultView == "issues" {
		activeView = usecase.ViewModeIssues
	} else if cfg.DefaultView == "notifications" {
		activeView = usecase.ViewModeNotifications
	}

	scope := usecase.AccountScope{
		IsUnified: cfg.DefaultMode != "scoped",
		AccountID: account.AccountID(cfg.ActiveAccount),
	}

	var sections []usecase.SectionFilter
	for _, s := range cfg.PRSections {
		sections = append(sections, usecase.SectionFilter{
			Title:   s.Title,
			Filters: s.Filters,
			Limit:   s.Limit,
		})
	}

	ti := textinput.New()
	ti.Placeholder = "Type filter query..."
	ti.CharLimit = 100

	ci := textinput.New()
	ci.Placeholder = "Write a comment..."
	ci.CharLimit = 500

	m := &DashboardModel{
		cfg:           cfg,
		dashboardUC:   dashboardUC,
		changeUC:      changeUC,
		issueUC:       issueUC,
		theme:         theme,
		keys:          keymap.DefaultKeyMap(),
		activeView:    activeView,
		activeScope:   scope,
		activeSection: 0,
		sections:      sections,
		previewOpen:   cfg.Preview.Open,
		previewPos:    cfg.Preview.Position,
		focusedPane:   PaneTable,
		searchInput:   ti,
		commentInput:  ci,
		header:        components.NewHeader(theme),
		sectionTabs:   components.NewSectionTabs(theme),
		itemTable:     components.NewItemTable(theme),
		previewPane:   components.NewPreviewPane(theme),
		statusBar:     components.NewStatusBar(theme),
		errors:        make(map[account.AccountID]error),
	}
	m.previewPane.SetIssueMode(activeView == usecase.ViewModeIssues)
	if activeView == usecase.ViewModeNotifications {
		m.focusedPane = PaneTable
	}
	return m
}

func (m *DashboardModel) Init() tea.Cmd {
	return tea.Batch(
		m.fetchDataCmd(),
		m.autoRefreshTick(),
	)
}

func (m *DashboardModel) autoRefreshTick() tea.Cmd {
	if !m.cfg.AutoRefresh {
		return nil
	}
	sec := m.cfg.RefreshInterval
	if sec <= 0 {
		sec = 60
	}
	return tea.Tick(time.Duration(sec)*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *DashboardModel) fetchDataCmd() tea.Cmd {
	return func() tea.Msg {
		if len(m.sections) == 0 {
			return dataLoadedMsg{result: &usecase.DashboardResult{}}
		}
		sec := m.sections[m.activeSection]
		res, err := m.dashboardUC.FetchView(context.Background(), m.activeView, m.activeScope, sec)
		return dataLoadedMsg{result: res, err: err}
	}
}

func (m *DashboardModel) fetchItemDetailCmd() tea.Cmd {
	return func() tea.Msg {
		if m.activeView == usecase.ViewModePRs {
			idx := m.itemTable.Cursor()
			if idx < 0 || idx >= len(m.changes) {
				return previewDetailLoadedMsg{}
			}
			cr := m.changes[idx]
			client, ok := m.dashboardUC.GetClient(cr.AccountID)
			if !ok {
				return previewDetailLoadedMsg{}
			}

			diffs, _ := client.GetDiff(context.Background(), cr.RepoOwner, cr.RepoName, cr.Number)
			commits, _ := client.ListCommits(context.Background(), cr.RepoOwner, cr.RepoName, cr.Number)
			checks, _ := client.ListChecks(context.Background(), cr.RepoOwner, cr.RepoName, fmt.Sprintf("%d", cr.Number))
			comments, _ := client.ListComments(context.Background(), cr.RepoOwner, cr.RepoName, cr.Number, true)

			return previewDetailLoadedMsg{
				diffs:    diffs,
				commits:  commits,
				checks:   checks,
				comments: comments,
			}
		}

		if m.activeView == usecase.ViewModeIssues {
			idx := m.itemTable.Cursor()
			if idx < 0 || idx >= len(m.issues) {
				return previewDetailLoadedMsg{}
			}
			iss := m.issues[idx]
			client, ok := m.dashboardUC.GetClient(iss.AccountID)
			if !ok {
				return previewDetailLoadedMsg{}
			}
			comments, _ := client.ListComments(context.Background(), iss.RepoOwner, iss.RepoName, iss.Number, false)
			return previewDetailLoadedMsg{
				comments: comments,
			}
		}

		return previewDetailLoadedMsg{}
	}
}

func (m *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recalculateLayout()
	case tickMsg:
		cmds = append(cmds, m.fetchDataCmd(), m.autoRefreshTick())

	case statusFlashMsg:
		m.statusFlash = string(msg)
		m.statusFlashExpiry = time.Now().Add(4 * time.Second)

	case dataLoadedMsg:
		m.loading = false
		if msg.err == nil && msg.result != nil {
			m.changes = msg.result.Changes
			m.issues = msg.result.Issues
			m.notifications = msg.result.Notifications
			m.errors = msg.result.Errors
		}
		m.itemTable.Clamp(m.activeTotalCount())
		cmds = append(cmds, m.fetchItemDetailCmd())

	case previewDetailLoadedMsg:
		m.curDiffs = msg.diffs
		m.curCommits = msg.commits
		m.curChecks = msg.checks
		m.curComments = msg.comments

	case tea.KeyMsg:
		if m.activeModal != ModalNone {
			return m.handleModalKey(msg)
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.activeModal = ModalHelp
			m.helpPage = 0
			return m, nil
		case key.Matches(msg, m.keys.AccountModal):
			m.activeModal = ModalAccountSwitch
			m.accountModalIndex = 0
			m.accountModalClients = m.sortedClients()
			return m, nil
		case key.Matches(msg, m.keys.Search):
			m.activeModal = ModalSearch
			m.searchInput.Focus()
			return m, nil
		case key.Matches(msg, m.keys.Refresh):
			m.loading = true
			return m, m.fetchDataCmd()
		case key.Matches(msg, m.keys.NextView):
			m.cycleView(true)
			m.loading = true
			return m, m.fetchDataCmd()
		case key.Matches(msg, m.keys.PrevView):
			m.cycleView(false)
			m.loading = true
			return m, m.fetchDataCmd()
		case key.Matches(msg, m.keys.Sec1):
			return m.switchSection(0)
		case key.Matches(msg, m.keys.Sec2):
			return m.switchSection(1)
		case key.Matches(msg, m.keys.Sec3):
			return m.switchSection(2)
		case key.Matches(msg, m.keys.Sec4):
			return m.switchSection(3)
		case key.Matches(msg, m.keys.Sec5):
			return m.switchSection(4)
		case key.Matches(msg, m.keys.TogglePreview):
			m.previewOpen = !m.previewOpen
			if !m.previewOpen {
				m.focusedPane = PaneTable
			}
			m.recalculateLayout()
			return m, nil
		case key.Matches(msg, m.keys.TogglePreviewPos):
			if m.previewPos == "right" {
				m.previewPos = "bottom"
			} else {
				m.previewPos = "right"
			}
			m.recalculateLayout()
			return m, nil
		}

		if m.effectivePreview() {
			switch {
			case key.Matches(msg, m.keys.FocusNext), key.Matches(msg, m.keys.ToggleFocus):
				if m.focusedPane == PaneTable {
					m.focusedPane = PanePreview
				} else {
					m.focusedPane = PaneTable
				}
				m.syncFocus()
				return m, nil
			case key.Matches(msg, m.keys.FocusPrev):
				if m.focusedPane == PanePreview {
					m.focusedPane = PaneTable
				} else {
					m.focusedPane = PanePreview
				}
				m.syncFocus()
				return m, nil
			}
		}

		if m.effectivePreview() && m.focusedPane == PanePreview {
			switch {
			case key.Matches(msg, m.keys.NextPreviewTab):
				m.previewPane.NextTab()
				return m, nil
			case key.Matches(msg, m.keys.PrevPreviewTab):
				m.previewPane.PrevTab()
				return m, nil
			case key.Matches(msg, m.keys.PreviewPageDown), key.Matches(msg, m.keys.Down), key.Matches(msg, m.keys.HalfPageDown):
				m.previewPane.ScrollDown()
				return m, nil
			case key.Matches(msg, m.keys.PreviewPageUp), key.Matches(msg, m.keys.Up), key.Matches(msg, m.keys.HalfPageUp):
				m.previewPane.ScrollUp()
				return m, nil
			case key.Matches(msg, m.keys.Esc):
				m.focusedPane = PaneTable
				m.syncFocus()
				return m, nil
			}
		}

		total := m.activeTotalCount()
		switch {
		case key.Matches(msg, m.keys.Down):
			m.itemTable.MoveDown(total)
			return m, m.fetchItemDetailCmd()
		case key.Matches(msg, m.keys.Up):
			m.itemTable.MoveUp()
			return m, m.fetchItemDetailCmd()
		case key.Matches(msg, m.keys.FirstLine), key.Matches(msg, m.keys.Top):
			m.itemTable.MoveTop()
			return m, m.fetchItemDetailCmd()
		case key.Matches(msg, m.keys.LastLine), key.Matches(msg, m.keys.Bottom):
			m.itemTable.MoveBottom(total)
			return m, m.fetchItemDetailCmd()
		case key.Matches(msg, m.keys.HalfPageDown):
			for range 5 {
				m.itemTable.MoveDown(total)
			}
			return m, m.fetchItemDetailCmd()
		case key.Matches(msg, m.keys.HalfPageUp):
			for range 5 {
				m.itemTable.MoveUp()
			}
			return m, m.fetchItemDetailCmd()
		case key.Matches(msg, m.keys.NextSection):
			if len(m.sections) > 0 {
				m.activeSection = (m.activeSection + 1) % len(m.sections)
				m.itemTable.MoveTop()
				m.loading = true
				return m, m.fetchDataCmd()
			}
		case key.Matches(msg, m.keys.PrevSection):
			if len(m.sections) > 0 {
				if m.activeSection == 0 {
					m.activeSection = len(m.sections) - 1
				} else {
					m.activeSection--
				}
				m.itemTable.MoveTop()
				m.loading = true
				return m, m.fetchDataCmd()
			}
		case key.Matches(msg, m.keys.NextPreviewTab):
			if m.effectivePreview() {
				m.previewPane.NextTab()
				return m, nil
			}
		case key.Matches(msg, m.keys.PrevPreviewTab):
			if m.effectivePreview() {
				m.previewPane.PrevTab()
				return m, nil
			}
		case key.Matches(msg, m.keys.PreviewPageDown):
			if m.effectivePreview() {
				m.previewPane.ScrollDown()
				return m, nil
			}
		case key.Matches(msg, m.keys.PreviewPageUp):
			if m.effectivePreview() {
				m.previewPane.ScrollUp()
				return m, nil
			}
		}

		switch {
		case key.Matches(msg, m.keys.OpenBrowser):
			url := m.activeItemURL()
			if url != "" {
				_ = openURL(url)
				return m, func() tea.Msg { return statusFlashMsg("Opened in browser: " + url) }
			}
		case key.Matches(msg, m.keys.CopyNumber):
			num := m.activeItemNumber()
			if num != "" {
				_ = clipboard.WriteAll(num)
				return m, func() tea.Msg { return statusFlashMsg("Copied: " + num) }
			}
		case key.Matches(msg, m.keys.CopyURL):
			url := m.activeItemURL()
			if url != "" {
				_ = clipboard.WriteAll(url)
				return m, func() tea.Msg { return statusFlashMsg("Copied URL: " + url) }
			}
		case key.Matches(msg, m.keys.DiffView):
			if m.activeView == usecase.ViewModePRs {
				m.previewOpen = true
				m.focusedPane = PanePreview
				m.previewPane.SetTab(components.PreviewTabFiles)
				m.recalculateLayout()
				return m, nil
			}
		case key.Matches(msg, m.keys.Comment):
			m.activeModal = ModalComment
			m.commentInput.Reset()
			m.commentInput.Focus()
			return m, nil
		case key.Matches(msg, m.keys.Approve):
			if cr, ok := m.activeChange(); ok {
				m.activeModal = ModalConfirm
				m.confirmPrompt = fmt.Sprintf("Approve PR %s: %s?", cr.DisplayID(), cr.Title)
				m.confirmAction = func() tea.Cmd {
					return func() tea.Msg {
						err := m.changeUC.Approve(context.Background(), cr, "Approved via ForgeDash")
						if err != nil {
							return statusFlashMsg(fmt.Sprintf("Failed to approve: %v", err))
						}
						return statusFlashMsg("Approved successfully!")
					}
				}
				return m, nil
			}
		case key.Matches(msg, m.keys.Merge):
			if cr, ok := m.activeChange(); ok {
				m.activeModal = ModalConfirm
				m.confirmPrompt = fmt.Sprintf("Merge PR %s: %s?", cr.DisplayID(), cr.Title)
				m.confirmAction = func() tea.Cmd {
					return func() tea.Msg {
						err := m.changeUC.Merge(context.Background(), cr, "merge", "", "")
						if err != nil {
							return statusFlashMsg(fmt.Sprintf("Failed to merge: %v", err))
						}
						return statusFlashMsg("Merged successfully!")
					}
				}
				return m, nil
			}
		case key.Matches(msg, m.keys.Close):
			if cr, ok := m.activeChange(); ok {
				m.activeModal = ModalConfirm
				m.confirmPrompt = fmt.Sprintf("Close PR %s: %s?", cr.DisplayID(), cr.Title)
				m.confirmAction = func() tea.Cmd {
					return func() tea.Msg {
						err := m.changeUC.Close(context.Background(), cr)
						if err != nil {
							return statusFlashMsg(fmt.Sprintf("Failed to close: %v", err))
						}
						return statusFlashMsg("Closed successfully!")
					}
				}
				return m, nil
			}
		case key.Matches(msg, m.keys.Reopen):
			if cr, ok := m.activeChange(); ok {
				return m, func() tea.Msg {
					err := m.changeUC.Reopen(context.Background(), cr)
					if err != nil {
						return statusFlashMsg(fmt.Sprintf("Failed to reopen: %v", err))
					}
					return statusFlashMsg("Reopened successfully!")
				}
			}
		case key.Matches(msg, m.keys.Checkout):
			if cr, ok := m.activeChange(); ok {
				return m, func() tea.Msg {
					err := m.changeUC.Checkout(context.Background(), cr)
					if err != nil {
						return statusFlashMsg(fmt.Sprintf("Checkout failed: %v", err))
					}
					return statusFlashMsg("Checked out branch: " + cr.SourceBranch)
				}
			}
		}
	}

	return m, tea.Batch(cmds...)
}
func (m *DashboardModel) handleModalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.activeModal {
	case ModalHelp:
		switch msg.String() {
		case "esc", "?", "q":
			m.activeModal = ModalNone
		case "right":
			if m.helpPage < 2 {
				m.helpPage++
			}
		case "left":
			if m.helpPage > 0 {
				m.helpPage--
			}
		}
		return m, nil

	case ModalSearch:
		switch msg.String() {
		case "esc":
			m.activeModal = ModalNone
			m.searchInput.Blur()
			return m, nil
		case "enter":
			query := m.searchInput.Value()
			m.activeModal = ModalNone
			m.searchInput.Blur()
			if len(m.sections) > 0 {
				m.sections[m.activeSection].Filters = query
				m.loading = true
				return m, m.fetchDataCmd()
			}
			return m, nil
		default:
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
		}

	case ModalComment:
		switch msg.String() {
		case "esc":
			m.activeModal = ModalNone
			m.commentInput.Blur()
			return m, nil
		case "enter":
			text := m.commentInput.Value()
			m.activeModal = ModalNone
			m.commentInput.Blur()
			if text != "" {
				return m, m.submitCommentCmd(text)
			}
			return m, nil
		default:
			var cmd tea.Cmd
			m.commentInput, cmd = m.commentInput.Update(msg)
			return m, cmd
		}

	case ModalConfirm:
		switch strings.ToLower(msg.String()) {
		case "y":
			m.activeModal = ModalNone
			if m.confirmAction != nil {
				cmd := m.confirmAction()
				return m, tea.Batch(cmd, m.fetchDataCmd())
			}
			return m, nil
		case "n", "esc":
			m.activeModal = ModalNone
			return m, nil
		}

	case ModalAccountSwitch:
		clients := m.accountModalClients
		totalOptions := len(clients) + 1

		switch msg.String() {
		case "esc", "q":
			m.activeModal = ModalNone
			m.accountModalClients = nil
			return m, nil
		case "j", "down":
			if m.accountModalIndex < totalOptions-1 {
				m.accountModalIndex++
			}
			return m, nil
		case "k", "up":
			if m.accountModalIndex > 0 {
				m.accountModalIndex--
			}
			return m, nil
		case "enter":
			m.activeModal = ModalNone
			if m.accountModalIndex == 0 {
				m.activeScope = usecase.AccountScope{IsUnified: true}
			} else if idx := m.accountModalIndex - 1; idx >= 0 && idx < len(clients) {
				client := clients[idx]
				m.activeScope = usecase.AccountScope{
					IsUnified: false,
					AccountID: client.Account().ID,
				}
			}
			m.accountModalClients = nil
			m.loading = true
			m.itemTable.MoveTop()
			return m, m.fetchDataCmd()
		}
	}

	return m, nil
}

func (m *DashboardModel) submitCommentCmd(text string) tea.Cmd {
	return func() tea.Msg {
		idx := m.itemTable.Cursor()
		if m.activeView == usecase.ViewModePRs && idx >= 0 && idx < len(m.changes) {
			cr := m.changes[idx]
			_, err := m.changeUC.AddComment(context.Background(), cr, text)
			if err != nil {
				return statusFlashMsg(fmt.Sprintf("Failed to comment: %v", err))
			}
			return statusFlashMsg("Comment posted!")
		} else if m.activeView == usecase.ViewModeIssues && idx >= 0 && idx < len(m.issues) {
			iss := m.issues[idx]
			_, err := m.issueUC.AddComment(context.Background(), iss, text)
			if err != nil {
				return statusFlashMsg(fmt.Sprintf("Failed to comment: %v", err))
			}
			return statusFlashMsg("Comment posted!")
		}
		return nil
	}
}

func (m *DashboardModel) switchSection(idx int) (tea.Model, tea.Cmd) {
	if idx >= 0 && idx < len(m.sections) {
		m.activeSection = idx
		m.itemTable.MoveTop()
		m.loading = true
		return m, m.fetchDataCmd()
	}
	return m, nil
}

func (m *DashboardModel) cycleView(forward bool) {
	views := []usecase.ViewMode{usecase.ViewModePRs, usecase.ViewModeIssues, usecase.ViewModeNotifications}
	currentIdx := 0
	for i, v := range views {
		if v == m.activeView {
			currentIdx = i
			break
		}
	}

	if forward {
		m.activeView = views[(currentIdx+1)%len(views)]
	} else {
		if currentIdx == 0 {
			m.activeView = views[len(views)-1]
		} else {
			m.activeView = views[currentIdx-1]
		}
	}

	m.activeSection = 0
	m.itemTable.MoveTop()
	m.setupSectionsForActiveView()
	m.previewPane.SetIssueMode(m.activeView == usecase.ViewModeIssues)
	if m.activeView == usecase.ViewModeNotifications {
		m.focusedPane = PaneTable
	}
	m.recalculateLayout()
}

func (m *DashboardModel) setupSectionsForActiveView() {
	var sections []usecase.SectionFilter
	if m.activeView == usecase.ViewModePRs {
		for _, s := range m.cfg.PRSections {
			sections = append(sections, usecase.SectionFilter{
				Title:   s.Title,
				Filters: s.Filters,
				Limit:   s.Limit,
			})
		}
	} else if m.activeView == usecase.ViewModeIssues {
		for _, s := range m.cfg.IssueSections {
			sections = append(sections, usecase.SectionFilter{
				Title:   s.Title,
				Filters: s.Filters,
				Limit:   s.Limit,
			})
		}
	} else {
		sections = []usecase.SectionFilter{
			{Title: "Unread", Filters: "", Limit: 30},
		}
	}
	m.sections = sections
}

func (m *DashboardModel) effectivePreview() bool {
	return m.previewOpen && m.activeView != usecase.ViewModeNotifications
}

func (m *DashboardModel) sortedClients() []ports.ForgeClient {
	clients := m.dashboardUC.ListClients()
	sort.Slice(clients, func(i, j int) bool {
		a, b := clients[i].Account(), clients[j].Account()
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return a.ID < b.ID
	})
	return clients
}

func (m *DashboardModel) activeChange() (change.ChangeRequest, bool) {
	if m.activeView != usecase.ViewModePRs {
		return change.ChangeRequest{}, false
	}
	idx := m.itemTable.Cursor()
	if idx < 0 || idx >= len(m.changes) {
		return change.ChangeRequest{}, false
	}
	return m.changes[idx], true
}

func (m *DashboardModel) syncFocus() {
	if !m.effectivePreview() {
		m.focusedPane = PaneTable
		m.itemTable.SetFocused(true)
		m.previewPane.SetFocused(false)
		return
	}
	if m.focusedPane == PanePreview {
		m.itemTable.SetFocused(false)
		m.previewPane.SetFocused(true)
	} else {
		m.itemTable.SetFocused(true)
		m.previewPane.SetFocused(false)
	}
}

func (m *DashboardModel) FocusedPane() FocusedPane {
	return m.focusedPane
}

func (m *DashboardModel) SetFocusedPane(pane FocusedPane) {
	m.focusedPane = pane
	m.syncFocus()
}
func (m *DashboardModel) PreviewOpen() bool {
	return m.previewOpen
}

func (m *DashboardModel) recalculateLayout() {
	m.syncFocus()

	if m.width > 0 {
		m.header.SetWidth(m.width)
		m.statusBar.SetWidth(m.width)
	}

	bodyHeight := m.height - 3
	if bodyHeight < 1 {
		bodyHeight = 1
	}

	w := m.width
	if w < 1 {
		w = 1
	}

	defer m.itemTable.Clamp(m.activeTotalCount())

	if !m.effectivePreview() {
		m.itemTable.SetSize(w, bodyHeight)
		m.previewPane.SetSize(w, bodyHeight)
		return
	}

	// Compact terminals show a single pane; both panes keep full-body sizes so
	// switching focus reveals the other without a relayout surprise.
	if w < 120 || m.height < 28 {
		m.itemTable.SetSize(w, bodyHeight)
		m.previewPane.SetSize(w, bodyHeight)
		return
	}

	if m.previewPos == "right" {
		tableWidth := (w - 1) * 55 / 100
		if tableWidth < 1 {
			tableWidth = 1
		}
		previewWidth := w - 1 - tableWidth
		if previewWidth < 1 {
			previewWidth = 1
			tableWidth = w - 2
			if tableWidth < 1 {
				tableWidth = 1
			}
		}
		m.itemTable.SetSize(tableWidth, bodyHeight)
		m.previewPane.SetSize(previewWidth, bodyHeight)
		return
	}

	tableHeight := (bodyHeight - 1) * 55 / 100
	if tableHeight < 1 {
		tableHeight = 1
	}
	previewHeight := bodyHeight - 1 - tableHeight
	if previewHeight < 1 {
		previewHeight = 1
		tableHeight = bodyHeight - 2
		if tableHeight < 1 {
			tableHeight = 1
		}
	}
	m.itemTable.SetSize(w, tableHeight)
	m.previewPane.SetSize(w, previewHeight)
}

func (m *DashboardModel) activeTotalCount() int {
	switch m.activeView {
	case usecase.ViewModePRs:
		return len(m.changes)
	case usecase.ViewModeIssues:
		return len(m.issues)
	case usecase.ViewModeNotifications:
		return len(m.notifications)
	}
	return 0
}

func (m *DashboardModel) ActiveView() usecase.ViewMode {
	return m.activeView
}

func (m *DashboardModel) activeItemURL() string {
	idx := m.itemTable.Cursor()
	switch m.activeView {
	case usecase.ViewModePRs:
		if idx >= 0 && idx < len(m.changes) {
			return m.changes[idx].HTMLURL
		}
	case usecase.ViewModeIssues:
		if idx >= 0 && idx < len(m.issues) {
			return m.issues[idx].HTMLURL
		}
	case usecase.ViewModeNotifications:
		if idx >= 0 && idx < len(m.notifications) {
			return m.notifications[idx].HTMLURL
		}
	}
	return ""
}

func (m *DashboardModel) activeItemNumber() string {
	idx := m.itemTable.Cursor()
	switch m.activeView {
	case usecase.ViewModePRs:
		if idx >= 0 && idx < len(m.changes) {
			return strconv.Itoa(m.changes[idx].Number)
		}
	case usecase.ViewModeIssues:
		if idx >= 0 && idx < len(m.issues) {
			return strconv.Itoa(m.issues[idx].Number)
		}
	case usecase.ViewModeNotifications:
		if idx >= 0 && idx < len(m.notifications) {
			return m.notifications[idx].ID
		}
	}
	return ""
}

func (m *DashboardModel) View() string {
	if m.width < 80 || m.height < 24 {
		return m.renderTooSmall()
	}

	accountName := "None"
	if !m.activeScope.IsUnified {
		if c, ok := m.dashboardUC.GetClient(m.activeScope.AccountID); ok {
			accountName = c.Account().Name
		}
	}
	headerView := m.header.Render(m.activeView, m.activeScope, accountName)

	counts := map[int]int{
		m.activeSection: m.activeTotalCount(),
	}
	sectionTabsView := m.sectionTabs.Render(m.sections, m.activeSection, counts, m.width)

	effectivePreview := m.effectivePreview()
	focusedPreview := effectivePreview && m.focusedPane == PanePreview
	compact := m.width < 120 || m.height < 28

	dividerStyle := lipgloss.NewStyle().Foreground(m.theme.P.BorderInactive)

	var tablePane []string
	if !effectivePreview || !compact || !focusedPreview {
		tablePane = fitBlock(m.renderActiveTable(), m.itemTable.Width(), m.itemTable.Height())
	}

	var bodyLines []string
	switch {
	case !effectivePreview:
		bodyLines = tablePane
	case compact:
		if focusedPreview {
			bodyLines = fitBlock(m.renderActivePreview(), m.previewPane.Width(), m.previewPane.Height())
		} else {
			bodyLines = tablePane
		}
	case m.previewPos == "right":
		previewLines := fitBlock(m.renderActivePreview(), m.previewPane.Width(), m.previewPane.Height())
		divider := dividerStyle.Render("│")
		bodyLines = make([]string, 0, len(tablePane))
		for i := range min(len(tablePane), len(previewLines)) {
			bodyLines = append(bodyLines, tablePane[i]+divider+previewLines[i])
		}
	default:
		previewLines := fitBlock(m.renderActivePreview(), m.previewPane.Width(), m.previewPane.Height())
		bodyLines = append(bodyLines, tablePane...)
		bodyLines = append(bodyLines, dividerStyle.Render(strings.Repeat("─", m.width)))
		bodyLines = append(bodyLines, previewLines...)
	}

	statusBarView := m.statusBar.Render(
		m.itemTable.Cursor(),
		m.activeTotalCount(),
		m.errors,
		m.loading,
		focusedPreview,
		effectivePreview,
	)
	if m.statusFlash != "" && time.Now().Before(m.statusFlashExpiry) {
		maxFlash := m.width - 2
		if maxFlash < 1 {
			maxFlash = 1
		}
		flashText := styles.Truncate("⚡ "+m.statusFlash, maxFlash)
		statusBarView = m.theme.StatusBar.Width(m.width).Render(
			lipgloss.NewStyle().Bold(true).Foreground(m.theme.P.Secondary).Render(flashText),
		)
	}

	lines := make([]string, 0, m.height)
	lines = append(lines, padDisplayLine(headerView, m.width))
	lines = append(lines, padDisplayLine(sectionTabsView, m.width))
	for _, line := range bodyLines {
		lines = append(lines, padDisplayLine(line, m.width))
	}
	lines = append(lines, padDisplayLine(statusBarView, m.width))

	for len(lines) > m.height {
		lines = lines[:m.height]
	}
	blank := strings.Repeat(" ", m.width)
	for len(lines) < m.height {
		lines = append(lines, blank)
	}

	fullView := strings.Join(lines, "\n")
	if m.activeModal != ModalNone {
		return placeOverlay(m.width, m.height, m.renderModal(), fullView)
	}

	return fullView
}

func (m *DashboardModel) renderActiveTable() string {
	switch m.activeView {
	case usecase.ViewModePRs:
		return m.itemTable.RenderChanges(m.changes)
	case usecase.ViewModeIssues:
		return m.itemTable.RenderIssues(m.issues)
	case usecase.ViewModeNotifications:
		return m.itemTable.RenderNotifications(m.notifications)
	}
	return ""
}

func (m *DashboardModel) renderActivePreview() string {
	idx := m.itemTable.Cursor()
	if m.activeView == usecase.ViewModePRs && idx >= 0 && idx < len(m.changes) {
		cr := m.changes[idx]
		return m.previewPane.RenderChange(&cr, m.curDiffs, m.curCommits, m.curChecks, m.curComments)
	}
	if m.activeView == usecase.ViewModeIssues && idx >= 0 && idx < len(m.issues) {
		iss := m.issues[idx]
		return m.previewPane.RenderIssue(&iss, m.curComments)
	}
	return m.previewPane.RenderChange(nil, nil, nil, nil, nil)
}

func (m *DashboardModel) renderModal() string {
	switch m.activeModal {
	case ModalHelp:
		return modals.RenderHelpModal(m.theme, m.width, m.height, m.helpPage)
	case ModalSearch:
		modalW := 54
		maxModalW := m.width - 6
		if modalW > maxModalW {
			modalW = maxModalW
		}
		if modalW < 10 {
			modalW = 10
		}
		inner := modalW - 4
		if inner < 1 {
			inner = 1
		}
		searchInput := m.searchInput
		searchInput.Width = inner
		searchContent := fmt.Sprintf("%s\n\n%s\n\n%s",
			m.theme.Title.Render("Filter Current Section"),
			searchInput.View(),
			m.theme.Subtitle.Render("Enter to apply • Esc to cancel"),
		)
		return m.theme.ModalBox.Width(modalW).Render(searchContent)
	case ModalComment:
		modalW := 64
		maxModalW := m.width - 6
		if modalW > maxModalW {
			modalW = maxModalW
		}
		if modalW < 10 {
			modalW = 10
		}
		inner := modalW - 4
		if inner < 1 {
			inner = 1
		}
		commentInput := m.commentInput
		commentInput.Width = inner
		commContent := fmt.Sprintf("%s\n\n%s\n\n%s",
			m.theme.Title.Render("Add Comment"),
			commentInput.View(),
			m.theme.Subtitle.Render("Enter to submit • Esc to cancel"),
		)
		return m.theme.ModalBox.Width(modalW).Render(commContent)
	case ModalConfirm:
		return modals.RenderConfirmModal(m.theme, "Confirmation Required", m.confirmPrompt, m.width, m.height)
	case ModalAccountSwitch:
		accounts := make([]account.Account, 0, len(m.accountModalClients))
		for _, c := range m.accountModalClients {
			accounts = append(accounts, c.Account())
		}
		return modals.RenderAccountModal(m.theme, accounts, m.activeScope.AccountID, m.activeScope.IsUnified, m.accountModalIndex, m.width, m.height)
	}
	return ""
}

func (m *DashboardModel) renderTooSmall() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(m.theme.P.Primary)
	wrapped := ansi.Wordwrap(
		fmt.Sprintf("Terminal window is too small (%dx%d). Minimum required: 80x24. Please resize.", m.width, m.height),
		m.width,
		"",
	)

	var b strings.Builder
	b.WriteString(titleStyle.Render("FORGEDASH"))
	b.WriteString("\n\n")
	b.WriteString(wrapped)

	lines := strings.Split(b.String(), "\n")
	for i := range lines {
		lines[i] = ansi.Truncate(lines[i], m.width, "")
	}

	return placeOverlay(m.width, m.height, strings.Join(lines, "\n"), "")
}

func padDisplayLine(s string, width int) string {
	if width <= 0 {
		return ""
	}
	w := ansi.StringWidth(s)
	if w > width {
		s = ansi.Truncate(s, width, "")
		w = ansi.StringWidth(s)
	}
	if w < width {
		s += strings.Repeat(" ", width-w)
	}
	return s
}

func fitBlock(content string, width, height int) []string {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	src := strings.Split(content, "\n")
	out := make([]string, 0, height)
	for i := range height {
		line := ""
		if i < len(src) {
			line = src[i]
		}
		out = append(out, padDisplayLine(line, width))
	}
	return out
}

func placeOverlay(w, h int, overlay, background string) string {
	if w <= 0 || h <= 0 {
		return ""
	}

	overlayLines := strings.Split(overlay, "\n")
	bgLines := strings.Split(background, "\n")

	ovW := 0
	for _, line := range overlayLines {
		if lw := ansi.StringWidth(line); lw > ovW {
			ovW = lw
		}
	}
	ovH := len(overlayLines)

	topPad := (h - ovH) / 2
	if topPad < 0 {
		topPad = 0
	}
	leftPad := (w - ovW) / 2
	if leftPad < 0 {
		leftPad = 0
	}
	if leftPad > w {
		leftPad = w
	}

	var b strings.Builder
	b.Grow((w + 1) * h)

	for y := range h {
		if y > 0 {
			b.WriteByte('\n')
		}

		if y >= topPad && y < topPad+ovH {
			line := overlayLines[y-topPad]
			avail := w - leftPad
			content := ""
			if avail > 0 {
				content = ansi.Truncate(line, avail, "")
			}
			if leftPad > 0 {
				b.WriteString(strings.Repeat(" ", leftPad))
			}
			b.WriteString(content)
			used := leftPad + ansi.StringWidth(content)
			if used < w {
				b.WriteString(strings.Repeat(" ", w-used))
			}
			continue
		}

		bg := ""
		if y < len(bgLines) {
			bg = bgLines[y]
		}
		b.WriteString(padDisplayLine(bg, w))
	}

	return b.String()
}

func openURL(rawURL string) error {
	return safebrowser.OpenURL(rawURL)
}

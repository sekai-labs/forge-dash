package screens

import (
	"context"
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

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
	"github.com/forgedash/forgedash/internal/interfaces/tui/styles"
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

	changes       []change.ChangeRequest
	issues        []issue.Issue
	notifications []notification.Notification
	errors        map[account.AccountID]error
	loading       bool

	curDiffs    []shared.DiffFile
	curCommits  []shared.Commit
	curChecks   []shared.CheckRun
	curComments []shared.Comment

	activeModal       ModalState
	searchInput       textinput.Model
	commentInput      textinput.Model
	accountModalIndex int
	confirmAction     func() tea.Cmd
	confirmPrompt     string
	statusFlash       string
	statusFlashExpiry time.Time

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

	return &DashboardModel{
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
		searchInput:   ti,
		commentInput:  ci,
		header:        components.NewHeader(theme),
		sectionTabs:   components.NewSectionTabs(theme),
		itemTable:     components.NewItemTable(theme),
		previewPane:   components.NewPreviewPane(theme),
		statusBar:     components.NewStatusBar(theme),
		errors:        make(map[account.AccountID]error),
	}
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
		m.header.SetWidth(m.width)
		m.statusBar.SetWidth(m.width)
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
			return m, nil
		case key.Matches(msg, m.keys.AccountModal):
			m.activeModal = ModalAccountSwitch
			m.accountModalIndex = 0
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
		}

		total := m.activeTotalCount()
		switch {
		case key.Matches(msg, m.keys.Down):
			m.itemTable.MoveDown(total)
			return m, m.fetchItemDetailCmd()
		case key.Matches(msg, m.keys.Up):
			m.itemTable.MoveUp()
			return m, m.fetchItemDetailCmd()
		case key.Matches(msg, m.keys.FirstLine):
			m.itemTable.MoveTop()
			return m, m.fetchItemDetailCmd()
		case key.Matches(msg, m.keys.LastLine):
			m.itemTable.MoveBottom(total)
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
		}

		switch {
		case key.Matches(msg, m.keys.TogglePreview):
			m.previewOpen = !m.previewOpen
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
		case key.Matches(msg, m.keys.NextPreviewTab):
			m.previewPane.NextTab()
			return m, nil
		case key.Matches(msg, m.keys.PrevPreviewTab):
			m.previewPane.PrevTab()
			return m, nil
		case key.Matches(msg, m.keys.PreviewPageDown):
			m.previewPane.ScrollDown()
			return m, nil
		case key.Matches(msg, m.keys.PreviewPageUp):
			m.previewPane.ScrollUp()
			return m, nil
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
			m.previewOpen = true
			m.previewPane.SetTab(components.PreviewTabFiles)
			m.recalculateLayout()
			return m, nil
		case key.Matches(msg, m.keys.Comment):
			m.activeModal = ModalComment
			m.commentInput.Reset()
			m.commentInput.Focus()
			return m, nil
		case key.Matches(msg, m.keys.Approve):
			if m.activeView == usecase.ViewModePRs && len(m.changes) > 0 {
				idx := m.itemTable.Cursor()
				cr := m.changes[idx]
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
			if m.activeView == usecase.ViewModePRs && len(m.changes) > 0 {
				idx := m.itemTable.Cursor()
				cr := m.changes[idx]
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
			if m.activeView == usecase.ViewModePRs && len(m.changes) > 0 {
				idx := m.itemTable.Cursor()
				cr := m.changes[idx]
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
			if m.activeView == usecase.ViewModePRs && len(m.changes) > 0 {
				idx := m.itemTable.Cursor()
				cr := m.changes[idx]
				return m, func() tea.Msg {
					err := m.changeUC.Reopen(context.Background(), cr)
					if err != nil {
						return statusFlashMsg(fmt.Sprintf("Failed to reopen: %v", err))
					}
					return statusFlashMsg("Reopened successfully!")
				}
			}
		case key.Matches(msg, m.keys.Checkout):
			if m.activeView == usecase.ViewModePRs && len(m.changes) > 0 {
				idx := m.itemTable.Cursor()
				cr := m.changes[idx]
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
		if msg.String() == "esc" || msg.String() == "?" || msg.String() == "q" {
			m.activeModal = ModalNone
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
		clients := m.dashboardUC.ListClients()
		totalOptions := len(clients) + 1

		switch msg.String() {
		case "esc", "q":
			m.activeModal = ModalNone
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
			} else {
				client := clients[m.accountModalIndex-1]
				m.activeScope = usecase.AccountScope{
					IsUnified: false,
					AccountID: client.Account().ID,
				}
			}
			m.loading = true
			m.itemTable.MoveTop()
			return m, m.fetchDataCmd()
		}
	}

	return m, nil
}

func (m *DashboardModel) submitCommentCmd(text string) tea.Cmd {
	return func() tea.Msg {
		if m.activeView == usecase.ViewModePRs && len(m.changes) > 0 {
			cr := m.changes[m.itemTable.Cursor()]
			_, err := m.changeUC.AddComment(context.Background(), cr, text)
			if err != nil {
				return statusFlashMsg(fmt.Sprintf("Failed to comment: %v", err))
			}
			return statusFlashMsg("Comment posted!")
		} else if m.activeView == usecase.ViewModeIssues && len(m.issues) > 0 {
			iss := m.issues[m.itemTable.Cursor()]
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

func (m *DashboardModel) recalculateLayout() {
	contentHeight := m.height - 5
	if contentHeight < 6 {
		contentHeight = 6
	}

	tableFullWidth := m.width - 2
	if tableFullWidth < 20 {
		tableFullWidth = 20
	}

	if !m.previewOpen {
		m.itemTable.SetSize(tableFullWidth, contentHeight)
		return
	}

	if m.previewPos == "right" {
		tableWidth := (m.width - 3) / 2
		if tableWidth < 25 {
			tableWidth = 25
		}
		previewWidth := m.width - tableWidth - 3
		if previewWidth < 20 {
			previewWidth = 20
		}
		m.itemTable.SetSize(tableWidth, contentHeight)
		m.previewPane.SetSize(previewWidth, contentHeight)
	} else {
		halfHeight := contentHeight / 2
		if halfHeight < 3 {
			halfHeight = 3
		}
		prevHeight := contentHeight - halfHeight
		if prevHeight < 3 {
			prevHeight = 3
		}
		m.itemTable.SetSize(tableFullWidth, halfHeight)
		m.previewPane.SetSize(tableFullWidth, prevHeight)
	}
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
	if m.width < 60 || m.height < 12 {
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			m.theme.Subtitle.Render("Terminal window is too small for ForgeDash. Please resize."),
		)
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
	sectionTabsView := m.sectionTabs.Render(m.sections, m.activeSection, counts)

	var mainContent string
	var tableContent string
	switch m.activeView {
	case usecase.ViewModePRs:
		tableContent = m.itemTable.RenderChanges(m.changes)
	case usecase.ViewModeIssues:
		tableContent = m.itemTable.RenderIssues(m.issues)
	case usecase.ViewModeNotifications:
		tableContent = m.itemTable.RenderNotifications(m.notifications)
	}

	tableBox := m.theme.Box.Render(tableContent)

	if !m.previewOpen {
		mainContent = tableBox
	} else {
		var previewContent string
		idx := m.itemTable.Cursor()
		if m.activeView == usecase.ViewModePRs && idx >= 0 && idx < len(m.changes) {
			cr := m.changes[idx]
			previewContent = m.previewPane.RenderChange(&cr, m.curDiffs, m.curCommits, m.curChecks, m.curComments)
		} else if m.activeView == usecase.ViewModeIssues && idx >= 0 && idx < len(m.issues) {
			iss := m.issues[idx]
			previewContent = m.previewPane.RenderIssue(&iss, m.curComments)
		} else {
			previewContent = m.previewPane.RenderChange(nil, nil, nil, nil, nil)
		}

		previewBox := m.theme.PreviewBox.Render(previewContent)

		if m.previewPos == "right" {
			mainContent = lipgloss.JoinHorizontal(lipgloss.Top, tableBox, " ", previewBox)
		} else {
			mainContent = lipgloss.JoinVertical(lipgloss.Left, tableBox, previewBox)
		}
	}

	statusBarView := m.statusBar.Render(m.itemTable.Cursor(), m.activeTotalCount(), m.errors, m.loading)
	if m.statusFlash != "" && time.Now().Before(m.statusFlashExpiry) {
		statusBarView = m.theme.StatusBar.Width(m.width).Render(
			lipgloss.NewStyle().Bold(true).Foreground(m.theme.P.Secondary).Render("⚡ " + m.statusFlash),
		)
	}

	fullView := lipgloss.JoinVertical(lipgloss.Left,
		headerView,
		sectionTabsView,
		mainContent,
		statusBarView,
	)

	if m.activeModal != ModalNone {
		var modalBox string
		switch m.activeModal {
		case ModalHelp:
			modalBox = modals.RenderHelpModal(m.theme, m.width, m.height)
		case ModalSearch:
			searchContent := fmt.Sprintf("%s\n\n%s\n\n%s",
				m.theme.Title.Render("Filter Current Section"),
				m.searchInput.View(),
				m.theme.Subtitle.Render("Enter to apply • Esc to cancel"),
			)
			modalBox = m.theme.ModalBox.Width(54).Render(searchContent)
		case ModalComment:
			commContent := fmt.Sprintf("%s\n\n%s\n\n%s",
				m.theme.Title.Render("Add Comment"),
				m.commentInput.View(),
				m.theme.Subtitle.Render("Enter to submit • Esc to cancel"),
			)
			modalBox = m.theme.ModalBox.Width(64).Render(commContent)
		case ModalConfirm:
			modalBox = modals.RenderConfirmModal(m.theme, "Confirmation Required", m.confirmPrompt)
		case ModalAccountSwitch:
			var accounts []account.Account
			for _, c := range m.dashboardUC.ListClients() {
				accounts = append(accounts, c.Account())
			}
			modalBox = modals.RenderAccountModal(m.theme, accounts, m.activeScope.AccountID, m.activeScope.IsUnified, m.accountModalIndex)
		}

		return placeOverlay(m.width, m.height, modalBox, fullView)
	}

	return fullView
}

func placeOverlay(w, h int, overlay, background string) string {
	overlayLines := strings.Split(overlay, "\n")
	bgLines := strings.Split(background, "\n")

	ovH := len(overlayLines)
	ovW := lipgloss.Width(overlay)

	topPad := (h - ovH) / 2
	if topPad < 0 {
		topPad = 0
	}
	leftPad := (w - ovW) / 2
	if leftPad < 0 {
		leftPad = 0
	}

	var res []string
	for y := range h {
		bgLine := ""
		if y < len(bgLines) {
			bgLine = bgLines[y]
		}

		if y >= topPad && y < topPad+ovH {
			ovLineIdx := y - topPad
			if ovLineIdx < len(overlayLines) {
				ovLine := overlayLines[ovLineIdx]
				leftSpacing := strings.Repeat(" ", leftPad)
				res = append(res, leftSpacing+ovLine)
				continue
			}
		}
		res = append(res, bgLine)
	}

	return strings.Join(res, "\n")
}

func openURL(rawURL string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", rawURL)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", rawURL)
	default:
		cmd = exec.Command("xdg-open", rawURL)
	}
	return cmd.Start()
}

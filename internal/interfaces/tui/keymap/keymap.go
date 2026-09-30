package keymap

import (
	"github.com/charmbracelet/bubbles/key"
)

type HelpGroup struct {
	Title    string
	Bindings []key.Binding
}

type KeyMap struct {
	Quit         key.Binding
	Help         key.Binding
	Refresh      key.Binding
	RefreshAll   key.Binding
	AccountModal key.Binding
	Search       key.Binding
	Esc          key.Binding

	Up           key.Binding
	Down         key.Binding
	Top          key.Binding
	Bottom       key.Binding
	FirstLine    key.Binding
	LastLine     key.Binding
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
	NextSection  key.Binding
	PrevSection  key.Binding
	Sec1         key.Binding
	Sec2         key.Binding
	Sec3         key.Binding
	Sec4         key.Binding
	Sec5         key.Binding
	NextView     key.Binding
	PrevView     key.Binding

	FocusNext   key.Binding
	FocusPrev   key.Binding
	ToggleFocus key.Binding

	TogglePreview    key.Binding
	TogglePreviewPos key.Binding
	NextPreviewTab   key.Binding
	PrevPreviewTab   key.Binding
	PreviewPageDown  key.Binding
	PreviewPageUp    key.Binding

	OpenBrowser key.Binding
	CopyNumber  key.Binding
	CopyURL     key.Binding
	DiffView    key.Binding
	Comment     key.Binding
	Approve     key.Binding
	Merge       key.Binding
	Close       key.Binding
	Reopen      key.Binding
	Checkout    key.Binding
}

func DefaultKeyMap() KeyMap {
	topBinding := key.NewBinding(
		key.WithKeys("g", "home"),
		key.WithHelp("g/home", "top"),
	)
	bottomBinding := key.NewBinding(
		key.WithKeys("G", "end"),
		key.WithHelp("G/end", "bottom"),
	)

	return KeyMap{

		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh section"),
		),
		RefreshAll: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "refresh all"),
		),
		AccountModal: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "accounts"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		Esc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel / close"),
		),

		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "down"),
		),
		Top:       topBinding,
		Bottom:    bottomBinding,
		FirstLine: topBinding,
		LastLine:  bottomBinding,
		HalfPageUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "half page up"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "half page down"),
		),
		NextSection: key.NewBinding(
			key.WithKeys("]", "right"),
			key.WithHelp("]", "next section"),
		),
		PrevSection: key.NewBinding(
			key.WithKeys("[", "left"),
			key.WithHelp("[", "prev section"),
		),
		Sec1: key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "section 1")),
		Sec2: key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "section 2")),
		Sec3: key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "section 3")),
		Sec4: key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "section 4")),
		Sec5: key.NewBinding(key.WithKeys("5"), key.WithHelp("5", "section 5")),
		NextView: key.NewBinding(
			key.WithKeys("L", "s"),
			key.WithHelp("L/s", "next view"),
		),
		PrevView: key.NewBinding(
			key.WithKeys("H", "S"),
			key.WithHelp("H/S", "prev view"),
		),

		FocusNext: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "focus next"),
		),
		FocusPrev: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "focus prev"),
		),
		ToggleFocus: key.NewBinding(
			key.WithKeys("h", "l"),
			key.WithHelp("h/l", "toggle focus"),
		),

		TogglePreview: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "toggle preview"),
		),
		TogglePreviewPos: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "toggle position"),
		),
		NextPreviewTab: key.NewBinding(
			key.WithKeys("]", "ctrl+n"),
			key.WithHelp("]/ctrl+n", "next tab"),
		),
		PrevPreviewTab: key.NewBinding(
			key.WithKeys("[", "ctrl+p"),
			key.WithHelp("[/ctrl+p", "prev tab"),
		),
		PreviewPageDown: key.NewBinding(
			key.WithKeys("J", "pgdown", "ctrl+d"),
			key.WithHelp("J/pgdn", "preview down"),
		),
		PreviewPageUp: key.NewBinding(
			key.WithKeys("K", "pgup", "ctrl+u"),
			key.WithHelp("K/pgup", "preview up"),
		),

		OpenBrowser: key.NewBinding(
			key.WithKeys("o", "enter"),
			key.WithHelp("o/enter", "open browser"),
		),
		CopyNumber: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "copy number"),
		),
		CopyURL: key.NewBinding(
			key.WithKeys("Y"),
			key.WithHelp("Y", "copy url"),
		),
		DiffView: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "view diff"),
		),
		Comment: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "comment"),
		),
		Approve: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "approve"),
		),
		Merge: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "merge"),
		),
		Close: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "close"),
		),
		Reopen: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "reopen"),
		),
		Checkout: key.NewBinding(
			key.WithKeys("C"),
			key.WithHelp("C", "checkout"),
		),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Help,
		k.Search,
		k.Refresh,
		k.TogglePreview,
		k.AccountModal,
		k.Quit,
	}
}

func (k KeyMap) HelpGroups() []HelpGroup {
	return []HelpGroup{
		{
			Title: "Navigation",
			Bindings: []key.Binding{
				k.Up,
				k.Down,
				k.Top,
				k.Bottom,
				k.HalfPageUp,
				k.HalfPageDown,
				k.NextSection,
				k.PrevSection,
			},
		},
		{
			Title: "Views & Sections",
			Bindings: []key.Binding{
				k.NextView,
				k.PrevView,
				k.Sec1,
				k.Sec2,
				k.Sec3,
				k.Sec4,
				k.Sec5,
			},
		},
		{
			Title: "Panels & Focus",
			Bindings: []key.Binding{
				k.FocusNext,
				k.FocusPrev,
				k.ToggleFocus,
			},
		},
		{
			Title: "Preview Pane",
			Bindings: []key.Binding{
				k.TogglePreview,
				k.TogglePreviewPos,
				k.NextPreviewTab,
				k.PrevPreviewTab,
				k.PreviewPageDown,
				k.PreviewPageUp,
			},
		},
		{
			Title: "Actions",
			Bindings: []key.Binding{
				k.OpenBrowser,
				k.CopyNumber,
				k.CopyURL,
				k.DiffView,
				k.Comment,
				k.Approve,
				k.Merge,
				k.Close,
				k.Reopen,
				k.Checkout,
			},
		},
		{
			Title: "Modals & Global",
			Bindings: []key.Binding{
				k.Search,
				k.AccountModal,
				k.Refresh,
				k.RefreshAll,
				k.Help,
				k.Esc,
				k.Quit,
			},
		},
	}
}

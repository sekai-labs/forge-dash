package keymap

import (
	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	Quit         key.Binding
	Help         key.Binding
	Refresh      key.Binding
	RefreshAll   key.Binding
	NextView     key.Binding
	PrevView     key.Binding
	AccountModal key.Binding
	Search       key.Binding
	Esc          key.Binding

	Up          key.Binding
	Down        key.Binding
	FirstLine   key.Binding
	LastLine    key.Binding
	NextSection key.Binding
	PrevSection key.Binding
	Sec1        key.Binding
	Sec2        key.Binding
	Sec3        key.Binding
	Sec4        key.Binding
	Sec5        key.Binding

	TogglePreview    key.Binding
	TogglePreviewPos key.Binding
	PrevPreviewTab   key.Binding
	NextPreviewTab   key.Binding
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
		NextView: key.NewBinding(
			key.WithKeys("tab", "s"),
			key.WithHelp("s/tab", "next view"),
		),
		PrevView: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev view"),
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
		FirstLine: key.NewBinding(
			key.WithKeys("g", "home"),
			key.WithHelp("g/home", "top"),
		),
		LastLine: key.NewBinding(
			key.WithKeys("G", "end"),
			key.WithHelp("G/end", "bottom"),
		),
		NextSection: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("l/→", "next section"),
		),
		PrevSection: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("h/←", "prev section"),
		),
		Sec1: key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "section 1")),
		Sec2: key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "section 2")),
		Sec3: key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "section 3")),
		Sec4: key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "section 4")),
		Sec5: key.NewBinding(key.WithKeys("5"), key.WithHelp("5", "section 5")),

		TogglePreview: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "toggle preview"),
		),
		TogglePreviewPos: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "toggle position"),
		),
		PrevPreviewTab: key.NewBinding(
			key.WithKeys("["),
			key.WithHelp("[", "prev tab"),
		),
		NextPreviewTab: key.NewBinding(
			key.WithKeys("]"),
			key.WithHelp("]", "next tab"),
		),
		PreviewPageDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "preview down"),
		),
		PreviewPageUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "preview up"),
		),

		OpenBrowser: key.NewBinding(
			key.WithKeys("enter", "o"),
			key.WithHelp("enter/o", "open browser"),
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
			key.WithKeys("d"),
			key.WithHelp("d", "view diff"),
		),
		Comment: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "comment"),
		),
		Approve: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "approve"),
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
			key.WithKeys("X"),
			key.WithHelp("X", "reopen"),
		),
		Checkout: key.NewBinding(
			key.WithKeys("C", " "),
			key.WithHelp("C/space", "checkout"),
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

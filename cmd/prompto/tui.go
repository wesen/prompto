package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/prompto/pkg"
	"github.com/go-go-golems/prompto/pkg/repositories"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Define styles
var (
	TitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFA500"))
	StatusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#BBBBBB"))
	SelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00"))
)

// PromptItem represents a selectable prompto item for the list
type PromptItem struct {
	prompto  pkg.Prompto
	filtered bool
}

func (i PromptItem) Title() string       { return i.prompto.Name }
func (i PromptItem) Description() string { return i.prompto.Type.String() }
func (i PromptItem) FilterValue() string  { return i.prompto.Name }

// RepositoryItem represents a selectable repository
type RepositoryItem struct {
	name      string
	path      string
	repository *pkg.Repository
	itemCount  int
}

func (i RepositoryItem) Title() string       { return i.name }
func (i RepositoryItem) Description() string { return fmt.Sprintf("%d prompts", i.itemCount) }
func (i RepositoryItem) FilterValue() string  { return i.name }

// KeyMap defines keybindings
type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Select   key.Binding
	Add      key.Binding
	Edit     key.Binding
	Delete   key.Binding
	DeleteConfirm key.Binding
	Filter   key.Binding
	Refresh  key.Binding
	Back     key.Binding
	Quit     key.Binding
	Help     key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Select},
		{k.Add, k.Edit, k.Delete, k.DeleteConfirm},
		{k.Filter, k.Refresh, k.Back},
		{k.Help, k.Quit},
	}
}

var DefaultKeyMap = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Add: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add directory"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	),
	Delete: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "delete"),
	),
	DeleteConfirm: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "confirm delete"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "backspace"),
		key.WithHelp("esc/backspace", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "q"),
		key.WithHelp("q/ctrl+c", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}

// Page represents different pages/states of the TUI
type Page int

const (
	PageRepositories Page = iota
	PagePrompts
	PageEditPrompt
	PageDeleteConfirm
	PageFilePicker
)

// Model holds the state of the TUI
type Model struct {
	repositories    []pkg.Repository
	repoList        list.Model
	promptList      list.Model
	filterInput     textinput.Model
	filePicker      filepicker.Model
	currentPage     Page
	selectedRepo    int
	selectedPrompt  int
	confirmDelete   bool
	status          string
	filtering       bool
	help            help.Model
	keyMap          KeyMap
}

// Init is called once when the program starts
func (m Model) Init() tea.Cmd {
	return nil
}

func NewModel(repositoryPaths []string) (Model, error) {
	// Create repository list
	repoConfig := repositories.NewRepositoryConfig(repositoryPaths)
	repos, err := repoConfig.LoadRepositories()
	if err != nil {
		return Model{}, err
	}

	// Initialize repository list items
	repoItems := make([]list.Item, len(repos))
	for i, repo := range repos {
		r := repo // local copy
		promptCount := len(r.GetPromptos())
		repoItems[i] = RepositoryItem{
			name:      r.Path,
			path:      r.Path,
			repository: &r,
			itemCount:  promptCount,
		}
	}

	// Create list delegates
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = SelectedStyle
	delegate.Styles.SelectedDesc = SelectedStyle

	// Initialize lists
	repoList := list.New(repoItems, delegate, 0, 0)
	repoList.Title = "Repositories"
	repoList.SetShowStatusBar(false)
	repoList.SetFilteringEnabled(false)
	repoList.Styles.Title = TitleStyle
	repoList.Styles.FilterPrompt = StatusStyle

	promptList := list.New([]list.Item{}, delegate, 0, 0)
	promptList.Title = "Prompts"
	promptList.SetShowStatusBar(false)
	promptList.SetFilteringEnabled(false)
	promptList.Styles.Title = TitleStyle
	promptList.Styles.FilterPrompt = StatusStyle

	// Initialize filter input
	filterInput := textinput.New()
	filterInput.Placeholder = "Filter..."
	filterInput.PromptStyle = StatusStyle

	// Initialize file picker
	fp := filepicker.New()
	fp.CurrentDirectory, _ = os.Getwd()
	fp.ShowHidden = false
	fp.AllowedTypes = []string{""} // Allow directories

	// Initialize help
	help := help.New()

	return Model{
		repositories: repos,
		repoList:     repoList,
		promptList:   promptList,
		filterInput:  filterInput,
		filePicker:   fp,
		status:       "Select a repository",
		currentPage:  PageRepositories,
		help:         help,
		keyMap:       DefaultKeyMap,
	}, nil
}

// Update handles all UI state changes
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global keybindings
		switch {
		case key.Matches(msg, m.keyMap.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keyMap.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		}

		// Page-specific keybindings
		switch m.currentPage {
		case PageRepositories:
			switch {
			case key.Matches(msg, m.keyMap.Select):
				i, ok := m.repoList.SelectedItem().(RepositoryItem)
				if ok && i.repository != nil {
					m.selectedRepo = m.repoList.Index()
					m.currentPage = PagePrompts
					m.status = fmt.Sprintf("Repository: %s", i.name)
					m.updatePromptList(*i.repository)
				}
				return m, nil

			case key.Matches(msg, m.keyMap.Add):
				// Show file picker to add a new repository
				m.currentPage = PageFilePicker
				m.status = "Select a directory to add as repository"
				return m, nil

			case key.Matches(msg, m.keyMap.Delete):
				i, ok := m.repoList.SelectedItem().(RepositoryItem)
				if ok {
					m.selectedRepo = m.repoList.Index()
					m.currentPage = PageDeleteConfirm
					m.status = fmt.Sprintf("Delete repository '%s'? (y/n)", i.name)
				}
				return m, nil

			case key.Matches(msg, m.keyMap.Refresh):
				return m, m.refreshRepositories()
			}

		case PageDeleteConfirm:
			switch {
			case key.Matches(msg, m.keyMap.DeleteConfirm):
				// Handle repository deletion
				_, ok := m.repoList.SelectedItem().(RepositoryItem)
				if ok {
					// This would require adding repository removal functionality
					// For now, we'll just remove it from our local list
					newRepos := make([]pkg.Repository, 0, len(m.repositories)-1)
					for idx, repo := range m.repositories {
						if idx != m.selectedRepo {
							newRepos = append(newRepos, repo)
						}
					}
					m.repositories = newRepos
					m.updateRepositoryList()
				}
				m.currentPage = PageRepositories
				m.status = "Repository removed"
				return m, nil

			case key.Matches(msg, m.keyMap.Back):
				// Cancel deletion
				m.currentPage = PageRepositories
				m.status = "Deletion cancelled"
				return m, nil
			}

		case PageFilePicker:
			// Handle file picker navigation
			var cmd tea.Cmd
			m.filePicker, cmd = m.filePicker.Update(msg)

			// Check if directory selected
			if msg.String() == "enter" && m.filePicker.CurrentDirectory != "" {
				// Add the selected directory as a repository
				selectedDir := m.filePicker.CurrentDirectory
				
				// Create a new repository
				newRepo := pkg.NewRepository(selectedDir)
				err := newRepo.LoadPromptos()
				if err != nil {
					m.status = fmt.Sprintf("Error loading repository: %v", err)
				} else {
					// Add to repositories and update the list
					m.repositories = append(m.repositories, *newRepo)
					m.updateRepositoryList()
					m.status = fmt.Sprintf("Added repository: %s", selectedDir)
				}
				m.currentPage = PageRepositories
			}

			// Handle back key
			if key.Matches(msg, m.keyMap.Back) {
				m.currentPage = PageRepositories
				m.status = "Repository addition cancelled"
			}

			return m, cmd

		case PagePrompts:
			switch {
			case key.Matches(msg, m.keyMap.Back):
				if m.filtering {
					m.filtering = false
					m.filterInput.Blur()
					m.status = fmt.Sprintf("Repository: %s", m.repositories[m.selectedRepo].Path)
				} else {
					m.currentPage = PageRepositories
					m.status = "Select a repository"
				}
				return m, nil

			case key.Matches(msg, m.keyMap.Filter):
				m.filtering = true
				m.filterInput.Focus()
				m.status = "Filter: "
				return m, nil

			case key.Matches(msg, m.keyMap.Select):
				if !m.filtering {
					i, ok := m.promptList.SelectedItem().(PromptItem)
					if ok {
						content := m.getPromptContent(i.prompto)
						m.status = fmt.Sprintf("Selected prompt: %s", i.prompto.Name)
						// Here you'd typically display the prompt content or handle executing it
						log.Info().Msgf("Selected prompt: %s\n%s", i.prompto.Name, content)
					}
				}
				return m, nil

			case key.Matches(msg, m.keyMap.Edit):
				i, ok := m.promptList.SelectedItem().(PromptItem)
				if ok {
					// Here you would trigger an editor to open the prompt file
					m.status = fmt.Sprintf("Editing prompt: %s", i.prompto.Name)
					return m, m.editPrompt(i.prompto)
				}
				return m, nil

			case key.Matches(msg, m.keyMap.Delete):
				i, ok := m.promptList.SelectedItem().(PromptItem)
				if ok {
					m.selectedPrompt = m.promptList.Index()
					m.currentPage = PageDeleteConfirm
					m.status = fmt.Sprintf("Delete prompt '%s'? (y/n)", i.prompto.Name)
					return m, nil
				}
				return m, nil

			case key.Matches(msg, m.keyMap.DeleteConfirm):
				if m.currentPage == PageDeleteConfirm {
					i, ok := m.promptList.SelectedItem().(PromptItem)
					if ok {
						m.status = fmt.Sprintf("Deleting prompt: %s", i.prompto.Name)
						m.currentPage = PagePrompts
						return m, m.deletePrompt(i.prompto)
					}
				}
				return m, nil

			case key.Matches(msg, m.keyMap.Add):
				// Add a new prompt
				return m, m.addNewPrompt()

			case key.Matches(msg, m.keyMap.Refresh):
				return m, m.refreshPrompts()
			}
		}

		// Handle filtering
		if m.filtering {
			var cmd tea.Cmd
			m.filterInput, cmd = m.filterInput.Update(msg)

			// Apply filter to prompt list
			m.filterPrompts(m.filterInput.Value())
			return m, cmd
		}

	case filteredPromptsMsg:
		m.updatePromptListItems(msg)
		return m, nil

	case refreshRepositoriesMsg:
		repoConfig := repositories.NewRepositoryConfig(getRepositoryPaths())
		repos, err := repoConfig.LoadRepositories()
		if err != nil {
			m.status = fmt.Sprintf("Error refreshing repositories: %v", err)
		} else {
			m.repositories = repos
			m.updateRepositoryList()
			m.status = "Repositories refreshed"
		}
		return m, nil

	case refreshPromptsMsg:
		repo := m.repositories[m.selectedRepo]
		err := repo.Refresh()
		if err != nil {
			m.status = fmt.Sprintf("Error refreshing prompts: %v", err)
		} else {
			m.repositories[m.selectedRepo] = repo
			m.updatePromptList(repo)
			m.status = "Prompts refreshed"
		}
		return m, nil

	case promptDeletedMsg:
		// Refresh the prompt list after deletion
		repo := m.repositories[m.selectedRepo]
		err := repo.Refresh()
		if err != nil {
			m.status = fmt.Sprintf("Error refreshing after delete: %v", err)
		} else {
			m.repositories[m.selectedRepo] = repo
			m.updatePromptList(repo)
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.repoList.SetWidth(msg.Width)
		m.repoList.SetHeight(msg.Height - 4) // Leave room for help and status
		m.promptList.SetWidth(msg.Width)
		m.promptList.SetHeight(msg.Height - 4)
		m.help.Width = msg.Width
		return m, nil
	}

	// Handle list updates
	switch m.currentPage {
	case PageRepositories:
		var cmd tea.Cmd
		m.repoList, cmd = m.repoList.Update(msg)
		return m, cmd

	case PagePrompts:
		if !m.filtering {
			var cmd tea.Cmd
			m.promptList, cmd = m.promptList.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// View renders the current UI state
func (m Model) View() string {
	var content string

	switch m.currentPage {
	case PageRepositories:
		content = BorderStyle.Render(m.repoList.View())

	case PagePrompts:
		var promptView string
		if m.filtering {
			promptView = m.promptList.View() + "\n" + StatusStyle.Render("Filter: ") + m.filterInput.View()
		} else {
			promptView = m.promptList.View()
		}
		content = BorderStyle.Render(promptView)
		
	case PageDeleteConfirm:
		// Show confirmation dialog based on which list we were viewing
		if m.repoList.Items() != nil && len(m.repoList.Items()) > 0 {
			content = BorderStyle.Render(m.repoList.View())
		} else {
			content = BorderStyle.Render(m.promptList.View())
		}
		
	case PageFilePicker:
		content = BorderStyle.Render(m.filePicker.View())
	}

	header := TitleStyle.Render("Prompto TUI") + "\n" + StatusStyle.Render(m.status)
	footer := m.help.View(m.keyMap)

	return fmt.Sprintf("%s\n\n%s\n\n%s", header, content, footer)
}

// Messages for asynchronous operations
type refreshRepositoriesMsg struct{}
type refreshPromptsMsg struct{}
type filteredPromptsMsg []list.Item
type promptDeletedMsg struct{}

// Helper methods
func (m *Model) updateRepositoryList() {
	items := make([]list.Item, len(m.repositories))
	for i, repo := range m.repositories {
		repo := repo // local copy
		items[i] = RepositoryItem{
			name:      repo.Path,
			path:      repo.Path,
			repository: &repo,
		}
	}
	m.repoList.SetItems(items)
}

func (m *Model) updatePromptList(repo pkg.Repository) {
	prompts := repo.GetPromptos()
	sort.Slice(prompts, func(i, j int) bool {
		return prompts[i].Name < prompts[j].Name
	})

	items := make([]list.Item, len(prompts))
	for i, prompt := range prompts {
		prompt := prompt // local copy
		items[i] = PromptItem{
			prompto: prompt,
		}
	}
	m.promptList.SetItems(items)
}

func (m *Model) filterPrompts(filter string) tea.Cmd {
	if filter == "" {
		// Reset filter
		repo := m.repositories[m.selectedRepo]
		m.updatePromptList(repo)
		return nil
	}

	// Apply filter asynchronously
	return func() tea.Msg {
		repo := m.repositories[m.selectedRepo]
		allPrompts := repo.GetPromptos()
		filtered := []list.Item{}

		filterLower := strings.ToLower(filter)
		for _, prompt := range allPrompts {
			if strings.Contains(strings.ToLower(prompt.Name), filterLower) {
				filtered = append(filtered, PromptItem{
					prompto: prompt,
				})
			}
		}

		return filteredPromptsMsg(filtered)
	}
}

func (m *Model) updatePromptListItems(items filteredPromptsMsg) {
	m.promptList.SetItems(items)
}

func (m *Model) getPromptContent(prompt pkg.Prompto) string {
	content, err := os.ReadFile(prompt.FilePath)
	if err != nil {
		return fmt.Sprintf("Error reading prompt: %v", err)
	}
	return string(content)
}

func (m *Model) refreshRepositories() tea.Cmd {
	return func() tea.Msg {
		return refreshRepositoriesMsg{}
	}
}

func (m *Model) refreshPrompts() tea.Cmd {
	return func() tea.Msg {
		return refreshPromptsMsg{}
	}
}

func (m *Model) editPrompt(prompt pkg.Prompto) tea.Cmd {
	return func() tea.Msg {
		// Here you'd integrate with an editor to open the file
		// For now, we'll use EDITOR environment variable
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi" // Default to vi if no EDITOR is set
		}

		// Execute the editor
		cmd := exec.Command(editor, prompt.FilePath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
		
		return refreshPromptsMsg{}
	}
}

func (m *Model) deletePrompt(prompt pkg.Prompto) tea.Cmd {
	return func() tea.Msg {
		err := os.Remove(prompt.FilePath)
		if err != nil {
			m.status = fmt.Sprintf("Error deleting prompt: %v", err)
		}
		return promptDeletedMsg{}
	}
}

func (m *Model) addNewPrompt() tea.Cmd {
	return func() tea.Msg {
		// This is a simplistic implementation - would need more UI to get user input
		repo := m.repositories[m.selectedRepo]
		promptDir := fmt.Sprintf("%s/prompto/new", repo.Path)
		
		// Ensure the directory exists
		err := os.MkdirAll(promptDir, 0755)
		if err != nil {
			m.status = fmt.Sprintf("Error creating directory: %v", err)
			return nil
		}
		
		// Create a new prompt file
		filePath := fmt.Sprintf("%s/new_prompt.txt", promptDir)
		f, err := os.Create(filePath)
		if err != nil {
			m.status = fmt.Sprintf("Error creating file: %v", err)
			return nil
		}
		f.Close()
		
		// Open the new file in editor
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}
		
		// Execute the editor
		cmd := exec.Command(editor, filePath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
		
		// Refresh the prompt list
		return refreshPromptsMsg{}
	}
}

// Helper function to get repository paths from config
func getRepositoryPaths() []string {
	return viper.GetStringSlice("repositories")
}

// NewTUICommand creates a new cobra command for the TUI
func NewTUICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Start the Prompto TUI",
		Long:  "Launch an interactive terminal UI for managing Prompto repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get repository paths from viper config
			repositories := viper.GetStringSlice("repositories")
			if len(repositories) == 0 {
				return fmt.Errorf("no repositories configured - add repositories to your config file")
			}

			// Create logger that doesn't interfere with the UI
			zlog := zerolog.New(zerolog.NewConsoleWriter(
				func(w *zerolog.ConsoleWriter) {
					w.Out = os.Stderr
					w.NoColor = false
				}))
			log.Logger = zlog

			// Create and start the model
			p := tea.NewProgram(
				initialized(repositories),
				tea.WithAltScreen(),
				tea.WithMouseCellMotion(),
			)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Start file watchers for repositories
			for _, repoPath := range repositories {
				repo := pkg.NewRepository(repoPath)
				err := repo.LoadPromptos()
				if err != nil {
					return err
				}

				go func(r *pkg.Repository) {
					err := r.Watch(ctx)
					if err != nil {
						log.Error().Err(err).Str("repository", r.Path).Msg("Error watching repository")
					}
				}(repo)
			}

			if _, err := p.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}

// initialized returns a ready-to-use model or exits on error
func initialized(repositories []string) Model {
	m, err := NewModel(repositories)
	if err != nil {
		fmt.Printf("Error initializing TUI: %v\n", err)
		os.Exit(1)
	}
	return m
}
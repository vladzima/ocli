package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

type EditMode int

const (
	EditModeNone EditMode = iota
	EditModeNew
	EditModeEdit
)

type AppMode int

const (
	AppModeNormal AppMode = iota
	AppModeSettings
	AppModeHelp
)

type Settings struct {
	ShowHierarchyLines bool
}

type Model struct {
	rootBullets     []*Bullet
	allBullets      []*Bullet
	selectedIndex   int
	editMode        EditMode
	appMode         AppMode
	textInput       textinput.Model
	editingBullet   *Bullet
	width           int
	height          int
	settings        Settings
	settingsIndex   int
	zoomedBullet    *Bullet
	breadcrumbs     []*Bullet
	configManager   *ConfigManager
	scrollOffset    int
}

func NewModel() Model {
	// Force color profile for SSH terminals
	lipgloss.SetColorProfile(termenv.ANSI256)
	
	ti := textinput.New()
	ti.Placeholder = "Enter text..."
	ti.Focus()
	ti.CharLimit = 256

	// Initialize config manager
	configManager, err := NewConfigManager()
	if err != nil {
		// Fallback to default if config fails
		configManager = nil
	}

	m := Model{
		rootBullets:   make([]*Bullet, 0),
		allBullets:    make([]*Bullet, 0),
		textInput:     ti,
		editMode:      EditModeNone,
		appMode:       AppModeNormal,
		settingsIndex: 0,
		zoomedBullet:  nil,
		breadcrumbs:   make([]*Bullet, 0),
		configManager: configManager,
	}

	// Load data from config or use defaults
	if configManager != nil {
		if data, err := configManager.Load(); err == nil {
			m.rootBullets = data.RootBullets
			m.settings = data.Settings
		} else {
			// Use defaults if loading fails
			m.loadDefaults()
		}
	} else {
		// Use defaults if config manager failed to initialize
		m.loadDefaults()
	}

	m.rebuildVisibleList()
	m.ensureSelectedVisible()
	return m
}

func (m *Model) loadDefaults() {
	m.settings = Settings{
		ShowHierarchyLines: true,
	}

	// Use the same comprehensive tutorial as persistence layer
	if m.configManager != nil {
		defaultData := m.configManager.createDefaultData()
		m.rootBullets = defaultData.RootBullets
		m.settings = defaultData.Settings
	} else {
		// Fallback if config manager is nil
		root := NewBullet("Welcome to OCLI!")
		root.AddChild(NewBullet("Press Enter to add a new bullet"))
		root.AddChild(NewBullet("Use arrow keys to navigate"))
		root.AddChild(NewBullet("Press 'h' for help"))
		m.rootBullets = []*Bullet{root}
	}
}

func (m *Model) saveData() error {
	if m.configManager == nil {
		return nil // No config manager, skip saving
	}

	data := &AppData{
		RootBullets: m.rootBullets,
		Settings:    m.settings,
	}

	return m.configManager.Save(data)
}

func (m *Model) ensureSelectedVisible() {
	if m.height == 0 {
		return
	}
	
	// Calculate available space for content (accounting for title, breadcrumbs, help text, etc.)
	availableHeight := m.height - 6 // Title (2 lines) + breadcrumbs (2 lines) + help (2 lines)
	if m.editMode == EditModeNew {
		availableHeight -= 2 // New bullet input
	}
	
	// Ensure selected item is visible in viewport
	if m.selectedIndex < m.scrollOffset {
		m.scrollOffset = m.selectedIndex
	} else if m.selectedIndex >= m.scrollOffset+availableHeight {
		m.scrollOffset = m.selectedIndex - availableHeight + 1
	}
	
	// Ensure scroll offset doesn't go negative
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
	
	// Ensure we don't scroll past the content
	maxScroll := len(m.allBullets) - availableHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scrollOffset > maxScroll {
		m.scrollOffset = maxScroll
	}
}

func (m *Model) rebuildVisibleList() {
	m.allBullets = make([]*Bullet, 0)
	
	if m.zoomedBullet != nil {
		// When zoomed, only show the zoomed bullet and its children
		m.allBullets = append(m.allBullets, m.zoomedBullet)
		m.allBullets = append(m.allBullets, m.zoomedBullet.GetVisibleDescendants()...)
	} else {
		// Normal view - show all bullets
		for _, root := range m.rootBullets {
			m.allBullets = append(m.allBullets, root)
			m.allBullets = append(m.allBullets, root.GetVisibleDescendants()...)
		}
	}
}

func (m *Model) getSelectedBullet() *Bullet {
	if m.selectedIndex >= 0 && m.selectedIndex < len(m.allBullets) {
		return m.allBullets[m.selectedIndex]
	}
	return nil
}

func (m *Model) addNewBullet(content string) {
	newBullet := NewBullet(content)
	selected := m.getSelectedBullet()

	// When zoomed in, new bullets should be children of the zoomed bullet
	if m.zoomedBullet != nil {
		if selected == nil || selected == m.zoomedBullet {
			// If no selection or selected is the zoomed bullet, add as child
			m.zoomedBullet.AddChild(newBullet)
		} else {
			// Add as sibling to selected bullet (child of same parent)
			parent := selected.Parent
			if parent == nil {
				// This shouldn't happen when zoomed, but fallback
				m.zoomedBullet.AddChild(newBullet)
			} else {
				index := 0
				for i, b := range parent.Children {
					if b.ID == selected.ID {
						index = i + 1
						break
					}
				}
				parent.InsertChildAt(index, newBullet)
			}
		}
	} else {
		// Normal (non-zoomed) behavior
		if selected == nil {
			m.rootBullets = append(m.rootBullets, newBullet)
		} else {
			if selected.Parent == nil {
				index := 0
				for i, b := range m.rootBullets {
					if b.ID == selected.ID {
						index = i + 1
						break
					}
				}
				m.rootBullets = append(m.rootBullets[:index], append([]*Bullet{newBullet}, m.rootBullets[index:]...)...)
			} else {
				parent := selected.Parent
				index := 0
				for i, b := range parent.Children {
					if b.ID == selected.ID {
						index = i + 1
						break
					}
				}
				parent.InsertChildAt(index, newBullet)
			}
		}
	}

	m.rebuildVisibleList()
	for i, b := range m.allBullets {
		if b.ID == newBullet.ID {
			m.selectedIndex = i
			break
		}
	}
	m.ensureSelectedVisible()
	
	// Auto-save after adding new bullet
	m.saveData()
}

func (m *Model) deleteBullet() {
	selected := m.getSelectedBullet()
	if selected == nil {
		return
	}

	if selected.Parent == nil {
		for i, b := range m.rootBullets {
			if b.ID == selected.ID {
				m.rootBullets = append(m.rootBullets[:i], m.rootBullets[i+1:]...)
				break
			}
		}
	} else {
		selected.Parent.RemoveChild(selected)
	}

	m.rebuildVisibleList()
	if m.selectedIndex >= len(m.allBullets) && m.selectedIndex > 0 {
		m.selectedIndex--
	}
	m.ensureSelectedVisible()
	
	// Auto-save after deleting bullet
	m.saveData()
}

func (m *Model) indentBullet() {
	selected := m.getSelectedBullet()
	if selected == nil || m.selectedIndex == 0 {
		return
	}

	var prevSibling *Bullet
	if selected.Parent == nil {
		for i, b := range m.rootBullets {
			if b.ID == selected.ID && i > 0 {
				prevSibling = m.rootBullets[i-1]
				m.rootBullets = append(m.rootBullets[:i], m.rootBullets[i+1:]...)
				break
			}
		}
	} else {
		parent := selected.Parent
		for i, b := range parent.Children {
			if b.ID == selected.ID && i > 0 {
				prevSibling = parent.Children[i-1]
				parent.RemoveChild(selected)
				break
			}
		}
	}

	if prevSibling != nil {
		prevSibling.AddChild(selected)
		if prevSibling.Collapsed {
			prevSibling.Collapsed = false
		}
	}

	m.rebuildVisibleList()
	m.ensureSelectedVisible()
}

func (m *Model) outdentBullet() {
	selected := m.getSelectedBullet()
	if selected == nil || selected.Parent == nil {
		return
	}

	parent := selected.Parent
	grandparent := parent.Parent

	parentIndex := 0
	if grandparent == nil {
		for i, b := range m.rootBullets {
			if b.ID == parent.ID {
				parentIndex = i
				break
			}
		}
	} else {
		for i, b := range grandparent.Children {
			if b.ID == parent.ID {
				parentIndex = i
				break
			}
		}
	}

	parent.RemoveChild(selected)

	if grandparent == nil {
		m.rootBullets = append(m.rootBullets[:parentIndex+1], append([]*Bullet{selected}, m.rootBullets[parentIndex+1:]...)...)
	} else {
		grandparent.InsertChildAt(parentIndex+1, selected)
	}

	m.rebuildVisibleList()
	m.ensureSelectedVisible()
}

func (m *Model) moveBulletUp() {
	selected := m.getSelectedBullet()
	if selected == nil {
		return
	}

	// Find the target depth to maintain indentation level
	var targetDepth int
	if m.zoomedBullet != nil {
		targetDepth = selected.GetDepthFrom(m.zoomedBullet)
	} else {
		targetDepth = selected.GetDepth()
	}

	// Look for the previous item at the same depth in the visible list
	var targetItem *Bullet
	for i := m.selectedIndex - 1; i >= 0; i-- {
		bullet := m.allBullets[i]
		var bulletDepth int
		if m.zoomedBullet != nil {
			bulletDepth = bullet.GetDepthFrom(m.zoomedBullet)
		} else {
			bulletDepth = bullet.GetDepth()
		}
		
		if bulletDepth == targetDepth {
			targetItem = bullet
			break
		}
	}

	if targetItem == nil {
		// If no item at same depth, try to find target at parent level
		if targetDepth > 0 {
			targetDepth--
			for i := m.selectedIndex - 1; i >= 0; i-- {
				bullet := m.allBullets[i]
				var bulletDepth int
				if m.zoomedBullet != nil {
					bulletDepth = bullet.GetDepthFrom(m.zoomedBullet)
				} else {
					bulletDepth = bullet.GetDepth()
				}
				
				if bulletDepth == targetDepth {
					targetItem = bullet
					break
				}
			}
		}
		
		if targetItem == nil {
			return // Still no target found
		}
	}

	// Remove selected from its current parent
	if selected.Parent == nil {
		for i, b := range m.rootBullets {
			if b.ID == selected.ID {
				m.rootBullets = append(m.rootBullets[:i], m.rootBullets[i+1:]...)
				break
			}
		}
	} else {
		selected.Parent.RemoveChild(selected)
	}

	// Insert selected before targetItem at the target's level
	var targetItemDepth int
	if m.zoomedBullet != nil {
		targetItemDepth = targetItem.GetDepthFrom(m.zoomedBullet)
	} else {
		targetItemDepth = targetItem.GetDepth()
	}
	
	if targetItemDepth == 0 {
		// Target is root level
		for i, b := range m.rootBullets {
			if b.ID == targetItem.ID {
				m.rootBullets = append(m.rootBullets[:i], append([]*Bullet{selected}, m.rootBullets[i:]...)...)
				break
			}
		}
		selected.Parent = nil
	} else {
		// Insert as sibling of targetItem
		parent := targetItem.Parent
		for i, b := range parent.Children {
			if b.ID == targetItem.ID {
				parent.InsertChildAt(i, selected)
				break
			}
		}
	}

	m.rebuildVisibleList()
	if m.selectedIndex > 0 {
		m.selectedIndex--
	}
	m.ensureSelectedVisible()
}

func (m *Model) moveBulletDown() {
	selected := m.getSelectedBullet()
	if selected == nil {
		return
	}

	// Find the target depth to maintain indentation level
	var targetDepth int
	if m.zoomedBullet != nil {
		targetDepth = selected.GetDepthFrom(m.zoomedBullet)
	} else {
		targetDepth = selected.GetDepth()
	}

	// Look for the next item at the same depth in the visible list
	var targetItem *Bullet
	for i := m.selectedIndex + 1; i < len(m.allBullets); i++ {
		bullet := m.allBullets[i]
		var bulletDepth int
		if m.zoomedBullet != nil {
			bulletDepth = bullet.GetDepthFrom(m.zoomedBullet)
		} else {
			bulletDepth = bullet.GetDepth()
		}
		
		if bulletDepth == targetDepth {
			targetItem = bullet
			break
		}
	}

	if targetItem == nil {
		// If no item at same depth, try to find target at parent level
		if targetDepth > 0 {
			targetDepth--
			for i := m.selectedIndex + 1; i < len(m.allBullets); i++ {
				bullet := m.allBullets[i]
				var bulletDepth int
				if m.zoomedBullet != nil {
					bulletDepth = bullet.GetDepthFrom(m.zoomedBullet)
				} else {
					bulletDepth = bullet.GetDepth()
				}
				
				if bulletDepth == targetDepth {
					targetItem = bullet
					break
				}
			}
		}
		
		if targetItem == nil {
			return // Still no target found
		}
	}

	// Remove selected from its current parent
	if selected.Parent == nil {
		for i, b := range m.rootBullets {
			if b.ID == selected.ID {
				m.rootBullets = append(m.rootBullets[:i], m.rootBullets[i+1:]...)
				break
			}
		}
	} else {
		selected.Parent.RemoveChild(selected)
	}

	// Insert selected after targetItem at the target's level
	var targetItemDepth int
	if m.zoomedBullet != nil {
		targetItemDepth = targetItem.GetDepthFrom(m.zoomedBullet)
	} else {
		targetItemDepth = targetItem.GetDepth()
	}
	
	if targetItemDepth == 0 {
		// Target is root level
		for i, b := range m.rootBullets {
			if b.ID == targetItem.ID {
				m.rootBullets = append(m.rootBullets[:i+1], append([]*Bullet{selected}, m.rootBullets[i+1:]...)...)
				break
			}
		}
		selected.Parent = nil
	} else {
		// Insert as sibling of targetItem
		parent := targetItem.Parent
		for i, b := range parent.Children {
			if b.ID == targetItem.ID {
				parent.InsertChildAt(i+1, selected)
				break
			}
		}
	}

	m.rebuildVisibleList()
	if m.selectedIndex < len(m.allBullets)-1 {
		m.selectedIndex++
	}
	m.ensureSelectedVisible()
}

func (m *Model) zoomIn() {
	selected := m.getSelectedBullet()
	if selected == nil {
		return
	}
	
	// Build breadcrumbs path to the selected bullet
	var path []*Bullet
	current := selected
	for current != nil {
		path = append([]*Bullet{current}, path...)
		current = current.Parent
	}
	
	// Remove the selected bullet from breadcrumbs (it becomes the zoomed view)
	if len(path) > 0 {
		m.breadcrumbs = path[:len(path)-1]
	}
	
	m.zoomedBullet = selected
	m.selectedIndex = 0
	m.scrollOffset = 0
	m.rebuildVisibleList()
	m.ensureSelectedVisible()
}

func (m *Model) zoomOut() {
	if len(m.breadcrumbs) > 0 {
		// Zoom out to parent
		m.zoomedBullet = m.breadcrumbs[len(m.breadcrumbs)-1]
		m.breadcrumbs = m.breadcrumbs[:len(m.breadcrumbs)-1]
	} else {
		// Zoom out to root view
		m.zoomedBullet = nil
		m.breadcrumbs = make([]*Bullet, 0)
	}
	
	m.selectedIndex = 0
	m.scrollOffset = 0
	m.rebuildVisibleList()
	m.ensureSelectedVisible()
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.appMode == AppModeSettings {
			switch msg.String() {
			case "q", "esc", "s":
				m.appMode = AppModeNormal
				return m, nil
				
			case "up", "k":
				if m.settingsIndex > 0 {
					m.settingsIndex--
				}
				
			case "down", "j":
				if m.settingsIndex < 0 { // We only have 1 setting for now
					m.settingsIndex++
				}
				
			case "enter", " ", "space":
				switch m.settingsIndex {
				case 0: // Toggle hierarchy lines
					m.settings.ShowHierarchyLines = !m.settings.ShowHierarchyLines
					// Auto-save after settings change
					m.saveData()
				}
			}
			return m, nil
		}
		
		if m.appMode == AppModeHelp {
			switch msg.String() {
			case "q", "esc", "h":
				m.appMode = AppModeNormal
				return m, nil
			}
			return m, nil
		}
		
		if m.editMode != EditModeNone {
			switch msg.String() {
			case "enter":
				content := m.textInput.Value()
				if m.editMode == EditModeNew {
					if content != "" {
						m.addNewBullet(content)
					}
				} else if m.editMode == EditModeEdit && m.editingBullet != nil {
					m.editingBullet.Content = content
					m.editingBullet.IsEditing = false
					// Auto-save after editing content
					m.saveData()
				}
				m.editMode = EditModeNone
				m.editingBullet = nil
				m.textInput.SetValue("")
				m.textInput.Blur()
				return m, nil

			case "esc":
				m.editMode = EditModeNone
				if m.editingBullet != nil {
					m.editingBullet.IsEditing = false
					m.editingBullet = nil
				}
				m.textInput.SetValue("")
				m.textInput.Blur()
				return m, nil

			default:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			// Save data before quitting
			m.saveData()
			return m, tea.Quit

		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
				m.ensureSelectedVisible()
			}

		case "down", "j":
			if m.selectedIndex < len(m.allBullets)-1 {
				m.selectedIndex++
				m.ensureSelectedVisible()
			}

		case "enter":
			m.editMode = EditModeNew
			m.textInput.SetValue("")
			m.textInput.Focus()
			return m, textinput.Blink

		case "e":
			if selected := m.getSelectedBullet(); selected != nil {
				m.editMode = EditModeEdit
				m.editingBullet = selected
				selected.IsEditing = true
				m.textInput.SetValue(selected.Content)
				m.textInput.Focus()
				m.textInput.SetCursor(len(selected.Content))
				return m, textinput.Blink
			}

		case "d":
			m.deleteBullet()

		case "tab":
			m.indentBullet()

		case "shift+tab":
			m.outdentBullet()

		case " ", "space":
			if selected := m.getSelectedBullet(); selected != nil {
				selected.Toggle()
				m.rebuildVisibleList()
				m.ensureSelectedVisible()
			}

		case "shift+up":
			m.moveBulletUp()

		case "shift+down":
			m.moveBulletDown()

		case "c":
			if selected := m.getSelectedBullet(); selected != nil {
				selected.CycleColor()
			}

		case "t":
			if selected := m.getSelectedBullet(); selected != nil {
				selected.ToggleTask()
			}

		case "x":
			if selected := m.getSelectedBullet(); selected != nil {
				selected.ToggleComplete()
			}
			
		case "s":
			m.appMode = AppModeSettings
			m.settingsIndex = 0
			
		case "h":
			m.appMode = AppModeHelp
			
		case "right":
			m.zoomIn()
			
		case "left":
			m.zoomOut()
		}
	}

	return m, cmd
}

func (m Model) View() string {
	if m.height == 0 {
		return "Loading..."
	}

	var s strings.Builder

	// Add padding to the entire app
	appStyle := lipgloss.NewStyle().
		PaddingTop(1).
		PaddingLeft(2)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("32")).
		MarginBottom(1)

	contentBuilder := strings.Builder{}
	
	if m.appMode == AppModeSettings {
		return m.renderSettings(appStyle, titleStyle)
	}
	
	if m.appMode == AppModeHelp {
		return m.renderHelp(appStyle, titleStyle)
	}
	
	contentBuilder.WriteString(titleStyle.Render("OCLI"))
	
	// Show breadcrumbs when zoomed
	if m.zoomedBullet != nil {
		breadcrumbStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Faint(true)
		
		var breadcrumbText strings.Builder
		for i, crumb := range m.breadcrumbs {
			if i > 0 {
				breadcrumbText.WriteString(" > ")
			}
			breadcrumbText.WriteString(crumb.Content)
		}
		if len(m.breadcrumbs) > 0 {
			breadcrumbText.WriteString(" > ")
		}
		breadcrumbText.WriteString(m.zoomedBullet.Content)
		
		contentBuilder.WriteString("\n")
		contentBuilder.WriteString(breadcrumbStyle.Render(breadcrumbText.String()))
		contentBuilder.WriteString("\n\n")
	} else {
		contentBuilder.WriteString("\n\n")
	}

	if m.editMode == EditModeNew {
		contentBuilder.WriteString("New bullet: " + m.textInput.View() + "\n\n")
	}

	// Define color styles
	colorStyles := map[BulletColor]lipgloss.Style{
		ColorDefault: lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		ColorBlue:    lipgloss.NewStyle().Foreground(lipgloss.Color("39")),
		ColorGreen:   lipgloss.NewStyle().Foreground(lipgloss.Color("42")),
		ColorYellow:  lipgloss.NewStyle().Foreground(lipgloss.Color("220")),
		ColorRed:     lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
	}

	completedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("232")).
		Faint(true)

	// Style for vertical hierarchy lines
	lineStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	// Calculate available space for content
	availableHeight := m.height - 6 // Title (2 lines) + breadcrumbs (2 lines) + help (2 lines)
	if m.editMode == EditModeNew {
		availableHeight -= 2 // New bullet input
	}

	// Calculate visible range
	startIndex := m.scrollOffset
	endIndex := m.scrollOffset + availableHeight
	if endIndex > len(m.allBullets) {
		endIndex = len(m.allBullets)
	}

	// Only render visible bullets
	for i := startIndex; i < endIndex; i++ {
		bullet := m.allBullets[i]
		var indent string
		var depth int
		
		// Calculate depth relative to zoom level
		if m.zoomedBullet != nil {
			depth = bullet.GetDepthFrom(m.zoomedBullet)
		} else {
			depth = bullet.GetDepth()
		}
		
		if m.settings.ShowHierarchyLines {
			// Build hierarchy lines
			var hierarchyLines strings.Builder
			
			// Add vertical lines for each level of indentation
			for level := 0; level < depth; level++ {
				if level == depth-1 {
					// Last level - use a branch character
					hierarchyLines.WriteString(lineStyle.Render("├── "))
				} else {
					// Not the last level - use a vertical line with spacing
					hierarchyLines.WriteString(lineStyle.Render("│   "))
				}
			}
			
			indent = hierarchyLines.String()
		} else {
			// Simple indentation without hierarchy lines
			indent = strings.Repeat("    ", depth)
		}

		prefix := ""

		// Handle caret for items with children
		if len(bullet.Children) > 0 {
			if bullet.Collapsed {
				prefix = "▶ "
			} else {
				prefix = "▼ "
			}
		}

		// Handle task checkbox or bullet
		if bullet.IsTask {
			if bullet.Completed {
				prefix += "☑ "
			} else {
				prefix += "☐ "
			}
		} else {
			// Only show bullet if there's no caret
			if len(bullet.Children) == 0 {
				prefix = "• "
			}
		}

		content := bullet.Content
		if bullet.IsEditing && m.editMode == EditModeEdit {
			content = m.textInput.View()
		}

		// Build the line with proper styling
		if i == m.selectedIndex {
			// For selected items, apply underline only to content, preserve original styling
			var baseStyle lipgloss.Style
			if bullet.IsTask && bullet.Completed {
				baseStyle = completedStyle
				// Also apply completed style to prefix for completed tasks
				styledPrefix := completedStyle.Render(prefix)
				styledContent := baseStyle.Copy().Underline(true).Render(content)
				line := fmt.Sprintf("%s%s%s", indent, styledPrefix, styledContent)
				contentBuilder.WriteString(line)
			} else {
				baseStyle = colorStyles[bullet.Color]
				// Apply underline to the content only
				styledContent := baseStyle.Copy().Underline(true).Render(content)
				line := fmt.Sprintf("%s%s%s", indent, prefix, styledContent)
				contentBuilder.WriteString(line)
			}
		} else if bullet.IsTask && bullet.Completed {
			// Apply completed style to both prefix and content
			styledPrefix := completedStyle.Render(prefix)
			styledContent := completedStyle.Render(content)
			line := fmt.Sprintf("%s%s%s", indent, styledPrefix, styledContent)
			contentBuilder.WriteString(line)
		} else {
			// Apply color based on bullet's color property only to content
			styledContent := colorStyles[bullet.Color].Render(content)
			line := fmt.Sprintf("%s%s%s", indent, prefix, styledContent)
			contentBuilder.WriteString(line)
		}
		contentBuilder.WriteString("\n")
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		MarginTop(2)

	help := "\n'h' for help • 's' for settings"
	
	// Add scroll indicators if there's more content
	if len(m.allBullets) > availableHeight {
		totalItems := len(m.allBullets)
		visibleStart := m.scrollOffset + 1
		visibleEnd := endIndex
		if visibleEnd > totalItems {
			visibleEnd = totalItems
		}
		
		scrollInfo := fmt.Sprintf(" • %d-%d of %d", visibleStart, visibleEnd, totalItems)
		help += scrollInfo
	}
	
	contentBuilder.WriteString(helpStyle.Render(help))

	// Apply padding to the entire content
	s.WriteString(appStyle.Render(contentBuilder.String()))

	return s.String()
}

func (m Model) renderSettings(appStyle, titleStyle lipgloss.Style) string {
	var contentBuilder strings.Builder
	
	contentBuilder.WriteString(titleStyle.Render("Settings"))
	contentBuilder.WriteString("\n\n")
	
	settingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	selectedSettingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Underline(true)
	
	settings := []struct {
		name   string
		value  bool
		toggle *bool
	}{
		{"Show hierarchy lines", m.settings.ShowHierarchyLines, &m.settings.ShowHierarchyLines},
	}
	
	for i, setting := range settings {
		status := "off"
		if setting.value {
			status = "on"
		}
		
		line := fmt.Sprintf("%s: %s", setting.name, status)
		
		if i == m.settingsIndex {
			contentBuilder.WriteString(selectedSettingStyle.Render(line))
		} else {
			contentBuilder.WriteString(settingStyle.Render(line))
		}
		contentBuilder.WriteString("\n")
	}
	
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		MarginTop(2)
	
	help := "\nKeys: ↑↓/jk:navigate • Enter/Space:toggle • s/esc/q:back"
	contentBuilder.WriteString(helpStyle.Render(help))
	
	return appStyle.Render(contentBuilder.String())
}

func (m Model) renderHelp(appStyle, titleStyle lipgloss.Style) string {
	var contentBuilder strings.Builder
	
	contentBuilder.WriteString(titleStyle.Render("Keyboard Shortcuts"))
	contentBuilder.WriteString("\n\n")
	
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)
	
	sections := []struct {
		title string
		items []string
	}{
		{
			"Navigation",
			[]string{
				"↑↓ or j/k    Navigate up/down",
				"←           Zoom out", 
				"→           Zoom in",
			},
		},
		{
			"Editing",
			[]string{
				"Enter       Create new bullet",
				"e           Edit selected bullet",
				"d           Delete selected bullet",
			},
		},
		{
			"Organization", 
			[]string{
				"Tab         Indent (move right)",
				"Shift+Tab   Outdent (move left)", 
				"Shift+↑↓    Move bullet up/down",
				"Space       Collapse/expand",
			},
		},
		{
			"Formatting",
			[]string{
				"c           Cycle bullet color",
				"t           Toggle task mode",
				"x           Mark task complete/incomplete",
			},
		},
		{
			"Other",
			[]string{
				"h           Show this help",
				"s           Open settings",
				"q           Quit application",
			},
		},
	}
	
	for _, section := range sections {
		contentBuilder.WriteString(sectionStyle.Render(section.title))
		contentBuilder.WriteString("\n")
		
		for _, item := range section.items {
			contentBuilder.WriteString("  ")
			contentBuilder.WriteString(helpStyle.Render(item))
			contentBuilder.WriteString("\n")
		}
		contentBuilder.WriteString("\n")
	}
	
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		MarginTop(1)
	
	footer := "Press 'h', 'esc', or 'q' to return"
	contentBuilder.WriteString(footerStyle.Render(footer))
	
	return appStyle.Render(contentBuilder.String())
}

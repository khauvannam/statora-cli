package switcher

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
)

// model is the Bubble Tea model for the SwitchPlan confirmation prompt.
type model struct {
	plan      *Plan
	confirmed bool
	aborted   bool
}

type confirmMsg bool

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y", "enter":
			m.confirmed = true
			return m, tea.Quit
		case "n", "N", "q", "ctrl+c", "esc":
			m.aborted = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.confirmed || m.aborted {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(color.CyanString("\n  Switch Plan\n"))
	sb.WriteString(color.CyanString("  ──────────\n"))

	for _, a := range m.plan.Actions {
		switch a.Kind {
		case "php":
			old := a.Current
			if old == "" {
				old = "(none)"
			}
			fmt.Fprintf(&sb, "  PHP       %s → %s\n",
	color.YellowString(old), color.GreenString(a.Name))
		case "composer":
			fmt.Fprintf(&sb, "  Composer  → %s\n", color.GreenString(a.Name))
		case "ext-enable":
			fmt.Fprintf(&sb, "  Enable    %s\n", color.GreenString(a.Name))
		case "ext-disable":
			fmt.Fprintf(&sb, "  Disable   %s\n", color.RedString(a.Name))
		}
	}

	sb.WriteString("\n  Apply? [y/N] ")
	return sb.String()
}

// PromptConfirm shows the plan and asks the user to confirm.
// Returns true if confirmed, false if aborted.
func PromptConfirm(plan *Plan) (bool, error) {
	if !plan.HasChanges() {
		fmt.Println("Nothing to switch — already up to date.")
		return false, nil
	}

	m := model{plan: plan}
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return false, err
	}

	final := result.(model)
	return final.confirmed, nil
}

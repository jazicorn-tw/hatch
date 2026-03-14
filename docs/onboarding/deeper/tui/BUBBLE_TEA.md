<!--
created_by:   jazicorn-tw
created_date: 2026-03-14
updated_by:   jazicorn-tw
updated_date: 2026-03-14
status:       active
tags:         [onboarding, tui, bubble-tea]
description:  "How Bubble Tea's Elm-style Model/Update/View pattern works and how hatch uses it for its terminal UI"
-->
# Bubble Tea — Hatch's Terminal UI Framework

Hatch's terminal interface (what juniors see when they SSH in and take quizzes) is built
with **Bubble Tea**, a Go TUI framework from [Charmbracelet](https://charm.sh). It
follows the **Elm architecture** — a pattern borrowed from functional programming that
makes complex UIs easy to reason about.

---

## The Elm architecture: Model / Update / View

Every Bubble Tea program is built around three things:

### Model — the state

```go
type Model struct {
    question string
    choices  []string
    cursor   int
    selected int
}
```

The model holds everything the UI needs to render. It's just a struct — no magic. When
you want to change what's on screen, you change the model.

### Update — handle events

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "up":
            if m.cursor > 0 {
                m.cursor--
            }
        case "down":
            if m.cursor < len(m.choices)-1 {
                m.cursor++
            }
        case "enter":
            m.selected = m.cursor
        case "q", "ctrl+c":
            return m, tea.Quit
        }
    }
    return m, nil
}
```

`Update` receives a message (a keypress, a network response, a timer tick — any event),
returns a new model and optionally a `Cmd`. It never mutates the model in place — it
returns a new value. This makes state changes traceable and testable.

### View — render to string

```go
func (m Model) View() string {
    s := m.question + "\n\n"
    for i, choice := range m.choices {
        cursor := " "
        if m.cursor == i {
            cursor = ">"
        }
        s += fmt.Sprintf("%s %s\n", cursor, choice)
    }
    return s
}
```

`View` converts the current model into a string. Bubble Tea handles writing that string
to the terminal. You never call `fmt.Print` directly — you just return the string you
want rendered.

---

## The event loop

Bubble Tea runs an internal loop:

```text
1. Render View(model) → terminal
2. Wait for event (keypress, Cmd result, etc.)
3. Call Update(model, event) → new model, optional Cmd
4. If Cmd is not nil, run it (possibly produces another event)
5. Go to 1
```

This loop runs until `Update` returns `tea.Quit`. For hatch, that happens when the quiz
session ends or the user presses `q`.

---

## Commands (`Cmd`)

A `Cmd` is a function that runs asynchronously and produces a message. This is how you
do I/O (API calls, timers, reads from stdin) without blocking the UI loop:

```go
// A Cmd — returns a Msg when done
func fetchNextQuestion() tea.Msg {
    // call the LLM, wait for response...
    return questionLoadedMsg{question: "..."}
}

// In Update, return the Cmd to start it
case loadingMsg:
    return m, fetchNextQuestion
```

When `fetchNextQuestion` returns, Bubble Tea delivers `questionLoadedMsg` to `Update`,
which updates the model to show the question.

---

## Charmbracelet ecosystem used in hatch

| Library                                                  | Purpose                                                               |
| -------------------------------------------------------- | --------------------------------------------------------------------- |
| [Bubble Tea](https://github.com/charmbracelet/bubbletea) | Core event loop and program runner                                    |
| [Bubbles](https://github.com/charmbracelet/bubbles)      | Pre-built components: text inputs, spinners, progress bars, viewports |
| [Lip Gloss](https://github.com/charmbracelet/lipgloss)   | Styling: colors, borders, padding, layout                             |
| [Glamour](https://github.com/charmbracelet/glamour)      | Markdown rendering in the terminal                                    |
| [Huh](https://github.com/charmbracelet/huh)              | Form components: dropdowns, confirms, text fields                     |
| [Wish](https://github.com/charmbracelet/wish)            | SSH server — delivers a Bubble Tea session over SSH                   |

---

## How Wish connects SSH to Bubble Tea

When a junior SSHes into the hatch server, **Wish** intercepts the connection and
creates a new Bubble Tea program instance for that session. Each SSH connection gets its
own independent `Model` and event loop running in a separate goroutine. Sessions don't
share state.

```text
SSH connection arrives
  └──► Wish middleware
         └──► new bubbletea.Program(initialModel)
                └──► runs Model/Update/View loop over that SSH session's terminal
```

The `agent.Runner` interface (`Run(ctx) error`) is what the server calls to hand off
control to the Bubble Tea program.

---

## What's not built yet

The TUI code (`internal/tui/`, `internal/server/`) is scaffolded for M3+. The current
M2 milestone focuses on the ingestion pipeline (`hatch ingest`), which is a regular CLI
command with a progress bar — no Bubble Tea involved there.

---

## Related

- [`CHARMBRACELET.md`](CHARMBRACELET.md) — the full Charmbracelet ecosystem, including Gum and Crush
- [`GOROUTINES.md`](../go/GOROUTINES.md) — how goroutines power concurrent SSH sessions
- [`INTERFACES.md`](../go/INTERFACES.md) — the `Runner` interface the server calls

## Resources

- [Bubble Tea GitHub](https://github.com/charmbracelet/bubbletea) — source, examples, tutorials
- [Charm.sh](https://charm.sh) — Charmbracelet's full library ecosystem
- [Elm architecture](https://guide.elm-lang.org/architecture/) — the original pattern Bubble Tea is based on
- [Wish: SSH apps with Bubble Tea](https://github.com/charmbracelet/wish) — how the SSH server works

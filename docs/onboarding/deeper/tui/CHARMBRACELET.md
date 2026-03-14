<!--
created_by:   jazicorn-tw
created_date: 2026-03-14
updated_by:   jazicorn-tw
updated_date: 2026-03-14
status:       active
tags:         [onboarding, tui, bubble-tea]
description:  "The full Charmbracelet ecosystem — libraries hatch will import and standalone tools worth knowing"
-->
# Charmbracelet — The Ecosystem

[Charmbracelet](https://charm.sh) is the company behind Bubble Tea and a family of
related libraries and tools for building beautiful terminal applications in Go. Hatch's
TUI (planned for M3) will be built entirely on this ecosystem.

The ecosystem splits into two categories: **Go libraries** you import in code, and
**standalone tools** you install and run from the shell.

---

## Go libraries

These are packages hatch will add to `go.mod` when the TUI milestone begins.

| Library                                                  | Role in hatch                                              |
| -------------------------------------------------------- | ---------------------------------------------------------- |
| [Bubble Tea](https://github.com/charmbracelet/bubbletea) | Core event loop — runs the Model/Update/View program       |
| [Bubbles](https://github.com/charmbracelet/bubbles)      | Pre-built components: text inputs, spinners, progress bars |
| [Lip Gloss](https://github.com/charmbracelet/lipgloss)   | Styling: colors, borders, padding, alignment               |
| [Glamour](https://github.com/charmbracelet/glamour)      | Renders Markdown to styled terminal output                 |
| [Huh](https://github.com/charmbracelet/huh)              | Form components: dropdowns, confirms, multi-select         |
| [Wish](https://github.com/charmbracelet/wish)            | SSH server — delivers a Bubble Tea session over SSH        |

None of these are in `go.mod` yet. The current M2 milestone covers the ingestion
pipeline only (`hatch ingest`), which uses a plain `schollz/progressbar` — no Bubble
Tea involved.

---

## Standalone tools

These are binaries you install separately. They are built *on top of* Charmbracelet
libraries but are not Go packages you import. Hatch does not depend on them directly.

### Gum

[Gum](https://github.com/charmbracelet/gum) makes shell scripts interactive. It wraps
Bubble Tea components behind a CLI so you can add prompts, spinners, and styled output
to bash/zsh scripts without writing any Go:

```bash
# Interactive prompt in a shell script
NAME=$(gum input --placeholder "Your name")
gum confirm "Deploy to production?" && deploy.sh
```

Useful for hatch contributors writing setup or release scripts that need user input.
Not used in the hatch binary itself.

### Crush

[Crush](https://github.com/charmbracelet/crush) is Charmbracelet's AI chat application
— a full-featured terminal chat client built with Bubble Tea. It is not a library and
has no relation to hatch's code. It demonstrates what a production-grade Bubble Tea
application looks like at scale, so it's worth browsing as a reference.

---

## How the libraries relate to each other

```text
Bubble Tea          ← program loop, events, Cmds
  ├── Bubbles       ← ready-made components (uses Bubble Tea internally)
  ├── Lip Gloss     ← styling primitives (used by Bubbles and your own views)
  ├── Glamour       ← Markdown renderer (uses Lip Gloss for styles)
  ├── Huh           ← form library (built on Bubble Tea + Lip Gloss)
  └── Wish          ← SSH layer (wraps a Bubble Tea program)
```

You can use Lip Gloss without Bubble Tea (for colorizing CLI output), and Glamour
without Bubble Tea (for rendering Markdown in any terminal program). But Bubbles, Huh,
and Wish are specifically designed to be composed with Bubble Tea programs.

---

## Related

- [`BUBBLE_TEA.md`](BUBBLE_TEA.md) — how the Model/Update/View pattern works in detail

## Resources

- [Charm.sh](https://charm.sh) — full library and tool catalog
- [Charmbracelet GitHub](https://github.com/charmbracelet) — all repositories
- [Gum](https://github.com/charmbracelet/gum) — shell script tool
- [Crush](https://github.com/charmbracelet/crush) — AI chat TUI (reference app)

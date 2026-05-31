package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/dariy/point-tui/internal/api"
)

// PreviewPane shows the ANSI photo + glamour-rendered post content.
type PreviewPane struct {
	vp         viewport.Model
	post       *api.Post
	ansis      map[string]string // url -> ansi string
	loadingImg bool
	width      int
	height     int
	tags       []api.TagRef
	tagIdx     int // index of the focused tag; -1 means none
}

func NewPreviewPane(width, height int) PreviewPane {
	vp := viewport.New(width, height)
	return PreviewPane{vp: vp, width: width, height: height, tagIdx: -1, ansis: make(map[string]string)}
}

// SetPost sets the displayed post and triggers a re-render.
func (pp *PreviewPane) SetPost(post *api.Post, ansis map[string]string) {
	pp.post = post
	pp.ansis = ansis
	if pp.ansis == nil {
		pp.ansis = make(map[string]string)
	}

	urls := api.MediaURLs(post)
	loadedCount := 0
	for _, u := range urls {
		if _, ok := pp.ansis[u]; ok {
			loadedCount++
		}
	}
	pp.loadingImg = post != nil && len(urls) > 0 && loadedCount < len(urls)

	pp.tagIdx = -1
	if post != nil {
		pp.tags = post.Tags
	} else {
		pp.tags = nil
	}
	pp.vp.GotoTop()
	pp.vp.SetContent(pp.renderContent())
}

// ANSI returns all current rendered ANSI image strings joined by newlines.
func (pp *PreviewPane) ANSI() string {
	var sb strings.Builder
	urls := api.MediaURLs(pp.post)
	for _, u := range urls {
		if ansi, ok := pp.ansis[u]; ok && ansi != "" {
			sb.WriteString(ansi)
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

// ANSIIdx returns the rendered ANSI string for the image at the given index.
func (pp *PreviewPane) ANSIIdx(idx int) string {
	urls := api.MediaURLs(pp.post)
	if idx < 0 || idx >= len(urls) {
		return ""
	}
	u := urls[idx]
	if ansi, ok := pp.ansis[u]; ok && ansi != "" {
		return ansi
	}
	return ""
}

// SetANSI updates the rendered image for the current post.
func (pp *PreviewPane) SetANSI(url, ansi string) {
	if pp.ansis == nil {
		pp.ansis = make(map[string]string)
	}
	pp.ansis[url] = ansi

	urls := api.MediaURLs(pp.post)
	loadedCount := 0
	for _, u := range urls {
		if _, ok := pp.ansis[u]; ok {
			loadedCount++
		}
	}
	pp.loadingImg = pp.post != nil && len(urls) > 0 && loadedCount < len(urls)

	if pp.post != nil {
		pp.vp.SetContent(pp.renderContent())
	}
}

// Resize updates viewport dimensions and re-renders.
func (pp *PreviewPane) Resize(width, height int) {
	pp.width = width
	pp.height = height
	pp.vp.Width = width
	pp.vp.Height = height
	pp.vp.SetContent(pp.renderContent())
}

func (pp *PreviewPane) renderContent() string {
	if pp.post == nil {
		return NormalItemStyle.Render("Select a post to preview")
	}

	var sb strings.Builder

	// 1. Tags
	var tagParts []string
	for i, t := range pp.post.Tags {
		label := "#" + t.Slug
		if i == pp.tagIdx {
			tagParts = append(tagParts, PreviewTagFocusedStyle.Render(label))
		} else {
			tagParts = append(tagParts, PreviewTagStyle.Render(label))
		}
	}
	if len(tagParts) > 0 {
		tagsRow := strings.Join(tagParts, " ")
		sb.WriteString(lipgloss.NewStyle().PaddingLeft(1).Render(tagsRow))
		sb.WriteString("\n\n")
	}

	// 2. Excerpt
	if pp.post.Excerpt != "" {
		wordWrap := pp.width - 2
		if wordWrap < 20 {
			wordWrap = 20
		}
		excerptSty := lipgloss.NewStyle().Italic(true).Foreground(colorStatus).Width(wordWrap).PaddingLeft(1)
		sb.WriteString(excerptSty.Render(pp.post.Excerpt))
		sb.WriteString("\n\n")
	}

	// 3. Image(s)
	urls := api.MediaURLs(pp.post)
	for _, u := range urls {
		if ansi, ok := pp.ansis[u]; ok && ansi != "" {
			sb.WriteString(ansi)
			sb.WriteByte('\n')
		}
	}
	if pp.loadingImg {
		sb.WriteString(NormalItemStyle.Render("Loading image(s)…"))
		sb.WriteByte('\n')
	}

	// 4. Content (Text)
	body := pp.post.Content
	// Avoid showing just the image URL if that's all there is
	if body != "" {
		isMediaOnly := true
		for _, u := range urls {
			if body != u {
				isMediaOnly = false
				break
			}
		}
		if isMediaOnly {
			body = ""
		}
	}

	if body != "" {
		wordWrap := pp.width - 2
		if wordWrap < 20 {
			wordWrap = 20
		}
		var rendered string
		r, err := glamour.NewTermRenderer(
			glamour.WithStandardStyle("dark"),
			glamour.WithWordWrap(wordWrap),
		)
		if err == nil {
			rendered, err = r.Render(body)
		}
		if err != nil {
			sb.WriteString(body)
		} else {
			sb.WriteString(lipgloss.NewStyle().PaddingLeft(1).Render(rendered))
		}
	}

	return sb.String()
}

func (pp *PreviewPane) Update(msg tea.Msg, focused bool) (PreviewPane, tea.Cmd) {
	if !focused {
		return *pp, nil
	}
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "t":
			// Advance to next tag (wraps around).
			if len(pp.tags) > 0 {
				pp.tagIdx = (pp.tagIdx + 1) % len(pp.tags)
				pp.vp.SetContent(pp.renderContent())
				return *pp, nil
			}
		case "T":
			// Go to previous tag (wraps around).
			if len(pp.tags) > 0 {
				if pp.tagIdx <= 0 {
					pp.tagIdx = len(pp.tags) - 1
				} else {
					pp.tagIdx--
				}
				pp.vp.SetContent(pp.renderContent())
				return *pp, nil
			}
		case "enter":
			if pp.tagIdx >= 0 && pp.tagIdx < len(pp.tags) {
				return *pp, selectPreviewTagCmd(pp.tags[pp.tagIdx])
			}
		}
	}
	var cmd tea.Cmd
	pp.vp, cmd = pp.vp.Update(msg)
	return *pp, cmd
}

func selectPreviewTagCmd(tag api.TagRef) tea.Cmd {
	return func() tea.Msg {
		return previewTagSelectedMsg{tag: tag}
	}
}

func (pp *PreviewPane) View() string {
	return lipgloss.NewStyle().Width(pp.width).Height(pp.height).Render(pp.vp.View())
}

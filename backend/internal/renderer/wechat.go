package renderer

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	gast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// ─── Style Mapping ──────────────────────────────────────────────

var baseFont = "-apple-system,BlinkMacSystemFont,'Helvetica Neue','PingFang SC','Hiragino Sans GB','Microsoft YaHei UI','Microsoft YaHei',Arial,sans-serif"
var monoFont = "'Geist Mono',ui-monospace,monospace"

// currentTheme holds the active theme name; defaults to "default".
var currentTheme string = "default"

// styles is the active style mapping, updated when theme changes.
var styles map[string]string

// themeStyles maps theme name → style overrides (only keys that differ from default).
var themeStyles map[string]map[string]string

// defaultStyles is the full default (green) theme style mapping.
var defaultStyles = map[string]string{
	// Headings: use <section> + <strong> instead of <h1>/<h2>/<h3> because WeChat overrides h-tag styles
	"h1":          "margin:0 0 16px;line-height:1.4;font-family:" + baseFont,
	"h1_strong":   "font-size:22px;font-weight:700;color:#111;letter-spacing:-0.01em;font-family:" + baseFont,
	"h2":          "margin:32px 0 14px;padding-left:12px;border-left:3px solid #07c160;line-height:1.45;font-family:" + baseFont,
	"h2_strong":   "font-size:18px;font-weight:600;color:#1a1a1a;font-family:" + baseFont,
	"h3":          "margin:24px 0 10px;font-family:" + baseFont,
	"h3_strong":   "font-size:16px;font-weight:600;color:#222;font-family:" + baseFont,

	"p":           "margin:14px 0;line-height:1.85;letter-spacing:0.01em;font-family:" + baseFont + ";color:#2c2c2c;font-size:16px",
	"blockquote_p": "margin:10px 0;line-height:1.85;font-family:" + baseFont + ";color:#555;font-size:15px",
	"blockquote":  "margin:18px 0;padding:14px 18px;background:#f8f9fa;border-left:3px solid #d1d5db;color:#555;font-size:15px;border-radius:0 6px 6px 0;line-height:1.85;font-family:" + baseFont,
	"code_inline": "background:#f3f4f6;padding:2px 6px;border-radius:4px;font-size:13px;font-family:" + monoFont + ";color:#dc2626",
	"pre":         "background:#1e1e2e;border-radius:8px;padding:16px;overflow-x:auto;margin:18px 0",
	"pre_code":    "background:none;color:#cdd6f4;padding:0;font-size:13px;font-family:" + monoFont + ";line-height:1.6",
	"img":         "max-width:100%;border-radius:8px;margin:14px 0;display:block",
	"a":           "color:#07c160;text-decoration:none;font-weight:500;font-family:" + baseFont,
	"strong":      "font-weight:700;font-family:" + baseFont,
	"em":          "font-style:italic;font-family:" + baseFont,
	"del":         "text-decoration:line-through;color:#999;font-family:" + baseFont,
	"ul":          "padding-left:22px;margin:14px 0;font-family:" + baseFont + ";color:#2c2c2c;font-size:16px;line-height:1.85",
	"ol":          "padding-left:22px;margin:14px 0;font-family:" + baseFont + ";color:#2c2c2c;font-size:16px;line-height:1.85",
	"li":          "margin:6px 0;line-height:1.85;font-family:" + baseFont + ";color:#2c2c2c;font-size:16px",
	"li_section":  "margin:6px 0;line-height:1.85;font-family:" + baseFont + ";color:#2c2c2c;font-size:16px;padding-left:8px",
	"bullet":      "color:#07c160;font-size:16px;margin-right:8px;font-family:" + baseFont,
	"ordered_num": "color:#07c160;font-size:16px;font-weight:600;margin-right:8px;font-family:" + baseFont,
	"hr":          "border:none;border-top:1px solid #eee;margin:28px 0",

	// Table: use <section> rows instead of <table>/<tr>/<td> because WeChat strips table styles
	"table_wrap":  "margin:18px 0;font-size:14px;font-family:" + baseFont,
	"table_row":   "display:flex;border-bottom:1px solid #e5e7eb;font-family:" + baseFont,
	"table_th_cell": "flex:1;padding:10px 14px;text-align:left;font-weight:600;color:#374151;background:#f9fafb;font-size:14px;font-family:" + baseFont,
	"table_td_cell": "flex:1;padding:10px 14px;text-align:left;color:#2c2c2c;font-size:14px;font-family:" + baseFont,
	"table_first_row": "display:flex;border:1px solid #e5e7eb;border-bottom:none;font-family:" + baseFont,
	"table_last_td": "flex:1;padding:10px 14px;text-align:left;border-right:1px solid #e5e7eb;color:#2c2c2c;font-size:14px;font-family:" + baseFont,
	"table_last_th": "flex:1;padding:10px 14px;text-align:left;border-right:1px solid #e5e7eb;font-weight:600;color:#374151;background:#f9fafb;font-size:14px;font-family:" + baseFont,

	"sup":         "font-size:12px;color:#07c160;vertical-align:super;line-height:0",
	"section_out": "font-size:16px;line-height:1.85;color:#2c2c2c;letter-spacing:0.01em;font-family:" + baseFont + ";word-break:break-word",
	"section_in":  "padding:0",
}

func init() {
	// Initialize styles with the default theme
	styles = copyStyles(defaultStyles)

	// Register theme overrides: each theme only stores keys that differ from defaultStyles
	themeStyles = map[string]map[string]string{
		"default": {}, // no overrides — identical to defaultStyles
		"blue": map[string]string{
			"h2":          "margin:32px 0 14px;padding-left:12px;border-left:3px solid #1e40af;line-height:1.45;font-family:" + baseFont,
			"h2_strong":   "font-size:18px;font-weight:600;color:#1e293b;font-family:" + baseFont,
			"a":           "color:#1e40af;text-decoration:none;font-weight:500;font-family:" + baseFont,
			"bullet":      "color:#1e40af;font-size:16px;margin-right:8px;font-family:" + baseFont,
			"ordered_num": "color:#1e40af;font-size:16px;font-weight:600;margin-right:8px;font-family:" + baseFont,
			"sup":         "font-size:12px;color:#1e40af;vertical-align:super;line-height:0",
			"code_inline": "background:#f3f4f6;padding:2px 6px;border-radius:4px;font-size:13px;font-family:" + monoFont + ";color:#1e40af",
			"blockquote":  "margin:18px 0;padding:14px 18px;background:#eff6ff;border-left:3px solid #93c5fd;color:#555;font-size:15px;border-radius:0 6px 6px 0;line-height:1.85;font-family:" + baseFont,
		},
		"gray": map[string]string{
			"h2":          "margin:32px 0 14px;padding-left:12px;border-left:3px solid #6b7280;line-height:1.45;font-family:" + baseFont,
			"h2_strong":   "font-size:18px;font-weight:600;color:#111827;font-family:" + baseFont,
			"a":           "color:#374151;text-decoration:none;font-weight:500;font-family:" + baseFont,
			"bullet":      "color:#6b7280;font-size:16px;margin-right:8px;font-family:" + baseFont,
			"ordered_num": "color:#6b7280;font-size:16px;font-weight:500;margin-right:8px;font-family:" + baseFont,
			"sup":         "font-size:12px;color:#6b7280;vertical-align:super;line-height:0",
			"code_inline": "background:#e5e7eb;padding:2px 6px;border-radius:4px;font-size:13px;font-family:" + monoFont + ";color:#6b7280",
			"blockquote":  "margin:18px 0;padding:14px 18px;background:#f8f9fa;border-left:3px solid #9ca3af;color:#555;font-size:15px;border-radius:0 6px 6px 0;line-height:1.85;font-family:" + baseFont,
		},
	}
}

// copyStyles returns a shallow copy of a style map.
func copyStyles(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// SetTheme switches the active theme. Invalid names fall back to "default".
func SetTheme(name string) {
	name = strings.ToLower(strings.TrimSpace(name))
	if _, ok := themeStyles[name]; !ok {
		name = "default"
	}
	currentTheme = name
	// Build styles: start from defaultStyles, then overlay theme overrides
	styles = copyStyles(defaultStyles)
	for k, v := range themeStyles[name] {
		styles[k] = v
	}
}

// AvailableThemes returns the list of registered theme names.
func AvailableThemes() []string {
	names := make([]string, 0, len(themeStyles))
	for name := range themeStyles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func styleAttr(key string) string {
	s, ok := styles[key]
	if !ok {
		return ""
	}
	return fmt.Sprintf(" style=\"%s\"", s)
}

// ─── Footnote Collector ────────────────────────────────────────

type footnoteCollector struct {
	links []string
}

func (fc *footnoteCollector) add(url string) int {
	fc.links = append(fc.links, url)
	return len(fc.links)
}

func (fc *footnoteCollector) renderFootnotes() string {
	if len(fc.links) == 0 {
		return ""
	}
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("<section%s>", styleAttr("p")))
	buf.WriteString("参考资料：</section>")
	for i, url := range fc.links {
		buf.WriteString(fmt.Sprintf("<section%s>[%d] %s</section>", styleAttr("p"), i+1, url))
	}
	return buf.String()
}

var currentCollector = &footnoteCollector{}

// ─── Table Cell Tracking ─────────────────────────────────────────

// We need to track cell positions within a row to apply border styling.
// Since goldmark walks the AST tree, we track row state globally.
var tableRowIsHeader bool
var tableCellIndex int
var tableCellCount int

// ─── Custom Node Renderers ──────────────────────────────────────

type wechatRenderer struct{}

func newWechatRenderer() *wechatRenderer {
	return &wechatRenderer{}
}

func (r *wechatRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindHeading, r.renderHeading)
	reg.Register(ast.KindParagraph, r.renderParagraph)
	reg.Register(ast.KindBlockquote, r.renderBlockquote)
	reg.Register(ast.KindFencedCodeBlock, r.renderCodeBlock)
	reg.Register(ast.KindCodeBlock, r.renderCodeBlock)
	reg.Register(ast.KindList, r.renderList)
	reg.Register(ast.KindListItem, r.renderListItem)
	reg.Register(ast.KindThematicBreak, r.renderThematicBreak)

	reg.Register(ast.KindText, r.renderText)
	reg.Register(ast.KindEmphasis, r.renderEmphasis)
	reg.Register(ast.KindCodeSpan, r.renderCodeSpan)
	reg.Register(ast.KindLink, r.renderLink)
	reg.Register(ast.KindImage, r.renderImage)
	reg.Register(ast.KindString, r.renderString)
	reg.Register(ast.KindAutoLink, r.renderAutoLink)

	reg.Register(gast.KindStrikethrough, r.renderStrikethrough)
	reg.Register(gast.KindTable, r.renderTable)
	reg.Register(gast.KindTableHeader, r.renderTableHeader)
	reg.Register(gast.KindTableRow, r.renderTableRow)
	reg.Register(gast.KindTableCell, r.renderTableCell)
	reg.Register(gast.KindTaskCheckBox, r.renderTaskCheckBox)
}

// ─── Block Node Renderers ────────────────────────────────────────

// Headings: <section> + <strong> instead of <h1>/<h2>/<h3>
// WeChat overrides styles on h-tags, so we use section+strong to preserve our styling.
func (r *wechatRenderer) renderHeading(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	level := n.Level
	if level > 3 {
		level = 3
	}
	tag := fmt.Sprintf("h%d", level)
	if entering {
		w.WriteString("<section")
		w.WriteString(styleAttr(tag))
		w.WriteString("><strong")
		w.WriteString(styleAttr(tag + "_strong"))
		w.WriteString(">")
	} else {
		w.WriteString("</strong></section>\n")
	}
	return ast.WalkContinue, nil
}

func (r *wechatRenderer) renderParagraph(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	styleKey := "p"
	if isInsideBlockquote(node) {
		styleKey = "blockquote_p"
	}
	if entering {
		w.WriteString("<section")
		w.WriteString(styleAttr(styleKey))
		w.WriteString(">")
	} else {
		w.WriteString("</section>\n")
	}
	return ast.WalkContinue, nil
}

func (r *wechatRenderer) renderBlockquote(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<section")
		w.WriteString(styleAttr("blockquote"))
		w.WriteString(">")
	} else {
		w.WriteString("</section>\n")
	}
	return ast.WalkContinue, nil
}

func (r *wechatRenderer) renderCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<section")
		w.WriteString(styleAttr("pre"))
		w.WriteString("><section")
		w.WriteString(styleAttr("pre_code"))
		w.WriteString(">")
		lines := node.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			w.Write(htmlEscape(line.Value(source)))
		}
		w.WriteString("</section></section>\n")
		return ast.WalkSkipChildren, nil
	}
	return ast.WalkContinue, nil
}

// Lists: use <section> + <span> with bullet/number prefix instead of <ul>/<ol>/<li>
// WeChat resets native list bullet styling, so we use explicit character prefixes.
var listIsOrdered bool
var listItemIndex int

func (r *wechatRenderer) renderList(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.List)
	listIsOrdered = n.IsOrdered()
	if entering {
		listItemIndex = 0
		w.WriteString("<section")
		w.WriteString(styleAttr("ul"))
		w.WriteString(">\n")
	} else {
		w.WriteString("</section>\n")
	}
	return ast.WalkContinue, nil
}

func (r *wechatRenderer) renderListItem(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		listItemIndex++
		prefix := "•" // bullet character •
		prefixStyle := "bullet"
		if listIsOrdered {
			prefix = fmt.Sprintf("%d.", listItemIndex)
			prefixStyle = "ordered_num"
		}
		w.WriteString("<section")
		w.WriteString(styleAttr("li_section"))
		w.WriteString("><span")
		w.WriteString(styleAttr(prefixStyle))
		w.WriteString(">")
		w.WriteString(prefix)
		w.WriteString("</span>")
	} else {
		w.WriteString("</section>\n")
	}
	return ast.WalkContinue, nil
}

func (r *wechatRenderer) renderThematicBreak(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<section")
		w.WriteString(styleAttr("hr"))
		w.WriteString("></section>\n")
	}
	return ast.WalkContinue, nil
}

// ─── Inline Node Renderers ──────────────────────────────────────

func (r *wechatRenderer) renderText(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Text)
	w.Write(htmlEscape(n.Segment.Value(source)))
	if n.SoftLineBreak() {
		w.WriteByte('\n')
	}
	if n.HardLineBreak() {
		w.WriteString("<br/>\n")
	}
	return ast.WalkContinue, nil
}

func (r *wechatRenderer) renderString(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	w.Write(htmlEscape(node.(*ast.String).Value))
	return ast.WalkContinue, nil
}

func (r *wechatRenderer) renderEmphasis(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Emphasis)
	if entering {
		if n.Level == 2 {
			w.WriteString("<strong")
			w.WriteString(styleAttr("strong"))
			w.WriteString("><em")
			w.WriteString(styleAttr("em"))
			w.WriteString(">")
		} else {
			tag := "em"
			if n.Level == 0 {
				tag = "strong"
			}
			w.WriteString("<")
			w.WriteString(tag)
			w.WriteString(styleAttr(tag))
			w.WriteString(">")
		}
	} else {
		if n.Level == 2 {
			w.WriteString("</em></strong>")
		} else {
			tag := "em"
			if n.Level == 0 {
				tag = "strong"
			}
			w.WriteString("</")
			w.WriteString(tag)
			w.WriteString(">")
		}
	}
	return ast.WalkContinue, nil
}

func (r *wechatRenderer) renderCodeSpan(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<code")
		w.WriteString(styleAttr("code_inline"))
		w.WriteString(">")
		for c := node.FirstChild(); c != nil; c = c.NextSibling() {
			w.Write(htmlEscape(c.(*ast.Text).Segment.Value(source)))
		}
		w.WriteString("</code>")
		return ast.WalkSkipChildren, nil
	}
	return ast.WalkContinue, nil
}

func (r *wechatRenderer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Link)
	url := string(n.Destination)
	if entering {
		idx := currentCollector.add(url)
		w.WriteString("<span")
		w.WriteString(styleAttr("a"))
		w.WriteString(">")
		linkFootnoteIdx = idx
	} else {
		w.WriteString("</span>")
		w.WriteString(fmt.Sprintf("<sup%s>[%d]</sup>", styleAttr("sup"), linkFootnoteIdx))
	}
	return ast.WalkContinue, nil
}

var linkFootnoteIdx int

func (r *wechatRenderer) renderAutoLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		n := node.(*ast.AutoLink)
		url := string(n.URL(source))
		idx := currentCollector.add(url)
		w.WriteString("<span")
		w.WriteString(styleAttr("a"))
		w.WriteString(">")
		w.Write(htmlEscape(n.Label(source)))
		w.WriteString("</span>")
		w.WriteString(fmt.Sprintf("<sup%s>[%d]</sup>", styleAttr("sup"), idx))
		return ast.WalkSkipChildren, nil
	}
	return ast.WalkContinue, nil
}

func (r *wechatRenderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		n := node.(*ast.Image)
		w.WriteString("<img")
		w.WriteString(styleAttr("img"))
		w.WriteString(fmt.Sprintf(" src=\"%s\"", htmlEscapeString(string(n.Destination))))
		if len(n.Title) > 0 {
			w.WriteString(fmt.Sprintf(" alt=\"%s\"", htmlEscapeString(string(n.Title))))
		}
		w.WriteString("/>")
		return ast.WalkSkipChildren, nil
	}
	return ast.WalkContinue, nil
}

func (r *wechatRenderer) renderStrikethrough(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<del")
		w.WriteString(styleAttr("del"))
		w.WriteString(">")
	} else {
		w.WriteString("</del>")
	}
	return ast.WalkContinue, nil
}

func (r *wechatRenderer) renderTaskCheckBox(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*gast.TaskCheckBox)
	if entering {
		checked := ""
		if n.IsChecked {
			checked = " checked"
		}
		w.WriteString(fmt.Sprintf("<input type=\"checkbox\"%s style=\"margin-right:4px;vertical-align:middle;\"", checked))
		w.WriteString(" disabled/>")
		return ast.WalkContinue, nil
	}
	return ast.WalkContinue, nil
}

func isInsideBlockquote(node ast.Node) bool {
	for p := node.Parent(); p != nil; p = p.Parent() {
		if p.Kind() == ast.KindBlockquote {
			return true
		}
	}
	return false
}

// ─── GFM Table Renderers (section-based) ──────────────────────────

// Tables: use <section> with display:flex instead of <table>/<tr>/<td>
// WeChat's WebView doesn't reliably render <table> with inline styles.
// Each row is a <section style="display:flex"> with cell <section style="flex:1"> children.

func (r *wechatRenderer) renderTable(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<section")
		w.WriteString(styleAttr("table_wrap"))
		w.WriteString(">\n")
	} else {
		w.WriteString("</section>\n")
	}
	return ast.WalkContinue, nil
}

func (r *wechatRenderer) renderTableHeader(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		tableRowIsHeader = true
		// Count cells in this header row
		tableCellCount = 0
		for c := node.FirstChild(); c != nil; c = c.NextSibling() {
			if c.Kind() == gast.KindTableCell {
				tableCellCount++
			}
		}
		tableCellIndex = 0
		w.WriteString("<section")
		w.WriteString(styleAttr("table_first_row"))
		w.WriteString(">\n")
	} else {
		tableRowIsHeader = false
		w.WriteString("</section>\n")
	}
	return ast.WalkContinue, nil
}

func (r *wechatRenderer) renderTableRow(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		tableCellCount = 0
		for c := node.FirstChild(); c != nil; c = c.NextSibling() {
			if c.Kind() == gast.KindTableCell {
				tableCellCount++
			}
		}
		tableCellIndex = 0
		w.WriteString("<section")
		w.WriteString(styleAttr("table_row"))
		w.WriteString(">\n")
	} else {
		w.WriteString("</section>\n")
	}
	return ast.WalkContinue, nil
}

func (r *wechatRenderer) renderTableCell(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	tableCellIndex++
	isLast := tableCellIndex == tableCellCount

	if entering {
		if tableRowIsHeader {
			if isLast {
				w.WriteString("<section")
				w.WriteString(styleAttr("table_last_th"))
				w.WriteString(">")
			} else {
				w.WriteString("<section")
				w.WriteString(styleAttr("table_th_cell"))
				w.WriteString(">")
			}
		} else {
			if isLast {
				w.WriteString("<section")
				w.WriteString(styleAttr("table_last_td"))
				w.WriteString(">")
			} else {
				w.WriteString("<section")
				w.WriteString(styleAttr("table_td_cell"))
				w.WriteString(">")
			}
		}
	} else {
		w.WriteString("</section>\n")
	}
	return ast.WalkContinue, nil
}

// ─── Utility ────────────────────────────────────────────────────

func htmlEscape(s []byte) []byte {
	var buf bytes.Buffer
	for _, b := range s {
		switch b {
		case '&':
			buf.WriteString("&amp;")
		case '<':
			buf.WriteString("&lt;")
		case '>':
			buf.WriteString("&gt;")
		case '"':
			buf.WriteString("&quot;")
		default:
			buf.WriteByte(b)
		}
	}
	return buf.Bytes()
}

func htmlEscapeString(s string) string {
	return string(htmlEscape([]byte(s)))
}

// ─── WeChat Inline-Style Extension ───────────────────────────────

type wechatExtender struct{}

func (e *wechatExtender) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.PrioritizedValue{Value: newWechatRenderer(), Priority: 1},
	))
}

// ─── Public API ─────────────────────────────────────────────────

// Render converts markdown to WeChat-compatible HTML with inline styles (default theme).
func Render(md string) string {
	html, _ := RenderWithFootnotes(md)
	return html
}

// RenderWithTheme converts markdown to WeChat-compatible HTML using the specified theme.
// Invalid or empty theme names fall back to "default".
func RenderWithTheme(md string, theme string) string {
	SetTheme(theme)
	html, _ := RenderWithFootnotes(md)
	return html
}

// RenderWithFootnotesAndTheme converts markdown and returns HTML plus collected footnote URLs,
// using the specified theme. Invalid or empty theme names fall back to "default".
func RenderWithFootnotesAndTheme(md string, theme string) (string, []string) {
	SetTheme(theme)
	return RenderWithFootnotes(md)
}

// RenderWithFootnotes converts markdown and returns HTML plus collected footnote URLs (default theme).
func RenderWithFootnotes(md string) (string, []string) {
	currentCollector = &footnoteCollector{}
	linkFootnoteIdx = 0
	tableRowIsHeader = false
	tableCellIndex = 0
	tableCellCount = 0

	gm := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.NewTable(),
			&wechatExtender{},
		),
	)

	source := []byte(md)
	var buf bytes.Buffer
	if err := gm.Convert(source, &buf); err != nil {
		return htmlEscapeString(md), nil
	}

	body := buf.String()
	footnotes := currentCollector.renderFootnotes()

	wrapped := fmt.Sprintf(
		"<section data-role=\"outer\" class=\"rich_media_content\"%s>\n"+
			"<section data-role=\"inner\"%s>\n"+
			"%s\n"+
			"%s\n"+
			"</section>\n"+
			"</section>",
		styleAttr("section_out"), styleAttr("section_in"), body, footnotes,
	)

	return wrapped, currentCollector.links
}
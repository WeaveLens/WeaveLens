package graph

import (
	"bytes"
	"fmt"
	"html"
	"strings"
)

const (
	svgNodeWidth  = 160
	svgNodeHeight = 60
	svgXSpacing   = 220
	svgYSpacing   = 120
	svgStartX     = 40
	svgStartY     = 40
	svgPadding    = 40
)

var svgCategoryColors = map[string]string{
	"compute":     "#dae8fc",
	"network":     "#d5e8d4",
	"database":     "#ffe6cc",
	"storage":     "#e1d5e7",
	"security":     "#f8cecc",
	"integration":  "#fff2cc",
	"other":        "#f5f5f5",
}

var svgCategoryStrokeColors = map[string]string{
	"compute":     "#6c8ebf",
	"network":     "#82b366",
	"database":     "#d79b00",
	"storage":     "#9673a6",
	"security":     "#b85450",
	"integration":  "#d6b656",
	"other":        "#666666",
}

func (g *ExportGraph) ToSVG() ([]byte, error) {
	nodePositions := make(map[string][2]int)
	rows := make(map[string]int)
	rowCounters := make(map[string]int)

	for _, node := range g.Nodes {
		category := strings.ToLower(node.Category)
		if _, exists := rows[category]; !exists {
			rows[category] = len(rows)
		}
		rowCounters[category]++
		nodePositions[node.ID] = [2]int{
			rows[category],
			rowCounters[category] - 1,
		}
	}

	maxX := 0
	maxY := 0
	for _, node := range g.Nodes {
		pos := nodePositions[node.ID]
		x := svgStartX + pos[1]*svgXSpacing + svgNodeWidth
		y := svgStartY + pos[0]*svgYSpacing + svgNodeHeight
		if x > maxX {
			maxX = x
		}
		if y > maxY {
			maxY = y
		}
	}

	width := maxX + svgPadding
	height := maxY + svgPadding

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, width, height, width, height))
	buf.WriteString("\n")

	buf.WriteString(`<defs>`)
	buf.WriteString("\n")
	buf.WriteString(`<marker id="arrowhead" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">`)
	buf.WriteString("\n")
	buf.WriteString(`<polygon points="0 0, 10 3.5, 0 7" fill="#999"/>`)
	buf.WriteString("\n")
	buf.WriteString(`</marker>`)
	buf.WriteString("\n")
	buf.WriteString(`</defs>`)
	buf.WriteString("\n")

	for _, edge := range g.Edges {
		srcPos, srcOk := nodePositions[edge.SourceID]
		tgtPos, tgtOk := nodePositions[edge.TargetID]
		if !srcOk || !tgtOk {
			continue
		}

		srcX := svgStartX + srcPos[1]*svgXSpacing + svgNodeWidth/2
		srcY := svgStartY + srcPos[0]*svgYSpacing + svgNodeHeight
		tgtX := svgStartX + tgtPos[1]*svgXSpacing + svgNodeWidth/2
		tgtY := svgStartY + tgtPos[0]*svgYSpacing

		buf.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#999" stroke-width="2" marker-end="url(#arrowhead)"/>`, srcX, srcY, tgtX, tgtY))
		buf.WriteString("\n")

		midX := (srcX + tgtX) / 2
		midY := (srcY + tgtY) / 2
		buf.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" font-size="9" fill="#666">%s</text>`, midX, midY-4, html.EscapeString(edge.Type)))
		buf.WriteString("\n")
	}

	for _, node := range g.Nodes {
		pos := nodePositions[node.ID]
		x := svgStartX + pos[1]*svgXSpacing
		y := svgStartY + pos[0]*svgYSpacing

		category := strings.ToLower(node.Category)
		fillColor := svgCategoryColors[category]
		if fillColor == "" {
			fillColor = svgCategoryColors["other"]
		}
		strokeColor := svgCategoryStrokeColors[category]
		if strokeColor == "" {
			strokeColor = svgCategoryStrokeColors["other"]
		}

		buf.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="8" ry="8" fill="%s" stroke="%s" stroke-width="2"/>`,
			x, y, svgNodeWidth, svgNodeHeight, fillColor, strokeColor))
		buf.WriteString("\n")

		displayName := node.Name
		if len(displayName) > 18 {
			displayName = displayName[:15] + "..."
		}
		buf.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" font-size="12" font-weight="bold" fill="#333">%s</text>`,
			x+svgNodeWidth/2, y+svgNodeHeight/2-6, html.EscapeString(displayName)))
		buf.WriteString("\n")

		buf.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" font-size="10" fill="#666">%s</text>`,
			x+svgNodeWidth/2, y+svgNodeHeight/2+12, html.EscapeString(node.Type)))
		buf.WriteString("\n")
	}

	buf.WriteString(`</svg>`)
	buf.WriteString("\n")

	return buf.Bytes(), nil
}

package graph

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
)

type drawioMXFile struct {
	XMLName xml.Name `xml:"mxfile"`
	Diagram drawioDiagram `xml:"diagram"`
}

type drawioDiagram struct {
	XMLName  xml.Name `xml:"diagram"`
	ID       string   `xml:"id,attr"`
	Name     string   `xml:"name,attr"`
	GraphModel drawioGraphModel `xml:"mxGraphModel"`
}

type drawioGraphModel struct {
	XMLName xml.Name `xml:"mxGraphModel"`
	Root    drawioRoot `xml:"root"`
}

type drawioRoot struct {
	XMLName xml.Name `xml:"root"`
	Cells   []drawioCell `xml:"mxCell"`
}

type drawioCell struct {
	XMLName  xml.Name `xml:"mxCell"`
	ID       string   `xml:"id,attr"`
	Value    string   `xml:"value,attr,omitempty"`
	Style    string   `xml:"style,attr,omitempty"`
	Vertex   string   `xml:"vertex,attr,omitempty"`
	Edge     string   `xml:"edge,attr,omitempty"`
	Source   string   `xml:"source,attr,omitempty"`
	Target   string   `xml:"target,attr,omitempty"`
	Geometry *drawioGeometry `xml:"mxGeometry,omitempty"`
}

type drawioGeometry struct {
	XMLName  xml.Name `xml:"mxGeometry"`
	X        string   `xml:"x,attr,omitempty"`
	Y        string   `xml:"y,attr,omitempty"`
	Width    string   `xml:"width,attr,omitempty"`
	Height   string   `xml:"height,attr,omitempty"`
	Relative string   `xml:"relative,attr,omitempty"`
	As       string   `xml:"as,attr,omitempty"`
}

const (
	nodeWidth  = 160
	nodeHeight = 60
	xSpacing   = 220
	ySpacing   = 120
	startX     = 40
	startY     = 40
)

var categoryColors = map[string]string{
	"compute":     "#dae8fc",
	"network":     "#d5e8d4",
	"database":     "#ffe6cc",
	"storage":     "#e1d5e7",
	"security":     "#f8cecc",
	"integration":  "#fff2cc",
	"other":        "#f5f5f5",
}

func (g *ExportGraph) ToDrawIO() ([]byte, error) {
	var cells []drawioCell

	cells = append(cells, drawioCell{
		ID: "0",
	})
	cells = append(cells, drawioCell{
		ID: "1",
		Vertex: "1",
	})

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

	for _, node := range g.Nodes {
		pos := nodePositions[node.ID]
		x := startX + pos[1]*xSpacing
		y := startY + pos[0]*ySpacing

		category := strings.ToLower(node.Category)
		color := categoryColors[category]
		if color == "" {
			color = categoryColors["other"]
		}

		style := fmt.Sprintf("rounded=1;whiteSpace=wrap;html=1;fillColor=%s;strokeColor=#666666;fontSize=11;fontStyle=1;", color)

		cells = append(cells, drawioCell{
			ID:     "node_" + node.ID,
			Value:  fmt.Sprintf("%s\n(%s)", node.Name, node.Type),
			Style:  style,
			Vertex: "1",
			Geometry: &drawioGeometry{
				X:      fmt.Sprintf("%d", x),
				Y:      fmt.Sprintf("%d", y),
				Width:  fmt.Sprintf("%d", nodeWidth),
				Height: fmt.Sprintf("%d", nodeHeight),
				As:     "geometry",
			},
		})
	}

	for _, edge := range g.Edges {
		style := "edgeStyle=orthogonalEdgeStyle;rounded=1;orthogonalLoop=1;jettySize=auto;targetPerimeterSpacing=0;strokeColor=#999999;fontSize=9;"

		cells = append(cells, drawioCell{
			ID:     "edge_" + edge.ID,
			Value:  edge.Type,
			Style:  style,
			Edge:   "1",
			Source: "node_" + edge.SourceID,
			Target: "node_" + edge.TargetID,
			Geometry: &drawioGeometry{
				Relative: "1",
				As:       "geometry",
			},
		})
	}

	mxfile := drawioMXFile{
		Diagram: drawioDiagram{
			ID:   "weavelens-export",
			Name: "Infrastructure Graph",
			GraphModel: drawioGraphModel{
				Root: drawioRoot{
					Cells: cells,
				},
			},
		},
	}

	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	buf.WriteString("\n")

	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")
	if err := encoder.Encode(mxfile); err != nil {
		return nil, fmt.Errorf("failed to encode draw.io XML: %w", err)
	}

	buf.WriteString("\n")
	return buf.Bytes(), nil
}

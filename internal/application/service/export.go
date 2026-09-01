package service

import (
	"context"
	"fmt"

	"github.com/elip/WeaveLens/internal/domain/graph"
	"github.com/elip/WeaveLens/internal/domain/relationship"
	"github.com/elip/WeaveLens/internal/domain/resource"
)

type ExportFormat string

const (
	ExportFormatJSON   ExportFormat = "json"
	ExportFormatDrawIO ExportFormat = "drawio"
	ExportFormatSVG    ExportFormat = "svg"
)

type ExportService interface {
	ExportGraph(ctx context.Context, scanID string, format ExportFormat) ([]byte, error)
}

type exportService struct {
	graphService GraphService
}

func NewExportService(gs GraphService) ExportService {
	return &exportService{
		graphService: gs,
	}
}

func (s *exportService) ExportGraph(ctx context.Context, scanID string, format ExportFormat) ([]byte, error) {
	resources, relationships, err := s.graphService.GetGraph(ctx, scanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	g := graph.NewGraph()
	for _, res := range resources {
		r, err := resource.NewResource(
			resource.ResourceID(res.ID),
			resource.ResourceType(res.Type),
			resource.ResourceCategory(res.Category),
			res.Name,
			resource.WithARN(res.ARN),
			resource.WithRegion(res.Region),
			resource.WithMetadata(res.Metadata),
		)
		if err != nil {
			continue
		}
		_ = g.AddNode(r)
	}

	for _, rel := range relationships {
		r, err := relationship.NewRelationship(
			relationship.RelationshipID(rel.ID),
			rel.SourceID,
			rel.TargetID,
			relationship.RelationshipType(rel.Type),
			relationship.WithRelationshipMetadata(rel.Metadata),
		)
		if err != nil {
			continue
		}
		_ = g.AddRelationship(r)
	}

	exportGraph := graph.NewExportGraph(g.Snapshot())

	switch format {
	case ExportFormatJSON:
		return exportGraph.ToJSON()
	case ExportFormatDrawIO:
		return exportGraph.ToDrawIO()
	case ExportFormatSVG:
		return exportGraph.ToSVG()
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

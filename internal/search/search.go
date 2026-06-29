package search

import (
	"fmt"

	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/vector"
)

func Search(p Params) (Results, error) {
	logger.Debug("search request",
		"categoryID", p.CategoryID,
		"userID", p.UserID,
		"q", p.Q,
		"expanded", p.Expanded,
		"limit", p.Limit,
		"offset", p.Offset,
		"hasGeo", p.HasGeo,
		"facetFilters", len(p.FacetFilters),
		"threshold", config.SearchThreshold,
	)

	embedding, err := vector.ResolveSearchEmbedding(p.UserID, p.CategoryID, p.Q)
	if err != nil {
		logger.Debug("search embedding error", "error", err)
		return Results{}, fmt.Errorf("search embedding: %w", err)
	}

	var pa pgArgs
	where := buildVectorMetadataWhere(p, &pa)
	orderPrefix := ""
	inAreaExpr := ""
	if p.HasGeo {
		inAreaExpr = geoInAreaExpr(p, &pa)
		orderPrefix = `CASE WHEN ` + inAreaExpr + ` THEN 0 ELSE 1 END, `
	}
	ids, inAreaCount, err := vector.QuerySimilarAdIDs(
		embedding,
		where,
		pa.args,
		orderPrefix,
		inAreaExpr,
		p.Limit,
		p.Offset,
		config.SearchThreshold,
	)
	if err != nil {
		logger.Debug("search vector query error", "error", err)
		return Results{}, err
	}

	logger.Debug("search complete",
		"categoryID", p.CategoryID,
		"q", p.Q,
		"resultCount", len(ids),
		"inAreaCount", inAreaCount,
	)
	return Results{IDs: ids, InAreaCount: inAreaCount}, nil
}

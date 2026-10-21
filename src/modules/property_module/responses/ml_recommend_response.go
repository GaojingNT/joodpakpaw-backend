package responses

type MLRecommendItem struct {
	PropertyID      int     `json:"property_id"`
	SimilarityScore float64 `json:"similarity_score"`
}

type MLPropertyRecommendItem struct {
	PropertyID      int      `json:"property_id"`
	SimilarityScore float64  `json:"similarity_score"`
	QualityScore    float64  `json:"quality_score"`
	HybridScore     float64  `json:"hybrid_score"`
	MatchedTags     []string `json:"matched_tags"`
}

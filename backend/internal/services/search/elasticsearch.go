package search

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/olivere/elastic/v7"
)

// ElasticsearchService handles all Elasticsearch operations
type ElasticsearchService struct {
	client *elastic.Client
	index  string
}

// NewElasticsearchService creates a new Elasticsearch service instance
func NewElasticsearchService(elasticURL, indexName string) (*ElasticsearchService, error) {
	// Create Elasticsearch client
	client, err := elastic.NewClient(
		elastic.SetURL(elasticURL),
		elastic.SetSniff(false),
		elastic.SetHealthcheckInterval(10*time.Second),
		elastic.SetErrorLog(nil),
		elastic.SetInfoLog(nil),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	// Check if Elasticsearch is reachable
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _, err = client.Ping(elasticURL).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch unreachable: %w", err)
	}

	service := &ElasticsearchService{
		client: client,
		index:  indexName,
	}

	// Create index with mapping if it doesn't exist
	if err := service.createIndex(); err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	return service, nil
}

// createIndex creates the equipment index with proper mappings
func (s *ElasticsearchService) createIndex() error {
	ctx := context.Background()

	// Check if index exists
	exists, err := s.client.IndexExists(s.index).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}

	if exists {
		return nil
	}

	// Define index mapping
	mapping := map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 1,
			"analysis": map[string]interface{}{
				"analyzer": map[string]interface{}{
					"autocomplete_analyzer": map[string]interface{}{
						"type":      "custom",
						"tokenizer": "standard",
						"filter":    []string{"lowercase", "autocomplete_filter"},
					},
				},
				"filter": map[string]interface{}{
					"autocomplete_filter": map[string]interface{}{
						"type":     "edge_ngram",
						"min_gram": 2,
						"max_gram": 20,
					},
				},
			},
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "long",
				},
				"name": map[string]interface{}{
					"type":     "text",
					"analyzer": "autocomplete_analyzer",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type":         "keyword",
							"ignore_above": 256,
						},
						"suggest": map[string]interface{}{
							"type":     "text",
							"analyzer": "autocomplete_analyzer",
						},
					},
				},
				"description": map[string]interface{}{
					"type":     "text",
					"analyzer": "standard",
				},
				"brand": map[string]interface{}{
					"type": "keyword",
				},
				"model": map[string]interface{}{
					"type": "keyword",
				},
				"category_name": map[string]interface{}{
					"type":     "text",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type":         "keyword",
							"ignore_above": 256,
						},
					},
				},
				"category_id": map[string]interface{}{
					"type": "long",
				},
				"daily_price": map[string]interface{}{
					"type": "double",
				},
				"weekly_price": map[string]interface{}{
					"type": "double",
				},
				"monthly_price": map[string]interface{}{
					"type": "double",
				},
				"location": map[string]interface{}{
					"type":     "text",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type":         "keyword",
							"ignore_above": 256,
						},
					},
				},
				"city": map[string]interface{}{
					"type":     "text",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type":         "keyword",
							"ignore_above": 256,
						},
					},
				},
				"latitude": map[string]interface{}{
					"type": "double",
				},
				"longitude": map[string]interface{}{
					"type": "double",
				},
				"vendor_id": map[string]interface{}{
					"type": "long",
				},
				"vendor_name": map[string]interface{}{
					"type":     "text",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type":         "keyword",
							"ignore_above": 256,
						},
					},
				},
				"vendor_rating": map[string]interface{}{
					"type": "float",
				},
				"total_quantity": map[string]interface{}{
					"type": "integer",
				},
				"available_quantity": map[string]interface{}{
					"type": "integer",
				},
				"availability_status": map[string]interface{}{
					"type": "keyword",
				},
				"image_url": map[string]interface{}{
					"type": "keyword",
				},
				"specs": map[string]interface{}{
					"type": "object",
				},
				"tags": map[string]interface{}{
					"type": "keyword",
				},
				"popularity_score": map[string]interface{}{
					"type": "float",
				},
				"created_at": map[string]interface{}{
					"type": "date",
				},
				"updated_at": map[string]interface{}{
					"type": "date",
				},
			},
		},
	}

	// Create index
	_, err = s.client.CreateIndex(s.index).BodyJson(mapping).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

// EquipmentDocument represents an equipment item in Elasticsearch
type EquipmentDocument struct {
	ID                 uint64                  `json:"id"`
	Name               string                  `json:"name"`
	Description        string                  `json:"description"`
	Brand              string                  `json:"brand"`
	Model              string                  `json:"model"`
	CategoryName       string                  `json:"category_name"`
	CategoryID         uint64                  `json:"category_id"`
	DailyPrice         float64                 `json:"daily_price"`
	WeeklyPrice        float64                 `json:"weekly_price"`
	MonthlyPrice       float64                 `json:"monthly_price"`
	Location           string                  `json:"location"`
	City               string                  `json:"city"`
	Latitude           float64                 `json:"latitude"`
	Longitude          float64                 `json:"longitude"`
	VendorID           uint64                  `json:"vendor_id"`
	VendorName         string                  `json:"vendor_name"`
	VendorRating       float32                 `json:"vendor_rating"`
	TotalQuantity      int                     `json:"total_quantity"`
	AvailableQuantity  int                     `json:"available_quantity"`
	AvailabilityStatus string                  `json:"availability_status"`
	ImageURL           string                  `json:"image_url"`
	Specs              map[string]interface{}  `json:"specs"`
	Tags               []string                `json:"tags"`
	PopularityScore    float32                 `json:"popularity_score"`
	CreatedAt          time.Time               `json:"created_at"`
	UpdatedAt          time.Time               `json:"updated_at"`
}

// IndexEquipment indexes a single equipment document
func (s *ElasticsearchService) IndexEquipment(equipment *EquipmentDocument) error {
	ctx := context.Background()

	_, err := s.client.Index().
		Index(s.index).
		Id(fmt.Sprintf("%d", equipment.ID)).
		BodyJson(equipment).
		Refresh("true").
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to index equipment: %w", err)
	}

	return nil
}

// BulkIndexEquipment indexes multiple equipment documents
func (s *ElasticsearchService) BulkIndexEquipment(equipment []*EquipmentDocument) error {
	if len(equipment) == 0 {
		return nil
	}

	ctx := context.Background()
	bulkRequest := s.client.Bulk()

	for _, item := range equipment {
		req := elastic.NewBulkIndexRequest().
			Index(s.index).
			Id(fmt.Sprintf("%d", item.ID)).
			Doc(item)
		bulkRequest = bulkRequest.Add(req)
	}

	_, err := bulkRequest.Refresh("true").Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to bulk index equipment: %w", err)
	}

	return nil
}

// SearchQuery represents a search query with all available filters
type SearchQuery struct {
	Query              string   `json:"query"`
	CategoryIDs        []uint64 `json:"category_ids"`
	Cities             []string `json:"cities"`
	MinDailyPrice      float64  `json:"min_daily_price"`
	MaxDailyPrice      float64  `json:"max_daily_price"`
	MinVendorRating    float32  `json:"min_vendor_rating"`
	AvailabilityStatus string   `json:"availability_status"`
	Latitude           float64  `json:"latitude"`
	Longitude          float64  `json:"longitude"`
	Radius             float64  `json:"radius"` // in kilometers
	Brand              string   `json:"brand"`
	Tags               []string `json:"tags"`
	SortBy             string   `json:"sort_by"` // "relevance", "price_asc", "price_desc", "rating", "popularity"
	Page               int      `json:"page"`
	PerPage            int      `json:"per_page"`
}

// SearchResult represents search results with metadata
type SearchResult struct {
	Hits      []*EquipmentDocument `json:"hits"`
	Total     int64                `json:"total"`
	Page      int                  `json:"page"`
	PerPage   int                  `json:"per_page"`
	MaxScore  float64              `json:"max_score"`
	Suggested []*EquipmentDocument `json:"suggested,omitempty"`
}

// Search performs a full-text search with filters
func (s *ElasticsearchService) Search(query *SearchQuery) (*SearchResult, error) {
	ctx := context.Background()

	// Build the search query
	boolQuery := elastic.NewBoolQuery()

	// Main text search (search in name, description, brand, city)
	if query.Query != "" {
		multiMatchQuery := elastic.NewMultiMatchQuery(query.Query, "name^3", "description^2", "brand^2", "city", "category_name", "tags").
			Type("best_fields").
			Fuzziness("AUTO").
			Operator("or")
		boolQuery.Must(multiMatchQuery)
	} else {
		// If no query, match all
		boolQuery.Must(elastic.NewMatchAllQuery())
	}

	// Apply filters
	if len(query.CategoryIDs) > 0 {
		categoryValues := make([]interface{}, len(query.CategoryIDs))
		for i, catID := range query.CategoryIDs {
			categoryValues[i] = catID
		}
		categoryFilter := elastic.NewTermsQuery("category_id", categoryValues...)
		boolQuery.Filter(categoryFilter)
	}

	if len(query.Cities) > 0 {
		cityValues := make([]interface{}, len(query.Cities))
		for i, city := range query.Cities {
			cityValues[i] = city
		}
		cityFilter := elastic.NewTermsQuery("city.keyword", cityValues...)
		boolQuery.Filter(cityFilter)
	}

	if query.MinDailyPrice > 0 {
		boolQuery.Filter(elastic.NewRangeQuery("daily_price").Gte(query.MinDailyPrice))
	}

	if query.MaxDailyPrice > 0 {
		boolQuery.Filter(elastic.NewRangeQuery("daily_price").Lte(query.MaxDailyPrice))
	}

	if query.MinVendorRating > 0 {
		boolQuery.Filter(elastic.NewRangeQuery("vendor_rating").Gte(query.MinVendorRating))
	}

	if query.AvailabilityStatus != "" {
		boolQuery.Filter(elastic.NewTermQuery("availability_status", query.AvailabilityStatus))
	}

	if query.Brand != "" {
		boolQuery.Filter(elastic.NewTermQuery("brand", query.Brand))
	}

	if len(query.Tags) > 0 {
		tagsValues := make([]interface{}, len(query.Tags))
		for i, tag := range query.Tags {
			tagsValues[i] = tag
		}
		tagsFilter := elastic.NewTermsQuery("tags", tagsValues...)
		boolQuery.Filter(tagsFilter)
	}

	// Geo-distance filter
	if query.Latitude != 0 && query.Longitude != 0 && query.Radius > 0 {
		geoDistanceQuery := elastic.NewGeoDistanceQuery("location")
		geoDistanceQuery = geoDistanceQuery.
			Distance(fmt.Sprintf("%dkm", int(query.Radius))).
			Point(query.Latitude, query.Longitude)
		boolQuery.Filter(geoDistanceQuery)
	}

	// Build the search request
	searchRequest := s.client.Search().
		Index(s.index).
		Query(boolQuery).
		From((query.Page - 1) * query.PerPage).
		Size(query.PerPage)

	// Add sorting
	switch query.SortBy {
	case "price_asc":
		searchRequest = searchRequest.Sort("daily_price", true)
	case "price_desc":
		searchRequest = searchRequest.Sort("daily_price", false)
	case "rating":
		searchRequest = searchRequest.Sort("vendor_rating", false)
	case "popularity":
		searchRequest = searchRequest.Sort("popularity_score", false)
	default:
		// Sort by relevance (score)
		searchRequest = searchRequest.Sort("_score", false)
	}

	// Execute search
	searchResult, err := searchRequest.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Parse results
	result := &SearchResult{
		Total:    searchResult.TotalHits(),
		Page:     query.Page,
		PerPage:  query.PerPage,
		MaxScore: 0.0,
		Hits:     make([]*EquipmentDocument, 0),
	}

	for _, hit := range searchResult.Hits.Hits {
		var doc EquipmentDocument
		if err := json.Unmarshal(hit.Source, &doc); err != nil {
			continue
		}
		result.Hits = append(result.Hits, &doc)
	}

	return result, nil
}

// AutocompleteQuery represents an autocomplete request
type AutocompleteQuery struct {
	Query  string `json:"query"`
	Field  string `json:"field"` // "name", "category", "city", "brand"
	Size   int    `json:"size"`
}

// Autocomplete returns search suggestions
func (s *ElasticsearchService) Autocomplete(query *AutocompleteQuery) ([]string, error) {
	ctx := context.Background()

	if query.Size == 0 {
		query.Size = 10
	}

	// Build field list for suggestions
	var fields []string
	switch query.Field {
	case "name":
		fields = []string{"name.suggest^3", "name^2"}
	case "category":
		fields = []string{"category_name"}
	case "city":
		fields = []string{"city.suggest^2", "city"}
	case "brand":
		fields = []string{"brand"}
	default:
		fields = []string{"name.suggest^3", "name^2", "brand^2", "category_name", "city"}
	}

	// Build search query
	boolQuery := elastic.NewBoolQuery()
	if query.Query != "" {
		prefixQuery := elastic.NewPrefixQuery("name.suggest", query.Query)
		boolQuery.Should(prefixQuery)
		boolQuery.Should(elastic.NewMultiMatchQuery(query.Query, fields...).Type("phrase_prefix"))
	} else {
		boolQuery.Must(elastic.NewMatchAllQuery())
	}

	// Execute search
	searchResult, err := s.client.Search().
		Index(s.index).
		Query(boolQuery).
		Size(query.Size).
		Aggregation("suggestions", elastic.NewTermsAggregation().Field("name.keyword").Size(query.Size)).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("autocomplete failed: %w", err)
	}

	// Extract suggestions
	suggestions := make([]string, 0)
	for _, hit := range searchResult.Hits.Hits {
		var doc EquipmentDocument
		if err := json.Unmarshal(hit.Source, &doc); err != nil {
			continue
		}

		// Add relevant field based on query type
		switch query.Field {
		case "name":
			suggestions = append(suggestions, doc.Name)
		case "category":
			suggestions = append(suggestions, doc.CategoryName)
		case "city":
			suggestions = append(suggestions, doc.City)
		case "brand":
			suggestions = append(suggestions, doc.Brand)
		default:
			if doc.Name != "" {
				suggestions = append(suggestions, doc.Name)
			}
		}
	}

	return suggestions, nil
}

// DeleteEquipment removes an equipment document from the index
func (s *ElasticsearchService) DeleteEquipment(equipmentID uint64) error {
	ctx := context.Background()

	_, err := s.client.Delete().
		Index(s.index).
		Id(fmt.Sprintf("%d", equipmentID)).
		Refresh("true").
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete equipment: %w", err)
	}

	return nil
}

// UpdateEquipment updates an existing equipment document
func (s *ElasticsearchService) UpdateEquipment(equipmentID uint64, updates map[string]interface{}) error {
	ctx := context.Background()

	_, err := s.client.Update().
		Index(s.index).
		Id(fmt.Sprintf("%d", equipmentID)).
		Doc(updates).
		Refresh("true").
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to update equipment: %w", err)
	}

	return nil
}

// GetAggregations returns aggregation data for filters
func (s *ElasticsearchService) GetAggregations(query *SearchQuery) (map[string]interface{}, error) {
	ctx := context.Background()

	// Build base query
	boolQuery := elastic.NewBoolQuery()
	if query.Query != "" {
		multiMatchQuery := elastic.NewMultiMatchQuery(query.Query, "name^3", "description^2", "brand^2").
			Type("best_fields").
			Fuzziness("AUTO")
		boolQuery.Must(multiMatchQuery)
	}

	// Build search request with aggregations
	searchRequest := s.client.Search().
		Index(s.index).
		Query(boolQuery).
		Size(0). // Don't return documents, just aggregations
		Aggregation("categories", elastic.NewTermsAggregation().Field("category_id").Size(50)).
		Aggregation("cities", elastic.NewTermsAggregation().Field("city.keyword").Size(50)).
		Aggregation("brands", elastic.NewTermsAggregation().Field("brand").Size(50)).
		Aggregation("price_ranges", elastic.NewRangeAggregation().Field("daily_price").
			AddRange(0, 500).
			AddRange(500, 1000).
			AddRange(1000, 2000).
			AddRange(2000, 5000).
			AddRange(5000, 10000).
			AddRange(10000, nil))

	// Execute search
	searchResult, err := searchRequest.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("aggregations failed: %w", err)
	}

	// Parse aggregations
	result := make(map[string]interface{})

	// Categories aggregation
	if categories, found := searchResult.Aggregations.Terms("categories"); found {
		var cats []map[string]interface{}
		for _, bucket := range categories.Buckets {
			cats = append(cats, map[string]interface{}{
				"category_id": bucket.Key.(float64),
				"count":       bucket.DocCount,
			})
		}
		result["categories"] = cats
	}

	// Cities aggregation
	if cities, found := searchResult.Aggregations.Terms("cities"); found {
		var cityList []map[string]interface{}
		for _, bucket := range cities.Buckets {
			cityList = append(cityList, map[string]interface{}{
				"city":  bucket.Key.(string),
				"count": bucket.DocCount,
			})
		}
		result["cities"] = cityList
	}

	// Brands aggregation
	if brands, found := searchResult.Aggregations.Terms("brands"); found {
		var brandList []map[string]interface{}
		for _, bucket := range brands.Buckets {
			brandList = append(brandList, map[string]interface{}{
				"brand": bucket.Key.(string),
				"count":  bucket.DocCount,
			})
		}
		result["brands"] = brandList
	}

	// Price ranges aggregation
	if priceRanges, found := searchResult.Aggregations.Range("price_ranges"); found {
		var ranges []map[string]interface{}
		for _, bucket := range priceRanges.Buckets {
			ranges = append(ranges, map[string]interface{}{
				"from":  bucket.From,
				"to":    bucket.To,
				"count": bucket.DocCount,
			})
		}
		result["price_ranges"] = ranges
	}

	return result, nil
}

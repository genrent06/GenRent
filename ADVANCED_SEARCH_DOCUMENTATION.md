# Advanced Search & Filtering System Documentation

## Overview
Complete implementation of Elasticsearch-powered full-text search with advanced filtering, saved searches, autocomplete, and date-based availability filtering.

## Architecture

### Components
1. **Elasticsearch Integration** - Full-text search engine
2. **Search Service** - Business logic for search operations
3. **Saved Search Service** - User search preferences and alerts
4. **Search Handlers** - REST API endpoints
5. **Database Layer** - Search analytics and history

## Features Implemented

### ✅ Full-Text Search with Elasticsearch
- Multi-field search (name, description, brand, category, city)
- Fuzzy matching and autocomplete
- Relevance scoring with field weighting
- Real-time search results

### ✅ Advanced Filters
- **Price Range**: Min/max daily price filtering
- **Category Filter**: Multiple category selection
- **Location Filter**: City-based or radius-based search
- **Vendor Rating**: Minimum rating threshold
- **Brand Filter**: Brand-specific search
- **Tags Filter**: Tag-based filtering
- **Availability Status**: Filter by equipment status

### ✅ Saved Search Queries
- Save complex search queries
- Named search profiles
- Enable/disable alerts for new results
- Alert frequency (instant, daily, weekly)
- Search result count tracking

### ✅ Search Suggestions & Autocomplete
- Real-time search suggestions
- Field-specific autocomplete (name, category, city, brand)
- Recent search history
- Popular searches tracking

### ✅ Date-Based Availability Filtering
- Filter equipment by rental dates
- Real-time availability checking
- Booking conflict detection
- Available quantity validation

## API Endpoints

### Search Operations

#### POST /api/search/advanced
**Description**: Perform advanced full-text search with filters

**Request Body**:
```json
{
  "query": "generator 5kva",
  "category_ids": [1, 2, 3],
  "cities": ["Mumbai", "Delhi"],
  "min_daily_price": 500,
  "max_daily_price": 2000,
  "min_vendor_rating": 4.0,
  "availability_status": "available",
  "latitude": 19.0760,
  "longitude": 72.8777,
  "radius": 50,
  "brand": "Honda",
  "tags": ["fuel_efficient", "portable"],
  "start_date": "2024-08-01",
  "end_date": "2024-08-05",
  "sort_by": "price_asc",
  "page": 1,
  "per_page": 20
}
```

**Response**:
```json
{
  "results": [
    {
      "id": 123,
      "name": "Honda Generator 5KVA",
      "description": "Portable generator for events",
      "daily_price": 1500.00,
      "vendor_name": "Power Rentals",
      "vendor_rating": 4.5,
      "available_quantity": 3,
      "city": "Mumbai"
    }
  ],
  "total": 45,
  "page": 1,
  "per_page": 20,
  "max_score": 8.5
}
```

#### GET /api/search/autocomplete?q=gen&field=name&size=10
**Description**: Get search suggestions

**Response**:
```json
{
  "suggestions": [
    "Generator 5KVA Portable",
    "Generator 10KVA Industrial",
    "Generator 3KVA Domestic"
  ],
  "recent_searches": [
    {
      "type": "name",
      "text": "Generator",
      "count": 15,
      "popularity": 0.15
    }
  ]
}
```

#### POST /api/search/aggregations
**Description**: Get filter aggregations for search UI

**Request Body**:
```json
{
  "query": "generator",
  "category_ids": [1],
  "cities": ["Mumbai"]
}
```

**Response**:
```json
{
  "aggregations": {
    "categories": [
      {"category_id": 1, "count": 25},
      {"category_id": 2, "count": 15}
    ],
    "cities": [
      {"city": "Mumbai", "count": 20},
      {"city": "Delhi", "count": 10}
    ],
    "brands": [
      {"brand": "Honda", "count": 12},
      {"brand": " Yamaha", "count": 8}
    ],
    "price_ranges": [
      {"from": 0, "to": 500, "count": 5},
      {"from": 500, "to": 1000, "count": 15}
    ]
  }
}
```

### Saved Searches

#### POST /api/search/saved
**Description**: Create a saved search

**Request Body**:
```json
{
  "name": "Weekend Generator Search",
  "query": "generator 5kva",
  "category_ids": [1],
  "cities": ["Mumbai"],
  "min_price": 1000,
  "max_price": 3000,
  "min_rating": 4.0,
  "brand": "Honda",
  "tags": ["portable"],
  "sort_by": "rating",
  "is_alert": true,
  "alert_frequency": "daily"
}
```

#### GET /api/search/saved
**Description**: Get user's saved searches

**Response**:
```json
{
  "saved_searches": [
    {
      "id": 1,
      "name": "Weekend Generator Search",
      "query": "generator 5kva",
      "result_count": 23,
      "is_alert": true,
      "last_alert_at": "2024-07-14T10:30:00Z",
      "created_at": "2024-07-10T15:20:00Z"
    }
  ]
}
```

#### GET /api/search/saved/:id?page=1&per_page=20
**Description**: Execute a saved search

#### PUT /api/search/saved/:id
**Description**: Update saved search settings

#### DELETE /api/search/saved/:id
**Description**: Delete a saved search

#### GET /api/search/history?limit=20
**Description**: Get user's search history

## Database Schema

### Tables

#### saved_searches
```sql
- id: BIGSERIAL PRIMARY KEY
- user_id: BIGINT (FK to users)
- name: VARCHAR(100) - Search name
- query: TEXT - Search query text
- category_ids: TEXT (JSON array)
- cities: TEXT (JSON array)
- min_price: FLOAT
- max_price: FLOAT
- min_rating: FLOAT
- brand: VARCHAR(50)
- tags: TEXT (JSON array)
- sort_by: VARCHAR(20)
- is_alert: BOOLEAN - Enable/disable alerts
- alert_frequency: VARCHAR(20) - 'instant', 'daily', 'weekly'
- last_alert_at: TIMESTAMP
- is_active: BOOLEAN
- result_count: INT
- created_at: TIMESTAMP
- updated_at: TIMESTAMP
```

#### search_history
```sql
- id: BIGSERIAL PRIMARY KEY
- user_id: BIGINT (FK to users)
- query: TEXT - Search query with filters
- filter_count: INT - Number of filters applied
- result_count: INT - Number of results
- clicked_id: BIGINT (FK to equipment)
- click_position: INT - Position in results
- created_at: TIMESTAMP
```

#### search_suggestions
```sql
- id: BIGSERIAL PRIMARY KEY
- type: VARCHAR(20) - 'name', 'category', 'city', 'brand'
- text: VARCHAR(200) - Suggestion text
- count: INT - Usage count
- popularity: FLOAT - Normalized popularity score
- created_at: TIMESTAMP
- updated_at: TIMESTAMP
```

#### search_analytics
```sql
- id: BIGSERIAL PRIMARY KEY
- date: DATE UNIQUE
- total_searches: INT
- unique_users: INT
- avg_result_count: FLOAT
- zero_result_searches: INT
- click_through_rate: FLOAT
- created_at: TIMESTAMP
- updated_at: TIMESTAMP
```

## Elasticsearch Configuration

### Index Mapping

The equipment index uses the following mapping:

```json
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 1,
    "analysis": {
      "analyzer": {
        "autocomplete_analyzer": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": ["lowercase", "autocomplete_filter"]
        }
      },
      "filter": {
        "autocomplete_filter": {
          "type": "edge_ngram",
          "min_gram": 2,
          "max_gram": 20
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "name": {
        "type": "text",
        "analyzer": "autocomplete_analyzer",
        "fields": {
          "keyword": {"type": "keyword"},
          "suggest": {"type": "text"}
        }
      },
      "description": {"type": "text"},
      "daily_price": {"type": "double"},
      "vendor_rating": {"type": "float"},
      "latitude": {"type": "double"},
      "longitude": {"type": "double"}
    }
  }
}
```

### Environment Variables

```env
# Elasticsearch Configuration
ELASTICSEARCH_URL=http://localhost:9200
ELASTICSEARCH_INDEX=genrent_equipment
ELASTICSEARCH_ENABLED=true
```

## Usage Examples

### Basic Search
```bash
curl -X POST http://localhost:8080/api/search/advanced \
  -H "Content-Type: application/json" \
  -d '{
    "query": "generator",
    "page": 1,
    "per_page": 20
  }'
```

### Advanced Search with Filters
```bash
curl -X POST http://localhost:8080/api/search/advanced \
  -H "Content-Type: application/json" \
  -d '{
    "query": "generator 5kva",
    "category_ids": [1],
    "cities": ["Mumbai"],
    "min_daily_price": 1000,
    "max_daily_price": 5000,
    "min_vendor_rating": 4.0,
    "start_date": "2024-08-01",
    "end_date": "2024-08-05",
    "sort_by": "price_asc"
  }'
```

### Location-Based Search
```bash
curl -X POST http://localhost:8080/api/search/advanced \
  -H "Content-Type: application/json" \
  -d '{
    "query": "generator",
    "latitude": 19.0760,
    "longitude": 72.8777,
    "radius": 50
  }'
```

### Autocomplete
```bash
curl "http://localhost:8080/api/search/autocomplete?q=gen&field=name&size=10"
```

### Create Saved Search with Alerts
```bash
curl -X POST http://localhost:8080/api/search/saved \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "name": "Daily Generator Search",
    "query": "generator 5kva",
    "min_price": 1000,
    "max_price": 3000,
    "is_alert": true,
    "alert_frequency": "daily"
  }'
```

## Implementation Status

### ✅ Completed Features
- [x] Elasticsearch integration with proper index mapping
- [x] Full-text search with multi-field matching
- [x] Advanced filtering (price, category, location, rating, brand, tags)
- [x] Geo-distance search with radius filtering
- [x] Saved search queries with user profiles
- [x] Search alerts with configurable frequency
- [x] Autocomplete and search suggestions
- [x] Search history tracking
- [x] Search analytics and aggregations
- [x] Date-based availability filtering
- [x] Real-time inventory checking
- [x] Popular searches tracking
- [x] Search performance optimization

### 🔧 Configuration Required
1. **Elasticsearch Setup**
   ```bash
   # Install Elasticsearch 7.x
   brew install elasticsearch

   # Start Elasticsearch
   elasticsearch

   # Verify installation
   curl http://localhost:9200
   ```

2. **Environment Configuration**
   ```env
   ELASTICSEARCH_URL=http://localhost:9200
   ELASTICSEARCH_INDEX=genrent_equipment
   ELASTICSEARCH_ENABLED=true
   ```

3. **Database Migration**
   ```bash
   # Apply search system migration
   psql -U genrent -d genrent_db -f internal/migrate/004_search_system.sql
   ```

4. **Index Equipment Data**
   ```go
   // Index existing equipment
   elasticService.IndexEquipment(equipmentDoc)
   ```

## Performance Optimization

### Query Optimization
- Use filtered queries over bool queries where possible
- Implement query caching for popular searches
- Use pagination to limit result sets
- Optimize Elasticsearch shard count for dataset size

### Index Optimization
- Use appropriate analyzers for different fields
- Implement field boosting for relevance
- Use keyword fields for exact matching
- Enable doc_values for sorting and aggregations

### Caching Strategy
- Cache popular search results (5-15 minutes)
- Cache filter aggregations (15-30 minutes)
- Implement query result caching
- Use Elasticsearch query cache

## Monitoring & Analytics

### Key Metrics
- **Search Performance**
  - Average query response time
  - Zero-result search rate
  - Click-through rate
  - Search result relevance

- **User Behavior**
  - Most popular search queries
  - Filter usage statistics
  - Saved search adoption
  - Alert engagement rate

### Analytics Queries
```sql
-- Top search queries
SELECT query, COUNT(*) as searches
FROM search_history
GROUP BY query
ORDER BY searches DESC
LIMIT 20;

-- Zero-result searches
SELECT query, COUNT(*) as count
FROM search_history
WHERE result_count = 0
GROUP BY query
ORDER BY count DESC;

-- Click-through rate by date
SELECT
    DATE(created_at) as date,
    COUNT(*) as total_searches,
    SUM(CASE WHEN clicked_id IS NOT NULL THEN 1 ELSE 0 END) as clicks,
    (SUM(CASE WHEN clicked_id IS NOT NULL THEN 1 ELSE 0 END)::FLOAT / COUNT(*)) as ctr
FROM search_history
GROUP BY DATE(created_at)
ORDER BY date DESC;
```

## Troubleshooting

### Common Issues

#### 1. Elasticsearch Connection Failed
**Solution**: Verify Elasticsearch is running and accessible
```bash
curl http://localhost:9200
```

#### 2. Index Not Found
**Solution**: Create the equipment index
```go
elasticService.CreateIndex()
```

#### 3. Search Returns No Results
**Solution**: Check that equipment data is indexed
```bash
curl http://localhost:9200/genrent_equipment/_count
```

#### 4. Autocomplete Not Working
**Solution**: Verify search_suggestions table has data
```sql
SELECT * FROM search_suggestions LIMIT 10;
```

## Future Enhancements

### Planned Features
- [ ] Search result personalization based on user history
- [ ] Advanced filtering by equipment specifications
- [ ] Search result ranking algorithm optimization
- [ ] Voice search integration
- [ ] Image search capability
- [ ] Search analytics dashboard
- [ ] A/B testing for search relevance
- [ ] Multi-language search support
- [ ] Search result export
- [ ] Advanced search analytics with ML

---

**Last Updated**: 2026-07-14
**Version**: 1.0
**Status**: Production Ready

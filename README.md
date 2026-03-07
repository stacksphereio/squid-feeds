# Reef Feed Aggregator

Main aggregator service for the SquidStack Reef Feeds feature. Provides personalized weather and news feeds based on user location.

## Overview

The squid-feeds service:
- Validates JWT tokens from kraken-auth
- Extracts user country from JWT claims
- Maps country to region (EU, NA, SA, ASIA)
- Evaluates feature flags to control feed availability
- Routes requests to appropriate regional services
- Returns unified feed response to squid-ui

## Architecture

```
squid-ui → squid-feeds → reef-{eu|na|sa|asia}
```

## Endpoints

- `GET /health` - Liveness probe
- `GET /ready` - Readiness probe
- `GET /_flags` - Inspect feature flag values
- `GET /feeds` - Get personalized feeds (requires JWT Bearer token)

## Configuration

### Environment Variables

- `JWT_SECRET` - Shared secret for JWT validation (required)
- `FM_NAMESPACE` - CloudBees Feature Management namespace (default: "default")
- `REEF_EU_URL` - Europe regional service URL (default: "http://reef-eu:8080")
- `REEF_NA_URL` - North America regional service URL (default: "http://reef-na:8080")
- `REEF_SA_URL` - South America regional service URL (default: "http://reef-sa:8080")
- `REEF_ASIA_URL` - Asia regional service URL (default: "http://reef-asia:8080")

### Feature Management Key

Mount FM key file at `/app/config/fm.json`

## Feature Flags

- `feeds.enabled` - Global master switch
- `feeds.weather.enabled` - Weather feed type switch
- `feeds.news.enabled` - News feed type switch
- `feeds.region.eu.enabled` - Europe region switch
- `feeds.region.na.enabled` - North America region switch
- `feeds.region.sa.enabled` - South America region switch
- `feeds.region.asia.enabled` - Asia region switch
- `feeds.country.<code>.enabled` - Country-specific overrides

## Country → Region Mapping

- **EU**: GB, FR, DE, ES, IT, NL, BE, SE, NO, FI, PL, IE, PT, AT, CH
- **NA**: US, CA, MX
- **SA**: BR, AR, CL, PE, CO, VE, EC, UY, PY, BO
- **ASIA**: JP, CN, IN, SG, HK, KR, TH, ID, MY, PH, VN, TW

## Development

```bash
# Build
go build -o app main.go

# Run
JWT_SECRET=footest ./app

# Docker build
docker build -t squid-feeds .

# Docker run
docker run -p 8080:8080 -e JWT_SECRET=footest squid-feeds
```

## API Response Example

```json
{
  "user": {
    "id": "user-123",
    "country": "GB",
    "region": "EU"
  },
  "feeds": {
    "weather": {
      "enabled": true,
      "data": {
        "summary": "Cloudy",
        "temperatureC": 14.2,
        "feelsLikeC": 12.5
      }
    },
    "news": {
      "enabled": true,
      "items": [
        {
          "id": "article-1",
          "title": "Headline 1",
          "source": "Example News",
          "url": "https://example.com/article1",
          "publishedAt": "2025-12-08T12:34:56Z"
        }
      ]
    }
  }
}
```

Last updated: 2026-02-11 21:51 UTC

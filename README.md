# Vibe-Drop

A Dropbox-style file storage application built with Go microservices architecture. This project demonstrates modern cloud-native development patterns including microservices, AWS integration, Docker containerization, and environment-based configuration.

## Architecture

### System Overview
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Gateway   │───▶│  File Service   │───▶│   AWS S3 /      │
│   Port: 8080    │    │   Port: 8081    │    │   LocalStack    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Middleware    │    │   S3 Client     │    │   File Storage  │
│ • Rate Limiting │    │ • Presigned URLs│    │ • Bucket: vibe- │
│ • CORS          │    │ • Upload/Download│    │   drop-bucket   │
│ • Logging       │    │ • AWS SDK v2    │    │                 │
│ • Recovery      │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Microservices
- **API Gateway**: Entry point with middleware stack, routes requests to appropriate services
- **File Service**: Handles file operations, generates S3 presigned URLs, manages file metadata
- **Storage Layer**: AWS S3 (or LocalStack for development) for actual file storage

### Tech Stack
- **Language**: Go 1.21+
- **HTTP Framework**: Gorilla Mux
- **Cloud Storage**: AWS S3 with AWS SDK v2
- **Development**: LocalStack for local AWS services
- **Configuration**: Environment-based with godotenv
- **Containerization**: Docker (LocalStack)

## Functional Requirements
- ✅ Users can upload files via presigned URLs
- ✅ Users can download files via presigned URLs  
- 🚧 File metadata management (planned)
- 🚧 User authentication and authorization (planned)

## Non-functional Requirements
- ✅ Prioritize availability over consistency
- 🚧 Documents can be up to 50GB with resumable uploads/downloads (planned)
- ✅ High data integrity through AWS S3
- ✅ Environment-aware configuration (dev/staging/prod)

## API Documentation

### API Gateway (Port 8080)
All requests are proxied through the API Gateway to the File Service.

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET    | `/health` | Health check for API Gateway |
| POST   | `/files/upload-url` | Get presigned URL for file upload |
| GET    | `/files` | List all files (mock data) |
| GET    | `/files/{id}` | Get file metadata |
| GET    | `/files/{id}/download-url` | Get presigned URL for file download |
| DELETE | `/files/{id}` | Delete file (mock implementation) |

### File Service (Port 8081)
Direct service endpoints (normally accessed via API Gateway).

#### Upload File
```http
POST /files/upload-url
Content-Type: application/json

{
  "filename": "document.pdf"
}
```

**Response:**
```json
{
  "url": "http://localhost:4566/vibe-drop-bucket/uuid-filename?X-Amz-Signature=...",
  "expires_at": "2025-10-28T16:15:00Z",
  "file_id": "uuid-generated-id"
}
```

#### Download File
```http
GET /files/{file_id}/download-url
```

**Response:**
```json
{
  "url": "http://localhost:4566/vibe-drop-bucket/uuid-filename?X-Amz-Signature=...",
  "expires_at": "2025-10-28T16:15:00Z", 
  "file_id": "uuid-generated-id"
}
```

## Setup Instructions

### Prerequisites
- Go 1.21 or higher
- Docker (for LocalStack)
- AWS CLI (for LocalStack testing)

### Development Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd vibe-drop
   ```

2. **Configure environment**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start LocalStack (for local S3)**
   ```bash
   docker run --rm -p 4566:4566 localstack/localstack
   ```

4. **Create S3 bucket in LocalStack**
   ```bash
   aws --endpoint-url=http://localhost:4566 s3 mb s3://vibe-drop-bucket
   ```

5. **Start the services**
   ```bash
   # Terminal 1: Start File Service
   make file-service
   
   # Terminal 2: Start API Gateway  
   make api-gateway
   ```

6. **Test the setup**
   ```bash
   # Health checks
   curl http://localhost:8080/health  # API Gateway
   curl http://localhost:8081/health  # File Service
   
   # Upload a file
   curl -X POST http://localhost:8080/files/upload-url \
     -H "Content-Type: application/json" \
     -d '{"filename": "test.txt"}'
   ```

### Environment Configuration

The application supports three environments: `dev`, `staging`, `prod`

**Development (.env):**
```env
ENVIRONMENT=dev
S3_BUCKET=vibe-drop-bucket
S3_ENDPOINT=http://localhost:4566  # LocalStack
FILE_SERVICE_URL=http://localhost:8081
```

**Production:**
```env
ENVIRONMENT=prod
S3_BUCKET=your-production-bucket
# S3_ENDPOINT=  # Leave empty for real AWS
FILE_SERVICE_URL=https://file-service.yourdomain.com
```

## Core Entities
- **Files**: Stored in S3 with unique keys
- **File Metadata**: File information, ownership, S3 key mapping (planned)
- **Users**: User accounts and authentication (planned)

## Development Status
- ✅ **Phase 1**: Basic microservices architecture with S3 integration
- 🚧 **Phase 2**: Database integration for metadata (DynamoDB planned)
- 🚧 **Phase 3**: Large file support with chunking/multipart uploads (up to 50GB)
- 🚧 **Phase 4**: User authentication and authorization  
- 🚧 **Phase 5**: Advanced features (resumable uploads, file sharing, versioning)

## Contributing
This is a learning project built with Claude Code. Feel free to explore the codebase to understand microservices patterns and AWS integration in Go.
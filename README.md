# Vibe-Drop
The point of this project is to create a dropbox style program leveraging Go and Claude Code. The goal is to learn how to leverage Vibecoding effectively while learning the technologies required to make a Dropbox clone.

## Functional Requirements
- users can upload files
- users can download files

## Non-functional requirements
- prioritize availability over consistency
- documents can be up to 50GB
    - resumable downloads / upload supported
- high data integrity

## Core Entities
- files
- file_metadata
- users

## Services
### API Gateway
- rate limiting
- api routing

## File Service
- generate presignedURL

## Database
- DynamoDB
- handle file metadata
- handle users

## S3 storage
- stores actual file
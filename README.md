# GitHub Repository Backup Tool

A Go application that clones all repositories from your primary GitHub account and backs them up to a secondary GitHub account.

## Features

- Clone all repositories (public and private) from your primary GitHub account
- Create mirrored copies in your secondary GitHub account
- Skip repositories that already exist in the destination
- Clean up temporary files after backup
- Support for organizations (with proper token permissions)

## Prerequisites

- Go 1.16 or higher
- GitHub accounts (primary and secondary)
- Personal access tokens for both accounts

## Installation

1. Clone this repository:
   ```bash
   git clone https://github.com/yourusername/github-backup-tool.git
   cd github-backup-tool
2. Install dependencies
   ```bash
   go mod download

## Configuration

### .env
```bash
PRIMARY_GITHUB_TOKEN=your_primary_personal_access_token_here
PRIMARY_GITHUB_USER=your_primary_github_username

SECONDARY_GITHUB_TOKEN=your_secondary_personal_access_token_here
SECONDARY_GITHUB_USER=your_secondary_github_username

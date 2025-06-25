# GitHub Integration Setup

The website can display live GitHub statistics on the about page. Without proper configuration, it shows fallback values.

## Current Status

**❌ GitHub integration is currently DISABLED**

The website is showing fallback stats:
- Public Repos: 42 (instead of real count)
- Total Stars: 1337 (instead of real count)  
- Contributions: 523 (instead of real count)
- Other stats: Default values

## Quick Setup

### 1. Generate GitHub Token

1. Go to [GitHub Settings > Tokens](https://github.com/settings/tokens)
2. Click "Generate new token (classic)"
3. Give it a descriptive name like "Blockhead Consulting Website"
4. Select these scopes:
   - ✅ `read:user` - Read user profile data
   - ✅ `repo` - Access repositories (for contribution data)
5. Click "Generate token"
6. **Copy the token immediately** (you won't see it again)

### 2. Set Environment Variables

Add these to your environment:

```bash
export GITHUB_USERNAME=lancekrogers
export GITHUB_TOKEN=ghp_your_token_here
```

Or create a `.env` file in the project root:

```env
GITHUB_USERNAME=lancekrogers
GITHUB_TOKEN=ghp_your_token_here
```

### 3. Test the Integration

Run the test tool to verify everything works:

```bash
go run cmd/test-github/main.go
```

Expected output when working:
```
✅ SUCCESS: Live GitHub data retrieved!
Public Repos: [real number]
Total Stars: [real number]
```

### 4. Restart the Server

The GitHub service is initialized at startup, so restart the server:

```bash
# Stop current server
# Start with environment variables
GITHUB_USERNAME=lancekrogers GITHUB_TOKEN=ghp_your_token_here go run main.go
```

## Troubleshooting

### Getting Default Values (42, 1337, 523)?

This means GitHub integration is not working. Check:

1. **Environment variables set?**
   ```bash
   echo $GITHUB_USERNAME
   echo $GITHUB_TOKEN
   ```

2. **Token valid?** Test with curl:
   ```bash
   curl -H "Authorization: Bearer $GITHUB_TOKEN" https://api.github.com/user
   ```

3. **Check server logs** for specific error messages:
   - `401` = Bad token
   - `403` = Rate limit or insufficient permissions
   - `404` = User not found

### Rate Limiting

GitHub API has rate limits:
- **Authenticated**: 5,000 requests/hour
- **Unauthenticated**: 60 requests/hour

The website caches results for 1 hour to avoid hitting limits.

### Token Security

🔒 **Important**: Never commit tokens to git!

- Add `.env` to `.gitignore` 
- Use environment variables in production
- Rotate tokens periodically

## Production Deployment

For production, set environment variables in your deployment system:

```bash
# Docker
docker run -e GITHUB_USERNAME=lancekrogers -e GITHUB_TOKEN=ghp_xxx ...

# Systemd service
Environment=GITHUB_USERNAME=lancekrogers
Environment=GITHUB_TOKEN=ghp_xxx

# Cloud platforms
# Set via platform's environment variable interface
```

## API Details

The integration uses GitHub's GraphQL API to fetch:

- User profile data (repos, followers, etc.)
- Contribution calendar data
- Repository statistics
- Recent activity events

Data is cached for 1 hour to improve performance and respect rate limits.

## Fallback Behavior

The system is designed to never break the website:

- **No credentials**: Shows default values
- **Invalid credentials**: Shows default values after logging error
- **API errors**: Shows default values after logging error
- **Rate limited**: Uses cached data if available, otherwise defaults

This ensures the about page always loads, even if GitHub is unavailable.
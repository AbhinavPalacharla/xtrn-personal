# Gmail MCP Server

This is a Gmail MCP server that provides tools for managing emails, sending messages, and searching through your Gmail inbox.

## Features

The Gmail MCP server provides the following tools:

### Email Management
- **List Emails**: Retrieve a list of emails from your inbox with optional filtering
- **Get Email**: Retrieve a specific email by its ID with various format options
- **Search Emails**: Advanced email search using Gmail's search syntax
- **Send Email**: Compose and send emails with support for CC and BCC

### Gmail Features
- **List Labels**: View all available Gmail labels for organization
- **Gmail Search Syntax**: Support for Gmail's powerful search operators

### Authentication
- OAuth2 authentication with Google
- Automatic token refresh
- Secure credential management

## Available Tools

### `list_emails`
Lists emails from your Gmail inbox with optional filtering.

**Parameters:**
- `maxResults` (optional): Maximum number of emails to return (default: 10)
- `query` (optional): Gmail search query to filter emails
- `labelIds` (optional): Array of label IDs to filter emails by

**Example:**
```json
{
  "maxResults": 20,
  "query": "from:important@company.com",
  "labelIds": ["INBOX", "IMPORTANT"]
}
```

### `get_email`
Retrieves a specific email message by its ID.

**Parameters:**
- `messageId` (required): The ID of the email message to retrieve
- `format` (optional): The format of the message to return
  - `minimal`: Basic message metadata
  - `full`: Full message content (default)
  - `raw`: Raw message data
  - `metadata`: Message metadata only

**Example:**
```json
{
  "messageId": "18c1a2b3d4e5f6g7h8i9j0",
  "format": "full"
}
```

### `send_email`
Sends an email using Gmail.

**Parameters:**
- `to` (required): Recipient email address
- `subject` (required): Email subject line
- `body` (required): Email body content
- `cc` (optional): CC recipient email address
- `bcc` (optional): BCC recipient email address

**Example:**
```json
{
  "to": "recipient@example.com",
  "subject": "Meeting Reminder",
  "body": "Don't forget about our meeting tomorrow at 2 PM.",
  "cc": "team@example.com"
}
```

### `search_emails`
Searches emails using Gmail's advanced search syntax.

**Parameters:**
- `query` (required): Gmail search query
- `maxResults` (optional): Maximum number of results to return (default: 10)
- `labelIds` (optional): Array of label IDs to filter search results by

**Search Examples:**
- `from:example@gmail.com` - Emails from a specific sender
- `subject:meeting` - Emails with "meeting" in the subject
- `has:attachment` - Emails with attachments
- `is:unread` - Unread emails
- `after:2024/01/01` - Emails after a specific date
- `label:important` - Emails with a specific label

**Example:**
```json
{
  "query": "from:boss@company.com has:attachment",
  "maxResults": 15
}
```

### `list_labels`
Lists all available Gmail labels for the authenticated user.

**Parameters:** None

**Example:**
```json
{}
```

### `greet`
A simple test tool to verify the server is working.

**Parameters:**
- `name` (optional): Name to greet

**Example:**
```json
{
  "name": "John"
}
```

## Gmail Search Syntax

The Gmail MCP server supports Gmail's powerful search operators:

### Basic Search
- `from:email@domain.com` - Search by sender
- `to:email@domain.com` - Search by recipient
- `subject:keyword` - Search in subject line
- `has:attachment` - Emails with attachments
- `is:unread` - Unread emails
- `is:read` - Read emails

### Date and Time
- `after:2024/01/01` - Emails after a date
- `before:2024/01/31` - Emails before a date
- `newer_than:2d` - Emails newer than 2 days
- `older_than:1w` - Emails older than 1 week

### Labels and Categories
- `label:important` - Emails with specific label
- `category:primary` - Emails in primary category
- `category:social` - Social media emails
- `category:promotions` - Promotional emails

### Advanced
- `has:userlabels` - Emails with custom labels
- `filename:document.pdf` - Emails with specific attachments
- `larger:10M` - Emails larger than 10MB
- `smaller:1M` - Emails smaller than 1MB

## Building and Running

```bash
# Install dependencies
npm install

# Build the project
npm run build

# Run in development mode
npm run dev

# Start the server
npm start
```

## Environment Variables

The server requires the following environment variables:

- `CLIENT_ID`: Google OAuth client ID
- `CLIENT_SECRET`: Google OAuth client secret
- `REFRESH_TOKEN`: Google OAuth refresh token
- `USER_EMAIL`: User's Gmail address

## OAuth Scopes

The server requests the following Gmail scopes:
- `https://www.googleapis.com/auth/gmail.readonly` - Read access to emails
- `https://www.googleapis.com/auth/gmail.send` - Send emails
- `https://www.googleapis.com/auth/gmail.modify` - Modify emails (labels, etc.)

## Error Handling

The server provides comprehensive error handling:

- **Authentication Errors**: Handles OAuth token refresh and validation
- **Gmail API Errors**: Graceful handling of Gmail API errors
- **Input Validation**: Zod schema validation for all tool parameters
- **MCP Error Responses**: Proper MCP protocol error responses

## Security

- OAuth2 authentication with Google
- Secure token storage and refresh
- No sensitive data logging
- Input validation and sanitization

## Contributing

When adding new tools or modifying existing ones:

1. Follow the existing tool structure
2. Use Zod schemas for parameter validation
3. Implement proper error handling
4. Add comprehensive documentation
5. Test with various Gmail scenarios
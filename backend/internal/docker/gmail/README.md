# Gmail MCP Server

This is a comprehensive Gmail MCP server that provides advanced tools for managing emails, sending messages, organizing labels, and handling attachments. Built following the same architectural pattern as the Google Calendar MCP server.

## Features

The Gmail MCP server provides the following comprehensive tools:

### Email Management
- **List Emails**: Retrieve a list of emails from your inbox with optional filtering
- **Get Email**: Retrieve a specific email by its ID with various format options
- **Search Emails**: Advanced email search using Gmail's search syntax
- **Send Email**: Compose and send emails with support for CC, BCC, and reply-to
- **Draft Email**: Create email drafts without sending immediately

### Email Organization & Modification
- **Modify Email**: Change email labels to move emails between folders
- **Delete Email**: Permanently delete individual emails
- **Batch Modify Emails**: Modify labels for multiple emails in batches
- **Batch Delete Emails**: Delete multiple emails in batches

### Label Management
- **List Labels**: View all available Gmail labels for organization
- **Create Label**: Create new custom labels for email categorization
- **Update Label**: Modify existing label properties
- **Delete Label**: Remove custom labels (system labels cannot be deleted)
- **Get or Create Label**: Find existing label or create if it doesn't exist

### Attachment Handling
- **Download Attachment**: Download email attachments to specified locations

### Gmail Features
- **Gmail Search Syntax**: Support for Gmail's powerful search operators
- **Batch Operations**: Efficient processing of large numbers of emails
- **Thread Support**: Maintain email conversation threads

### Authentication
- OAuth2 authentication with Google
- Automatic token refresh
- Secure credential management

## Available Tools

### `list_emails`
Lists emails from your Gmail inbox with optional filtering.

**Parameters:**
- `maxResults` (optional): Maximum number of emails to return (default: 10, min: 1, max: 100)
- `query` (optional): Gmail search query to filter emails
- `labelIds` (optional): Array of label IDs to filter emails by
- `includeSpamTrash` (optional): Whether to include emails from SPAM and TRASH (default: false)
- `pageToken` (optional): Page token for pagination

**Example:**
```json
{
  "maxResults": 20,
  "query": "from:important@company.com",
  "labelIds": ["INBOX", "IMPORTANT"],
  "includeSpamTrash": false
}
```

### `get_email`
Retrieves a specific email message by its ID.

**Parameters:**
- `messageId` (required): The unique ID of the email message to retrieve
- `format` (optional): The format of the message to return
  - `minimal`: Basic message metadata
  - `full`: Full message content (default)
  - `raw`: Raw message data
  - `metadata`: Message metadata only
- `metadataHeaders` (optional): Array of specific header names to include (only applies when format is 'metadata')

**Example:**
```json
{
  "messageId": "18c1a2b3d4e5f6g7h8i9j0",
  "format": "full",
  "metadataHeaders": ["From", "Subject", "Date"]
}
```

### `send_email`
Sends an email using Gmail.

**Parameters:**
- `to` (required): Array of recipient email addresses
- `subject` (required): Email subject line
- `body` (required): Email body content in plain text
- `cc` (optional): Array of CC recipient email addresses
- `bcc` (optional): Array of BCC recipient email addresses
- `replyTo` (optional): Reply-to email address

**Example:**
```json
{
  "to": ["recipient@example.com", "team@example.com"],
  "subject": "Meeting Reminder",
  "body": "Don't forget about our meeting tomorrow at 2 PM.",
  "cc": ["manager@example.com"],
  "replyTo": "noreply@company.com"
}
```

### `draft_email`
Creates a draft email in Gmail without sending it immediately.

**Parameters:**
- `to` (required): Array of recipient email addresses
- `subject` (required): Email subject line
- `body` (required): Email body content in plain text
- `cc` (optional): Array of CC recipient email addresses
- `bcc` (optional): Array of BCC recipient email addresses
- `replyTo` (optional): Reply-to email address
- `threadId` (optional): Thread ID to add this draft to an existing conversation

**Example:**
```json
{
  "to": ["client@example.com"],
  "subject": "Project Update",
  "body": "Here's the latest update on our project...",
  "threadId": "thread123"
}
```

### `search_emails`
Searches emails using Gmail's advanced search syntax.

**Parameters:**
- `query` (required): Gmail search query using Gmail's search syntax
- `maxResults` (optional): Maximum number of results to return (default: 10, min: 1, max: 100)
- `labelIds` (optional): Array of label IDs to filter search results by
- `includeSpamTrash` (optional): Whether to include emails from SPAM and TRASH (default: false)
- `pageToken` (optional): Page token for pagination
- `orderBy` (optional): Sort order for results (default: 'internalDate')

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
  "maxResults": 15,
  "orderBy": "date"
}
```

### `modify_email`
Modifies email labels to move emails between folders.

**Parameters:**
- `messageId` (required): The unique ID of the email message to modify
- `addLabelIds` (optional): Array of label IDs to add to the email
- `removeLabelIds` (optional): Array of label IDs to remove from the email
- `labelIds` (optional): Legacy parameter - array of label IDs to add

**Example:**
```json
{
  "messageId": "18c1a2b3d4e5f6g7h8i9j0",
  "addLabelIds": ["IMPORTANT", "WORK"],
  "removeLabelIds": ["INBOX", "UNREAD"]
}
```

### `delete_email`
Permanently deletes an email message.

**Parameters:**
- `messageId` (required): The unique ID of the email message to delete

**Example:**
```json
{
  "messageId": "18c1a2b3d4e5f6g7h8i9j0"
}
```

### `batch_modify_emails`
Modifies labels for multiple emails in batches.

**Parameters:**
- `messageIds` (required): Array of email message IDs to modify (1-1000 IDs)
- `addLabelIds` (optional): Array of label IDs to add to all specified emails
- `removeLabelIds` (optional): Array of label IDs to remove from all specified emails
- `batchSize` (optional): Number of emails to process in each batch (default: 50, min: 1, max: 100)

**Example:**
```json
{
  "messageIds": ["id1", "id2", "id3"],
  "addLabelIds": ["IMPORTANT"],
  "removeLabelIds": ["UNREAD"],
  "batchSize": 25
}
```

### `batch_delete_emails`
Deletes multiple emails in batches.

**Parameters:**
- `messageIds` (required): Array of email message IDs to delete (1-1000 IDs)
- `batchSize` (optional): Number of emails to process in each batch (default: 50, min: 1, max: 100)

**Example:**
```json
{
  "messageIds": ["id1", "id2", "id3"],
  "batchSize": 25
}
```

### `list_labels`
Lists all available Gmail labels.

**Parameters:** None

**Example:**
```json
{}
```

### `create_label`
Creates a new custom Gmail label.

**Parameters:**
- `name` (required): Name for the new label (1-100 characters)
- `messageListVisibility` (optional): Whether to show the label in the message list (default: 'show')
- `labelListVisibility` (optional): Whether to show the label in the label list (default: 'labelShow')

**Example:**
```json
{
  "name": "Work Projects",
  "messageListVisibility": "show",
  "labelListVisibility": "labelShow"
}
```

### `update_label`
Updates an existing Gmail label.

**Parameters:**
- `id` (required): The unique ID of the label to update
- `name` (optional): New name for the label
- `messageListVisibility` (optional): Whether to show the label in the message list
- `labelListVisibility` (optional): Whether to show the label in the label list

**Example:**
```json
{
  "id": "Label_123",
  "name": "Updated Work Projects",
  "messageListVisibility": "hide"
}
```

### `delete_label`
Deletes a Gmail label.

**Parameters:**
- `id` (required): The unique ID of the label to delete

**Example:**
```json
{
  "id": "Label_123"
}
```

### `get_or_create_label`
Gets an existing label by name or creates it if it doesn't exist.

**Parameters:**
- `name` (required): Name of the label to find or create
- `messageListVisibility` (optional): Whether to show the label in the message list (default: 'show')
- `labelListVisibility` (optional): Whether to show the label in the label list (default: 'labelShow')

**Example:**
```json
{
  "name": "Work Projects"
}
```

### `download_attachment`
Downloads an email attachment to a specified location.

**Parameters:**
- `messageId` (required): ID of the email message containing the attachment
- `attachmentId` (required): ID of the attachment to download
- `filename` (optional): Filename to save the attachment as
- `savePath` (optional): Directory path to save the attachment (defaults to current directory)

**Example:**
```json
{
  "messageId": "18c1a2b3d4e5f6g7h8i9j0",
  "attachmentId": "attachment123",
  "filename": "document.pdf",
  "savePath": "/downloads"
}
```

### `greet`
A simple test tool to verify the server is working.

**Parameters:**
- `name` (optional): Name to greet (1-100 characters)

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
- **Batch Operation Errors**: Detailed reporting of batch operation successes and failures

## Security

- OAuth2 authentication with Google
- Secure token storage and refresh
- No sensitive data logging
- Input validation and sanitization
- File system access controls for attachment downloads

## Contributing

When adding new tools or modifying existing ones:

1. Follow the existing tool structure
2. Use Zod schemas for parameter validation
3. Implement proper error handling
4. Add comprehensive documentation
5. Test with various Gmail scenarios
6. Ensure proper batch processing for bulk operations
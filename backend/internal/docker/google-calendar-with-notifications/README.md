# Google Calendar MCP Server with Notifications

This is an enhanced version of the Google Calendar MCP server that supports creating and updating events with custom reminders/notifications.

## New Features

### Reminders/Notifications Support

Both the `create_event` and `update_event` tools now support a `reminders` parameter that allows you to:

1. **Use Default Reminders**: Set `useDefault: true` to use the calendar's default reminder settings
2. **Custom Reminders**: Set `useDefault: false` and provide custom reminder times

### Reminder Schema

```typescript
reminders: {
  useDefault?: boolean;  // Use default reminders (true) or custom reminders (false)
  overrides?: Array<{
    method: "email" | "popup";  // Reminder method
    minutes: number;            // Minutes before event to send reminder
  }>;
}
```

### Examples

#### Create Event with Email Reminder 30 minutes before
```json
{
  "summary": "Team Meeting",
  "start": {
    "dateTime": "2024-01-15T10:00:00-08:00",
    "timeZone": "America/Los_Angeles"
  },
  "end": {
    "dateTime": "2024-01-15T11:00:00-08:00",
    "timeZone": "America/Los_Angeles"
  },
  "reminders": {
    "useDefault": false,
    "overrides": [
      {
        "method": "email",
        "minutes": 30
      }
    ]
  }
}
```

#### Create Event with Multiple Reminders
```json
{
  "summary": "Important Presentation",
  "start": {
    "dateTime": "2024-01-20T14:00:00-08:00",
    "timeZone": "America/Los_Angeles"
  },
  "end": {
    "dateTime": "2024-01-20T15:00:00-08:00",
    "timeZone": "America/Los_Angeles"
  },
  "reminders": {
    "useDefault": false,
    "overrides": [
      {
        "method": "popup",
        "minutes": 15
      },
      {
        "method": "email",
        "minutes": 60
      },
      {
        "method": "email",
        "minutes": 1440  // 24 hours before
      }
    ]
  }
}
```

#### Use Default Calendar Reminders
```json
{
  "summary": "Casual Meeting",
  "start": {
    "dateTime": "2024-01-25T16:00:00-08:00",
    "timeZone": "America/Los_Angeles"
  },
  "end": {
    "dateTime": "2024-01-25T17:00:00-08:00",
    "timeZone": "America/Los_Angeles"
  },
  "reminders": {
    "useDefault": true
  }
}
```

## Available Tools

- `create_event`: Creates a new calendar event with optional reminders
- `update_event`: Updates an existing calendar event including reminders
- `list_events`: Lists events in a time range
- `get_event`: Gets details of a specific event
- `delete_event`: Deletes an event
- `greet`: Test tool for basic functionality

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

## Changes from Original

This version extends the original Google Calendar MCP server with:

1. Enhanced `create_event` tool with reminders support
2. Enhanced `update_event` tool with reminders support
3. Proper handling of Google Calendar API reminder structure
4. **Consistent error handling across all tools** - all tools now have comprehensive error handling with:
   - Specific auth error handling (`handleAuthErrorToMCP`)
   - LLM-handleable error responses (`NewMCPLLMErrorResponse`)
   - Fallback for unknown errors (`NewMCPError`)
5. Updated package name and description
6. Comprehensive documentation and examples

The original functionality remains unchanged - this is purely additive.
import { z } from "zod";
import { type InferSchema } from "xmcp";
import { gmailOauthClient } from "../auth";
import { google } from "googleapis";
import { NewMCPResponse } from "../utils/error";

export const schema = {
  query: z.string().min(1).describe("Gmail search query using Gmail's search syntax (required, cannot be empty). Examples: 'from:example@gmail.com', 'subject:meeting', 'has:attachment', 'is:unread', 'after:2024/01/01', 'label:important'"),
  maxResults: z.number().min(1).max(100).optional().describe("Maximum number of search results to return (optional, default: 10, min: 1, max: 100)"),
  labelIds: z.array(z.string()).optional().describe("Array of Gmail label IDs to filter search results by (optional, e.g., ['INBOX', 'IMPORTANT', 'SENT', 'DRAFT'])"),
  includeSpamTrash: z.boolean().optional().describe("Whether to include emails from SPAM and TRASH in the search results (optional, default: false)"),
  pageToken: z.string().optional().describe("Page token for pagination to get the next page of search results (optional)"),
  orderBy: z.enum(["internalDate", "date"]).optional().describe("Sort order for search results (optional, default: 'internalDate'). 'internalDate' sorts by Gmail's internal date, 'date' sorts by the email's Date header"),
};

export const metadata = {
  name: "search_emails",
  description: "Searches emails using Gmail's powerful search syntax with advanced filtering options, label filtering, and pagination support",
  annotations: {
    title: "Search Gmail emails",
    readOnlyHint: true,
    idempotentHint: true,
  },
};

export default async function searchEmails(args: InferSchema<typeof schema>) {
  const client = await gmailOauthClient();
  const gmail = google.gmail({ version: "v1", auth: client });

  const res = await gmail.users.messages.list({
    userId: "me",
    q: args.query,
    maxResults: args.maxResults ?? 10,
    labelIds: args.labelIds,
    includeSpamTrash: args.includeSpamTrash ?? false,
    pageToken: args.pageToken,
    orderBy: args.orderBy ?? "internalDate",
  });

  return NewMCPResponse(JSON.stringify(res.data));
}
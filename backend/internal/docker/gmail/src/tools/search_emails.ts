import { z } from "zod";
import { type InferSchema } from "xmcp";
import { gmailOauthClient } from "../auth";
import { google } from "googleapis";
import { NewMCPResponse } from "../utils/error";

export const schema = {
  query: z.string().describe("Gmail search query (e.g., 'from:example@gmail.com', 'subject:meeting', 'has:attachment')"),
  maxResults: z.number().optional().describe("Maximum number of results to return (default: 10)"),
  labelIds: z.array(z.string()).optional().describe("Label IDs to filter search results by"),
};

export const metadata = {
  name: "search_emails",
  description: "Searches emails using Gmail's search syntax with advanced filtering options",
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
  });

  return NewMCPResponse(JSON.stringify(res.data));
}
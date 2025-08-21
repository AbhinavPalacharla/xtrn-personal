import { z } from "zod";
import { type InferSchema } from "xmcp";
import { gmailOauthClient } from "../auth";
import { google } from "googleapis";
import { NewMCPResponse } from "../utils/error";

export const schema = {
  to: z.array(z.string().email()).describe("Array of recipient email addresses (required)"),
  subject: z.string().min(1).describe("Email subject line (required, cannot be empty)"),
  body: z.string().min(1).describe("Email body content in plain text (required, cannot be empty)"),
  cc: z.array(z.string().email()).optional().describe("Array of CC recipient email addresses (optional)"),
  bcc: z.array(z.string().email()).optional().describe("Array of BCC recipient email addresses (optional)"),
  replyTo: z.string().email().optional().describe("Reply-to email address (optional, defaults to sender)"),
  threadId: z.string().optional().describe("Thread ID to add this draft to an existing conversation (optional)"),
};

export const metadata = {
  name: "draft_email",
  description: "Creates a draft email in Gmail without sending it immediately. The draft can be reviewed and sent later from the Gmail interface.",
  annotations: {
    title: "Draft Gmail email",
    readOnlyHint: false,
    idempotentHint: false,
  },
};

export default async function draftEmail(args: InferSchema<typeof schema>) {
  const client = await gmailOauthClient();
  const gmail = google.gmail({ version: "v1", auth: client });

  // Create email message
  const message = [
    `To: ${args.to.join(", ")}`,
    `Subject: ${args.subject}`,
    ...(args.cc && args.cc.length > 0 ? [`Cc: ${args.cc.join(", ")}`] : []),
    ...(args.bcc && args.bcc.length > 0 ? [`Bcc: ${args.bcc.join(", ")}`] : []),
    ...(args.replyTo ? [`Reply-To: ${args.replyTo}`] : []),
    "",
    args.body,
  ].join("\n");

  // Encode the message in base64
  const encodedMessage = Buffer.from(message).toString("base64").replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");

  const messageRequest: any = {
    raw: encodedMessage,
  };

  // Add threadId if specified
  if (args.threadId) {
    messageRequest.threadId = args.threadId;
  }

  const res = await gmail.users.drafts.create({
    userId: "me",
    requestBody: {
      message: messageRequest,
    },
  });

  return NewMCPResponse(JSON.stringify(res.data));
}
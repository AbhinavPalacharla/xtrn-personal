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
};

export const metadata = {
  name: "send_email",
  description: "Sends an email using Gmail with support for multiple recipients, CC, BCC, and reply-to addresses",
  annotations: {
    title: "Send Gmail email",
    readOnlyHint: false,
    idempotentHint: false,
  },
};

export default async function sendEmail(args: InferSchema<typeof schema>) {
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

  const res = await gmail.users.messages.send({
    userId: "me",
    requestBody: {
      raw: encodedMessage,
    },
  });

  return NewMCPResponse(JSON.stringify(res.data));
}
import { z } from "zod";
import { type InferSchema } from "xmcp";
import { gmailOauthClient } from "../auth";
import { google } from "googleapis";
import { NewMCPResponse } from "../utils/error";

export const schema = {
  to: z.string().describe("Recipient email address"),
  subject: z.string().describe("Email subject line"),
  body: z.string().describe("Email body content"),
  cc: z.string().optional().describe("CC recipient email address"),
  bcc: z.string().optional().describe("BCC recipient email address"),
};

export const metadata = {
  name: "send_email",
  description: "Sends an email using Gmail",
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
    `To: ${args.to}`,
    `Subject: ${args.subject}`,
    ...(args.cc ? [`Cc: ${args.cc}`] : []),
    ...(args.bcc ? [`Bcc: ${args.bcc}`] : []),
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
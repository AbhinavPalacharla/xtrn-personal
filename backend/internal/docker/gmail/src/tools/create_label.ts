import { z } from "zod";
import { type InferSchema } from "xmcp";
import { gmailOauthClient } from "../auth";
import { google } from "googleapis";
import { NewMCPResponse } from "../utils/error";

export const schema = {
  name: z.string().min(1).max(100).describe("Name for the new label (required, 1-100 characters)"),
  messageListVisibility: z.enum(["show", "hide"]).optional().describe("Whether to show the label in the message list (optional, default: 'show')"),
  labelListVisibility: z.enum(["labelShow", "labelHide"]).optional().describe("Whether to show the label in the label list (optional, default: 'labelShow')"),
};

export const metadata = {
  name: "create_label",
  description: "Creates a new custom Gmail label for organizing emails. Labels can be used to categorize and filter emails.",
  annotations: {
    title: "Create Gmail label",
    readOnlyHint: false,
    idempotentHint: false,
  },
};

export default async function createLabel(args: InferSchema<typeof schema>) {
  const client = await gmailOauthClient();
  const gmail = google.gmail({ version: "v1", auth: client });

  const labelRequest: any = {
    name: args.name,
    messageListVisibility: args.messageListVisibility ?? "show",
    labelListVisibility: args.labelListVisibility ?? "labelShow",
  };

  const res = await gmail.users.labels.create({
    userId: "me",
    requestBody: labelRequest,
  });

  return NewMCPResponse(JSON.stringify({
    success: true,
    label: res.data,
    message: `Label '${args.name}' created successfully with ID: ${res.data.id}`
  }));
}
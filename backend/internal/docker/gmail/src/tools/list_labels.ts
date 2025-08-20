import { z } from "zod";
import { type InferSchema } from "xmcp";
import { gmailOauthClient } from "../auth";
import { google } from "googleapis";
import { NewMCPResponse } from "../utils/error";

export const schema = {
  // No parameters needed for this tool
};

export const metadata = {
  name: "list_labels",
  description: "Lists all available Gmail labels for the authenticated user",
  annotations: {
    title: "List Gmail labels",
    readOnlyHint: true,
    idempotentHint: true,
  },
};

export default async function listLabels(args: InferSchema<typeof schema>) {
  const client = await gmailOauthClient();
  const gmail = google.gmail({ version: "v1", auth: client });

  const res = await gmail.users.labels.list({
    userId: "me",
  });

  return NewMCPResponse(JSON.stringify(res.data));
}
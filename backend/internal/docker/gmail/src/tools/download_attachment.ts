import { z } from "zod";
import { type InferSchema } from "xmcp";
import { gmailOauthClient } from "../auth";
import { google } from "googleapis";
import { NewMCPResponse } from "../utils/error";
import * as fs from "fs";
import * as path from "path";

export const schema = {
  messageId: z.string().min(1).describe("ID of the email message containing the attachment (required, cannot be empty)"),
  attachmentId: z.string().min(1).describe("ID of the attachment to download (required, cannot be empty)"),
  filename: z.string().optional().describe("Filename to save the attachment as (optional, if not provided, uses original filename)"),
  savePath: z.string().optional().describe("Directory path to save the attachment (optional, defaults to current working directory)"),
};

export const metadata = {
  name: "download_attachment",
  description: "Downloads an email attachment to a specified location. Supports custom filenames and save paths.",
  annotations: {
    title: "Download Gmail attachment",
    readOnlyHint: false,
    idempotentHint: false,
  },
};

export default async function downloadAttachment(args: InferSchema<typeof schema>) {
  const client = await gmailOauthClient();
  const gmail = google.gmail({ version: "v1", auth: client });

  try {
    // Get the attachment data from Gmail API
    const attachmentResponse = await gmail.users.messages.attachments.get({
      userId: "me",
      messageId: args.messageId,
      id: args.attachmentId,
    });

    if (!attachmentResponse.data.data) {
      throw new Error("No attachment data received from Gmail API");
    }

    // Decode the base64 data
    const data = attachmentResponse.data.data;
    const buffer = Buffer.from(data, "base64url");

    // Determine save path and filename
    const savePath = args.savePath || process.cwd();
    let filename = args.filename;
    
    if (!filename) {
      // Get original filename from message if not provided
      const messageResponse = await gmail.users.messages.get({
        userId: "me",
        id: args.messageId,
        format: "full",
      });
      
      // Find the attachment part to get original filename
      const findAttachment = (part: any): string | null => {
        if (part.body && part.body.attachmentId === args.attachmentId) {
          return part.filename || `attachment-${args.attachmentId}`;
        }
        if (part.parts) {
          for (const subpart of part.parts) {
            const found = findAttachment(subpart);
            if (found) return found;
          }
        }
        return null;
      };
      
      filename = findAttachment(messageResponse.data.payload) || `attachment-${args.attachmentId}`;
    }

    // Ensure save directory exists
    if (!fs.existsSync(savePath)) {
      fs.mkdirSync(savePath, { recursive: true });
    }

    // Write file
    const fullPath = path.join(savePath, filename);
    fs.writeFileSync(fullPath, buffer);

    return NewMCPResponse(JSON.stringify({
      success: true,
      message: `Attachment downloaded successfully`,
      filename: filename,
      size: buffer.length,
      savedTo: fullPath,
      attachmentId: args.attachmentId,
      messageId: args.messageId
    }));
  } catch (error: any) {
    throw new Error(`Failed to download attachment: ${error.message}`);
  }
}
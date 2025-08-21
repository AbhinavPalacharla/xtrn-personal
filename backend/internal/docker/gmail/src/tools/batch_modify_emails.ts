import { z } from "zod";
import { type InferSchema } from "xmcp";
import { gmailOauthClient } from "../auth";
import { google } from "googleapis";
import { NewMCPResponse } from "../utils/error";

export const schema = {
  messageIds: z.array(z.string().min(1)).min(1).max(1000).describe("Array of email message IDs to modify (required, 1-1000 IDs)"),
  addLabelIds: z.array(z.string()).optional().describe("Array of label IDs to add to all specified emails (optional, e.g., ['IMPORTANT', 'WORK'])"),
  removeLabelIds: z.array(z.string()).optional().describe("Array of label IDs to remove from all specified emails (optional, e.g., ['INBOX', 'UNREAD'])"),
  batchSize: z.number().min(1).max(100).optional().describe("Number of emails to process in each batch (optional, default: 50, min: 1, max: 100)"),
};

export const metadata = {
  name: "batch_modify_emails",
  description: "Modifies labels for multiple emails in batches for efficient bulk operations. Useful for organizing large numbers of emails at once.",
  annotations: {
    title: "Batch modify Gmail email labels",
    readOnlyHint: false,
    idempotentHint: false,
  },
};

export default async function batchModifyEmails(args: InferSchema<typeof schema>) {
  const client = await gmailOauthClient();
  const gmail = google.gmail({ version: "v1", auth: client });

  const messageIds = args.messageIds;
  const batchSize = args.batchSize ?? 50;
  
  // Prepare request body
  const requestBody: any = {};
  
  if (args.addLabelIds && args.addLabelIds.length > 0) {
    requestBody.addLabelIds = args.addLabelIds;
  }
  
  if (args.removeLabelIds && args.removeLabelIds.length > 0) {
    requestBody.removeLabelIds = args.removeLabelIds;
  }

  // Ensure we have at least one operation to perform
  if (Object.keys(requestBody).length === 0) {
    throw new Error("At least one of addLabelIds or removeLabelIds must be provided");
  }

  // Process messages in batches
  const successes: string[] = [];
  const failures: Array<{ messageId: string; error: string }> = [];
  
  for (let i = 0; i < messageIds.length; i += batchSize) {
    const batch = messageIds.slice(i, i + batchSize);
    
    try {
      // Process batch in parallel
      const batchResults = await Promise.all(
        batch.map(async (messageId) => {
          try {
            await gmail.users.messages.modify({
              userId: "me",
              id: messageId,
              requestBody: requestBody,
            });
            return { messageId, success: true };
          } catch (error: any) {
            return { messageId, success: false, error: error.message };
          }
        })
      );

      // Collect results
      batchResults.forEach(result => {
        if (result.success) {
          successes.push(result.messageId);
        } else {
          failures.push({ messageId: result.messageId, error: result.error });
        }
      });
    } catch (error: any) {
      // If batch fails, try individual items
      for (const messageId of batch) {
        try {
          await gmail.users.messages.modify({
            userId: "me",
            id: messageId,
            requestBody: requestBody,
          });
          successes.push(messageId);
        } catch (itemError: any) {
          failures.push({ messageId, error: itemError.message });
        }
      }
    }
  }

  // Generate summary
  const successCount = successes.length;
  const failureCount = failures.length;
  
  let resultText = `Batch label modification complete.\n`;
  resultText += `Successfully processed: ${successCount} messages\n`;
  
  if (failureCount > 0) {
    resultText += `Failed to process: ${failureCount} messages\n\n`;
    resultText += `Failed message IDs:\n`;
    resultText += failures.map(f => `- ${f.messageId.substring(0, 16)}... (${f.error})`).join('\n');
  }

  return NewMCPResponse(JSON.stringify({
    success: true,
    summary: resultText,
    totalProcessed: messageIds.length,
    successful: successCount,
    failed: failureCount,
    successfulIds: successes,
    failedIds: failures.map(f => ({ messageId: f.messageId, error: f.error }))
  }));
}
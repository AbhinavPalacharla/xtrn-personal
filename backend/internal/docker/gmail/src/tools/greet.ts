import { z } from "zod";
import { type InferSchema } from "xmcp";
import { NewMCPResponse } from "../utils/error";

export const schema = {
  name: z.string().optional().describe("Name to greet (optional)"),
};

export const metadata = {
  name: "greet",
  description: "A simple greeting tool to test the Gmail MCP server",
  annotations: {
    title: "Greet user",
    readOnlyHint: true,
    idempotentHint: true,
  },
};

export default async function greet(args: InferSchema<typeof schema>) {
  const greeting = args.name ? `Hello, ${args.name}! Welcome to the Gmail MCP server.` : "Hello! Welcome to the Gmail MCP server.";
  
  return NewMCPResponse(greeting);
}
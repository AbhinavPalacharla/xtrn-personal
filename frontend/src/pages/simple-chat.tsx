"use client";

import { useChat } from "@ai-sdk/react";
import { DefaultChatTransport } from "ai";
import type { UIMessage } from "ai";
import { useState } from "react";

// No tools passed: use the base UIMessage type
type Msg = UIMessage;
type Part = Msg["parts"][number];

export default function SimpleChat() {
  const [input, setInput] = useState("");
  const [authRequests, setAuthRequests] = useState<Array<{provider: string, authorization_url: string}>>([]);

  // Simple transport always pointing to new chat endpoint
  const transport = new DefaultChatTransport({
    api: "http://localhost:8080/chats",
    prepareSendMessagesRequest: ({ messages }) => ({
      body: { message: messages[messages.length - 1] },
      headers: {
        'Content-Type': 'application/json',
      }
    }),
    // Custom fetch to handle auth requests
    fetch: async (url, options) => {
      const response = await fetch(url, options);

      // Create a custom response that can handle auth requests
      const originalReader = response.body?.getReader();
      if (!originalReader) return response;

      const stream = new ReadableStream({
        start(controller) {
          function pump(): Promise<void> {
            return originalReader!.read().then(({ done, value }) => {
              if (done) {
                controller.close();
                return;
              }

              // Convert the chunk to string and check for auth requests
              const chunkStr = new TextDecoder().decode(value);
              console.log("Received chunk:", chunkStr);

              // Check if this chunk contains an auth request
              if (chunkStr.includes('"type":"data-auth-request"')) {
                try {
                  const lines = chunkStr.split('\n');
                  for (const line of lines) {
                    if (line.startsWith('data: ')) {
                      const data = JSON.parse(line.substring(6));
                      if (data.type === 'data-auth-request') {
                        console.log("Found auth request:", data);
                        setAuthRequests(prev => [...prev, data.data]);
                      }
                    }
                  }
                } catch (e) {
                  console.error("Error parsing auth request:", e);
                }
              }

              controller.enqueue(value);
              return pump();
            });
          }
          return pump();
        }
      });

      return new Response(stream, {
        status: response.status,
        statusText: response.statusText,
        headers: response.headers
      });
    }
  });

  const { messages, sendMessage } = useChat<Msg>({
    transport,
    onError: (e) => console.error("useChat error:", e),
  });

  const renderPart = (part: Part, key: string) => {
    switch (part.type) {
      case "text":
        return <div key={key}>{part.text}</div>;

      case "dynamic-tool": {
        const name = part.toolName ?? "tool";
        switch (part.state) {
          case "input-streaming":
            return (
              <div key={key}>
                <em>{name}</em>: preparing input…
              </div>
            );
          case "input-available":
            return (
              <div key={key}>
                <strong>↪ {name} request</strong>
                <pre>{JSON.stringify(part.input, null, 2)}</pre>
              </div>
            );
          case "output-available":
            return (
              <div key={key}>
                <strong>⬅ {name} response</strong>
                <pre>{JSON.stringify(part.output, null, 2)}</pre>
              </div>
            );
          case "output-error":
            return (
              <div key={key} style={{ color: "crimson" }}>
                <strong>{name} error:</strong> {String(part.errorText ?? "An error occurred")}
              </div>
            );
          default:
            return <div key={key}>{name}…</div>;
        }
      }



      case "step-start":
        return <hr key={key} />;

      default:
        return (
          <pre key={key}>
            {JSON.stringify(part, null, 2)}
          </pre>
        );
    }
  };

  return (
    <div style={{ maxWidth: '800px', margin: '0 auto', padding: '20px' }}>
      {/* Auth Requests */}
      {authRequests.map((authReq, index) => (
        <div key={`auth-${index}`} style={{
          backgroundColor: '#fff3cd',
          border: '1px solid #ffeaa7',
          borderRadius: '8px',
          padding: '12px',
          margin: '8px 0'
        }}>
          <div style={{ fontWeight: 'bold', color: '#856404', marginBottom: '8px' }}>
            🔐 Authentication Required
          </div>
          <div style={{ marginBottom: '8px' }}>
            The tool needs access to your {authReq.provider} account.
          </div>
          <a
            href={authReq.authorization_url}
            target="_blank"
            rel="noopener noreferrer"
            style={{
              display: 'inline-block',
              backgroundColor: '#007bff',
              color: 'white',
              padding: '8px 16px',
              textDecoration: 'none',
              borderRadius: '4px',
              fontSize: '14px'
            }}
          >
            Authorize {authReq.provider}
          </a>
          <div style={{ marginTop: '8px', fontSize: '12px', color: '#6c757d' }}>
            After authorization, return here and try your request again.
          </div>
        </div>
      ))}

      {/* Messages */}
      <div style={{ marginBottom: '20px' }}>
        {messages.map((m, mi) => (
          <div key={m.id ?? mi} style={{ marginBottom: 16 }}>
            <div style={{ fontWeight: 600, marginBottom: 4 }}>
              {m.role === "user" ? "User" : "AI"}:
            </div>
            {m.parts.map((p, pi) => renderPart(p as Part, `${m.id ?? mi}-${pi}`))}
          </div>
        ))}
      </div>

      {/* Input Form */}
      <form
        onSubmit={(e) => {
          e.preventDefault();
          if (input.trim()) {
            sendMessage({ text: input });
            setInput("");
          }
        }}
        style={{ display: 'flex', gap: '10px' }}
      >
        <input
          value={input}
          placeholder="say something..."
          onChange={(e) => setInput(e.currentTarget.value)}
          style={{
            flex: 1,
            padding: '10px',
            border: '1px solid #ddd',
            borderRadius: '4px'
          }}
        />
        <button
          type="submit"
          style={{
            padding: '10px 20px',
            backgroundColor: '#28a745',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer'
          }}
        >
          Send
        </button>
      </form>
    </div>
  );
}
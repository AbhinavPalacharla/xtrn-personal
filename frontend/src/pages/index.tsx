"use client";

import { useChat } from "@ai-sdk/react";
import { DefaultChatTransport } from "ai";
import type { UIMessage } from "ai";
import { useState, useEffect, useRef, useMemo } from "react";

// No tools passed: use the base UIMessage type
type Msg = UIMessage;
type Part = Msg["parts"][number];

export default function Chat() {
  const [input, setInput] = useState("");
  const [chatID, setChatID] = useState<string | null>(null);
  const [manualChatID, setManualChatID] = useState("");
  const chatIDRef = useRef<string | null>(null);

  // Keep ref in sync with state
  useEffect(() => {
    chatIDRef.current = chatID;
  }, [chatID]);

  // Check for chatID in headers, URL params, or other sources on component mount
  useEffect(() => {
    // Try to get chatID from URL params first
    const urlParams = new URLSearchParams(window.location.search);
    const urlChatID = urlParams.get('chatID');
    
    if (urlChatID) {
      setChatID(urlChatID);
      return;
    }

    // Try to get chatID from meta tag (if set by server-side rendering or parent)
    const headerChatID = document.querySelector('meta[name="chatID"]')?.getAttribute('content');
    if (headerChatID) {
      setChatID(headerChatID);
      return;
    }

    // Try to get chatID from data attribute on body or html
    const bodyChatID = document.body.getAttribute('data-chat-id') || 
                      document.documentElement.getAttribute('data-chat-id');
    if (bodyChatID) {
      setChatID(bodyChatID);
      return;
    }

    // Check for chatID in window object (if set by parent frame)
    if (typeof window !== 'undefined' && (window as any).chatID) {
      setChatID((window as any).chatID);
      return;
    }

    // Check localStorage for existing chatID
    const storedChatID = localStorage.getItem('currentChatID');
    if (storedChatID) {
      setChatID(storedChatID);
    }
  }, []);

  // Update localStorage and URL when chatID changes
  useEffect(() => {
    if (chatID) {
      localStorage.setItem('currentChatID', chatID);
      // Update URL params without causing a page reload
      const url = new URL(window.location.href);
      url.searchParams.set('chatID', chatID);
      window.history.replaceState({}, '', url.toString());
    }
  }, [chatID]);

  // Do not fetch message history directly here.
  // The backend streams responses and maintains history server-side.
  // Fetching /messages/{chatID} would fail CORS as that route lacks CORS headers.

  // Create transport that updates when chatID changes
  const transport = useMemo(() => {
    return new DefaultChatTransport({
      // Use the appropriate endpoint based on current chatID state
      api: chatID 
        ? `http://localhost:8080/chats/${chatID}/messages`
        : "http://localhost:8080/chats",
      prepareSendMessagesRequest: ({ messages }) => ({
        body: { message: messages[messages.length - 1] },
        headers: {
          'Content-Type': 'application/json',
          ...(chatID && { 'X-Chat-ID': chatID })
        }
      }),
      // Custom fetch to extract chatID from response headers
      fetch: async (url, options) => {
        console.log(`Making request to: ${url}`);
        console.log(`Current chatID: ${chatID}`);
        
        const response = await fetch(url, options);
        
        // Extract chatID from response headers if this was a new chat
        if (!chatID) {
          const responseChatID = response.headers.get('x-xtrn-chat-id');
          console.log(`Extracted chatID from response: ${responseChatID}`);
          if (responseChatID) {
            setChatID(responseChatID);
          }
        }
        
        return response;
      }
    });
  }, [chatID]); // Recreate transport when chatID changes

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
                <pre className="whitespace-pre-wrap break-words">{JSON.stringify(part.input, null, 2)}</pre>
              </div>
            );
          case "output-available":
            return (
              <div key={key}>
                <strong>⬅ {name} response</strong>
                <pre className="whitespace-pre-wrap break-words">{JSON.stringify(part.output, null, 2)}</pre>
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

      // Optional: show step boundaries if your backend emits them
      case "step-start":
        return <hr key={key} />;

      default:
        // Debug fallback so you can see any other parts you stream (sources, files, etc.)
        return (
          <pre key={key} className="whitespace-pre-wrap break-words text-xs opacity-70">
            {JSON.stringify(part, null, 2)}
          </pre>
        );
    }
  };

  const startNewChat = () => {
    setChatID(null);
    localStorage.removeItem('currentChatID');
    const url = new URL(window.location.href);
    url.searchParams.delete('chatID');
    window.history.replaceState({}, '', url.toString());
    // Optionally clear messages - note: this might require reinitialization
    window.location.reload();
  };

  const loadExistingChat = (existingChatID: string) => {
    setChatID(existingChatID);
  };

  return (
    <div style={{ maxWidth: '800px', margin: '0 auto', padding: '20px' }}>
      {/* Chat Header */}
      <div style={{ 
        marginBottom: '20px', 
        padding: '10px', 
        border: '1px solid #ddd', 
        borderRadius: '8px',
        backgroundColor: '#f9f9f9',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center'
      }}>
        <div>
          <strong>Chat ID:</strong> {chatID || 'New Chat'}
          <br />
          <small style={{ color: '#666' }}>
            API Endpoint: {chatID 
              ? `http://localhost:8080/chats/${chatID}/messages`
              : "http://localhost:8080/chats"
            }
          </small>
        </div>
        <div style={{ display: 'flex', gap: '10px', alignItems: 'center' }}>
          <input
            value={manualChatID}
            onChange={(e) => setManualChatID(e.target.value)}
            placeholder="Enter Chat ID"
            style={{
              padding: '6px 10px',
              border: '1px solid #ddd',
              borderRadius: '4px',
              fontSize: '14px'
            }}
          />
          <button 
            onClick={() => {
              if (manualChatID.trim()) {
                loadExistingChat(manualChatID.trim());
                setManualChatID("");
              }
            }}
            style={{
              padding: '6px 12px',
              backgroundColor: '#17a2b8',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
              fontSize: '14px'
            }}
          >
            Load Chat
          </button>
          <button 
            onClick={startNewChat}
            style={{
              padding: '6px 12px',
              backgroundColor: '#007bff',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
              fontSize: '14px'
            }}
          >
            New Chat
          </button>
        </div>
      </div>

      {/* Messages */}
      <div style={{ marginBottom: '20px' }}>
        {messages.map((m, mi) => (
          <div key={m.id ?? mi} style={{ marginBottom: 16 }}>
            <div style={{ fontWeight: 600, marginBottom: 4 }}>{m.role === "user" ? "User" : "AI"}:</div>
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

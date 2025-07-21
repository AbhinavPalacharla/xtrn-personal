FROM node:18-slim

RUN apt-get update && apt-get install -y --no-install-recommends curl ca-certificates && rm -rf /var/lib/apt/lists/*

CMD ["npx", "-y", "@openbnb/mcp-server-airbnb", "--ignore-robots-txt"]
FROM node:24-slim

RUN apt-get update && apt-get install -y --no-install-recommends zsh curl ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY dist/ ./dist/

CMD ["node", "dist/stdio.js"]
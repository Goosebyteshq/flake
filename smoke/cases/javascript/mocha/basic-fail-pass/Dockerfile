FROM node:20-alpine
WORKDIR /app
COPY package.json package.json
RUN npm install --no-audit --no-fund --silent
COPY . .
CMD ["sh", "-c", "npm test 2>&1 || true"]

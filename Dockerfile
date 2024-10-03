# Use the official Bun image as a base
FROM oven/bun:latest

# Set the working directory inside the container
WORKDIR /app

# Copy package files first for better caching
COPY bun.lockb ./
COPY package.json ./

# Install dependencies
RUN bun install --frozen-lockfile --production

# Copy the rest of the application files (non-ignored)
COPY . .

# Run the app
USER bun
EXPOSE 3000/tcp
ENTRYPOINT [ "bun", "run", "src/server.ts" ]
# Use an official Node.js runtime as a parent image
FROM node:20-slim

# Set the working directory in the container
WORKDIR /app

# Copy package.json first for better cache
COPY package*.json ./

# Install dependencies
RUN npm install --omit=dev

# Copy source code
COPY . .

# Use tini as init system to handle signals properly
RUN apt-get update && apt-get install -y tini
ENTRYPOINT ["/usr/bin/tini", "--"]

# The command will be provided by docker-compose
CMD ["npm", "start"]

# Add node_modules/.bin to PATH
ENV PATH /app/node_modules/.bin:$PATH

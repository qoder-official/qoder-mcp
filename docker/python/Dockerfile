# Use a minimal, secure base image for Python
FROM python:3.12-slim

# Set up a non-root user for security
RUN groupadd --gid 1001 python \
    && useradd --uid 1001 --gid 1001 --shell /bin/bash --create-home python

# Create and set the working directory
WORKDIR /usr/src/app

# Copy dependency files and install them
# This layer is cached to speed up builds when only source code changes.
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

# Copy the rest of the application source code
COPY . .

# Switch to the non-root user
USER python

# The command to run the application will be provided by the 'docker run' command
# using the 'entry' field from the mcp.yml manifest.
CMD [ "python", "main.py" ]

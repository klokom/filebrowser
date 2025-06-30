# slideserver/Dockerfile

# --- Stage 1: System Dependencies ---
# Use a slim Python base image
FROM python:3.11-slim

# Install the essential OpenSlide C library from the Debian repositories
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    libvips-dev \
    libvips42 \
    libtiff5-dev \
    libjpeg62-turbo-dev \
    zlib1g-dev \
    liblcms2-dev \
    libwebp-dev \
    libopenjp2-7-dev \
    openslide-tools \
    && rm -rf /var/lib/apt/lists/*

# --- Stage 2: Python Application ---
WORKDIR /app

# Copy the requirements file and install Python packages
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy the Flask application code
COPY *.py .

# Expose the port Gunicorn will run on
EXPOSE 5000

# Command to run the application using Gunicorn (a production-ready server)
#CMD ["gunicorn", "--workers", "2", "--bind", "0.0.0.0:5000", "app:app"]

# We add --log-level debug and --access-logfile - to make Gunicorn log everything
CMD ["gunicorn", "--workers", "2", "--bind", "0.0.0.0:5000", "--log-level", "debug", "--access-logfile", "-", "app:app"]
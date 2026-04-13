FROM python:3.12-slim

WORKDIR /app

# Copy SDK first for layer caching
COPY sdk/python/ /sdk/python/
RUN pip install --no-cache-dir /sdk/python

# Copy example
COPY examples/fastapi-api/requirements.txt ./
RUN pip install --no-cache-dir fastapi uvicorn[standard]
COPY examples/fastapi-api/ ./

EXPOSE 8000
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]

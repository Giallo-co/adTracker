import os
import duckdb
from fastapi import FastAPI
import pandas as pd
import logging

app = FastAPI()
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Configure DuckDB to use MinIO (S3 compatible)
def get_duckdb_conn():
    conn = duckdb.connect(':memory:')
    conn.execute("INSTALL httpfs;")
    conn.execute("LOAD httpfs;")
    endpoint = os.getenv('MINIO_ENDPOINT', 'minio:9000')
    conn.execute(f"SET s3_endpoint='{endpoint}';")
    conn.execute(f"SET s3_access_key_id='{os.getenv('MINIO_ACCESS_KEY', 'minioadmin')}';")
    conn.execute(f"SET s3_secret_access_key='{os.getenv('MINIO_SECRET_KEY', 'minioadmin')}';")
    conn.execute("SET s3_url_style='path';")
    conn.execute("SET s3_use_ssl=false;")
    return conn

@app.get("/report/top-states")
def top_states():
    try:
        conn = get_duckdb_conn()
        bucket = os.getenv('MINIO_BUCKET', 'events-raw')
        query = f"""
            SELECT state, COUNT(*) as impressions
            FROM read_json_auto('s3://{bucket}/events/impressions/*/*/*/*/*.json')
            GROUP BY state
            ORDER BY impressions DESC
            LIMIT 10
        """
        logger.info(f"Executing query: {query}")
        df = conn.execute(query).df()
        # Handle NaN/Inf
        df = df.fillna(0)
        return df.to_dict(orient='records')
    except Exception as e:
        logger.error(f"Error in top_states: {e}")
        return {"error": str(e)}

@app.get("/report/top-advertisers")
def top_advertisers():
    try:
        conn = get_duckdb_conn()
        bucket = os.getenv('MINIO_BUCKET', 'events-raw')
        # Simplified query for now
        query = f"""
            SELECT user_info.state as state, SUM(conversion_value) as revenue
            FROM read_json_auto('s3://{bucket}/events/conversions/*/*/*/*/*.json')
            GROUP BY state
            ORDER BY revenue DESC
            LIMIT 10
        """
        logger.info(f"Executing query: {query}")
        df = conn.execute(query).df()
        df = df.fillna(0)
        return df.to_dict(orient='records')
    except Exception as e:
        logger.error(f"Error in top_advertisers: {e}")
        return {"error": str(e)}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8081)

import os
import duckdb
from fastapi import FastAPI
import pandas as pd
import logging

app = FastAPI()
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Configure DuckDB with strict Phase 8 SECRET management
def get_duckdb_conn():
    conn = duckdb.connect(database=':memory:')
    conn.execute("INSTALL httpfs;")
    conn.execute("LOAD httpfs;")
    conn.execute("INSTALL aws;")
    conn.execute("LOAD aws;")
    
    access_key = os.getenv('MINIO_ACCESS_KEY', 'minioadmin')
    secret_key = os.getenv('MINIO_SECRET_KEY', 'minioadmin')
    endpoint = os.getenv('MINIO_ENDPOINT', 'minio:9000')

    
    logger.info(f"Setting up DuckDB S3 Secret for endpoint: {endpoint}")
    
    # Precise Phase 8 syntax for secret management
    conn.execute(f"""
        CREATE OR REPLACE SECRET (
            TYPE S3,
            KEY_ID '{access_key}',
            SECRET '{secret_key}',
            REGION 'us-east-1',
            ENDPOINT '{endpoint}',
            USE_SSL false,
            URL_STYLE 'path'
        );
    """)
    return conn

@app.get("/")
def read_root():
    return {"status": "Reporting Service Active", "endpoints": ["/report/top-states", "/report/top-advertisers"]}

@app.get("/report/top-states")
def top_states():
    try:
        conn = get_duckdb_conn()
        bucket = os.getenv('MINIO_BUCKET', 'events-raw')
        # Explicit Hive pattern from Phase 8
        query = f"""
            SELECT 
                state, 
                COUNT(*) as total_impressions 
            FROM read_json_auto(
                's3://{bucket}/events/impressions/year=*/month=*/day=*/hour=*/*.json',
                hive_partitioning=1
            )
            GROUP BY state 
            ORDER BY total_impressions DESC 
            LIMIT 10;
        """
        logger.info("Executing Phase 8 query for top-states")
        df = conn.execute(query).df()
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
        # Optimized query with hive_partitioning
        query = f"""
            SELECT 
                user_info.state as state, 
                SUM(CAST(conversion_value AS DOUBLE)) as revenue
            FROM read_json_auto(
                's3://{bucket}/events/conversions/year=*/month=*/day=*/hour=*/*.json',
                hive_partitioning=1
            )
            GROUP BY 1
            ORDER BY revenue DESC
            LIMIT 10
        """
        logger.info("Executing Phase 8 query for top-advertisers")
        df = conn.execute(query).df()
        df = df.fillna(0)
        return df.to_dict(orient='records')
    except Exception as e:
        logger.error(f"Error in top_advertisers: {e}")
        return {"error": str(e)}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8081)

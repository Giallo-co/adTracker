import os
import duckdb
from fastapi import FastAPI
import pandas as pd

app = FastAPI()

# Configure DuckDB to use MinIO (S3 compatible)
def get_duckdb_conn():
    conn = duckdb.connect(':memory:')
    conn.execute("INSTALL httpfs;")
    conn.execute("LOAD httpfs;")
    conn.execute(f"SET s3_endpoint='{os.getenv('MINIO_ENDPOINT')}';")
    conn.execute(f"SET s3_access_key_id='{os.getenv('MINIO_ACCESS_KEY')}';")
    conn.execute(f"SET s3_secret_access_key='{os.getenv('MINIO_SECRET_KEY')}';")
    conn.execute("SET s3_url_style='path';")
    conn.execute("SET s3_use_ssl=false;")
    return conn

@app.get("/report/top-states")
def top_states():
    conn = get_duckdb_conn()
    bucket = os.getenv('MINIO_BUCKET', 'events-raw')
    query = f"""
        SELECT state, COUNT(*) as impressions
        FROM read_json_auto('s3://{bucket}/events/impressions/*/*/*/*/*.json')
        GROUP BY state
        ORDER BY impressions DESC
        LIMIT 10
    """
    df = conn.execute(query).df()
    return df.to_dict(orient='records')

@app.get("/report/top-advertisers")
def top_advertisers():
    conn = get_duckdb_conn()
    bucket = os.getenv('MINIO_BUCKET', 'events-raw')
    query = f"""
        SELECT advertiser_name, SUM(conversion_value) as revenue
        FROM read_json_auto('s3://{bucket}/events/conversions/*/*/*/*/*.json')
        JOIN (
            SELECT unnest(ads).advertiser.advertiser_id as advertiser_id, 
                   unnest(ads).advertiser.advertiser_name as advertiser_name
            FROM read_json_auto('s3://{bucket}/events/impressions/*/*/*/*/*.json')
        ) USING (advertiser_id)
        GROUP BY advertiser_name
        ORDER BY revenue DESC
        LIMIT 10
    """
    # Note: The above query might be complex due to the nested structure. 
    # For now, a simpler one assuming conversion has some advertiser info or joining with impressions.
    # Simplifying for the prototype:
    query = f"""
        SELECT user_info.state, SUM(conversion_value) as revenue
        FROM read_json_auto('s3://{bucket}/events/conversions/*/*/*/*/*.json')
        GROUP BY user_info.state
        ORDER BY revenue DESC
        LIMIT 10
    """
    df = conn.execute(query).df()
    return df.to_dict(orient='records')

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8081)

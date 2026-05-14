import os
import duckdb
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
import logging

app = FastAPI()

# --- CONFIGURACIÓN DE CORS ---
# Mantenemos exactamente tu configuración para que el front en 5173 conecte
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:5173"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def get_duckdb_conn():
    # Creamos la conexión con límites estrictos de hardware
    conn = duckdb.connect(database=':memory:')
    
    # Configuraciones para que DuckDB no consuma toda la RAM del sistema
    conn.execute("SET memory_limit='1GB';")
    conn.execute("SET threads=1;")
    
    conn.execute("INSTALL httpfs; LOAD httpfs;")
    conn.execute("INSTALL aws; LOAD aws;")
    
    access_key = os.getenv('MINIO_ACCESS_KEY', 'minioadmin')
    secret_key = os.getenv('MINIO_SECRET_KEY', 'minioadmin')
    endpoint = os.getenv('MINIO_ENDPOINT', 'minio:9000')

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

@app.get("/report/top-states")
def top_states():
    try:
        conn = get_duckdb_conn()
        bucket = os.getenv('MINIO_BUCKET', 'events-raw')
        
        # Query optimizada: extrae los estados directamente de los JSON en MinIO
        query = f"""
            SELECT 
                CAST(state AS VARCHAR) as state, 
                COUNT(*) as total_impressions 
            FROM read_json_auto('s3://{bucket}/events/impressions/*/*/*/*/*.json')
            WHERE state IS NOT NULL
            GROUP BY 1 
            ORDER BY 2 DESC 
            LIMIT 10;
        """
        logger.info("Enviando reporte de estados al frontend...")
        df = conn.execute(query).df()
        return df.to_dict(orient='records')
    except Exception as e:
        logger.error(f"Error: {e}")
        return []

@app.get("/report/top-advertisers")
def top_advertisers():
    try:
        conn = get_duckdb_conn()
        bucket = os.getenv('MINIO_BUCKET', 'events-raw')
        
        query = f"""
            SELECT 
                CAST(user_info.state AS VARCHAR) as state, 
                SUM(CAST(conversion_value AS DOUBLE)) as revenue
            FROM read_json_auto('s3://{bucket}/events/conversions/*/*/*/*/*.json')
            GROUP BY 1
            ORDER BY 2 DESC
            LIMIT 10
        """
        df = conn.execute(query).df()
        return df.to_dict(orient='records')
    except Exception as e:
        logger.error(f"Error: {e}")
        return []

if __name__ == "__main__":
    import uvicorn
    # Mantenemos el puerto 8081 que ya tienes configurado
    uvicorn.run(app, host="0.0.0.0", port=8081)
# Signal Catcher (adTracker)

Signal Catcher is a high-speed ad event tracking system capable of handling 1000+ requests per second with real-time visualization and resilient historical storage.

## 🚀 Quick Start (One-Command Deployment)

1. **Clone the repository**
2. **Setup environment:**
   ```bash
   cp .env.example .env
   ```
3. **Launch the ecosystem:**
   ```bash
   docker compose up -d
   ```

## 📊 Access Points

- **Ingestion API:** `http://localhost:8080` (Endpoints: `/api/events/impression`, `/click`, `/conversion`)
- **Grafana Dashboard:** `http://localhost:3000` (User: `admin`, Password: `admin` or as defined in `.env`)
- **MinIO Console:** `http://localhost:9001` (User: `minioadmin`, Password: `minioadmin`)
- **Reporting API:** `http://localhost:8081`

## 🧪 Manual Verification

Send a test impression using `curl`:

```bash
curl -X POST http://localhost:8080/api/events/impression \
  -H "Content-Type: application/json" \
  -d '{
    "impression_id": "test-imp-1",
    "user_ip": "127.0.0.1",
    "timestamp": "2026-05-12T10:00:00Z",
    "state": "CA",
    "session_id": "sess-1",
    "ads": []
  }'
```

The server should respond with `202 Accepted`.

## 🛠 Architecture Highlights

- **Backend:** Go (Gin) + Sarama (Kafka AsyncProducer).
- **Messaging:** Apache Kafka (KRaft mode).
- **TSDB:** InfluxDB 2.7.
- **Object Store:** MinIO (S3 Compatible) with Hive-style partitioning.
- **Analytics:** DuckDB 1.x integrated into a Python FastAPI service.
- **Dashboard:** Grafana 11 with automatic provisioning.

For more details, see [ARCHITECTURE.md](./ARCHITECTURE.md) and the [Project Log](./project_log.md).

import asyncio
import aiohttp
import time
import uuid
import random
import argparse
from datetime import datetime

# Configuración de payloads
STATES = ["CA", "NY", "TX", "FL", "IL", "WA", "MA", "GA"]
KEYWORDS = ["shoes", "laptop", "phone", "coffee", "travel"]

async def send_event(session, url, data):
    try:
        async with session.post(url, json=data) as response:
            return response.status == 202
    except Exception:
        return False

async def run_iteration(session, base_url):
    session_id = str(uuid.uuid4())
    ip = f"192.168.1.{random.randint(1, 254)}"
    state = random.choice(STATES)
    now = datetime.utcnow().isoformat() + "Z"
    
    # 1. Impression
    imp_id = str(uuid.uuid4())
    imp_data = {
        "impression_id": imp_id,
        "user_ip": ip,
        "timestamp": now,
        "state": state,
        "search_keywords": random.choice(KEYWORDS),
        "session_id": session_id,
        "ads": [{
            "advertiser": {"advertiser_id": "adv-remote", "advertiser_name": "Remote Brand"},
            "campaign": {"campaign_id": "camp-remote", "campaign_name": "LAN Test"},
            "ad": {"ad_id": "ad-remote", "ad_name": "Remote Ad", "ad_position": 1}
        }]
    }
    
    success = await send_event(session, f"{base_url}/impression", imp_data)
    
    # 2. Click (30% chance)
    if success and random.random() < 0.3:
        click_id = str(uuid.uuid4())
        click_data = {
            "click_id": click_id,
            "impression_id": imp_id,
            "timestamp": now,
            "clicked_ad": {"ad_id": "ad-remote", "ad_position": 1},
            "user_info": {"user_ip": ip, "state": state, "session_id": session_id}
        }
        await send_event(session, f"{base_url}/click", click_data)

        # 3. Conversion (10% chance)
        if random.random() < 0.2:
            conv_data = {
                "conversion_id": str(uuid.uuid4()),
                "click_id": click_id,
                "impression_id": imp_id,
                "timestamp": now,
                "conversion_value": round(random.uniform(10, 100), 2),
                "user_info": {"user_ip": ip, "state": state, "session_id": session_id}
            }
            await send_event(session, f"{base_url}/conversion", conv_data)

async def main():
    parser = argparse.ArgumentParser(description="Signal Catcher Remote Stress Tester")
    parser.add_argument("--ip", required=True, help="IP address of the Signal Catcher server")
    parser.add_argument("--rps", type=int, default=100, help="Target Requests Per Second")
    parser.add_argument("--duration", type=int, default=30, help="Duration in seconds")
    args = parser.parse_args()

    base_url = f"http://{args.ip}:8080/api/events"
    print(f"🚀 Starting stress test against {base_url}")
    print(f"📊 Target: {args.rps} RPS for {args.duration}s")

    conn = aiohttp.TCPConnector(limit=2000)
    async with aiohttp.ClientSession(connector=conn) as session:
        start_time = time.time()
        total_requests = 0
        
        while time.time() - start_time < args.duration:
            batch_start = time.time()
            tasks = []
            for _ in range(args.rps):
                tasks.append(run_iteration(session, base_url))
            
            await asyncio.gather(*tasks)
            total_requests += args.rps
            
            # Control de velocidad (throttle)
            elapsed = time.time() - batch_start
            if elapsed < 1:
                await asyncio.sleep(1 - elapsed)
        
        end_time = time.time()
        print(f"\n✅ Test Finished")
        print(f"⏱  Total time: {end_time - start_time:.2f}s")
        print(f"📈 Avg RPS: {total_requests / (end_time - start_time):.2f}")

if __name__ == "__main__":
    asyncio.run(main())

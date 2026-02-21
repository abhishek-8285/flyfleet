from fastapi import FastAPI
from contextlib import asynccontextmanager
from apscheduler.schedulers.background import BackgroundScheduler
from app.core.config import settings
from app.api.v1.api import api_router
from app.services.messaging import nats_client
from app.services.etl_service import ETLService
import logging

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def run_scheduled_training():
    logger.info("Executing scheduled ETL and ML training job...")
    service = ETLService()
    try:
        service.run_training_pipeline()
    except Exception as e:
        logger.error(f"Scheduled training failed: {e}")

scheduler = BackgroundScheduler()
scheduler.add_job(run_scheduled_training, 'cron', hour=0)  # Run daily at midnight

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup
    scheduler.start()
    await nats_client.connect()
    yield
    # Shutdown
    scheduler.shutdown()
    await nats_client.close()

app = FastAPI(
    title=settings.PROJECT_NAME,
    openapi_url=f"{settings.API_V1_STR}/openapi.json",
    lifespan=lifespan
)

app.include_router(api_router, prefix=settings.API_V1_STR)

@app.get("/health")
def health_check():
    return {"status": "ok", "service": settings.PROJECT_NAME}

from fastapi import APIRouter
from app.services.etl_service import ETLService

router = APIRouter()

@router.post("/train")
def trigger_training():
    service = ETLService()
    result = service.run_training_pipeline()
    return result

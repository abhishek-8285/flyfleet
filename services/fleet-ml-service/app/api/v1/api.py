from app.api.v1.endpoints import demand, pricing, jobs

api_router = APIRouter()
api_router.include_router(demand.router, prefix="/demand", tags=["demand"])
api_router.include_router(pricing.router, prefix="/pricing", tags=["pricing"])
api_router.include_router(jobs.router, prefix="/jobs", tags=["jobs"])

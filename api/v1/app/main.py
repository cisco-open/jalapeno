from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from .config.settings import Settings
from .routes import graphs, instances, collections, vpns, rpo

app = FastAPI(
    title="Network UI API",
    description="API for network topology visualization",
    version="1.0.0"
)

# CORS configuration
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Configure this appropriately for production
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Load settings
settings = Settings()

# Include routers
app.include_router(instances.router, prefix="/api/v1", tags=["instances"])
app.include_router(graphs.router, prefix="/api/v1", tags=["graphs"])
app.include_router(collections.router, prefix="/api/v1", tags=["collections"])
app.include_router(vpns.router, prefix="/api/v1", tags=["vpns"])
app.include_router(rpo.router, prefix="/api/v1", tags=["rpo"])

@app.get("/health")
async def health_check():
    return {
        "status": "healthy",
        "database_server": settings.database_server,
        "database_name": settings.database_name
    } 


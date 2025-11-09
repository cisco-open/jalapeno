from fastapi import APIRouter, HTTPException
from arango import ArangoClient
from ..config.settings import Settings

router = APIRouter()
settings = Settings()

def get_db():
    client = ArangoClient(hosts=settings.database_server)
    try:
        db = client.db(
            settings.database_name,
            username=settings.username,
            password=settings.password
        )
        return db
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"Could not connect to database: {str(e)}"
        )

@router.get("/instances")
async def get_instances():
    try:
        db = get_db()
        # Get list of collections that are graphs
        collections = [c['name'] for c in db.collections() 
                      if not c['name'].startswith('_') 
                      and c['type'] == 'edge']
        return collections
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=str(e)
        ) 
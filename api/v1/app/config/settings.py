from pydantic_settings import BaseSettings
import json
import os

class Settings(BaseSettings):
    # Default values for k8s deployment
    database_server: str = "http://arangodb:8529"
    database_name: str = "jalapeno"
    credentials_path: str = "/credentials/auth"
    username: str = "root"
    password: str = "jalapeno"

    # Environment-based configuration
    # export LOCAL_DEV=1
    def __init__(self, **kwargs):
        super().__init__(**kwargs)
        if os.getenv("LOCAL_DEV"):
            self.database_server = "http://198.18.133.112:30852"
            self.credentials_path = None

    class Config:
        env_prefix = "JALAPENO_"

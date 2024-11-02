import os

from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    team_id: str = "a698e896-231a-4299-94e6-a000a634f6c6"
    backend_api: str = "https://aes-agniachallenge-case.olymp.innopolis.university/"
    root_directory: str = os.path.dirname(__file__)


settings = Settings()

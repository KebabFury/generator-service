import os
import re
import requests
from typing import Dict
import importlib.util
import shutil
import zipfile

from team_actions.src.settings import settings
from team_actions.src.systems_config import systems_info, available_actions


def collect_systems_documentation() -> Dict[str, str]:
    actions_directory: str = os.path.join(settings.root_directory, "actions")
    systems_documentation: Dict[str, str] = {}
    uppercase_dir_pattern = re.compile(r"^[A-Z]")

    for dirpath, dirnames, filenames in os.walk(actions_directory):
        current_dir_name: str = os.path.basename(dirpath)
        if uppercase_dir_pattern.match(current_dir_name):
            action_file_path: str = os.path.join(dirpath, "actions.py")
            if "actions.py" in filenames:
                spec = importlib.util.spec_from_file_location(
                    "actions", action_file_path
                )
                module = importlib.util.module_from_spec(spec)
                spec.loader.exec_module(module)

            doc_path: str = os.path.join(dirpath, "documentation.md")
            if os.path.exists(doc_path):
                with open(doc_path, "r") as doc_file:
                    system_documentation: str = doc_file.read()
                systems_documentation[current_dir_name] = system_documentation

    registration_data: Dict[str, str] = {
        "team_id": settings.team_id,
        "actions_info": available_actions,
        "systems_info": systems_info,
        "actions_doc": systems_documentation,
    }

    # Save registred systems and actions
    response = requests.post(
        url=f"{settings.backend_api}/submit-documentation", json=registration_data
    )

    if response.status_code == 200:
        print("Successfully registered systems and actions")
    else:
        print(f"Failed to register systems and actions: {response.status_code}")

    return systems_documentation


def check_user_settings() -> None:
    assert settings.team_id != "", "Укажите ID команды в settings.py"


def create_zip_archive():
    root = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
    archive_path = os.path.join(root, "team_actions_archive")
    shutil.make_archive(archive_path, "zip", root, "team_actions")
    return f"{archive_path}.zip"


def upload_zip(zip_file_path: str):
    files = {}
    try:
        with zipfile.ZipFile(zip_file_path, "r") as zip_file:
            zip_contents = zip_file.namelist()
            if "team_actions/requirements.txt" not in zip_contents:
                print("Warning: requirements.txt not found. ")
            else:
                with zip_file.open("team_actions/requirements.txt") as req_file:
                    files["requirements.txt"] = req_file.read()

        files["code_zip"] = open(zip_file_path, "rb")
        response = requests.post(
            url=f"{settings.backend_api}/upload-actions-code",
            files=files,
            data={"team_id": settings.team_id},
        )

        if response.status_code == 200:
            print(f"Response from server: {response.status_code} - {response.text}")
        else:
            print(
                f"Failed to upload the zip file : {response.status_code}- {response.text}"
            )
    finally:
        if "code_zip" in files and not files["code_zip"].closed:
            files["code_zip"].close()


def main() -> None:
    check_user_settings()
    collect_systems_documentation()
    zip_file_path = create_zip_archive()
    upload_zip(zip_file_path)


if __name__ == "__main__":
    main()

# python -m team_actions.src.initial_setup

import os
import requests
from functools import wraps
from typing import Any, Callable, Optional, Dict, List

from team_actions.src.settings import settings
from team_actions.src.utils.exceptions import TeamHackathonException
from team_actions.src.utils.action_router import ActionRouter


def fetch_available_actions() -> Dict[str, Any]:
    """Fetches systems configuration data from a specified endpoint."""
    try:
        response = requests.get(
            f"{settings.backend_api}/actions/base-available-actions"
        )
        response.raise_for_status()  # Raises an error for HTTP codes 4xx/5xx
        return response.json()
    except requests.RequestException as e:
        print(f"Error fetching systems configuration: {e}")


def infer_system_name(func: Callable[..., Any]) -> str:
    """Infer system_name from the function's file location."""
    # Get the full path of the module where the function is defined
    file_path = func.__globals__.get("__file__")

    if file_path is None:
        raise Exception(
            f"Cannot infer system_name: no __file__ attribute for {func.__name__}"
        )

    # Extract the parent directory (system name) from the path
    system_name = os.path.basename(os.path.dirname(file_path))
    return system_name


def register_action(
    system_type: str,
    include_in_plan: bool,
    signature: str,
    arguments: List[str],
    description: str,
):
    """Decorator for logging the operation of a component."""

    def decorator(func: Callable[..., Any]):

        from team_actions.src.systems_config import available_actions, systems_info

        system_name: str = infer_system_name(func)

        try:
            action_name = func.__name__
        except AttributeError:
            raise Exception(
                f"Seems function '{func}' doesn't have a name, so can't be registered"
            )
        try:
            system_specific_description: str = systems_info[system_type]["systems"][
                system_name
            ]
        except KeyError:
            print(f"System {system_name} not found in systems_info")
            system_specific_description: str = ""

        action_info: Dict[str, Any] = {
            system_type: {
                system_name: {
                    "description": system_specific_description,
                    "actions": {
                        action_name: {
                            "is_general": include_in_plan,
                            "signature": signature,
                            "arguments": arguments,
                            "description": description,
                        }
                    },
                }
            }
        }

        system_types: Optional[Dict[str, Any]] = available_actions.get(
            system_type, None
        )

        if system_types is None:
            available_actions.update(action_info)
        else:
            system_names: Optional[Dict[str, Any]] = system_types.get(system_name, None)
            if system_names is None:
                available_actions[system_type].update(action_info[system_type])
            else:
                available_actions[system_type][system_name]["actions"].update(
                    action_info[system_type][system_name]["actions"]
                )

        @wraps(func)
        def wrapper(*args: Any, **kwargs: Any):
            try:
                return func(*args, **kwargs)
            except Exception as e:
                raise TeamHackathonException(
                    f"Error in action for system '{system_name}' with action '{action_name}'"
                ) from e

        ActionRouter.register_new_action_function(
            system_name=system_name, function_name=action_name, action_func=wrapper
        )
        return wrapper

    return decorator

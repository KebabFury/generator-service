from typing import Callable
from types import ModuleType


class ActionRouter:
    systems_to_functions_mapping: dict[str, dict[str, Callable]] = dict()

    @classmethod
    def register_new_action_function(
        cls, system_name: str, function_name: str, action_func: Callable
    ) -> None:
        if system_name not in cls.systems_to_functions_mapping:
            cls.systems_to_functions_mapping[system_name] = dict()
        cls.systems_to_functions_mapping[system_name][function_name] = action_func

    @classmethod
    def add_actions_for_module(cls, action_module: ModuleType) -> None:
        """
        Needed so the action functions from action module are added to router.
        Actual logic happens in @register_action decorator, added to action functions.
        Without adding action module to ActionRouter using this method, action module is never imported,
        so decorators for action functions are not called, and action functions are not registered in ActionRouter
        """
        pass

    def get_action_functions_by_system_name(
        self, system_name: str
    ) -> dict[str, Callable] | None:
        return self.systems_to_functions_mapping.get(system_name, {})

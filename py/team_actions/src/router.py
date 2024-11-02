from team_actions.src.utils.action_router import ActionRouter

# Import actions module for new system here
from team_actions.src.actions.Todoist import actions as todoist_actions

# Add your module to the router
ActionRouter.add_actions_for_module(todoist_actions)

# Keep it as is
action_router = ActionRouter()
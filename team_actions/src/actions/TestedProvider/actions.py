
import requests


from team_actions.src.registration import register_action


from typing import *
from pydantic import *


authorization_data = {}
# Держите это поле пустым изначально.
# После регистрации действий в системе, сюда будут автоматически
# добавлены авторизационные данные участников.

Id = Annotated[str, Field(pattern="^[0-9]+$")]
TaskName = Annotated[str, Field(description="Task name")]
Datetime = Annotated[
    str, Field(pattern=r"^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?Z$")
]
ShortDatetime = Annotated[str, Field(pattern=r"^\d{4}-\d{2}-\d{2}$")]
DueString = Annotated[str, Field(description="Due date in English", example="tomorrow")]
DueLang = Annotated[str, Field(pattern="^[a-z]{2}$", default="en")]
DurationUnit = Literal["minute", "day"]
Priority = Annotated[int, Field(ge=1, le=4)]




class Due(BaseModel):
    string: Optional[DueString]
    date: Optional[ShortDatetime]
    is_recurring: bool
    datetime: Optional[Datetime]
    timezone: Optional[str]


class Duration(BaseModel):
    amount: Optional[int]
    unit: Optional[DurationUnit]


class Task(BaseModel):
    id: Id
    assigner_id: Optional[Id]
    assignee_id: Optional[Id]
    project_id: Id
    section_id: Optional[Id]
    parent_id: Optional[Id]
    order: int
    content: TaskName
    description: Optional[str]
    is_completed: bool
    labels: Optional[List[str]]
    priority: Priority
    comment_count: int
    creator_id: Id
    created_at: Datetime
    due: Optional[Due]
    url: HttpUrl
    duration: Optional[Duration]





@register_action(
    system_type="task_tracker",
    include_in_plan=True,  # Действие может быть использовано в плане
    signature="( content: TaskName = None, ) -> Task",
    arguments=[
       "content",
       
    ],
    description="Creates a new task in the system with optional parameters like descriptions, project and section IDs, and duration.",
)
def create_task(
    content: TaskName = None,
    
) -> Task:
    # Логика вызова API Todoist для создания задачи
    response = requests.post(
        "https://our_back/provider/create_task",
        headers={"Authorization": f"Bearer {authorization_data['Todoist']}"},
        json={
            "content": content,
            
        },
    )
    response.raise_for_status()
    data = response.json()
    return data



    
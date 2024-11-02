# TypeHints Definition
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


# Models definition
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


# API Calls Documentation
## `create_task`

**Description**:
Creates a new task in the system with optional parameters like descriptions, project and section IDs, and duration.

**Parameters**:
- content (TaskName): The name of the task.

**Returns**:
- created_task (Task)
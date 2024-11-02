# TypeHints Definition
Id = Annotated[str, Field(description="Any ID represented as a string.")]
TaskName = Annotated[str, Field(description="A well-crafted task name.")]
MongoID = Annotated[str, Field(regex="^[0-9a-f]{24}$")]
PriorityChoice = Annotated[str, Field(regex="^(low|middle|high)$")]
Datetime = Annotated[
    str,
    Field(
        regex=r"^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$",
        description="ISO 8601 datetime format with milliseconds, e.g., '2024-10-02T10:42:16.822Z'",
    ),
]
StatusValue = Annotated[str, Field(default="empty", regex="^(empty|bug|epic)$")]


# Models definition
class Entity(BaseModel):
    name: str
    id: Id


class Creator(BaseModel):
    id: MongoID
    email: EmailStr
    name: str


class Space(BaseModel):
    id: MongoID
    name: str
    users: List[MongoID]
    owner: MongoID
    description: str
    createdAt: Datetime
    updatedAt: Datetime


class Task(BaseModel):
    id: MongoID
    name: TaskName
    creator: Creator
    description: Optional[str] = ""
    task_number: str
    status: str
    end_date: Optional[Datetime]
    created_at: Datetime
    updated_at: Datetime


class User(BaseModel):
    id: MongoID
    email: str
    name: str


class Board(BaseModel):
    id: MongoID
    name: str
    users: List[MongoID]
    owner: MongoID
    columns: List[MongoID]
    spaceId: MongoID
    projectId: MongoID
    createdAt: Datetime
    updatedAt: Datetime


class Project(BaseModel):
    id: MongoID
    name: str
    description: str
    owner: User
    spaceId: MongoID
    boards: List[Board]
    users: List[MongoID]
    createdAt: Datetime
    updatedAt: Datetime


class Column(BaseModel):
    id: MongoID
    name: str


class DeleteResult(BaseModel):
    result: bool


class SpaceTasks:
    result: List[Task]


class Spaces:
    result: List[Space]


class SpaceBoard:
    id: MongoID
    name: str
    users: List[MongoID]
    owner: MongoID
    columns: List[Column]
    spaceId: MongoID
    projectId: MongoID
    createdAt: Datetime
    updatedAt: Datetime


class Boards(BaseModel):
    result: List[Board]


class SpaceBoards(BaseModel):
    result: List[SpaceBoard]


class SpaceProject(BaseModel):
    id: MongoID
    name: str
    description: str
    owner: User
    spaceId: MongoID
    boards: List[SpaceBoard]
    users: List[MongoID]
    createdAt: Datetime
    updatedAt: Datetime


class SpaceProjects(BaseModel):
    result: List[SpaceProject]


class Users(BaseModel):
    result: List[User]


class BoardColumn(BaseModel):
    id: MongoID
    name: str
    tasks: List[Task]


class BoardColumns(BaseModel):
    result: List[Column]


# API Calls Documentation

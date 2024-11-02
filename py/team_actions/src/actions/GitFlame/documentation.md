# TypeHints Definition
UserName = Annotated[str, Field(description="Exact name of a Gitflame user")]
RepoName = Annotated[str, Field(description="Exact repository name")]
RepoFullName = Annotated[str, Field(description="The full name of the repository in the format 'owner/repo_name'")]
Datetime = Annotated[
    str,
    Field(
        regex=r"^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$",
        description="ISO 8601 UTC datetime format without milliseconds, e.g., '2024-02-14T07:22:56Z'",
    ),
]
IssueState = Annotated[Literal["open", "closed"]]
ContentBody = Annotated[str, Field(description="The body content of the issue, can include HTML or code.")]
ContentType = Annotated[Literal["text", "code"]]


# Models Definition
class IssueRepo(BaseModel):
    id: int
    name: RepoName
    owner: UserName
    full_name: RepoFullName

class UserInfo(BaseModel):
    id: int
    login: UserName
    full_name: str
    email: EmailStr

class Dependency(BaseModel):
    id: int
    number: int
    url: HttpUrl
    html_url: HttpUrl
    title: str
    state: IssueState

class RepoInfo(BaseModel):
    id: int
    owner: UserInfo
    name: str
    full_name: str
    description: Optional[str]
    private: bool
    language: Optional[str]
    html_url: HttpUrl
    clone_url: HttpUrl
    stars_count: int
    forks_count: int
    open_issues_count: int
    default_branch: str
    archived: bool
    created_at: Datetime
    updated_at: Datetime
    has_issues: bool
    has_wiki: bool
    has_pull_requests: bool
    has_projects: bool
    
class Issue(BaseModel):
    id: int
    url: HttpUrl
    html_url: HttpUrl
    number: int
    user: UserInfo
    original_author: Optional[str]
    original_author_id: Optional[int]
    title: str
    body: List[Content]
    ref: Optional[str]
    labels: List[str]
    milestone: Optional[str]
    assignees: Optional[List[UserInfo]] 
    dependencies: Optional[List[Dependency]]
    state: IssueState
    is_locked: bool
    comments: int
    created_at: Datetime
    updated_at: Datetime
    closed_at: Optional[Datetime]
    due_date: Optional[Datetime]
    repository: IssueRepo

class GetRepoIssuesResponse(BaseModel):
    open_issue_count: int
    close_issue_count: int
    list: List[Issue]

class GetRepoIssues(BaseModel):
    issues: GetRepoIssuesResponse


# API Calls Documentation
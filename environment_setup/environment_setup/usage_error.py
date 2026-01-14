import typing as t

class UsageError(RuntimeError):
    def __init__(self, problem: str, subproblem: t.Optional[str] = None):
        self.problem = f"{problem}"
        if subproblem:
            self.problem += f": {subproblem}"
        super().__init__(self.problem, subproblem)
        

class UsageErrorList(RuntimeError):
    def __init__(self, header: str, errors: t.Iterable[UsageError]) -> None:
        self.header = header
        self.errors = list(errors)
        assert self.errors
        super().__init__()
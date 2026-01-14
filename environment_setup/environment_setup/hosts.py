from dataclasses import dataclass

from .usage_error import UsageError

@dataclass
class HostDeclaration:
    name: str
    description: str


def host_declaration_from_line(line: str) -> HostDeclaration:
    try:
        name, description = map(str.strip, line.split("|"))
    except ValueError:
        raise UsageError(f"each line must have two values, separated by |, like HOST_NAME|A description")

    
    return HostDeclaration(
            name=name,
            description=description
        )
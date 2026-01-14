"""Recursively searches a root path and validates environment variable declarations.

Different actions may be specified with flags;
however, if no actions are specified, then validation alone is performed.
"""

import argparse
import enum
from pathlib import Path
import sys
import typing as t

from .environment import raise_if_bad_environment_variable_declarations
from .usage_error import UsageError, UsageErrorList
from .variable_declaration import EnvironmentVariableDeclaration, get_environment_variable_declarations_from_files

def main():
    try:
        cli()
    except UsageError as e:
        print(f"Usage Error: {e.problem}", file=sys.stderr)
        sys.exit(1)
    except UsageErrorList as e:
        print(e.header + "\n" + "\n".join(map(lambda x: f"{'':>4}{x}", map(lambda y: y.problem, e.errors))))
        sys.exit(1)

def cli():
    parser = argparse.ArgumentParser(
        description=__doc__
    )

    parser.add_argument("root", default=".", help="the root path, which will be searched recursively for environment variable and host files.")
    parser.add_argument("--list", "-l", action="store_true", help="if set, lists all variables for which there are declarations")
    parser.add_argument("--validate", "-v", action="store_true", help="if set, validates that all environment variables are well defined; leads an an undefined or improperly defined environment variable to cause the progra to exit with a non-zero exit status.")
    parser.add_argument("--env-file-name", default="env_vars.txt", help="the name of the files that contain environment variable declarations.")

    args = parser.parse_args()
    root = Path(args.root)
    env_file_name = str(args.env_file_name)
    actions_to_perform = actions if (actions := get_actions_to_perform(
        list_declarations=args.list,
        validate_environment=args.validate)) else {Action.ValidateDeclarations}

    environment_files = list(filter(lambda f: not f.is_dir(), root.glob(f"**/{env_file_name}")))


    try:
        var_declarations = get_environment_variable_declarations_from_files(environment_files)
    except UsageError as e:
        raise UsageError("parsing environment variable declarations", e.problem)
    
    if Action.ListVars in actions_to_perform:
        print(make_table(variable_declarations_to_columns(list(map(lambda x: x.data, var_declarations)))))
    
    if Action.ValidateDeclarations in actions_to_perform:
        raise_if_bad_environment_variable_declarations(var_declarations)
        print("All environment variables are defined and passed type validation.")


class Action(enum.Enum):
    ListVars = 1
    ValidateDeclarations = 2

def get_actions_to_perform(*, list_declarations: bool, validate_environment: bool) -> set[Action]:
    def not_none(v: Action | None) -> t.TypeGuard[Action]:
        return v is not None
    return set[Action](filter(not_none, [
        Action.ListVars if list_declarations else None,
        Action.ValidateDeclarations if validate_environment else None,
    ]))


def variable_declarations_to_columns(variable_declarations: t.Sequence[EnvironmentVariableDeclaration]) -> dict[str, list[str]]:
    return {
        "VAR NAME": [vd.name for vd in variable_declarations],
        "TYPE": [vd.var_type.__name__ for vd in variable_declarations],
        "DESCRIPTION": [vd.description for vd in variable_declarations]
    }

def make_table(columns: dict[str, list[str]], column_spacing: int = 4) -> str:

    assert len(set([len(rows) for rows in columns.values()])) == 1

    num_rows = len(next(v for v in columns.values()))

    if len(columns) == 0:
        return ""
    column_lengths = {
        column: max(map(len, [column] + rows)) for column, rows in columns.items()
        }
    column_length_with_spaces = {
        column: length + column_spacing for column, length in column_lengths.items()
    }

    table = (
        make_spaced_row([(column, length) for column, length in column_length_with_spaces.items()])
        + "\n"
        + "\n".join(make_spaced_row(
            [(columns[column][row_index], column_length_with_spaces[column]) for column in columns]
        ) for row_index in range(num_rows)))
    return table

            

def make_spaced_row(row: t.Sequence[tuple[str, int]]):
    
    str_row = ""
    for column, length_with_spacing in row:
        str_row += f"{column:<{length_with_spacing}}"
    return str_row




if __name__ == "__main__":

    try:
        main()
    except UsageError as e:
        print(f"Usage Error: {e.problem}", file=sys.stderr)
        sys.exit(1)

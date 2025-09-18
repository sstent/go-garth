## Brief overview
Ensures Architect mode focuses on planning and design without making direct code edits. Guides collaboration between Architect and Code modes.

## Mode-specific Instructions
- Architect mode must only create implementation plans and documentation
- All code/file changes must be delegated to Code mode via switch_mode
- Use update_todo_list for task tracking instead of direct implementation

## Development Workflow
- Generate detailed technical specifications in Architect mode
- Use read_file and list_code_definition_names for context gathering
- Create comprehensive todo lists with update_todo_list
- Transition to Code mode via switch_mode for implementation

## Collaboration Guidelines
- Architect and Code modes should maintain separate responsibilities
- Architectural plans must include:
  - File structure diagrams
  - API specifications
  - Data flow documentation
- Code mode implementations must reference architectural docs

## Enforcement Rules
- Block apply_diff/write_to_file usage in Architect mode
- Require switch_mode to Code for any code modifications
- Automatic todo list validation against architectural specs
## Brief overview
Git workflow rules for projects with Git repositories, ensuring secure handling of sensitive data and proper version control practices.

## Git repository setup
- Ensure `.gitignore` exists in the project root when the directory is a Git repository
- Verify `.gitignore` includes patterns for sensitive files and data directories
- Check that `.git` directory exists to confirm Git repository status

## Sensitive file protection
- Always include `.env` files in `.gitignore` to prevent password/token exposure
- Ensure `data/` directories are ignored to prevent data file commits
- Add patterns for database files: `*.db`, `*.sqlite`, `*.sqlite3`
- Include common sensitive file patterns:
  - `.env.local`
  - `.env.production`
  - `config/secrets.json`
  - `*.key`
  - `*.pem`

## Pre-change workflow
- Before starting any new changes, ensure current state is committed
- Run `git status` to check for uncommitted changes
- Commit all pending changes with descriptive messages
- Push commits to remote repository before starting new work
- Use `git add . && git commit -m "WIP: save state before [feature/change]"` for work-in-progress saves

## Post-completion workflow
- When a task is finished, commit all changes with meaningful commit messages
- Use conventional commit format: `type(scope): description`
- Examples:
  - `feat(auth): add Garmin Connect integration`
  - `fix(data): handle missing activity files`
  - `docs(readme): update setup instructions`
- Push commits immediately after committing: `git push origin [branch-name]`

## Branch management
- Create feature branches for new work: `git checkout -b feature/description`
- Keep main/master branch clean and deployable
- Use descriptive branch names that reflect the task
- Delete merged branches to keep repository clean

## Verification steps
- After each commit, verify with `git log --oneline -5`
- Before pushing, run `git status` to confirm clean working directory
- Check remote repository to ensure pushes were successful
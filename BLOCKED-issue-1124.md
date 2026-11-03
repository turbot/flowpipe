## Kind: environment

# #1124 add-to-project / add-to-project is red on develop

## What I was doing
Looking for a fix in turbot/flowpipe for the failing add-to-project check.

## Finding
No code or workflow change in this repo can make it pass.

- .github/workflows/31-add-issues-to-pipeling-issue-tracker.yaml only calls the reusable workflow
  turbot/steampipe-workflows/.github/workflows/assign-issue-to-pipeling-issue-tracker.yml@main
  (inputs: issue number, repository) with secrets: inherit.
- That workflow's step is actions/add-to-project with project-url https://github.com/orgs/turbot/projects/31
  and the token from the secret GH_PROJECT_PAT. The failure is
  "Could not resolve to a ProjectV2 with the number 31." That GraphQL answer means the PAT used
  cannot see org project 31 (expired, revoked, lost org or project access, or the project was deleted or renumbered).
- The "Unexpected input(s) 'fields'" message is only a warning, and the called workflow
  has had that line since 2025-08; it is not the failure. The job's GITHUB_TOKEN permissions are not used
  (the PAT is).
- The called workflow last changed 2026-03-24 (before the 2026-03-31 success). Runs on this repo
  succeeded until 2026-03-31 and failed on 2026-05-18, 2026-06-29 and twice on 2026-11-03,
  with no change to either workflow in between. So the cause is the secret or the project, not the code.
- Changing the call here would need a .github/workflows/ change (parked, never pushed by a run)
  and would only hide the failure.

## Next step (a person with org admin)
Check that the secret GH_PROJECT_PAT is valid, has project scope and access to
turbot org project 31 (pipeling issue tracker), and that the project still exists at
https://github.com/orgs/turbot/projects/31. Rotate the PAT or fix the project URL in
turbot/steampipe-workflows, then re-run the failed run.

import type { HarnessStatus, Project } from './types'

// missingReason is why nothing can start in a project whose checkout cannot
// be worked in, as the service words its refusal, or null when it can.
export function missingReason(project: Project | undefined): string | null {
  if (!project?.checkout_missing) {
    return null
  }
  return `The checkout is missing: '${project.checkout}' (${project.checkout_missing})`
}

// claudeCannotRun is why Claude Code cannot run in a project, or null: what
// the project's latest resolution found missing, or before one, what mega's
// start-up check found in the operator's shell environment.
export function claudeCannotRun(project: Project | undefined, harnesses: HarnessStatus[] | undefined): string | null {
  const env = project?.environment
  if (env?.resolved) {
    return env.missing.find((m) => m.harness === 'claude')?.reason ?? null
  }
  const status = harnesses?.find((h) => h.name === 'claude')
  return status && !status.found ? (status.reason ?? 'Claude Code cannot run') : null
}

// environmentReason is why nothing can start in a project whose environment
// may not run (its .mega/environment.yaml unapproved, changed or malformed),
// as the service words its refusal, or null when it may.
export function environmentReason(project: Project | undefined): string | null {
  return project?.environment?.refusal?.message ?? null
}

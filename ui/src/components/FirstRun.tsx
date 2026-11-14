import { harnesses } from '../harnesses'
import type { HarnessStatus } from '../types'

// FirstRun is the empty workspace's page: what mega is for, the control to
// begin with, named in words, and what mega needs, each harness found or
// missing and why, as the service's start-up check found it in the
// operator's shell environment and re-checks on every read.
export function FirstRun({ status }: { status: HarnessStatus[] | undefined }) {
  return (
    <div className="first-run">
      <p>
        mega runs Claude Code and Codex conversations in the checkouts you register. To begin, press <strong>Register a project</strong>,
        the folder button at the top of the side panel, and give it a checkout&apos;s absolute path.
      </p>
      <h2>What mega needs</h2>
      {status === undefined ? (
        <p className="muted">Looking for the harnesses…</p>
      ) : (
        <ul>
          {harnesses.map((h) => {
            const s = status.find((item) => item.name === h.value)
            if (!s) {
              return null
            }
            return (
              <li key={h.value} data-harness={h.value} className={s.found ? 'found' : 'missing notice'}>
                <strong>{h.label}</strong> {s.found ? `can run: ${s.where ?? ''}` : `cannot run: ${s.reason ?? ''}`}
              </li>
            )
          })}
        </ul>
      )}
      <p className="muted">
        mega looks for them in your shell&apos;s environment, as a project with no activation of its own gets it. A project&apos;s page says what
        its own environment is missing.
      </p>
    </div>
  )
}

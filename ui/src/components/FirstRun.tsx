import { asSentence, harnesses } from '../harnesses'
import type { HarnessStatus } from '../types'

// needsLook is a found harness whose version needs the operator's look: one
// older than the tested range, or one that could not be read. a newer build
// is the normal state for a harness that updates itself, and is said plainly.
const needsLook = (s: HarnessStatus) => s.standing === 'older' || s.standing === 'unknown'

// Found says a harness that can run in a sentence: its build, where it was
// found, and the range mega was tested with, in plain words; for one that
// needs a look, the sentence saying why, in the notice colour. node and the
// shim are the start-up log's to name.
function Found({ status }: { status: HarnessStatus }) {
  const at = `'${status.path ?? status.where ?? ''}'`
  if (status.standing === 'tested' || status.standing === 'newer') {
    const relation = status.standing === 'newer' ? 'newer than' : 'within'
    return (
      <>
        can run: version {status.version}, at {at}, {relation} the {status.tested} mega was tested with.
      </>
    )
  }
  return (
    <>
      can run: at {at}.{needsLook(status) && <span className="harness-version"> {asSentence(status.note)}.</span>}
    </>
  )
}

// FirstRun is the empty workspace's page: what mega is for, the control to
// begin with, named in words, and what mega needs, each harness found or
// missing and why, as the service's start-up check found it in the
// operator's shell environment and re-checks on every read.
export function FirstRun({ status }: { status: HarnessStatus[] | undefined }) {
  return (
    <div className="first-run">
      <p>
        mega runs Claude Code, Codex and pi conversations in the checkouts you register. To begin, press <strong>Register a project</strong>,
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
              <li key={h.value} data-harness={h.value} className={s.found ? (needsLook(s) ? 'found notice' : 'found') : 'missing notice'}>
                <strong>{h.label}</strong> {s.found ? <Found status={s} /> : `cannot run: ${s.reason ?? ''}`}
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

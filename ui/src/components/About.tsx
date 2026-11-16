import { Fragment } from 'react'
import { newerLabel } from '../build'
import { asSentence, harnesses } from '../harnesses'
import type { About } from '../types'
import { Dialog } from './Dialog'

// AboutPanel is what 'mega version' prints, from the service's build
// information, with where it logs and the Claude Agent SDK the shim runs,
// each harness's tested range and the version the service found, and
// whether this page runs the interface the service now serves.
export function AboutPanel({ about, newer, onClose }: { about: About; newer: boolean; onClose: () => void }) {
  return (
    <Dialog title="This mega" onClose={onClose}>
      <dl className="about">
        <dt>Version</dt>
        <dd>
          {about.version}
          {about.developer && ' (developer build)'}
        </dd>
        {about.commit && (
          <>
            <dt>Commit</dt>
            <dd>
              <code>{about.commit}</code>
            </dd>
          </>
        )}
        {about.built && (
          <>
            <dt>Built</dt>
            <dd>{about.built}</dd>
          </>
        )}
        {about.branch && (
          <>
            <dt>Branch</dt>
            <dd>{about.branch}</dd>
          </>
        )}
        <dt>Go</dt>
        <dd>{about.go}</dd>
        <dt>Target</dt>
        <dd>{about.target}</dd>
        <dt>Log file</dt>
        <dd>
          <code>{about.log_file}</code>
          <span className="muted">
            {about.logs_to_file ? 'This mega is writing to it.' : 'Written when mega serves with no terminal; this one logs to its terminal.'}
          </span>
        </dd>
        <dt>Claude Agent SDK</dt>
        <dd>{about.claude_sdk}</dd>
        {harnesses.map((h) => {
          const status = about.harnesses?.find((s) => s.name === h.value)
          return (
            <Fragment key={h.value}>
              <dt>{h.label}</dt>
              <dd data-harness={h.value}>
                {!status && `Tested ${about.tested[h.value] ?? 'with no range'}; not looked for yet.`}
                {status && !status.found && `Tested ${about.tested[h.value]}; cannot run: ${status.reason ?? ''}`}
                {status?.found && (
                  <>
                    {status.version ? `Version ${status.version}` : status.standing ? 'Version unknown' : 'Version not read yet'}; tested {status.tested}
                    {status.standing === 'newer' ? ', so this build is newer.' : '.'}
                    {(status.standing === 'older' || status.standing === 'unknown') && <span className="notice-text"> {asSentence(status.note)}.</span>}
                  </>
                )}
              </dd>
            </Fragment>
          )
        })}
        <dt>This page</dt>
        <dd>{newer ? newerLabel : 'Running the interface this mega serves.'}</dd>
      </dl>
    </Dialog>
  )
}

import { Fragment, useEffect, useState } from 'react'
import { getAbout } from '../api'
import { dateTime } from '../zone'
import { newerLabel } from '../build'
import { asSentence, harnesses } from '../harnesses'
import type { About, HarnessStatus } from '../types'
import { Dialog } from './Dialog'

// AboutPanel is what 'mega version' prints, from the service's build
// information, with where it logs and the Claude Agent SDK the shim runs,
// each harness's tested range and the version the service found, and
// whether this page runs the interface the service now serves, and where the
// configuration snapshots are and when the newest was taken. it reads the
// service again as it opens, since a snapshot may have been taken since the
// page connected. the harnesses come from the workspace when it has them,
// which the service announces again when a version read goes stale and is
// read anew, so the panel follows an update while it is open.
export function AboutPanel({
  about: connected,
  harnesses: live,
  newer,
  onClose,
}: {
  about: About
  harnesses: HarnessStatus[] | undefined
  newer: boolean
  onClose: () => void
}) {
  const [about, setAbout] = useState(connected)
  const statuses = live ?? about.harnesses
  useEffect(() => {
    getAbout().then(setAbout, () => undefined)
  }, [])
  const snapshots = about.snapshots
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
          const status = statuses?.find((s) => s.name === h.value)
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
        <dt>Configuration snapshots</dt>
        <dd data-snapshots="">
          {!snapshots && 'None are taken (snapshots.keep is 0).'}
          {snapshots?.newest && (
            <>
              Newest {dateTime(new Date(snapshots.newest.taken))}, at {snapshots.newest.reason === 'version' ? 'a version change' : 'start-up'}; {snapshots.count}{' '}
              kept, in <code>{snapshots.dir}</code>
            </>
          )}
          {snapshots && !snapshots.newest && (
            <>
              None yet, in <code>{snapshots.dir}</code>
            </>
          )}
        </dd>
        <dt>This page</dt>
        <dd>{newer ? newerLabel : 'Running the interface this mega serves.'}</dd>
      </dl>
    </Dialog>
  )
}

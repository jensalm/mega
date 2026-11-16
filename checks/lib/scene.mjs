// appends hand-written turns to a scratch mega event log, for a picture of
// the page that reads as real work: a prompt, the actions a turn took with
// their results, and a reply in markdown. it writes mega's own events, as
// gen.mjs does. a tool call's payload is the message the fake Claude shim
// sends for it, built by the shim's own helper (shim/claude/fake/messages.mjs),
// so an expanded row shows the call and the Claude protocol stays with the
// shim. it writes only scratch logs.
//
//   appendScene(path, turns) -> last seq
//   turns: [{ prompt, actions: [{ tool, title, input, result }], reply }]
import { appendFileSync, readFileSync } from 'node:fs'
import { randomUUID } from 'node:crypto'
import { toolUseMessage } from '../../shim/claude/fake/messages.mjs'

export function appendScene(path, turns) {
  const last = JSON.parse(readFileSync(path, 'utf8').trim().split('\n').at(-1))
  let seq = last.seq
  let at = Date.now() - turns.length * 4 * 60000
  const out = []
  const emit = (ev) => {
    seq++
    at += 2000 + Math.floor(seq % 7) * 1500
    out.push(JSON.stringify({ origin: 'managed', ...ev, seq, at: new Date(at).toISOString() }))
  }
  for (const turn of turns) {
    emit({ kind: 'user', native_id: randomUUID(), text: turn.prompt })
    for (const action of turn.actions ?? []) {
      const id = `toolu_${randomUUID().slice(0, 8)}`
      emit({
        kind: 'tool_use',
        tool: action.tool,
        tool_use_id: id,
        native_id: randomUUID(),
        text: action.title,
        raw: toolUseMessage({ id, name: action.tool, input: action.input, uuid: randomUUID(), session: 'scene' }),
      })
      emit({ kind: 'tool_result', tool_use_id: id, native_id: randomUUID(), text: action.result })
    }
    emit({ kind: 'assistant', native_id: randomUUID(), text: turn.reply })
    emit({ kind: 'turn_end', status: 'success', text: 'success', native_id: randomUUID() })
  }
  appendFileSync(path, out.join('\n') + '\n')
  return seq
}

// writes a long synthetic Claude Code transcript, for timing a read of one:
// where Claude Code keeps it under a config directory (CLAUDE_CONFIG_DIR),
// in the shape its transcripts take (2.1.325), each turn a prompt, a tool
// call, its result and a reply, of filler text. kept here, beside the fake
// shim, so the Claude protocol stays with the shim; probes/look/gen.mjs
// calls it. it returns the session's id.
import { mkdirSync, writeFileSync } from 'node:fs'
import { randomUUID } from 'node:crypto'
import { join } from 'node:path'

const filler = (n) => 'filler text standing in for a long reply. '.repeat(n)

export function writeLongTranscript(configDir, cwd, turns) {
  const claudeID = randomUUID()
  const project = join(configDir, 'projects', cwd.replace(/[^a-zA-Z0-9]/g, '-'))
  mkdirSync(project, { recursive: true })
  const lines = []
  let parent = null
  let at = Date.parse('2026-11-18T09:00:00Z')
  const entry = (type, message) => {
    const uuid = randomUUID()
    lines.push(JSON.stringify({ type, uuid, parentUuid: parent, sessionId: claudeID, cwd, timestamp: new Date((at += 1000)).toISOString(), version: '2.1.325', isSidechain: false, userType: 'external', message }))
    parent = uuid
  }
  for (let i = 1; i <= turns; i++) {
    entry('user', { role: 'user', content: `turn ${i}: ${filler(4)}` })
    const tool = `toolu_${i}`
    entry('assistant', { role: 'assistant', model: 'claude-fake', content: [{ type: 'tool_use', id: tool, name: 'Bash', input: { command: `echo ${i}` } }] })
    entry('user', { role: 'user', content: [{ type: 'tool_result', tool_use_id: tool, content: filler(10) }] })
    entry('assistant', { role: 'assistant', model: 'claude-fake', content: [{ type: 'text', text: `reply ${i}: ${filler(20)}` }] })
  }
  writeFileSync(join(project, `${claudeID}.jsonl`), lines.join('\n') + '\n')
  return claudeID
}

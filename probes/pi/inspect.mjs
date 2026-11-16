// what a pi RPC process reports before any prompt, without a model turn:
// get_state (model, thinking level, session), the model list, the thinking
// levels the model allows, the commands (extension commands show whether the
// project's extensions loaded), and the MCP servers' status, which pi gives
// only as the text of a notify answering the /mcp command. no session file is
// written, since no prompt reaches a model.
//
//   node probes/pi/inspect.mjs <checkout>
import { spawn } from 'node:child_process'
import { randomUUID } from 'node:crypto'

const cwd = process.argv[2] ?? '.'
const child = spawn('pi', ['--mode', 'rpc', '--session-id', randomUUID()], { cwd, stdio: ['pipe', 'pipe', 'pipe'] })
let buf = ''
const frames = []
child.stdout.on('data', (d) => {
  buf += d
  let i
  while ((i = buf.indexOf('\n')) >= 0) {
    const line = buf.slice(0, i)
    buf = buf.slice(i + 1)
    try {
      frames.push(JSON.parse(line))
    } catch {
      frames.push({ raw: line.slice(0, 200) })
    }
  }
})
let stderr = ''
child.stderr.on('data', (d) => (stderr += d))
const send = (o) => child.stdin.write(JSON.stringify(o) + '\n')
const sleep = (ms) => new Promise((r) => setTimeout(r, ms))

await sleep(1500)
send({ type: 'get_state', id: 'state' })
send({ type: 'get_available_models', id: 'models' })
send({ type: 'get_available_thinking_levels', id: 'levels' })
send({ type: 'get_commands', id: 'commands' })
// mcp servers connect in the background as the session starts.
await sleep(12000)
send({ type: 'prompt', id: 'mcp', message: '/mcp' })
await sleep(8000)
child.stdin.end()
await sleep(1500)

for (const f of frames) {
  if (f.type === 'response' && f.command === 'get_state') {
    console.log('state:', JSON.stringify({ model: `${f.data?.model?.provider}/${f.data?.model?.id}`, thinkingLevel: f.data?.thinkingLevel, keys: Object.keys(f.data ?? {}) }))
  } else if (f.type === 'response' && f.command === 'get_available_models') {
    console.log('models:', JSON.stringify((f.data?.models ?? []).map((m) => `${m.provider}/${m.id}${m.reasoning ? '' : ' (no reasoning)'}`)))
  } else if (f.type === 'response' && f.command === 'get_available_thinking_levels') {
    console.log('thinking levels:', JSON.stringify(f.data?.levels))
  } else if (f.type === 'response' && f.command === 'get_commands') {
    console.log('commands:', JSON.stringify((f.data?.commands ?? []).map((c) => `${c.name} (${c.source ?? ''})`)))
  } else if (f.type === 'extension_ui_request') {
    console.log(`ui ${f.method}:`, JSON.stringify(f.message ?? f.title ?? ''))
  } else if (f.type === 'response' && f.command === 'prompt') {
    console.log('/mcp:', JSON.stringify(f.data ?? f.error))
  }
}
console.log('stderr:', stderr.trim().slice(0, 300))

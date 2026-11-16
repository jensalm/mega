// how pi ends when its input closes while an extension dialog is pending:
// a new session in the fixture asks for `sleep 5`, the fixture's extension
// raises its approval select, and the dialog is never answered. then stdin is
// closed, and if pi has not exited within 30 s, SIGTERM is sent, then a second
// SIGTERM 30 s later, then SIGKILL. with abort-then-close, an abort is sent
// first. it prints when each step happened and how pi exited.
//
//   cd <trusted pi-rt1 fixture> && node probes/pi/close-with-dialog.mjs . close
//   cd <trusted pi-rt1 fixture> && node probes/pi/close-with-dialog.mjs . abort-then-close
//
// pi 1.1.0 exited within about 40 ms of stdin closing; 1.5.2 waits until a
// second signal (capability report, 2026-11-16). it uses the operator's
// default model, never passes --approve, and answers no dialog.
import { spawn } from 'node:child_process'
import { randomUUID } from 'node:crypto'
const [cwd, how] = process.argv.slice(2)
const t0 = Date.now()
const at = () => ((Date.now() - t0) / 1000).toFixed(1) + 's'
const child = spawn('pi', ['--mode', 'rpc', '--session-id', randomUUID()], { cwd, stdio: ['pipe', 'pipe', 'pipe'] })
let exited = null
child.on('exit', (code, sig) => { exited = { code, sig }; console.log(at(), 'exit', code, sig) })
let buf = ''
let dialog = null
child.stdout.on('data', (d) => {
  buf += d
  let i
  while ((i = buf.indexOf('\n')) >= 0) {
    const line = buf.slice(0, i); buf = buf.slice(i + 1)
    try { const f = JSON.parse(line); if (f.type === 'extension_ui_request' && f.method === 'select') { dialog = f; console.log(at(), 'dialog', f.title) } } catch {}
  }
})
const sleep = (ms) => new Promise((r) => setTimeout(r, ms))
child.stdin.write(JSON.stringify({ type: 'prompt', id: 'p', message: 'Using the bash tool, run exactly `sleep 5` and then reply CLOSE-TEST-1. Nothing else.' }) + '\n')
while (!dialog && Date.now() - t0 < 60000) await sleep(200)
if (!dialog) { console.log('no dialog'); child.kill('SIGKILL'); process.exit(1) }
await sleep(1000)
if (how === 'abort-then-close') { child.stdin.write(JSON.stringify({ type: 'abort', id: 'a' }) + '\n'); console.log(at(), 'abort sent'); await sleep(2000) }
child.stdin.end(); console.log(at(), 'stdin closed')
for (let i = 0; i < 300 && !exited; i++) await sleep(100)
if (!exited) { child.kill('SIGTERM'); console.log(at(), 'SIGTERM 1') }
for (let i = 0; i < 300 && !exited; i++) await sleep(100)
if (!exited) { child.kill('SIGTERM'); console.log(at(), 'SIGTERM 2') }
for (let i = 0; i < 100 && !exited; i++) await sleep(100)
if (!exited) { child.kill('SIGKILL'); console.log(at(), 'SIGKILL') }
await sleep(500)

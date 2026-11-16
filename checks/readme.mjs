// browser check: the README's picture of a workspace, against the fakes.
//
//   make checks SUITE=readme      # the check, and shots/workspace.png
//   make readme-image             # the same, copied to docs/images/workspace.png
//
// a scratch workspace, never anyone's own: three projects, and conversations
// on all three harnesses in different states. one Claude conversation holds
// hand-written turns (lib/scene.mjs), so the history reads as real work; the
// rest are the fakes in the states the side panel shows: a Claude approval
// waiting, a Codex reply unread, a pi turn running, and one idle. the mega is
// the checks' own, mega-check. what it checks: the needs-you list holds the
// approval and the unread reply, the running pi turn shows as working, and
// the open conversation shows the scene's last reply; then it saves the page
// at 1440x900 as shots/workspace.png.
import { join } from 'node:path'
import { appendScene } from './lib/scene.mjs'
import { suite } from './lib/mega.mjs'

const t = await suite('readme')
const { url, check, until, api, convOf, workspace } = t

const web = await t.register('website', t.checkout('website', { git: true }))
const apiProject = await t.register('api-service', t.checkout('api-service', { git: true }))
const notes = await t.register('field-notes', t.checkout('field-notes', { git: true }))

// the conversation the picture opens, given its history while the service is
// stopped.
const nav = (await t.launch(web.id, 'claude', 'Make the nav bar collapse on small screens.', { name: 'nav on small screens' })).id
await t.idle(nav)
await api('POST', `/api/conversations/${nav}/release`)
await until('nav released', async () => (await convOf(nav)).state === 'none')
await t.stop()
appendScene(join(t.dir, 'data', 'conversations', nav, 'events.jsonl'), [
  {
    prompt: 'Where is the nav bar defined, and what styles it?',
    actions: [
      { tool: 'Grep', title: 'Grep <Nav src', input: { pattern: '<Nav', path: 'src' }, result: 'src/App.tsx:12\nsrc/components/Nav.tsx:8' },
      { tool: 'Read', title: 'Read src/styles/nav.css', input: { file_path: 'src/styles/nav.css' }, result: '.links { display: flex; gap: 24px; }' },
    ],
    reply: 'In `src/components/Nav.tsx`, rendered once from `App.tsx`. Its styles are in `src/styles/nav.css`: a single flex row with no breakpoint, which is why it overflows on a narrow screen.',
  },
  {
    prompt: 'The nav bar overflows on phones. Make it collapse into a menu button below 720px, and keep the keyboard order sensible.',
    actions: [
      { tool: 'Read', title: 'Read src/components/Nav.tsx', input: { file_path: 'src/components/Nav.tsx' }, result: 'export function Nav() { … }' },
      { tool: 'Edit', title: 'Edit src/components/Nav.tsx', input: { file_path: 'src/components/Nav.tsx', old_string: '<ul className="links">', new_string: '<button aria-expanded={open}>Menu</button>\n<ul className="links">' }, result: 'The file src/components/Nav.tsx has been updated.' },
      { tool: 'Edit', title: 'Edit src/styles/nav.css', input: { file_path: 'src/styles/nav.css', old_string: '.links {', new_string: '@media (max-width: 720px) {\n  .links { display: none; }\n}\n.links {' }, result: 'The file src/styles/nav.css has been updated.' },
      { tool: 'Bash', title: 'Bash npm test -- nav', input: { command: 'npm test -- nav', description: 'Run the nav tests' }, result: '✓ 14 tests passed' },
    ],
    reply:
      'The nav now collapses below 720px.\n\n- **`Nav.tsx`** renders a menu button that shows and hides the links, with `aria-expanded`, and focus goes back to the button when the menu closes.\n- **`nav.css`** hides the links behind the button under the breakpoint; the desktop layout is unchanged.\n\nThe nav tests pass. I left the footer links alone: they already wrap.',
  },
  {
    prompt: 'Good. Does Escape close the menu?',
    actions: [
      { tool: 'Grep', title: 'Grep Escape src/components', input: { pattern: 'Escape', path: 'src/components' }, result: 'src/components/Dialog.tsx:41' },
      { tool: 'Edit', title: 'Edit src/components/Nav.tsx', input: { file_path: 'src/components/Nav.tsx', old_string: 'onClick={toggle}', new_string: 'onClick={toggle}\nonKeyDown={(e) => e.key === "Escape" && close()}' }, result: 'The file src/components/Nav.tsx has been updated.' },
      { tool: 'Bash', title: 'Bash npm test -- nav', input: { command: 'npm test -- nav', description: 'Run the nav tests' }, result: '✓ 15 tests passed' },
    ],
    reply: 'It does now. Escape closes the menu and returns focus to the button, as the dialog does. I added a test for it, so there are 15.',
  },
])
await t.start()

// the rest, in the states the side panel shows.
const review = (await t.launch(web.id, 'claude', 'ask', { name: 'review the login form' })).id
const limits = (await t.launch(apiProject.id, 'codex', 'hello', { name: 'rate limits' })).id
const flaky = (await t.launch(apiProject.id, 'claude', 'echo\nhello', { name: 'flaky retry test' })).id
const summary = (await t.launch(notes.id, 'pi', 'hold', { name: 'weekly summary' })).id
await until('the approval waiting', async () => (await convOf(review)).pending.length === 1)
await t.idle(limits)
await t.idle(flaky)
await until('pi working', async () => (await convOf(summary)).turn_active)
// the idle one has been read; the codex reply has not.
const flakyLast = (await api('GET', `/api/conversations/${flaky}/events?turns=1`)).body.through
await api('POST', `/api/conversations/${flaky}/read`, { seq: flakyLast })

const page = await t.page({ width: 1440, height: 900 })
await page.goto(`${url}/#/c/${nav}`)
await until('the scene shown', async () => (await page.locator('.events .message.assistant', { hasText: 'Escape closes the menu' }).count()) === 1)
const needs = (await workspace()).needs_you
check(needs.includes(review) && needs.includes(limits) && !needs.includes(flaky), `needs you holds the approval and the unread reply: ${needs.length}`)
check((await convOf(summary)).turn_active, 'the pi turn is running')
// every project unfolded, and the last action of the last turn open, so the
// picture shows what an expanded row holds.
const folded = page.locator('nav.sidebar button[aria-expanded="false"]')
while ((await folded.count()) > 0) await folded.first().click()
const lastAction = page.locator('.events details.tool').last()
await lastAction.locator('summary').click()
await until('the action read', async () => (await lastAction.locator('pre.command').count()) === 1)
await page.locator('.main').evaluate((m) => {
  m.scrollTop = m.scrollHeight
})
// focus and the pointer away, so no turn's line, which carries the time,
// shows: the picture does not change with the clock.
await page.evaluate(() => (document.activeElement instanceof HTMLElement ? document.activeElement.blur() : undefined))
await page.mouse.move(2, 2)
await page.waitForTimeout(500)
await page.screenshot({ path: join(t.dir, 'shots', 'workspace.png') })
check(true, 'the workspace saved as shots/workspace.png')

await t.done()

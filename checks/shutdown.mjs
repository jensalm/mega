// browser check: the shutdown gate, against the fakes.
//
//   make checks SUITE=shutdown
//
// rebuilt from what the journals recorded of the lost shutdown script and
// from docs/current. two conversations hold the gate: R with a request for the
// operator, W with a turn in flight. nothing is approved: each is brought to
// idle by an interrupt. what it checks: arming from the footer,
// the banner naming what it waits for by count and by name, the request
// marked as the operator's; a second page seeing it; new work refused with the
// reason, on the page and by the API; a blocker's name opening it; Shut down
// not offered, and refused, before idle; cancelling lifting it on both pages;
// at idle, Shut down offered, the title saying so and one notification; Shut
// down ending the service with status 0; the page that asked saying when, the
// other showing only what it last knew; and both recovering when the service
// returns, the gate lifted. the service runs in one time zone and the pages
// in another, UTC, as a hardened browser reports; the time the asking page
// gives is the service's.
import { reconnectable, suite } from './lib/mega.mjs'

const serviceZone = 'Asia/Kolkata'
const t = await suite('shutdown', { tz: serviceZone })
const { url, check, until, api, convOf, workspace } = t

const P = await t.register('gate', t.checkout('gate'))
const R = (await t.launch(P.id, 'claude', 'echo\nhello', { name: 'asker' })).id
const W = (await t.launch(P.id, 'claude', 'echo\nhello', { name: 'worker' })).id
await t.idle(R)
await t.idle(W)
const prompt = async (id, text) => {
  const r = await api('POST', `/api/conversations/${id}/prompt`, { text })
  if (r.status !== 202) throw new Error(`prompting: ${r.status} ${JSON.stringify(r.body)}`)
}
await prompt(R, 'ask')
await prompt(W, 'tool 600000')
await until('R waiting and W working', async () => (await convOf(R)).pending.length === 1 && (await convOf(W)).turn_active)

// notifications are recorded, not shown: the page sees a granted permission.
const notes = () => {
  window.__notes = []
  class Recorded {
    static permission = 'granted'
    static requestPermission() {
      return Promise.resolve('granted')
    }
    constructor(title, options) {
      window.__notes.push({ title, body: options?.body })
    }
  }
  window.Notification = Recorded
}
const page = await t.page({ init: notes, timezoneId: 'UTC' })
const other = await t.page({ init: notes, timezoneId: 'UTC' })
const pageStream = await reconnectable(page)
await page.goto(`${url}/#/p/${P.id}`)
await other.goto(`${url}/#/c/${W}`)
for (const p of [page, other]) await p.locator('.side-foot .stream-connected').waitFor()

const banner = (p) => p.locator('section.gate-banner[aria-label=Shutdown]')
const pendingReason = 'Waiting to shut down once every agent is idle; cancel the shutdown to start new work'

// arming from the footer: the banner names what it waits for.
await page.getByRole('button', { name: 'Shut down when every agent is idle' }).click()
await banner(page).waitFor()
const head = await banner(page).locator('strong').textContent()
check(head === 'Waiting to shut down: 1 request for you to answer · 1 turn in flight', `armed, the banner counts what it waits for: "${head}"`)
const items = banner(page).locator('ul.gate-blockers li')
const named = await items.allTextContents()
check(named.length === 2 && named.some((n) => n.startsWith('asker') && n.includes('a request for you to answer')) && named.some((n) => n.startsWith('worker') && n.includes('a turn in flight')), `and names each: ${JSON.stringify(named)}`)
check((await items.filter({ hasText: 'asker' }).locator('.gate-kinds.yours').count()) === 1 && (await items.filter({ hasText: 'worker' }).locator('.gate-kinds.yours').count()) === 0, 'the request is marked as yours to answer')
check((await banner(page).getByRole('button', { name: 'Shut down mega now' }).count()) === 0, 'Shut down is not offered before idle')
const early = await api('POST', '/api/shutdown')
check(early.status === 409 && early.body.code === 'conflict', `and the service refuses it: ${early.status} ${early.body.code}`)

// a second page sees it.
await banner(other).waitFor()
check((await banner(other).locator('strong').textContent()) === head && (await other.locator('.side-gate').count()) === 1, 'a second page sees it, in its banner and above its footer')

// new work is refused, with the reason.
const launch = page.getByRole('button', { name: pendingReason })
await until('New conversation held', async () => (await launch.count()) >= 1)
check(await launch.first().isDisabled(), 'New conversation is disabled with the shutdown as its reason')
const send = other.locator('form.reply button[type=submit]')
check((await send.getAttribute('aria-label')) === pendingReason && (await send.isDisabled()), 'and so is Send')
const refused = await api('POST', `/api/projects/${P.id}/conversations`, { harness: 'claude', prompt: 'echo\nrefused' })
check(refused.status === 409 && refused.body.code === 'shutdown_pending', `the service refuses new work: ${refused.status} ${refused.body.code}`)

// a blocker's name opens it.
await items.filter({ hasText: 'asker' }).getByRole('link', { name: 'asker' }).click()
await until('asker open', async () => (await page.locator('.conversation-head h1').textContent()) === 'asker')
check(page.url().endsWith(`#/c/${R}`), 'a blocker’s name opens its conversation')

// cancelling lifts it on both pages.
await banner(page).getByRole('button', { name: 'Cancel the shutdown' }).click()
await until('the banner gone on both pages', async () => (await banner(page).count()) === 0 && (await banner(other).count()) === 0)
await until('Send free', async () => (await send.getAttribute('aria-label')) === 'Send (Ctrl+Enter)')
check(!(await workspace()).shutdown.armed, 'cancelling lifts it on both pages, and new work is offered again')

// armed again and brought to idle by the operator, each turn interrupted
// from its own page; nothing is answered or interrupted for them. the request
// is not approved to get there: interrupting its turn aborts it.
await page.getByRole('button', { name: 'Shut down when every agent is idle' }).click()
await banner(page).waitFor()
await page.getByRole('button', { name: 'Interrupt', exact: true }).click()
await until('one blocker left', async () => (await items.count()) === 1)
check((await banner(page).locator('strong').textContent()) === 'Waiting to shut down: 1 turn in flight' && (await convOf(R)).pending.length === 0, 'interrupting the asker aborts its request and leaves the other turn')
await other.getByRole('button', { name: 'Interrupt', exact: true }).click()
await until('idle', async () => (await banner(page).locator('strong').textContent()) === 'Every agent is idle')
const shutDown = banner(page).getByRole('button', { name: 'Shut down mega now' })
check((await shutDown.count()) === 1 && (await shutDown.locator('.word').textContent()) === 'Shut down', 'at idle the banner says so and offers Shut down')
await until('the title', async () => (await page.title()) === 'Ready to shut down · mega · mega-check')
check(true, 'the title says it is ready')
await until('the notification', async () => (await page.evaluate(() => window.__notes.length)) >= 1)
const sent = await page.evaluate(() => window.__notes)
check(sent.length === 1 && sent[0].title === 'mega · mega-check' && sent[0].body === 'Every agent is idle. Shut down from the banner when you are ready.', `one notification: ${JSON.stringify(sent)}`)

// Shut down ends the service cleanly.
const askedFrom = Date.now()
await shutDown.click()
const ended = await t.stopped()
check(ended && ended.code === 0, `Shut down ends the service with status 0: ${JSON.stringify(ended)}`)
await until('the asking page says when', async () => /^mega was shut down from this page at \d/.test((await page.locator('section.gate-banner.shut-down strong').textContent().catch(() => '')) ?? ''))
const askedTo = Date.now()
const saidAt = ((await page.locator('section.gate-banner.shut-down strong').textContent()) ?? '').replace('mega was shut down from this page at ', '')
// the time of day, to the second, in a zone, as the page formats it.
const clockIn = (ms, zone) => page.evaluate(([ms, zone]) => new Intl.DateTimeFormat([], { hour: '2-digit', minute: '2-digit', second: '2-digit', timeZone: zone }).format(new Date(ms)), [ms, zone])
const could = []
for (let ms = askedFrom - 1000; ms <= askedTo + 1000; ms += 500) could.push(await clockIn(ms, serviceZone))
check(could.includes(saidAt) && !could.includes(await clockIn(askedFrom, 'UTC')), `the page that asked says when, in the service's zone, not the page's UTC: "${saidAt}"`)
await until('its footer', async () => (await page.locator('.side-foot .stream').textContent()) === 'Shut down from here, retrying')
check(true, 'and its footer says it keeps trying')
// the footer's longest status: the mega's name is the last thing to give
// way, shown whole, while the status wraps and the version yields.
const footName = page.locator('.side-foot .mega-name')
const nameFits = await footName.evaluate((e) => ({ text: e.textContent, whole: e.scrollWidth <= e.clientWidth, width: e.clientWidth }))
check(nameFits.text === 'mega-check' && nameFits.whole && nameFits.width > 0, `beside it, the mega's name is shown whole: ${JSON.stringify(nameFits)}`)
await page.locator('.side-foot').screenshot({ path: `${t.dir}/shots/foot-shut-down.png` })
await until('the other page loses it', async () => (await other.locator('.side-foot .stream').textContent()) !== 'Live')
await until('the other page shows the last known gate', async () => (await banner(other).locator('strong').textContent()) === 'Last known: Every agent is idle')
check((await banner(other).getByRole('button', { name: 'Shut down mega now' }).count()) === 0, `the other page shows only what it last knew, offering nothing: "${await other.locator('.side-foot .stream').textContent()}"`)
await t.shot(page, 'shut-down')

// the service returns: both pages recover, the gate lifted. the asking page
// does not report its old shutdown as under way before a read of the new
// service's workspace says one is: its first read is held, and meanwhile the
// banner says it is reading.
await pageStream.hold()
await t.start()
await until('the asking page back and catching up', async () => (await page.locator('.side-foot .stream').textContent()) === 'Catching up' && pageStream.held() > 0, 30000)
const backSays = (await page.locator('section.gate-banner.shut-down').textContent()) ?? ''
check(!backSays.includes('Shutting down') && !backSays.includes('being released') && !backSays.includes('was shut down') && backSays.includes('reading the workspace'), `reconnected before its workspace read, the banner claims neither a shutdown under way nor one done: "${backSays}"`)
// another page arms the new service before that read lands: the read shows a
// gate armed afresh, not the old shutdown, and the asking page shows that.
check((await api('POST', '/api/shutdown/arm')).status === 202, 'the new service is armed from elsewhere before the asking page has read it')
await pageStream.catchUp()
await until('the asking page shows the new gate', async () => (await banner(page).count()) === 1)
const rearmed = (await page.locator('section.gate-banner').textContent()) ?? ''
check(!rearmed.includes('Shutting down') && !rearmed.includes('being released') && (await page.locator('section.gate-banner.shut-down').count()) === 0, `a gate armed afresh is shown as itself, never as the old shutdown: "${rearmed}"`)
check((await api('POST', '/api/shutdown/cancel')).status === 202, 'and that arming is cancelled')
for (const p of [page, other]) await p.locator('.side-foot .stream-connected').waitFor({ timeout: 30000 })
await until('the banners gone', async () => (await page.locator('section.gate-banner').count()) === 0 && (await banner(other).count()) === 0)
await until('the asking page free', async () => (await page.locator('form.reply button[type=submit]').getAttribute('aria-label')) === 'Send (Ctrl+Enter)')
const after = await workspace()
check(!after.shutdown.armed && after.conversations.every((c) => c.state === 'none'), 'when the service returns both pages recover, the gate unarmed and every runtime released')
check((await page.title()) !== 'Ready to shut down · mega · mega-check', `and the title is back: "${await page.title()}"`)

await t.done()

// times of day on the page are said in the zone the service runs in, the
// workspace's time_zone, not the browser's own: a hardened browser may tell
// the page it is in UTC, hours off the operator's clock. until the
// workspace has named a zone, or where it names none, the browser's is used.

let zone: string | undefined

// setZone takes the service's zone, as the workspace names it. a name this
// browser cannot format in is passed over for the browser's own.
export function setZone(name: string | undefined) {
  if (name === zone) {
    return
  }
  try {
    if (name) {
      new Intl.DateTimeFormat([], { timeZone: name })
    }
    zone = name || undefined
  } catch {
    zone = undefined
  }
  formats.clear()
}

const formats = new Map<string, Intl.DateTimeFormat>()

function format(at: Date, options: Intl.DateTimeFormatOptions): string {
  const key = JSON.stringify(options)
  let f = formats.get(key)
  if (!f) {
    f = new Intl.DateTimeFormat([], { ...options, timeZone: zone })
    formats.set(key, f)
  }
  return f.format(at)
}

const day = (at: Date) => format(at, { year: 'numeric', month: '2-digit', day: '2-digit' })

// clockTime is a time of day, to the second.
export const clockTime = (at: Date) => format(at, { hour: '2-digit', minute: '2-digit', second: '2-digit' })

// shortTime is a time of day to the minute, with its date before it when it
// is not today, and its year when that is not this year.
export function shortTime(at: Date, now = new Date()): string {
  const time = format(at, { hour: '2-digit', minute: '2-digit' })
  if (day(at) === day(now)) {
    return time
  }
  const sameYear = format(at, { year: 'numeric' }) === format(now, { year: 'numeric' })
  return `${format(at, sameYear ? { month: 'short', day: 'numeric' } : { year: 'numeric', month: 'short', day: 'numeric' })} ${time}`
}

// fullTime is a date and time to the second, with the zone named.
export const fullTime = (at: Date) =>
  format(at, { weekday: 'short', year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit', second: '2-digit', timeZoneName: 'short' })

// dateTime is a date and time to the second, as a list or a panel says when
// something happened.
export const dateTime = (at: Date) => format(at, { year: 'numeric', month: 'numeric', day: 'numeric', hour: 'numeric', minute: '2-digit', second: '2-digit' })

// parse is an ISO time from the service, or undefined when absent or not a
// time.
export function parse(at: string | undefined): Date | undefined {
  if (!at) {
    return undefined
  }
  const d = new Date(at)
  return Number.isNaN(d.getTime()) ? undefined : d
}

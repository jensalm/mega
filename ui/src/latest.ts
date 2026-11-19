import { useCallback, useLayoutEffect, useRef, useState, type RefObject } from 'react'

// how far from the bottom still counts as at the bottom, for rounding and for
// the last line settling.
const slack = 32

// the main column is what scrolls; a conversation is one element inside it.
function scrollerOf(content: HTMLElement | null): HTMLElement | null {
  return content?.closest<HTMLElement>('.main') ?? null
}

// where the view was left last, by a scroll the page read or by mega's own
// move, per scrolling column. a scroll event that finds the view higher than
// this was the reader's: mega moves a following view only through place,
// which records where it put it, so a stale scroll event queued by one of
// mega's own moves finds the view where mega left it.
const left = new WeakMap<HTMLElement, number>()

// place moves the view as mega does, to top, and records where it ended:
// every move of mega's own goes through it. a move to the end of a column
// whose content has shrunk lands higher than the view was, which the scroll
// event it queues must not read as the reader's.
export function place(scroller: HTMLElement, top: number) {
  scroller.scrollTop = top
  left.set(scroller, scroller.scrollTop)
}

function atBottom(scroller: HTMLElement): boolean {
  return scroller.scrollHeight - scroller.scrollTop - scroller.clientHeight <= slack
}

// moveToLatest scrolls to the newest activity and reports whether the view
// ended at the bottom. pending request cards are docked in the input panel,
// which stays in view, so the bottom of the history is all there is to reach.
function moveToLatest(scroller: HTMLElement): boolean {
  place(scroller, scroller.scrollHeight)
  return atBottom(scroller)
}

// useLatest keeps an open conversation at its newest activity. content is the
// conversation's element inside the scrolling main column, and dock the input
// panel pinned to its foot. the view starts at the top, so nothing carries
// over from whatever was shown before; once the history has loaded it moves to
// the newest event; pending request cards are docked in the input panel,
// always in view. after that, new activity (a change in activity) is followed
// only while the view is at the bottom: a reader who scrolled up stays put,
// and newer is true until they return, by scrolling or through toLatest.
//
// the dock changes height with no new activity: the activity line appears as
// a turn starts, a request docks or leaves, the reply box opens, closes or
// grows. while the view follows the bottom, any change in the dock's size, or
// in the scrolling column's, keeps it at the bottom; a reader who scrolled up
// is not moved.
//
// following is the one reckoning of whether the view follows. it ends when
// a scroll event finds the view higher than where it was left last (left)
// and not at the bottom: a move up that was not mega's, whatever made it, a
// wheel, a key, the scrollbar, a touch, the find bar. nothing else ends it:
// content growing under the view moves nothing, so the scroll event queued
// by mega's own pin and read after the content grew, which once found the
// view off the bottom and ended following for good, finds it where mega
// left it and ends nothing. it begins whenever a scroll finds the view at the
// bottom, however it got there. it is set by scrolling alone, so it still
// holds the state from before the dock changed, which fires no scroll event,
// and a tall docked request cannot read as having scrolled away.
//
// onFollowing hears each change in following, and its value as the
// conversation opens: the page reads a conversation only while its view
// follows the newest output (App), so this one reckoning is also what says
// the newest output is on screen.
export function useLatest(
  content: RefObject<HTMLElement | null>,
  dock: RefObject<HTMLElement | null>,
  loaded: boolean,
  activity: string,
  onFollowing?: (following: boolean) => void,
) {
  const following = useRef(true)
  const seen = useRef<string | null>(null)
  const [newer, setNewer] = useState(false)
  // the newest onFollowing, which is a new function each render.
  const heard = useRef(onFollowing)
  useLayoutEffect(() => {
    heard.current = onFollowing
  })
  // follow sets following, and says so when it changes.
  const follow = useCallback((now: boolean) => {
    const was = following.current
    following.current = now
    if (now !== was) {
      heard.current?.(now)
    }
  }, [])
  useLayoutEffect(() => {
    heard.current?.(following.current)
  }, [])

  useLayoutEffect(() => {
    const scroller = scrollerOf(content.current)
    if (!scroller) {
      return
    }
    place(scroller, 0)
    const onScroll = () => {
      const top = scroller.scrollTop
      const higher = top < (left.get(scroller) ?? top) - 1
      left.set(scroller, top)
      if (atBottom(scroller)) {
        follow(true)
        setNewer(false)
      } else if (higher) {
        follow(false)
      }
    }
    scroller.addEventListener('scroll', onScroll, { passive: true })
    return () => scroller.removeEventListener('scroll', onScroll)
  }, [content, follow])

  useLayoutEffect(() => {
    const scroller = scrollerOf(content.current)
    if (!content.current || !scroller || !loaded || seen.current === activity) {
      return
    }
    const first = seen.current === null
    seen.current = activity
    if (first) {
      follow(moveToLatest(scroller))
    } else if (following.current) {
      place(scroller, scroller.scrollHeight)
    } else {
      setNewer(true)
    }
  }, [content, loaded, activity, follow])

  // a ResizeObserver's callbacks run after layout and before paint, so the
  // re-pin lands in the same frame as the change: no line slides under the
  // dock and back. its first callback, as observing begins, pins nothing
  // until the history has loaded and been moved to.
  useLayoutEffect(() => {
    const scroller = scrollerOf(content.current)
    const panel = dock.current
    if (!scroller || !panel) {
      return
    }
    const observer = new ResizeObserver(() => {
      if (seen.current !== null && following.current) {
        place(scroller, scroller.scrollHeight)
      }
    })
    observer.observe(panel)
    observer.observe(scroller)
    return () => observer.disconnect()
  }, [content, dock])

  const toLatest = () => {
    const scroller = scrollerOf(content.current)
    if (content.current && scroller) {
      follow(moveToLatest(scroller))
    }
    setNewer(false)
  }

  // isFollowing tells the input panel whether the view follows the bottom, so
  // it keeps the reader's place only when it does not. it is the same
  // function for the conversation's life.
  const isFollowing = useCallback(() => following.current, [])

  return { newer, toLatest, isFollowing }
}

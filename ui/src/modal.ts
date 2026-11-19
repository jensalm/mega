import { useSyncExternalStore } from 'react'

// dialogOpen is whether a modal dialog is open anywhere on the page, covering
// what is behind it. it is the page's one test for that, read from the page
// itself, so a dialog added later counts without being named: the keys that
// wait on dialogs, the focus that never lands behind one, and the read that
// is held while one is open (App).
export const dialogOpen = () => document.querySelector('dialog[open]') !== null

// watch calls changed whenever a dialog may have opened or closed: shown,
// closed, or removed from the page while open.
function watch(changed: () => void): () => void {
  const observer = new MutationObserver(changed)
  observer.observe(document.body, { subtree: true, childList: true, attributes: true, attributeFilter: ['open'] })
  return () => observer.disconnect()
}

// useDialogOpen is dialogOpen followed as the page changes.
export const useDialogOpen = (): boolean => useSyncExternalStore(watch, dialogOpen)

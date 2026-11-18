import { useEffect, useId, useRef, useState, type FormEvent, type KeyboardEvent } from 'react'
import { createProject, getFolders, type Folders } from '../api'
import type { Project } from '../types'
import { IconButton } from './Icon'

interface RegisterProjectDialogProps {
  onClose: () => void
  onRegistered: (project: Project) => void
}

// where a checkout field's text points: the folder to list, and the start
// of the name being typed in it. "~/Repos/al" lists ~/Repos for names
// starting "al"; "~/Repos/" lists ~/Repos whole.
function pointsAt(text: string): { dir: string; prefix: string } | null {
  const cut = text.lastIndexOf('/')
  if (cut < 0) {
    return null
  }
  return { dir: cut === 0 ? '/' : text.slice(0, cut), prefix: text.slice(cut + 1) }
}

const join = (dir: string, name: string) => (dir === '/' ? `/${name}` : `${dir}/${name}`)

// the folder a checkout names, without a trailing slash.
const trimmed = (text: string) => (text.length > 1 && text.endsWith('/') ? text.slice(0, -1) : text).trim()

const baseName = (text: string) => {
  const path = trimmed(text)
  return path.slice(path.lastIndexOf('/') + 1)
}

// RegisterProjectDialog registers a checkout as a project. its field takes a
// path, ~ for the home directory, and completes it from the service: below
// it, the folder being looked at, its folders to click through, and a way up
// that stops at the projects root (config projects_root). a path typed in
// full is registered wherever it is; only what is listed is confined to the
// root. the name follows the folder until it is changed by hand.
export function RegisterProjectDialog({ onClose, onRegistered }: RegisterProjectDialogProps) {
  const dialogRef = useRef<HTMLDialogElement>(null)
  const checkoutRef = useRef<HTMLInputElement>(null)
  const listRef = useRef<HTMLUListElement>(null)
  const listId = useId()
  const [name, setName] = useState('')
  const [named, setNamed] = useState(false)
  const [checkout, setCheckout] = useState('')
  const [listing, setListing] = useState<{ dir: string; folders: Folders | null; error: string | null } | null>(null)
  const [selected, setSelected] = useState(0)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  // showModal moves the focus to the dialog's first control, its close
  // button, so the field is focused after it. the field starts at the
  // projects root.
  useEffect(() => {
    dialogRef.current?.showModal()
    checkoutRef.current?.focus()
    getFolders('').then(
      (root) => {
        setCheckout((current) => current || `${root.path}/`)
        setListing({ dir: root.path, folders: root, error: null })
      },
      (failure: unknown) => setListing({ dir: '', folders: null, error: failure instanceof Error ? failure.message : String(failure) }),
    )
  }, [])

  // the folder the field points at is listed, a moment after typing settles.
  const at = pointsAt(checkout)
  const dir = at?.dir ?? null
  useEffect(() => {
    if (dir === null || dir === listing?.dir) {
      return
    }
    let current = true
    const timer = window.setTimeout(() => {
      getFolders(dir).then(
        (folders) => current && setListing({ dir, folders, error: null }),
        (failure: unknown) => current && setListing({ dir, folders: null, error: failure instanceof Error ? failure.message : String(failure) }),
      )
    }, 80)
    return () => {
      current = false
      window.clearTimeout(timer)
    }
  }, [dir, listing?.dir])

  const prefix = (at?.prefix ?? '').toLowerCase()
  const shown = listing?.dir === dir && listing.folders ? listing.folders.entries.filter((f) => f.name.toLowerCase().startsWith(prefix)) : []
  const pick = Math.min(selected, Math.max(shown.length - 1, 0))
  // a folder the field names exactly, as listed: registered or not.
  const exact = listing?.dir === dir && at ? listing.folders?.entries.find((f) => f.name === at.prefix) : undefined
  const registeredAs = exact?.registered
  const shownName = named ? name : baseName(checkout)

  const go = (to: string) => {
    setCheckout(`${to}/`)
    setSelected(0)
  }
  const goIn = (index: number) => {
    const folder = shown[index]
    if (folder && !folder.registered && dir !== null) {
      go(join(dir, folder.name))
    }
  }
  const goUp = () => {
    const parent = listing?.folders?.parent
    if (parent) {
      go(parent)
    }
  }

  // the field: Down into the list, Tab to complete the name being typed.
  const onFieldKey = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'ArrowDown' && shown.length > 0) {
      event.preventDefault()
      setSelected(0)
      listRef.current?.focus()
    } else if (event.key === 'Tab' && !event.shiftKey && prefix !== '' && shown.length > 0 && !shown[pick].registered && dir !== null) {
      event.preventDefault()
      go(join(dir, shown[pick].name))
    }
  }
  // the list: arrows move, Enter or Right goes in, Left or Backspace goes up,
  // Up from the first row goes back to the field.
  const onListKey = (event: KeyboardEvent<HTMLUListElement>) => {
    switch (event.key) {
      case 'ArrowDown':
        event.preventDefault()
        setSelected(Math.min(pick + 1, shown.length - 1))
        break
      case 'ArrowUp':
        event.preventDefault()
        if (pick === 0) {
          checkoutRef.current?.focus()
        } else {
          setSelected(pick - 1)
        }
        break
      case 'Enter':
      case 'ArrowRight':
        event.preventDefault()
        goIn(pick)
        break
      case 'ArrowLeft':
      case 'Backspace':
        event.preventDefault()
        goUp()
        break
    }
  }

  const onSubmit = async (event: FormEvent) => {
    event.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      onRegistered(await createProject(shownName, trimmed(checkout)))
    } catch (failure) {
      setError(failure instanceof Error ? failure.message : String(failure))
      setSubmitting(false)
    }
  }

  const folders = listing?.dir === dir ? listing.folders : null
  return (
    <dialog ref={dialogRef} onClose={onClose} className="register-dialog">
      <form onSubmit={(event) => void onSubmit(event)}>
        <div className="dialog-head">
          <h2>Register a project</h2>
          <IconButton icon="close" label="Cancel" onClick={onClose} />
        </div>
        <label>
          Checkout
          <input
            ref={checkoutRef}
            required
            placeholder="~/path/to/checkout"
            value={checkout}
            spellCheck={false}
            autoComplete="off"
            aria-controls={listId}
            onChange={(event) => {
              setCheckout(event.target.value)
              setSelected(0)
            }}
            onKeyDown={onFieldKey}
          />
        </label>
        <div className="folder-picker">
          <div className="folder-head">
            <IconButton icon="arrow_upward" label="Up a folder (Left or Backspace in the list)" disabled={!folders?.parent} onClick={goUp} />
            <span className="folder-at">{folders ? folders.path : dir ?? ''}</span>
          </div>
          {/* the list stays in the page whatever it holds, so focus in it
              survives going into an empty folder or one being read. */}
          <ul
            ref={listRef}
            id={listId}
            className="folder-list"
            role="listbox"
            tabIndex={0}
            aria-label={`Folders in ${folders?.path ?? dir ?? 'the projects root'}`}
            aria-activedescendant={shown.length > 0 ? `${listId}-${pick}` : undefined}
            onKeyDown={onListKey}
          >
            {listing?.dir === dir && listing.error ? (
              <li role="presentation" className="muted folder-note">
                {listing.error}
              </li>
            ) : !folders ? (
              <li role="presentation" className="muted folder-note">
                Looking…
              </li>
            ) : folders.note ? (
              <li role="presentation" className="muted folder-note">
                {folders.note}
              </li>
            ) : shown.length === 0 ? (
              <li role="presentation" className="muted folder-note">
                {prefix ? 'No folder here starts with that.' : 'No folders here.'}
              </li>
            ) : (
              shown.map((folder, index) => (
                <li
                  key={folder.name}
                  id={`${listId}-${index}`}
                  role="option"
                  aria-selected={index === pick}
                  aria-disabled={folder.registered ? true : undefined}
                  className={`${index === pick ? 'selected' : ''}${folder.registered ? ' registered' : ''}`}
                  title={folder.registered ? `Already registered as ${folder.registered}` : `Look in ${folder.name}`}
                  onPointerDown={(event) => event.preventDefault()}
                  onClick={() => {
                    setSelected(index)
                    goIn(index)
                  }}
                >
                  <span className="folder-name">{folder.name}</span>
                  {folder.git && <span className="tag">git</span>}
                  {folder.registered && <span className="muted folder-registered">registered as {folder.registered}</span>}
                </li>
              ))
            )}
          </ul>
          {folders?.stopped && <p className="muted folder-note">Listed only the first 2,000 folders here; type more of a name to narrow them.</p>}
        </div>
        <label>
          Name
          <input
            placeholder="Defaults to the folder's name"
            value={shownName}
            onChange={(event) => {
              setName(event.target.value)
              setNamed(true)
            }}
          />
        </label>
        {registeredAs && <p className="error">This folder is already registered as {registeredAs}.</p>}
        {error && <p className="error">{error}</p>}
        <div className="actions">
          <IconButton type="submit" icon="create_new_folder" label="Register" disabled={submitting || !!registeredAs} />
        </div>
      </form>
    </dialog>
  )
}

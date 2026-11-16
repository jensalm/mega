import type { PathSearch } from '../mention'

// PathList is the reply box's @ list, above the caret: the checkout's paths
// matching what follows the @, or, for ~/ or /, the paths under the
// directories the operator named beyond it, newest first, the picked one
// marked. a press on an entry picks
// it without taking focus from the box.
export function PathList({ search, id }: { search: PathSearch; id: string }) {
  const { mention, paths, error, selected, outside } = search
  if (!mention || !search.ready) {
    return null
  }
  return (
    <div className="path-list" style={{ left: `max(0px, min(${mention.left}px, calc(100% - var(--path-list-width))))`, bottom: `calc(100% - ${mention.top}px + 12px)` }}>
      {error ? (
        <p className="path-note">Could not search {outside ? 'the named directories' : 'the checkout'}: {error}</p>
      ) : paths.length === 0 ? (
        <p className="path-note">{mention.query ? 'No matching paths' : 'No paths'}</p>
      ) : (
        <ul role="listbox" id={id} aria-label={outside ? 'Paths in the directories named beyond the checkout' : 'Paths in the checkout'}>
          {paths.map((path, index) => {
            const cut = path.lastIndexOf('/', path.length - 2) + 1
            return (
              <li
                key={path}
                id={`${id}-${index}`}
                role="option"
                aria-selected={index === selected}
                className={index === selected ? 'selected' : undefined}
                onPointerDown={(event) => event.preventDefault()}
                onPointerMove={() => index !== selected && search.hover(index)}
                onClick={() => search.pick(path)}
              >
                <span className="path-dir">{path.slice(0, cut)}</span>
                <span className="path-base">{path.slice(cut)}</span>
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}

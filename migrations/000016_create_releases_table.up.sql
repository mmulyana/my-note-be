CREATE TABLE releases (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    version      text NOT NULL DEFAULT '',
    title        text NOT NULL,
    tag          text NOT NULL DEFAULT 'new',
    summary      text NOT NULL DEFAULT '',
    content      text NOT NULL DEFAULT '',
    image        text NOT NULL DEFAULT '',
    published_at timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT releases_tag_check CHECK (tag IN ('new', 'improved', 'fixed'))
);

-- note: the widget reads newest-first and never filters, so one index carries
-- every query. NULL published_at means a draft and sorts out of the way.
CREATE INDEX releases_published_at_idx ON releases (published_at DESC NULLS LAST);

-- note: seed rows so the widget has something to show on a fresh database.
-- `content` is editor HTML, the same shape a note stores, so these entries can
-- be re-edited later with the ordinary rich editor.
INSERT INTO releases (version, title, tag, summary, content, published_at) VALUES
(
    '1.4.0',
    'Today page',
    'new',
    'Every day now opens on its own page: a daily note on top and the tasks due that day right under it, so the first thing you see is what the day actually asks of you.',
    $html$<p>Every day now opens on its own page: a daily note on top and the tasks due that day right under it, so the first thing you see is what the day actually asks of you.</p><ul><li><p>A daily note is created the first time you open a date — nothing to set up.</p></li><li><p>Tasks pulled from every note, grouped by the note they live in.</p></li><li><p>Arrow keys move between days; deadlines follow along.</p></li></ul>$html$,
    '2026-09-10T09:00:00Z'
),
(
    '1.3.2',
    'Task deadlines in the editor',
    'improved',
    'Give any checklist item a date without leaving the editor. Overdue items turn red in the note, on the card and on the todos page.',
    $html$<p>Give any checklist item a date without leaving the editor. Overdue items turn red in the note, on the card and on the todos page.</p><ul><li><p>A date chip sits at the end of each task row.</p></li><li><p>Cards in the grid show the deadline in their preview.</p></li><li><p>Finished tasks dim their chip instead of shouting at you.</p></li></ul>$html$,
    '2026-08-28T09:00:00Z'
),
(
    '1.3.0',
    'Folders and secret notes',
    'new',
    'Group notes into folders, colour them, and mark a folder secret so its notes stay out of search and out of the grid until you open it.',
    $html$<p>Group notes into folders, colour them, and mark a folder secret so its notes stay out of search and out of the grid until you open it.</p><ul><li><p>Drag a note onto a folder tile to move it.</p></li><li><p>Secret folders are hidden from search results.</p></li><li><p>Folder colours carry over to the note cards inside.</p></li></ul>$html$,
    '2026-08-14T09:00:00Z'
),
(
    '1.2.4',
    'Faster editor, fewer lost keystrokes',
    'fixed',
    'Autosave no longer fights with typing. Long notes stay smooth past a few thousand words, and a failed save now tells you instead of silently retrying.',
    $html$<p>Autosave no longer fights with typing. Long notes stay smooth past a few thousand words, and a failed save now tells you instead of silently retrying.</p><ul><li><p>Saves are debounced and batched per note.</p></li><li><p>A save failure shows a banner with a retry.</p></li><li><p>Image uploads no longer block the editor while they finish.</p></li></ul>$html$,
    '2026-07-30T09:00:00Z'
);

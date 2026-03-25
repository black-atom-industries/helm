# Visual Testing

Capture and verify helm UI screenshots after code changes.

## Quick Reference

```bash
scripts/test-visual.sh sessions    # Capture session view
scripts/test-visual.sh bookmarks   # Capture bookmarks view
scripts/test-visual.sh projects    # Capture projects view
scripts/test-visual.sh clone       # Capture clone view
scripts/test-visual.sh all         # Capture all views
scripts/test-visual.sh clean       # Remove screenshots
```

Screenshots land in `.screenshots/<mode>.png` (gitignored).

## When To Use

Run visual tests after any change to:

- View rendering (`viewSessionList`, `viewBookmarks`, `viewPickDirectory`, `viewClone*`)
- Layout logic (`renderWithSidebar`, sidebar, section box, footer)
- Styles (`styles.go`, colors, borders)
- Row rendering (`columns.go`, `RenderSessionRow`, etc.)

## How To Verify

1. Run the script for the relevant mode(s)
2. Read the screenshot with the Read tool
3. Check: sidebar box borders intact, footer at bottom, content aligned, no clipping

## The `--initial-view` Flag

The script uses `helm --initial-view <mode>` to launch directly into a view without code changes. Valid modes: `sessions` (default), `bookmarks`, `projects`, `clone`.

This flag is also available for production use (e.g., tmux bindings that open directly to bookmarks).

## Environment Overrides

| Var                | Default | Purpose                        |
| ------------------ | ------- | ------------------------------ |
| `HELM_TEST_WIDTH`  | 150     | Popup width                    |
| `HELM_TEST_HEIGHT` | 40      | Popup height                   |
| `HELM_TEST_DELAY`  | 0.8     | Seconds to wait before capture |

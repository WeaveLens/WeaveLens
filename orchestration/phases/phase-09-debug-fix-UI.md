The Phase 09 implementation does not currently satisfy
the responsive layout requirements.

Do NOT rewrite Phase 09.

Perform a root-cause investigation of the existing layout.

The current observed problems are:

1. The UI does not expand when the viewport is larger than Full HD.
2. The Graph does not consume the additional available width.
3. The Resource panel may show a scrollbar when empty.
4. Increasing browser zoom causes the Resource panel and/or Sidebar
   to become clipped or disappear.

Inspect the actual DOM and CSS/layout hierarchy.

Trace:

Viewport
→ body
→ #root
→ App
→ Header
→ Workspace
→ Sidebar
→ Graph
→ Resource Panel

Check for:

- fixed width
- fixed height
- max-width
- max-height
- min-width
- min-height
- flex-shrink
- flex-grow
- flex-basis
- grid sizing
- overflow
- position
- viewport units
- parent sizing constraints

Find the ROOT CAUSE before modifying the code.

The expected behavior is:

1920×1080:
Sidebar + Graph + Resource

2560×1440:
Sidebar + larger Graph + Resource

3840×2160:
Sidebar + significantly larger Graph + Resource

The Sidebar and Resource panel should remain bounded,
while the Graph is the primary flexible region.

Do not make Sidebar or Resource unnecessarily wider
just because the viewport becomes larger.

The Graph must expand into the remaining available space.

An empty Resource panel must not create a scrollbar.

When the viewport becomes too narrow because of browser zoom,
progressively adapt the layout instead of clipping functionality.

Do not solve the issue by blindly adding overflow:hidden.

Do not hide components merely to make the layout fit.

Do not hard-code a 1920px layout.

Validate the fix at:

1920×1080
2560×1440
3840×2160

and browser zoom:

100%
125%
150%
175%
200%

After fixing, explain the root cause and list the files/components
that were changed.

Run the project's frontend tests, build and lint.

Create one focused commit:

fix(web): fix responsive workspace layout
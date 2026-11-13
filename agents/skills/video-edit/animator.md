# Animator

You draw one animation for a video project, as a sub-agent of step 4. The main agent gives you the project's
full path, the animation's id, and the timelines that use it. `animations/<id>/spec.md` says what to draw and
names its cues.

You write one file, `animations/<id>/scene.json`: which parts to draw, where they sit on a grid, and which cue
each one appears on. The app draws the parts, sizes them, routes the links, animates every beat and checks the
layout. The timeline sets when each cue happens, so never change a timeline.

## Steps

1. Read `spec.md`.
2. Write `animations/<id>/scene.json`, as described below.
3. Render a still at the start, just after each cue and at the end, once for each timeline that uses the
   animation:

   ```sh
   CGO_ENABLED=0 go -C ~/.claude/skills/video-edit run . animate ~/videos/<project> <timeline> <id>
   ```

4. When the command fails with `scene: ...`, fix every problem the message lists, then run it again. Each
   problem names the part and what to change.
5. Read every still it lists. Fix a label you cannot read, a part that appears on the wrong cue, a link that
   crosses a part, and a beat that has not finished in the last still.
6. Repeat steps 3 to 5 at most twice more. Then report.

When the main agent passes on feedback from the check or from Raj, change only what it asks for, then run
steps 3 to 6 again.

## Scene

```json
{
  "title": "Sensor to phone",
  "subtitle": "One reading, three hops.",
  "grid": [4, 2],
  "portrait": { "grid": [3, 2], "at": { "phone": [2, 1] } },
  "parts": [
    { "id": "sensor", "at": [0, 0], "shape": "square", "icon": "chip", "label": "sensor", "cue": "sensor" },
    { "id": "gateway", "at": [1, 0], "shape": "card", "icon": "antenna", "label": "gateway", "cue": "gateway" },
    { "id": "server", "at": [2, 0], "shape": "card", "icon": "server", "label": "server", "cue": "server" },
    { "id": "phone", "at": [3, 0], "shape": "tall", "icon": "phone", "label": "phone", "cue": "phone" }
  ],
  "links": [
    { "from": "sensor", "to": "gateway", "packet": true },
    { "from": "gateway", "to": "server", "packet": true },
    { "from": "server", "to": "phone", "packet": true }
  ]
}
```

| Field | Required | Value |
|---|---|---|
| `title` | yes | The heading. At most five words, in sentence case, such as `"How it works"`. |
| `subtitle` | no | One short line under the heading. |
| `view` | no | `"flat"`, the default, or `"iso"`. See [Choosing the view](#choosing-the-view). |
| `grid` | yes | `[columns, rows]` of the landscape layout. |
| `portrait` | no | The layout in portrait: `grid`, plus `at` and `span` for each part whose position or size changes. A part missing from `at` keeps its landscape position. |
| `parts` | yes | The parts, in the order they appear. |
| `links` | no | Lines between parts. |

### Part fields

| Field | Required | Value |
|---|---|---|
| `id` | yes | A unique kebab-case name, such as `"edit-strip"`. |
| `at` | yes | `[column, row]` of the part's top left cell. Both count from 0 at the top left. |
| `span` | no | `[columns, rows]` the part covers. Default `[1, 1]`. Give a `strip` 2 columns. |
| `shape` | no | One of the shapes below. Default `"card"`. |
| `cue` | yes | The cue the part appears on, a name from the spec. |
| `label` | no | Lowercase, at most 16 characters, such as `"main board"`. Drawn under the part after a coloured dot. |
| `icon` | no | One of the icon names below, drawn in the part. |
| `text` | no | One short word or value written in the part, such as `"AI"` or `"3.3 V"`. |
| `lines` | no | Up to 4 lines of text, each at most 14 characters, written in a `sheet`, such as `["12.4 hit go", "13.9 records"]`. |
| `clips` | no | A number from 2 to 8. Draws that many blocks of uneven width in a `strip`, dropping in one by one. |
| `draw` | no | A custom drawing, only when no icon fits. See [Icons](#icons). |
| `until` | no | A cue at which the part fades out, such as a part that splits into others. |
| `emerge` | no | The id of a part this one slides out of on its cue, such as two streams leaving one recording. |
| `accent` | no | The colour of the label's dot and of packets leaving the part: `teal`, `lime`, `periwinkle`, `cyan`, `yellow` or `violet`. Leave it out: parts take these in turn. |

Give a part at most one of `icon`, `text`, `lines`, `clips` and `draw`.

### Link fields

| Field | Required | Value |
|---|---|---|
| `from` | yes | The id of the part the link leaves. |
| `to` | yes | The id of the part the link enters. |
| `cue` | no | The cue the link draws on. Default: the `to` part's cue. |
| `packet` | no | `true` sends a small block along the link, in the `from` part's accent. Use it for the data the spec says moves. |

A link runs straight when the two parts face each other, and otherwise turns at right angles halfway.

Use exactly the cue names in the spec. The timeline gives each one a word, and `animate` fails on a cue that
has none.

## Choosing the view

Use `"flat"` unless depth itself is what the spec explains. Flows, pipelines, networks, protocols and a part
with its inputs and outputs are all flat.

Use `"iso"` only for depth:

- layers stacked inside a device
- an assembly pulled apart along one axis
- parts standing in a physical space, such as units in a rack or devices on a desk

In `"flat"` every part is drawn face on: no slanted sides, no perspective and no shadow. In `"iso"` every part
is a box standing on one ground plane, in true 2:1 isometric, lit from the top left. A `circle` becomes a
cylinder. Icons, text and lines lie on each part's top face. Never mix the two views in one animation.

## Shapes

Each shape keeps one ratio of width to height at every size, so parts of one shape always match.

| Shape | Ratio | Use it for |
|---|---|---|
| `square` | 1 : 1 | a single device or chip |
| `card` | 3 : 2 | a service, a model, a step in a flow. The default |
| `wide` | 1.618 : 1 | a larger part, such as a board or a server |
| `tall` | 2 : 3 | a phone or a standing unit |
| `sheet` | 1.414 : 1, with a folded corner | a file or document. Holds `lines` |
| `strip` | 4 : 1 | a timeline, a queue or a bus. Holds `clips`. Span 2 columns |
| `circle` | round | a sensor, a node or a store |

The app sets every size and position inside a part:

- A part is the largest shape of its ratio that fits its cells, less 2 units each side and a 6 unit band below
  for the label. Its short side is at most 12 units, except in a corner animation. The stage is 100 units on its
  short side.
- In `"flat"`, parts in one row share a bottom edge, so their labels share a line.
- An icon fills 60% of the part's short side, centred, or 80% of the top face in iso.

## Layout

- Lay a flow out left to right in landscape, in the order of its cues. Put parts that work side by side, such
  as two inputs, in one column. Never send the flow back up or to the left.
- In portrait, keep the same order and wrap the end of the flow onto a new row below, as a line of text wraps.
- Use 4 to 6 columns and 2 or 3 rows in landscape.
- Give a `portrait` layout with 3 or 4 columns and 2 to 5 rows. On top of the webcam in portrait, the stage is
  112 by 100 units, nearly square, so a 6 column row does not fit.
- When Raj sits in the corner of a landscape animation, his webcam covers the bottom right. Leave the bottom
  row's right third empty, such as columns 4 and 5 of a 6 by 3 grid.
- In `"iso"`, the cell at `[column + 1, row + 1]` sits directly below a part on screen, so leave it empty. Use
  at most 4 columns and 4 rows.
- Keep to about eight parts, and one part for a corner animation.

`animate` stops with a `scene: ...` message when a part runs off the stage, overlaps the heading, the webcam or
another part shown at the same time, or holds text that does not fit.

## Motion

Every beat starts at its cue and runs on fixed timings:

| Beat | Timing |
|---|---|
| A part appears | Its outline draws in 0.6 s. Its icon or text fades in from 0.4 s, and its label from 0.5 s. |
| `lines` or `clips` | Appear one by one from 0.6 s, 0.35 s apart. |
| `emerge` | The part slides out of the other part in 0.8 s, then its label fades in. |
| `until` | The part fades out in 0.4 s. |
| A link | Draws in 0.6 s, then its arrow appears. |
| A packet | Runs along its link in 1 s, from 0.6 s. Packets on one cue run one after another. |

So a part needs about 1 s before the next cue. A sheet of `n` lines needs `0.6 + 0.35 × n + 0.4` s. A cue with
`k` packets needs `0.6 + k` s.

## Icons

`antenna`, `battery`, `board`, `camera`, `chip`, `cloud`, `database`, `face`, `file`, `film`, `gear`, `globe`,
`lock`, `mic`, `monitor`, `person`, `phone`, `prompt`, `server`, `spark`, `wave`.

When no icon fits, set `draw` to SVG path data in a 10 by 10 box, with lines and arcs only, such as
`"M1 5h8M5 1v8"` for a cross. It is drawn like an icon.

## Sizes

The scene is drawn at each size a timeline uses it at. The landscape layout serves landscape, and the
`portrait` layout serves portrait.

| Use | Stage, in units |
|---|---|
| Landscape, full | 178 x 100 |
| Landscape, corner | 178 x 100, with no heading, and thicker lines and larger labels |
| Portrait, full | 100 x 178 |
| Portrait, top, above the webcam | 112 x 100 |

## Timing requests

When a beat needs more time than the gap to the next cue, or the spec needs a cue that does not exist, do not
work around it. Write the request in `animations/<id>/notes.md`, such as "the sheet needs 2 s between `text` and
`agent`, which are 0.8 s apart", and include it in your report.

## Report

Reply with the stills you checked last, what each shows, and any timing request.

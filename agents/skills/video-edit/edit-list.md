# Edit list

`edit/landscape.json` and `edit/portrait.json` each hold one edit of the
project. A variant, such as `edit/portrait-five-minute-fix.json`, holds another
cut in the format before the dash. `check` reads one and prints its timeline,
and `render` turns it into `out/<format>.mp4`, where `<format>` is the edit
list's name without `.json`.

```json
{
  "segments": [
    {"source": "recording-1.mkv", "start": 2.10, "end": 14.80, "layout": "webcam",
     "title": "Ada Lovelace · Engineer · Example Ltd"},
    {"source": "recording-1.mkv", "start": 15.30, "end": 61.00, "layout": "screen+webcam"},
    {"source": "broll/IMG_0412.MOV", "start": 3.00, "end": 9.00,
     "title": "Demo day · Copenhagen · 5th May 2026"},
    {"source": "recording-1.mkv", "start": 64.20, "end": 90.50, "layout": "webcam"}
  ],
  "cutaways": [
    {"source": "photos/IMG_0420.HEIC", "over": "recording-1.mkv", "from": 20.40, "to": 23.10},
    {"source": "broll/IMG_0413.MOV", "start": 12.00, "over": "recording-1.mkv", "from": 30.00, "to": 34.50},
    {"source": "animations/data-flow", "over": "recording-1.mkv", "from": 40.10, "to": 58.00,
     "layout": "inset", "cues": {"device": 41.30, "server": 47.85, "phone": 52.10}}
  ]
}
```

## Read first

1. `brief.md`. It overrides every default here.
2. Every `.analysis/*/transcript.txt`, one phrase per line with its start and
   end. Take exact cut times from the matching `transcript.json`.
3. `.analysis/descriptions.txt`. Each `==` line gives a source's kind, size,
   length, and when and where it was made. Each line under it describes one
   frame and lists its subjects in brackets.

Open a frame with Read only when its description leaves a decision unclear,
such as whether code is readable or what a photo shows.

## Segments

Segments play in order. `start` and `end` are seconds in `source`, a path in
`.sources/`.

A recording segment takes a `layout`:

| Layout | Landscape | Portrait | Use it when |
|---|---|---|---|
| `webcam` | the webcam | the webcam | Raj talks to the camera or shows hardware |
| `screen` | the screen | the screen | the screen is the whole point, such as code too small to read with the webcam beside it |
| `screen+webcam` | the screen, with the webcam in the bottom right corner | the screen above the webcam | Raj talks over what the screen shows |

A b-roll segment plays with its own sound. Use one for a clip whose sound
matters, such as a speaker at an event. It takes no layout, unless its picture
holds a screen and a webcam side by side, such as a webinar export with Raj's
camera beside his slides. Then it takes a layout like a recording, plus a
region for each picture the layout shows: `screen` and `webcam`, each
`[left, top, right, bottom]` in thousandths of the picture. Measure them on a
frame in `.analysis/`. Set them on every segment, and change them where the
picture's layout changes, such as when a talk moves to its Q&A:

```json
{"source": "broll/talk.mp4", "start": 742.20, "end": 745.90, "layout": "screen+webcam",
 "screen": [235, 243, 973, 981], "webcam": [32, 411, 222, 711]}
```

Framing uses the subjects `analyse` found inside each region, so regions need
no new analysis. A photo is never a segment.

## Cutting

Cut on word boundaries from the source's `transcript.json`. Keep every cut at
least 0.15 s from any word, so no word is clipped and none of the neighbouring
word leaks in. Where two words sit closer than 0.3 s, cut elsewhere. `check`
warns about a cut inside or within 0.15 s of a word. Cut filler words, false
starts, long silences, and every take but the last of a repeated line.

- Tell one story, in the order the brief asks for or the order Raj told it.
- Before rendering, read the words of every segment in order as one script.
  `check` prints the first and last words of each. It must read as Raj telling
  one story: each segment follows from the one before, and no sentence ends
  mid-thought, such as on "connected to their,".
- Change layout only when what matters changes, such as from talking to a demo
  on the screen. Switching back and forth is jarring, so hold each layout for
  at least 15 s. `check` warns about a shorter one. A long talking shot is
  fine. Vary it with b-roll or a photo, not by switching layout.
- Put b-roll and photos where Raj's words are about them: an event over the
  sentence about the event, a photo of a board over the sentence that names it.
  Use a b-roll segment instead of a cutaway when its own sound matters.

## Landscape

The landscape edit is the full video. It keeps everything the brief needs.

## Portrait

The portrait edit is a fresh cut for Shorts and Reels, not a crop of the
landscape edit:

- Keep it to the length the brief gives. Otherwise aim for 30 to 60 s, but
  let a story that needs it run to 2 minutes rather than stop short.
- Open on the most striking line or picture within the first 2 s. It does not
  have to come first in the recording.
- Keep one idea. Drop asides, setup and anything the hook does not need.
- End on the payoff: the point the hook promises, such as the result, the
  number or the lesson. Never end on setup, a problem without its answer, or a
  list that trails off. When the payoff comes after the passage you picked,
  extend the cut to it and drop setup earlier instead.
- Cut tighter than landscape: no asides, and shorter pauses between sentences.
- Prefer `webcam` and `screen+webcam`. Use `screen` alone only when the screen
  has subjects to frame, because code at phone width is unreadable. For the
  same reason, give b-roll of a screen a `box` around the part worth reading,
  or leave it out.

When Raj asks for clips or Shorts, make only portrait variants, one per clip,
all in one project, such as `edit/portrait-five-minute-fix.json`. Make no
`landscape.json` or `portrait.json` unless Raj also asks for a full video.

## Cutaways

A cutaway covers the picture of the segments under it while their sound plays
on. `over`, `from` and `to` name the stretch of a source it covers, in that
source's seconds. Use the times of the words it belongs with, so a recut keeps
the cutaway on those words. A cutaway may span a cut. Cutaways never overlap.

- **B-roll**: `start` is the first second of the clip to show. Its own sound
  plays quietly under the narration.
- **Photo**: show it for 2 to 4 s. It drifts slowly towards its subjects, or
  pushes in on its centre when it has none.
- **Animation**: see below.

`layout` places the cutaway:

- `full`, the default, covers the whole picture.
- `inset` keeps the webcam in view: in the bottom right corner in landscape,
  and below the cutaway in portrait. It can only cover recording segments.
- `corner` shows the cutaway small in the top right corner, and leaves the
  picture under it as it is, such as Raj full screen.

## Animations

An animation is a line drawing of a system Raj explains, such as a device, a
server and a phone with data flowing between them. Add one wherever Raj
explains something technical that a drawing shows better than words: how data
flows between parts, what is inside a device, or how a protocol, algorithm or
mechanism works. Add none while Raj shows something physical, such as a board
held up to the camera, or works on the screen. There the picture is the
point, so keep the viewer on it. The brief overrides this.

Pick the layout from what the viewer should watch:

| Layout | Use it when |
|---|---|
| `full` | the drawing is the point and Raj's face adds nothing, such as a system with many parts |
| `inset` | the drawing leads, and Raj's face keeps the explanation personal |
| `corner` | Raj or the screen stays the point, and a small drawing of one thing helps, such as a part he names while talking to the camera |

An animator draws it from your spec alone, so write the spec for someone who
has not heard the recording.

1. Create `animations/<id>/spec.md`, with a short kebab-case id such as
   `data-flow`. Name the layout. Describe what to draw: each part, the links
   between them, and what moves along the links. List the cues in order, each
   with its name, the words that trigger it, and its time. Keep it to about
   eight parts, or one for a corner animation.
2. Add a cutaway with `"source": "animations/<id>"`, 8 to 30 s long, whose
   `from` and `to` cover the explanation. Its `cues` give each beat's time in
   `over`: the start of the word that names it. Cue names are short kebab-case
   words, such as `device` or `upload`. The animation reads each cue as the CSS
   property `--cue-<name>`.

The portrait edit reuses the animation with the same cue names when its passage
is in the cut.

## Framing

`render` frames every shot from the subjects that `analyse` found in its
frames:

- Consecutive segments of one source, layout and regions form a run, which
  keeps one fixed framing, so jump cuts never move the camera. The crop holds
  every subject seen during the run, such as Raj and a board held up partway
  through, and never zooms past 1.2 times. When the subjects range wider than
  the panel, as in a portrait crop, it centres on where they usually are.
- When fewer than half of a run's frames have subjects, it shows the whole
  picture. Where the picture's shape differs from the panel's, it sits over a
  blurred copy of itself.
- A b-roll or photo cutaway is framed the same way, on its own.

A segment that sets `frame` or `box` starts a new run, and the setting applies
to the whole run:

- `"frame": "auto"` reframes at a change of topic. Use it rarely.
- `"frame": "full"` shows the whole picture.
- `"frame": "push"` pushes in by 5% across the run, eased at both ends, so it is
  barely noticed. Use it only on a run of at least 10 s that builds up to
  something, such as the lead-up to a result. A cutaway never takes it.
- `"box": [left, top, right, bottom]`, in thousandths of the picture, shows
  that region, widened around its centre to the panel's shape. In a
  `screen+webcam` segment it frames the screen.

A subject that lasts only a few seconds, such as a small settings window, is
framed only when it fills most of a run. Give it a run of its own with
`"frame": "auto"`, and set `"frame": "auto"` again on the segment after it.

### Cropping the screen

When the screen shows more than Raj is working in, such as other windows, a
panel or empty desktop, crop it out with a `box` around the window or region
he uses. Open a screen frame from `.analysis/<source>/frames/` to measure it.

Keep one crop for the whole passage. Set the box on the first segment of each
screen run, and change it only when Raj moves to another window. The box is
for a clean frame, not for zooming in and out on details.

## Titles and captions

A segment or cutaway takes a `title`: one short line, its parts joined with
` · `, such as `"Ada Lovelace · Engineer · Example Ltd"`. It fades in 0.3 s
after the shot starts and shows for 6 s, in the bottom left corner in
landscape and above the captions in portrait. Keep a portrait title under
about 40 characters, so it fits on one line.

Title Raj's first shot `Raj Nakarja · Founder · Silicon Witchery`, unless the
brief gives another title or asks for none. Title b-roll of an event with the
event, place and date. Take every other fact from Raj, the brief, or a
source's details in `.analysis/descriptions.txt`. Never invent a name, role,
place or date. When a fact is missing, leave the title out.

Portrait gets captions from the transcript automatically. Landscape has none,
because YouTube captions it.

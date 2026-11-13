"use strict"

// Builds a video-edit animation from a scene description. animator.md documents every field. The renderer
// pauses every Web Animation on the page at each frame's time, so all motion here uses element.animate().
// CSS transitions and timers never render.

// Seconds
const timing = { draw: 0.6, content: 0.4, label: 0.5, fade: 0.4, stagger: 0.35, packet: 1, move: 0.8 }
const easeInOut = "cubic-bezier(0.65, 0, 0.35, 1)"
const easeOut = "cubic-bezier(0.33, 1, 0.68, 1)"

// Width to height, or width to depth in the iso view
const ratios = { square: 1, card: 1.5, wide: 1.618, tall: 2 / 3, sheet: 1.414, strip: 4, circle: 1 }

const accents = ["teal", "lime", "periwinkle", "cyan", "yellow", "violet"]

// Line icons, drawn in a 10 by 10 box
const icons = {
  antenna: "M5 5v4.5M3.4 3.4a2.26 2.26 0 0 0 0 3.2M6.6 3.4a2.26 2.26 0 0 1 0 3.2M2.2 2.2a3.96 3.96 0 0 0 0 5.6M7.8 2.2a3.96 3.96 0 0 1 0 5.6",
  battery: "M1 3h7.5v4h-7.5zM8.5 4.2h1v1.6h-1M2.4 4.2v1.6M3.8 4.2v1.6M5.2 4.2v1.6",
  board: "M0.5 2h9v6h-9zM3.5 3.5h3v3h-3zM1.3 3h0.6M8.1 3h0.6M1.3 7h0.6M8.1 7h0.6",
  camera: "M0.5 2.5h6v5h-6zM6.5 4.2l3-1.5v4.6l-3-1.5",
  chip: "M2.5 2.5h5v5h-5zM3.8 2.5v-1.5M5 2.5v-1.5M6.2 2.5v-1.5M3.8 7.5v1.5M5 7.5v1.5M6.2 7.5v1.5M2.5 3.8h-1.5M2.5 5h-1.5M2.5 6.2h-1.5M7.5 3.8h1.5M7.5 5h1.5M7.5 6.2h1.5",
  cloud: "M2.6 7.5a2 2 0 0 1 0.1-4a2.8 2.8 0 0 1 5.3 0.6a1.7 1.7 0 0 1-0.3 3.4z",
  database: "M1.5 2.5a3.5 1.3 0 1 0 7 0a3.5 1.3 0 1 0-7 0M1.5 2.5v5a3.5 1.3 0 0 0 7 0v-5M1.5 5a3.5 1.3 0 0 0 7 0",
  face: "M1 5a4 4 0 1 0 8 0a4 4 0 1 0-8 0M3.3 4.1h0.6M6.1 4.1h0.6M3.3 6.2q1.7 1.4 3.4 0",
  file: "M2 0.5h4.5l2 2v7h-6.5zM6.5 0.5v2h2M3.5 5h3.5M3.5 7h3.5",
  film: "M1 1.5h8v7h-8zM2.5 1.5v7M7.5 1.5v7M1 3.5h1.5M1 5h1.5M1 6.5h1.5M7.5 3.5h1.5M7.5 5h1.5M7.5 6.5h1.5",
  gear: "M2.8 5a2.2 2.2 0 1 0 4.4 0a2.2 2.2 0 1 0-4.4 0M5 0.8v1.6M5 7.6v1.6M0.8 5h1.6M7.6 5h1.6M2 2l1.1 1.1M6.9 6.9l1.1 1.1M8 2l-1.1 1.1M3.1 6.9l-1.1 1.1",
  globe: "M1 5a4 4 0 1 0 8 0a4 4 0 1 0-8 0M1 5h8M5 1a2 4 0 0 0 0 8a2 4 0 0 0 0-8",
  lock: "M2 4.5h6v5h-6zM3.2 4.5v-1.3a1.8 1.8 0 0 1 3.6 0v1.3M5 6.3v1.5",
  mic: "M3.5 2a1.5 1.5 0 0 1 3 0v3a1.5 1.5 0 0 1-3 0zM2 4.8a3 3 0 0 0 6 0M5 7.8v1.7M3.5 9.5h3",
  monitor: "M1 1.5h8v5.5h-8zM5 7v1.5M3 8.5h4",
  person: "M3 3a2 2 0 1 0 4 0a2 2 0 1 0-4 0M1.5 9.5q0-4 3.5-4q3.5 0 3.5 4",
  phone: "M3 0.5h4v9h-4zM4.4 8.3h1.2",
  prompt: "M0.5 1.5h9v7h-9zM2.2 4l1.5 1l-1.5 1M4.6 6.5h2.6",
  server: "M1 1h8v2.4h-8zM1 3.8h8v2.4h-8zM1 6.6h8v2.4h-8zM2 2.2h0.6M2 5h0.6M2 7.8h0.6",
  spark: "M5 0.5L6.1 3.9L9.5 5L6.1 6.1L5 9.5L3.9 6.1L0.5 5L3.9 3.9Z",
  wave: "M1.5 4v2M3 2.5v5M4.5 3.5v3M6 1.5v7M7.5 3v4M9 4.2v1.6",
}

// settings.cues gives each cue's second from the animation's start. settings.webcam is the box the webcam
// covers, as fractions of the stage, or null.
function scene(description, settings) {
  const problems = []
  const unit = Math.min(innerWidth, innerHeight) / 100
  const stageWidth = innerWidth / unit
  const stageHeight = innerHeight / unit
  const corner = matchMedia("(max-width: 480px)").matches
  const landscape = stageWidth > 150
  const iso = description.view == "iso"

  const cue = name => {
    const seconds = settings.cues[name]
    if (seconds === undefined) {
      problems.push(`no cue "${name}" in the timeline's cues`)
      return 0
    }
    return seconds
  }

  if (!document.body) {
    throw new Error("scene: put <body> before the scripts")
  }

  // Heading, top left
  const header = document.createElement("header")
  header.className = "title"
  const heading = document.createElement("h1")
  heading.textContent = description.title
  header.append(heading)
  if (description.subtitle) {
    const subtitle = document.createElement("p")
    subtitle.textContent = description.subtitle
    header.append(subtitle)
  }
  document.body.append(header)

  const svg = element(document.body, "svg", { class: "stage", viewBox: `0 0 ${stageWidth} ${stageHeight}` })
  const linkLayer = element(svg, "g", {})
  const partLayer = element(svg, "g", {})

  // The drawing area sits below the heading, 6 units in from the stage edges
  const margin = 6
  const headerBox = header.getBoundingClientRect()
  const top = headerBox.height > 0 ? headerBox.bottom / unit + 4 : margin
  const area = { x: margin, y: top, width: stageWidth - 2 * margin, height: stageHeight - top - margin }

  const webcam = settings.webcam
    ? { x: settings.webcam[0] * stageWidth - 2, y: settings.webcam[1] * stageHeight - 2, width: stageWidth, height: stageHeight }
    : null

  // Grid and projection. Grid units: column u and row v, from the grid's top left corner.
  const portrait = !landscape && description.portrait ? description.portrait : {}
  const [columns, rows] = portrait.grid || description.grid
  let project
  let cellWidth = 0
  let cellHeight = 0
  let isoScale = 0
  if (iso) {
    isoScale = Math.min(area.width / (columns + rows), (area.height - 6) / ((columns + rows) / 2 + 0.5))
    const originX = area.x + (area.width - (columns + rows) * isoScale) / 2 + rows * isoScale
    const originY = area.y + 0.5 * isoScale
    project = (u, v, z = 0) => [originX + (u - v) * isoScale, originY + ((u + v) * isoScale) / 2 - z * isoScale]
  } else {
    cellWidth = area.width / columns
    cellHeight = area.height / rows
    project = (u, v) => [area.x + u * cellWidth, area.y + v * cellHeight]
  }

  // Parts
  const parts = new Map()
  description.parts.forEach((part, index) => {
    if (parts.has(part.id)) {
      problems.push(`two parts have id "${part.id}"`)
      return
    }

    const at = (portrait.at || {})[part.id] || part.at
    const span = (portrait.span || {})[part.id] || part.span || [1, 1]
    const shape = part.shape || "card"
    const ratio = ratios[shape]
    if (!ratio) {
      problems.push(`part "${part.id}" has shape "${shape}", which is not ${Object.keys(ratios).join(", ")}`)
      return
    }
    if (at[0] < 0 || at[1] < 0 || at[0] + span[0] > columns || at[1] + span[1] > rows) {
      problems.push(`part "${part.id}" at [${at}] with span [${span}] lies outside the ${columns} by ${rows} grid`)
      return
    }

    // Footprint: centre and half extents in grid units, height in grid units for iso
    let centreU = at[0] + span[0] / 2
    let centreV = at[1] + span[1] / 2
    let halfU
    let halfV
    let height = 0
    if (iso) {
      const width = Math.min(span[0] * 0.62, span[1] * 0.62 * ratio)
      halfU = width / 2
      halfV = width / ratio / 2
      height = shape == "sheet" ? 0.03 : shape == "strip" ? 0.08 : 0.45 * Math.min(2 * halfU, 2 * halfV)
    } else {
      // A part fills its cells less 2 units each side and a 6 unit band below for its label, and its
      // short side stays at most 12 units, except in a corner animation.
      const roomWidth = span[0] * cellWidth - 4
      const roomHeight = span[1] * cellHeight - 8
      let width = Math.min(roomWidth, roomHeight * ratio)
      let height2d = width / ratio
      const largest = corner ? Infinity : 12
      const shrink = Math.min(1, largest / Math.min(width, height2d))
      width *= shrink
      height2d *= shrink
      // Parts in a row share a bottom edge, so their labels share a line
      centreV = at[1] + (2 + roomHeight - height2d / 2) / cellHeight
      halfU = width / 2 / cellWidth
      halfV = height2d / 2 / cellHeight
    }

    parts.set(part.id, { ...part, index, at, span, shape, centreU, centreV, halfU, halfV, height, accent: part.accent || accents[index % accents.length] })
  })

  for (const part of parts.values()) {
    const outer = element(partLayer, "g", {})
    const inner = element(outer, "g", {})
    const shapes = []
    let local
    let labelAnchor

    if (iso) {
      const u0 = part.centreU - part.halfU
      const u1 = part.centreU + part.halfU
      const v0 = part.centreV - part.halfV
      const v1 = part.centreV + part.halfV
      const h = part.height
      const k = isoScale

      if (part.shape == "circle") {
        const [cx, cy] = project(part.centreU, part.centreV, h)
        const [, baseY] = project(part.centreU, part.centreV, 0)
        const rx = part.halfU * k * Math.SQRT2
        const ry = rx / 2
        shapes.push(element(inner, "path", { class: "part left", d: `M${cx - rx} ${cy} V${baseY} A${rx} ${ry} 0 0 0 ${cx + rx} ${baseY} V${cy} A${rx} ${ry} 0 0 1 ${cx - rx} ${cy}Z` }))
        shapes.push(element(inner, "ellipse", { class: "part", cx, cy, rx, ry }))
        const side = part.halfU * Math.SQRT2
        const [ox, oy] = project(part.centreU - side / 2, part.centreV - side / 2, h)
        local = { width: side * k, height: side * k, transform: `matrix(1 0.5 -1 0.5 ${ox} ${oy})` }
        labelAnchor = [cx, baseY + ry]
      } else {
        const corners = [[u0, v0, h], [u1, v0, h], [u1, v1, h], [u0, v1, h], [u1, v0, 0], [u1, v1, 0], [u0, v1, 0]].map(([u, v, z]) => project(u, v, z))
        const [a, b, c, d, bBase, cBase, dBase] = corners
        const polygon = points => "M" + points.map(p => p.join(" ")).join("L") + "Z"
        shapes.push(element(inner, "path", { class: "part right", d: polygon([b, c, cBase, bBase]) }))
        shapes.push(element(inner, "path", { class: "part left", d: polygon([d, c, cBase, dBase]) }))
        shapes.push(element(inner, "path", { class: "part", d: polygon([a, b, c, d]) }))
        local = { width: (u1 - u0) * k, height: (v1 - v0) * k, transform: `matrix(1 0.5 -1 0.5 ${a[0]} ${a[1]})` }
        labelAnchor = cBase
      }
    } else {
      const [x, y] = project(part.centreU - part.halfU, part.centreV - part.halfV)
      const [right, bottom] = project(part.centreU + part.halfU, part.centreV + part.halfV)
      const width = right - x
      const height = bottom - y

      if (part.shape == "circle") {
        shapes.push(element(inner, "circle", { class: "part", cx: x + width / 2, cy: y + height / 2, r: width / 2 }))
        const side = width / Math.SQRT2
        local = { width: side, height: side, transform: `translate(${x + (width - side) / 2} ${y + (height - side) / 2})` }
      } else if (part.shape == "sheet") {
        const fold = 0.22 * height
        shapes.push(element(inner, "path", { class: "part", d: `M${x} ${y}H${right - fold}L${right} ${y + fold}V${bottom}H${x}Z` }))
        shapes.push(element(inner, "path", { class: "part", d: `M${right - fold} ${y}V${y + fold}H${right}` }))
        local = { width: width - fold, height, transform: `translate(${x} ${y})` }
      } else {
        const r = Math.min(1, 0.12 * Math.min(width, height))
        shapes.push(element(inner, "path", {
          class: "part",
          d: `M${x + r} ${y}H${right - r}A${r} ${r} 0 0 1 ${right} ${y + r}V${bottom - r}A${r} ${r} 0 0 1 ${right - r} ${bottom}H${x + r}A${r} ${r} 0 0 1 ${x} ${bottom - r}V${y + r}A${r} ${r} 0 0 1 ${x + r} ${y}Z`,
        }))
        local = { width, height, transform: `translate(${x} ${y})` }
      }
      labelAnchor = [x + width / 2, bottom]
    }

    // Content, laid out in the part's face: the front in 2D, the top in iso
    const content = element(inner, "g", { transform: local.transform })
    const pad = 1.2
    const lines = []
    const clips = []
    if (part.icon || part.draw) {
      const path = part.draw || icons[part.icon]
      if (!path) {
        problems.push(`part "${part.id}" has icon "${part.icon}", which is not ${Object.keys(icons).join(", ")}`)
      } else {
        // The top face of an iso part is foreshortened, so its icon fills more of it
        const size = (iso ? 0.8 : 0.6) * Math.min(local.width, local.height)
        const scale = size / 10
        const group = element(content, "g", { transform: `translate(${(local.width - size) / 2} ${(local.height - size) / 2}) scale(${scale})` })
        element(group, "path", { class: "mark", d: path, style: `stroke-width: ${0.18 / scale}` })
      }
    }
    if (part.text) {
      const text = element(content, "text", { class: "text", x: local.width / 2, y: local.height / 2, "text-anchor": "middle", "dominant-baseline": "central", "font-size": 4 })
      text.textContent = part.text
      const size = Math.min(4, (4 * (local.width - 2 * pad)) / text.getComputedTextLength(), local.height * 0.5)
      if (size < 2) {
        problems.push(`the text of part "${part.id}" does not fit. Shorten it, or give the part a wider shape or span`)
      }
      text.setAttribute("font-size", size)
    }
    if (part.lines) {
      const texts = part.lines.map(line => {
        const text = element(content, "text", { class: "lines", x: pad, "font-size": 1.6 })
        text.textContent = line
        return text
      })
      const widest = Math.max(...texts.map(text => text.getComputedTextLength()))
      const size = Math.min(1.6, (1.6 * (local.width - 2 * pad)) / widest, (local.height - 2 * pad) / (1.5 * texts.length))
      if (size < 1) {
        problems.push(`the lines of part "${part.id}" do not fit. Shorten them, use fewer, or give the part a larger span`)
      }
      texts.forEach((text, i) => {
        text.setAttribute("font-size", size)
        text.setAttribute("y", pad + size + i * 1.5 * size)
        lines.push(text)
      })
    }
    if (part.clips) {
      const widths = [3, 2, 4, 2.5, 3.5, 2, 3, 2.5].slice(0, part.clips)
      const gap = 0.6
      const inset = 0.8
      const scale = (local.width - 2 * inset - gap * (widths.length - 1)) / widths.reduce((sum, width) => sum + width, 0)
      let x = inset
      for (const width of widths) {
        clips.push(element(content, "rect", { class: "clip", x, y: inset, width: width * scale, height: local.height - 2 * inset, rx: 0.3 }))
        x += width * scale + gap
      }
    }

    // Label, centred below the part, after a dot in the part's accent
    let label = null
    if (part.label) {
      label = element(inner, "g", {})
      const [labelX, labelY] = [labelAnchor[0], labelAnchor[1] + 3.4]
      const text = element(label, "text", { class: "label", x: labelX, y: labelY, "text-anchor": "middle" })
      text.textContent = part.label
      element(label, "circle", { class: "dot", cx: labelX - text.getComputedTextLength() / 2 - 1.2, cy: labelY - 0.8, style: `fill: var(--${part.accent})` })
    }

    Object.assign(part, { outer, inner, shapes, content, lines, clips, label, box: inner.getBBox() })
  }

  // Links, routed along the grid with right-angled turns
  const links = description.links || []
  const routes = []
  for (const link of links) {
    const from = parts.get(link.from)
    const to = parts.get(link.to)
    if (!from || !to) {
      problems.push(`link from "${link.from}" to "${link.to}" names a part that does not exist`)
      continue
    }

    // Leaves below a 2D part's label, which sits under the part
    const labelGap = iso ? 0 : 6 / cellHeight
    // Where the two parts' sides face each other, the link runs straight across the middle of that
    // stretch. Otherwise it leaves each part's centre and turns halfway.
    const facing = (centreA, halfA, centreB, halfB) => {
      const low = Math.max(centreA - halfA, centreB - halfB)
      const high = Math.min(centreA + halfA, centreB + halfB)
      return low < high ? [(low + high) / 2, (low + high) / 2] : [centreA, centreB]
    }
    let points
    if (to.at[0] >= from.at[0] + from.span[0] || to.at[0] + to.span[0] <= from.at[0]) {
      const direction = to.at[0] > from.at[0] ? 1 : -1
      const [startV, endV] = facing(from.centreV, from.halfV, to.centreV, to.halfV)
      const start = [from.centreU + direction * from.halfU, startV]
      const end = [to.centreU - direction * to.halfU, endV]
      const turn = (start[0] + end[0]) / 2
      points = startV == endV ? [start, end] : [start, [turn, start[1]], [turn, end[1]], end]
    } else {
      const direction = to.at[1] > from.at[1] ? 1 : -1
      const [startU, endU] = facing(from.centreU, from.halfU, to.centreU, to.halfU)
      const start = [startU, from.centreV + direction * (from.halfV + (direction > 0 ? labelGap : 0))]
      const end = [endU, to.centreV - direction * (to.halfV + (direction < 0 ? labelGap : 0))]
      const turn = (start[1] + end[1]) / 2
      points = startU == endU ? [start, end] : [start, [start[0], turn], [end[0], turn], end]
    }
    points = points.map(([u, v]) => project(u, v, 0))

    // The arrow's tip stops half a unit short of the part
    const [lastX, lastY] = points[points.length - 1]
    const [beforeX, beforeY] = points[points.length - 2]
    const length = Math.hypot(lastX - beforeX, lastY - beforeY)
    const along = [(lastX - beforeX) / length, (lastY - beforeY) / length]
    if (!iso) {
      points[points.length - 1] = [lastX - along[0] * 0.5, lastY - along[1] * 0.5]
    }

    let d = `M${points[0][0]} ${points[0][1]}`
    for (let i = 1; i < points.length - 1; i++) {
      const [px, py] = points[i - 1]
      const [x, y] = points[i]
      const [nx, ny] = points[i + 1]
      const inLength = Math.hypot(x - px, y - py)
      const outLength = Math.hypot(nx - x, ny - y)
      const r = Math.min(1.5, inLength / 2, outLength / 2)
      d += `L${x - ((x - px) / inLength) * r} ${y - ((y - py) / inLength) * r}Q${x} ${y} ${x + ((nx - x) / outLength) * r} ${y + ((ny - y) / outLength) * r}`
    }
    const [tipX, tipY] = points[points.length - 1]
    d += `L${tipX} ${tipY}`

    const path = element(linkLayer, "path", { class: "link", d })
    // Iso links end behind their part, where an arrow would hide
    const arrow = iso ? null : element(linkLayer, "path", {
      class: "link",
      d: `M${tipX - along[0] * 1.1 + along[1] * 0.75} ${tipY - along[1] * 1.1 - along[0] * 0.75}L${tipX} ${tipY}L${tipX - along[0] * 1.1 - along[1] * 0.75} ${tipY - along[1] * 1.1 + along[0] * 0.75}`,
    })
    let packet = null
    if (link.packet) {
      packet = element(linkLayer, "g", { class: "packet", style: `offset-path: path("${d}"); offset-rotate: 0deg` })
      element(packet, "rect", { x: -1.6, y: -1, width: 3.2, height: 2, rx: 0.6, style: `fill: var(--${from.accent})` })
    }
    routes.push({ start: cue(link.cue || to.cue), path, arrow, packet })
  }

  // Layout checks, before any motion moves a part
  const headerUnits = { x: headerBox.left / unit, y: headerBox.top / unit, width: headerBox.width / unit, height: headerBox.height / unit }
  const overlaps = (a, b, clearance) => a.x < b.x + b.width + clearance && b.x < a.x + a.width + clearance && a.y < b.y + b.height + clearance && b.y < a.y + a.height + clearance
  const visible = part => [cue(part.cue), part.until ? cue(part.until) : Infinity]
  const list = [...parts.values()]
  for (const part of list) {
    const { x, y, width, height } = part.box
    if (x < 2 || y < 2 || x + width > stageWidth - 2 || y + height > stageHeight - 2) {
      problems.push(`part "${part.id}" or its label runs off the stage`)
    }
    if (headerBox.height > 0 && overlaps(part.box, headerUnits, 1)) {
      problems.push(`part "${part.id}" or its label overlaps the heading`)
    }
    if (webcam && overlaps(part.box, webcam, 0)) {
      problems.push(`part "${part.id}" or its label lies under the webcam in the bottom right corner. Move it to another cell, or leave that cell empty`)
    }
    if (part.emerge && !parts.has(part.emerge)) {
      problems.push(`part "${part.id}" emerges from "${part.emerge}", which does not exist`)
    }
  }
  for (let i = 0; i < list.length; i++) {
    for (let j = i + 1; j < list.length; j++) {
      const [startA, endA] = visible(list[i])
      const [startB, endB] = visible(list[j])
      if (startA < endB && startB < endA && overlaps(list[i].box, list[j].box, 1)) {
        problems.push(`parts "${list[i].id}" and "${list[j].id}" overlap, with their labels. Spread them apart, or shorten a label`)
      }
    }
  }

  if (problems.length > 0) {
    throw new Error("scene: " + problems.join("; "))
  }

  // Motion
  for (const part of parts.values()) {
    const start = cue(part.cue)
    let contentStart = start + timing.content
    if (part.emerge) {
      const source = parts.get(part.emerge)
      const [fromX, fromY] = project(source.centreU, source.centreV, source.height)
      const [toX, toY] = project(part.centreU, part.centreV, part.height)
      animate(part.inner, [{ translate: `${fromX - toX}px ${fromY - toY}px`, opacity: 0 }, { opacity: 1, offset: 0.25 }, { translate: "0px 0px", opacity: 1 }], start, timing.move, easeInOut)
      contentStart = start
    } else {
      for (const shape of part.shapes) {
        draw(shape, start)
      }
    }

    if (part.lines.length > 0 || part.clips.length > 0) {
      part.lines.forEach((line, i) => animate(line, [{ opacity: 0 }, { opacity: 1 }], start + timing.draw + i * timing.stagger, timing.fade, easeOut))
      part.clips.forEach((clip, i) => animate(clip, [{ translate: "0px -3px", opacity: 0 }, { translate: "0px 0px", opacity: 1 }], start + timing.draw + i * timing.stagger, timing.fade, easeOut))
    } else {
      animate(part.content, [{ opacity: 0 }, { opacity: 1 }], contentStart, timing.fade, easeOut)
    }

    if (part.label) {
      animate(part.label, [{ opacity: 0 }, { opacity: 1 }], (part.emerge ? start + timing.move : start + timing.label), timing.fade, easeOut)
    }
    if (part.until) {
      animate(part.outer, [{ opacity: 1 }, { opacity: 0 }], cue(part.until), timing.fade, easeOut)
    }
  }

  // Packets on one cue run one after another, so one thing moves at a time
  const packetsAt = new Map()
  for (const route of routes) {
    draw(route.path, route.start)
    if (route.arrow) {
      animate(route.arrow, [{ opacity: 0 }, { opacity: 1 }], route.start + timing.draw, timing.fade, easeOut)
    }
    if (route.packet) {
      const order = packetsAt.get(route.start) || 0
      packetsAt.set(route.start, order + 1)
      animate(route.packet, [{ offsetDistance: "0%", opacity: 0 }, { opacity: 1, offset: 0.1 }, { opacity: 1, offset: 0.9 }, { offsetDistance: "100%", opacity: 0 }], route.start + timing.draw + order * timing.packet, timing.packet, easeInOut)
    }
  }
}

function element(parent, name, attributes) {
  const node = document.createElementNS("http://www.w3.org/2000/svg", name)
  for (const [key, value] of Object.entries(attributes)) {
    node.setAttribute(key, value)
  }
  parent.append(node)
  return node
}

function animate(target, keyframes, start, duration, easing) {
  target.animate(keyframes, { delay: start * 1000, duration: duration * 1000, easing, fill: "both" })
}

// Draws a shape like a pen, then fills it. Chromium measures a curved path slightly short, so the dash
// is longer than the path, and the gap after it keeps the next dash off the end of the path.
function draw(shape, start) {
  shape.setAttribute("pathLength", 1)
  shape.setAttribute("stroke-dasharray", "1.05 2")
  animate(shape, [{ strokeDashoffset: 1.05, opacity: 0, fillOpacity: 0 }, { opacity: 1, offset: 0.01 }, { fillOpacity: 0, offset: 0.7 }, { strokeDashoffset: 0, opacity: 1, fillOpacity: 1 }], start, timing.draw, easeInOut)
}

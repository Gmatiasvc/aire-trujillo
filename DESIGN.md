---
name: Aire Trujillo
description: Urban environmental and air-quality telemetry slide rack on an illuminated amber light table for Trujillo and Huanchaco.
colors:
  vermilion-signal: "#b83a28"
  vermilion-signal-hover: "#9d2f20"
  vermilion-ink: "#fff9f2"
  bottle-green: "#265d43"
  bottle-ink: "#f7f1e3"
  amber-gold-fill: "#e7b043"
  amber-ochre-accent: "#8a5607"
  amber-bed-light: "#dda126"
  amber-bed-deep-light: "#c98c19"
  ivory-sheet-light: "#f7f1e3"
  ivory-well-light: "#efe6d2"
  carbon-surface-light: "#191d20"
  slate-surface-light: "#29373e"
  glass-slate-light: "rgba(25, 34, 38, 0.91)"
  ink-primary-light: "#181a1c"
  ink-secondary-light: "#433a2e"
  ink-muted-carbon-light: "#cfc3af"
  rule-soft-light: "#c7b698"
  vermilion-dark: "#bf4128"
  vermilion-dark-hover: "#d35035"
  vermilion-ink-dark: "#fff8f0"
  bottle-green-dark: "#2f6e4f"
  bottle-ink-dark: "#f5ecdb"
  amber-gold-dark: "#e6aa3c"
  amber-gold-fill-dark: "#d99b26"
  amber-bed-dark: "#15120e"
  amber-bed-deep-dark: "#0e0c09"
  ivory-sheet-dark: "#211c16"
  ivory-well-dark: "#2b241c"
  carbon-surface-dark: "#100d0a"
  slate-surface-dark: "#332c24"
  glass-slate-dark: "rgba(17, 14, 11, 0.88)"
  ink-primary-dark: "#f5ecdb"
  ink-secondary-dark: "#cbbca3"
  ink-muted-carbon-dark: "#c2b298"
  rule-dark: "#6b563a"
  rule-soft-dark: "#3d3224"
  pcb-board-light: "#23332d"
  pcb-board-dark: "#1b2420"
  pcb-copper: "#d99b26"
  pcb-silk: "#f1e6cf"
  nodata-gray: "#9ca3af"
  nodata-ink: "#111827"
  status-active-green: "#22c55e"
  saved-series-light: "#166534"
  saved-series-dark: "#4ade80"
typography:
  display:
    fontFamily: "'Barlow Condensed', sans-serif"
    fontSize: "clamp(2.55rem, 4.7vw, 4.15rem)"
    fontWeight: 800
    lineHeight: 0.98
    letterSpacing: "-0.02em"
  headline:
    fontFamily: "'Barlow Condensed', sans-serif"
    fontSize: "clamp(1.9rem, 3.2vw, 2.75rem)"
    fontWeight: 800
    lineHeight: 1.02
    letterSpacing: "0em"
  title:
    fontFamily: "'Barlow Condensed', sans-serif"
    fontSize: "1.625rem"
    fontWeight: 700
    lineHeight: 1.05
    letterSpacing: "0em"
  body:
    fontFamily: "'Archivo', sans-serif"
    fontSize: "1rem"
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: "0em"
  label:
    fontFamily: "'Azeret Mono', monospace"
    fontSize: "0.78125rem"
    fontWeight: 600
    lineHeight: 1.2
    letterSpacing: "0.05em"
    fontFeature: "tabular-nums"
rounded:
  none: "0px"
  pcb-pad: "4px"
  pcb-silk: "6px"
  pcb-board: "10px"
spacing:
  xs: "6px"
  sm: "10px"
  md: "18px"
  lg: "24px"
  xl: "32px"
  "2xl": "42px"
components:
  button-primary:
    backgroundColor: "{colors.vermilion-signal}"
    textColor: "{colors.vermilion-ink}"
    typography: "{typography.label}"
    rounded: "{rounded.none}"
    padding: "12px 22px"
  button-primary-hover:
    backgroundColor: "{colors.vermilion-signal-hover}"
    textColor: "{colors.vermilion-ink}"
  button-secondary:
    backgroundColor: "{colors.ivory-well-light}"
    textColor: "{colors.ink-primary-light}"
    typography: "{typography.label}"
    rounded: "{rounded.none}"
    padding: "12px 22px"
  button-secondary-hover:
    backgroundColor: "{colors.carbon-surface-light}"
    textColor: "{colors.ivory-sheet-light}"
  button-carbon:
    backgroundColor: "{colors.carbon-surface-light}"
    textColor: "{colors.ivory-sheet-light}"
    typography: "{typography.label}"
    rounded: "{rounded.none}"
    padding: "12px 22px"
  button-carbon-hover:
    backgroundColor: "{colors.vermilion-signal}"
    textColor: "{colors.vermilion-ink}"
  nav-slide-tab:
    backgroundColor: "transparent"
    textColor: "{colors.ink-secondary-light}"
    typography: "{typography.label}"
    rounded: "{rounded.none}"
    padding: "6px 11px"
  nav-slide-tab-active:
    backgroundColor: "{colors.carbon-surface-light}"
    textColor: "{colors.ivory-sheet-light}"
  rail-tick:
    backgroundColor: "transparent"
    textColor: "{colors.ink-muted-carbon-light}"
    typography: "{typography.label}"
    rounded: "{rounded.none}"
    padding: "10px 4px 8px"
  rail-tick-active:
    backgroundColor: "{colors.amber-gold-fill}"
    textColor: "{colors.ink-primary-light}"
  slide-sheet:
    backgroundColor: "{colors.ivory-sheet-light}"
    textColor: "{colors.ink-primary-light}"
    typography: "{typography.body}"
    rounded: "{rounded.none}"
    padding: "38px"
  timetable-glass-slate:
    backgroundColor: "{colors.glass-slate-light}"
    textColor: "{colors.ivory-sheet-light}"
    typography: "{typography.label}"
    rounded: "{rounded.none}"
  nodata-chip:
    backgroundColor: "{colors.nodata-gray}"
    textColor: "{colors.nodata-ink}"
    typography: "{typography.label}"
    rounded: "{rounded.none}"
    padding: "5px 12px"
  input-gate:
    backgroundColor: "{colors.ivory-well-light}"
    textColor: "{colors.ink-primary-light}"
    typography: "{typography.label}"
    rounded: "{rounded.none}"
    padding: "10px 12px"
  modal-card:
    backgroundColor: "{colors.ivory-sheet-light}"
    textColor: "{colors.ink-primary-light}"
    typography: "{typography.body}"
    rounded: "{rounded.none}"
    width: "min(680px, 100%)"
  map-topbar:
    backgroundColor: "{colors.carbon-surface-light}"
    textColor: "{colors.ivory-sheet-light}"
    typography: "{typography.label}"
    rounded: "{rounded.none}"
    padding: "8px 16px"
    height: "56px"
  map-std-tab:
    backgroundColor: "transparent"
    textColor: "{colors.ink-muted-carbon-light}"
    typography: "{typography.label}"
    rounded: "{rounded.none}"
    padding: "7px 13px"
  map-std-tab-active:
    backgroundColor: "{colors.vermilion-signal}"
    textColor: "{colors.vermilion-ink}"
  map-layer-select:
    backgroundColor: "{colors.ivory-sheet-light}"
    textColor: "{colors.ink-primary-light}"
    typography: "{typography.label}"
    rounded: "{rounded.none}"
    padding: "0 10px"
    height: "33px"
  map-legend:
    backgroundColor: "{colors.ivory-sheet-light}"
    textColor: "{colors.ink-primary-light}"
    typography: "{typography.body}"
    rounded: "{rounded.none}"
    width: "268px"
  playback-bar:
    backgroundColor: "{colors.ivory-sheet-light}"
    textColor: "{colors.ink-primary-light}"
    typography: "{typography.label}"
    rounded: "{rounded.none}"
    padding: "10px 14px"
  playback-live-synced:
    backgroundColor: "{colors.bottle-green}"
    textColor: "{colors.bottle-ink}"
    typography: "{typography.label}"
    rounded: "{rounded.none}"
    padding: "9px 16px"
  point-drawer:
    backgroundColor: "{colors.ivory-sheet-light}"
    textColor: "{colors.ink-primary-light}"
    typography: "{typography.body}"
    rounded: "{rounded.none}"
    width: "min(392px, calc(100vw - 24px))"
  button-gate-submit:
    backgroundColor: "{colors.carbon-surface-light}"
    textColor: "{colors.ivory-sheet-light}"
    rounded: "{rounded.none}"
    padding: "10px 16px"
  button-gate-submit-hover:
    backgroundColor: "{colors.vermilion-signal}"
    textColor: "{colors.vermilion-ink}"
  admin-topbar:
    backgroundColor: "{colors.carbon-surface-light}"
    textColor: "{colors.ivory-sheet-light}"
    typography: "{typography.label}"
    rounded: "{rounded.none}"
    padding: "8px 18px"
    height: "56px"
  offline-demo-banner:
    backgroundColor: "{colors.amber-gold-fill}"
    textColor: "{colors.ink-primary-light}"
    typography: "{typography.label}"
    rounded: "{rounded.none}"
    padding: "9px 18px"
  admin-panel:
    backgroundColor: "{colors.ivory-sheet-light}"
    textColor: "{colors.ink-primary-light}"
    typography: "{typography.body}"
    rounded: "{rounded.none}"
    padding: "16px"
  directory-item-selected:
    backgroundColor: "{colors.carbon-surface-light}"
    textColor: "{colors.ivory-sheet-light}"
    typography: "{typography.body}"
    rounded: "{rounded.none}"
    padding: "8px 10px"
  admin-metric-card:
    backgroundColor: "{colors.ivory-well-light}"
    textColor: "{colors.ink-primary-light}"
    typography: "{typography.body}"
    rounded: "{rounded.none}"
    padding: "10px 12px"
  admin-mode-tab:
    backgroundColor: "{colors.ivory-well-light}"
    textColor: "{colors.ink-secondary-light}"
    typography: "{typography.label}"
    rounded: "{rounded.none}"
    padding: "5px 10px"
  admin-mode-tab-active:
    backgroundColor: "{colors.carbon-surface-light}"
    textColor: "{colors.ivory-sheet-light}"
---

# Design System: Aire Trujillo

## Overview

**Creative North Star: "The Amber Timetable Slide Rack"**

Aire Trujillo organizes urban environmental monitoring for Trujillo and Huanchaco across three coordinated expressions of the `rw-timetable-slide-rack` system: an archival editorial landing rack (`/`) of six progressive slide sheets resting on an illuminated light table (`#dda126` in light polarity; `#15120e` umber charcoal in dark polarity), a full-viewport operative cartographic console (`/map`), and a discreet unlinked operator gate and administrative workbench (`/edit`) for managing virtual stations and 5-minute batch telemetry. Across all modes, the interface balances immediate civic legibility—live coastal timetables, `50 m` IDW spatial heatmaps, `ECA / INCA (Perú)` ↔ `AQI (EPA)` switching, and 24-hour point inspection—with engineering inspection across inverse-distance weighting (IDW), custom PCB hardware, Go/PostGIS backend math, and IEEE documentation.

Every sheet, dock, workbench panel, and control plate pairs crisp `0px`-cornered warm ivory (`#f7f1e3`) or warm slate-charcoal (`#211c16`) surfaces and sunken wells (`#efe6d2` / `#2b241c`) with solid carbon anchors (`#191d20` / `#100d0a`), hairline structural rules, and `-9deg` raked parallelogram triggers whose counter-skewed labels stay upright while signaling instrument precision. Warm light and warm dark themes operate as equal, switchable environmental polarities across `/`, the WebGL basemap on `/map`, and the `/edit` workbench.

The system explicitly rejects generic SaaS landing pages built from identical icon-plus-text cards, rounded pill buttons, cold dark-mode blues, decorative drop shadows, and fabricated zero readings where sensor coverage ends.

**Key Characteristics:**
- Illuminated warm-amber light-table bed supporting stacked, unrounded (`0px`) archival slide sheets (`01` through `06`) on `/` and full-bleed `100svh` WebGL cartographic framing on `/map`.
- Sticky `86px` graduated left ruler rail synchronized via scroll observation on `/`, collapsing cleanly into a 3-column mobile strip below `960px`.
- Operative `/map` desk governed by a `56px` solid carbon top toolbar, collapsible bottom-left dynamic legend, anchored 24h playback scrubber bar, and right-docked point inspection drawer (bottom sheet below `900px`).
- Discreet `/edit` route isolated from all public navigation (`/` and `/map`), opening onto an unadorned single-field passphrase gate before unlocking a `1560px` two-pane operator workbench with bi-directional MapLibre coordinate picking, Peru `America/Lima` 24h batch ingestion, and individual Chart.js variable preview cards.
- Raked `-9deg` parallelogram buttons and standard/preview switches with counter-skewed upright labels, balanced by unskewed `0deg` instrumental readouts, selects, directory rows, and playback scrubbers.
- Single authored translucent glass-slate timetable panel (`-2.5deg` tilt on desktop) anchored in the opening viewport of `/`.
- Cartographic honesty encoded as a normative token: neutral `#9ca3af` and `"No hay datos"` strictly represent areas beyond `600 m` without coverage.
- Three self-hosted typefaces (`Barlow Condensed`, `Archivo`, `Azeret Mono`) enforcing tabular numeric lock across all telemetry cells, coordinates, and timestamps.

## Colors

The palette balances warm mineral paper and carbon ink with signal vermilion actions, bottle-green geospatial and live-sync accents, golden-amber active indicators, and an immutable neutral gray boundary for uncovered areas.

### Primary
- **Signal Vermilion** (`#b83a28` light / `#bf4128` dark; hover `#9d2f20` light / `#d35035` dark; ink `#fff9f2` light / `#fff8f0` dark): Primary map CTAs, active index/standard switches (`ECA / INCA` vs `AQI EPA` on both `/` and `/map`), active IDW distance checkpoint tabs, selected BOM blocks on the PCB canvas, `/map` playback progress track, non-live `"En vivo"` catch-up button, selected coordinate target ring/dot, light-mode 24h chart curve, `/edit` gate submit hover fill, `/edit` primary node/batch submission CTAs, selected coordinate picker badge (`.picker-node-tag.is-selected`), inline 2-step delete controls, `"Limpiar campos"` action, validation error alerts, light-mode Draft preview curves (`#B83A28`), node badge hover fill, and light-mode text selection caret.
- **Bottle Green** (`#265d43` light / `#2f6e4f` dark; ink `#f7f1e3` light / `#f5ecdb` dark): Geospatial and environmental stamps (`GRILLA 50 M · INCA`, backend layer tag, PCB subsystem badges) and the synced live-state `"En vivo"` CTA fill on `/map`. Shipped CSS tightened the original `#2C6B4F` contract hue for stronger ink contrast in both light and dark polarities.
- **Golden Amber Fill** (`#e7b043` light / `#d99b26` dark; ochre border accent `#8a5607` light / `#e6aa3c` dark): Active `01–06` rail tick fills, active PCB subsystem filter pills, `America/Lima` timestamp stamps, station code chips, active `/map` playback speed ink, hover tooltip frame border, `/edit` offline demo warning banner (`Sin conexión al backend :8080 — sesión de muestra local` with `#181a1c` ink and `#8a5607`/`#e6aa3c` bottom rule), `"Virtual"` node badges, `"Inactiva"` chosen toggle ink on carbon, dark-mode Draft preview dashed curves (`#E6AA3C`), dark-mode 24h chart curve (`#e6aa3c`), dark-mode focus rings, formula display ink on carbon, and custom scrollbar thumbs.

### Secondary
- **Tinted Glass Slate** (`rgba(25, 34, 38, 0.91)` light / `rgba(17, 14, 11, 0.88)` dark; row hover `rgba(44, 58, 64, 0.88)` light / `rgba(50, 41, 29, 0.88)` dark; inner rule `rgba(247, 241, 227, 0.16)` light / `rgba(245, 236, 219, 0.15)` dark): Translucent smoked surface reserved exclusively for the opening Trujillo–Huanchaco timetable overlay on `/`; its subtle inner rule variable also frames hairlines inside solid carbon `/map` and `/edit` headers.
- **PCB Solder-Mask & Copper** (substrate `#23332d` light / `#1b2420` dark; copper traces `#d99b26` light / `#e6aa3c` dark; silkscreen `#f1e6cf` light / `#f5ecdb` dark): Dedicated hardware diagram palette representing the marine-coated coastal sensor board on `/`.
- **Cartographic Basemap Surface** (MapLibre GL canvas shared by `/map` and `/edit`'s `AdminCoordPickerMap`): Warm parchment land (`#EFE6D2` light / `#1A1611` dark), muted Pacific Ocean water fill (`#D5DFDC` light / `#111A1C` dark), sepia coastline (`#8C7D65` / `#4B3E2D`), dashed graticule (`#D6C9AF` / `#2B241C`), and arterial strokes (`#B7A78B` / `#544533`), gently blended beneath desaturated CARTO raster tiles (`0.72` opacity, `-0.45`/`-0.42` saturation). Active WebGL station dots (`nodes-core`) render `#F7F1E3`/`#191D20` (light) and `#E6AA3C`/`#BF4128` (dark), whereas inactive station dots use warm sunken-well `#EFE6D2` fill with `#433A2E` stroke (light) and `#2B241C` fill with `#6B563A` stroke (dark)—preserving `#9CA3AF` strictly for uncovered `"No hay datos"` regions.
- **Station Status & Saved Telemetry Greens** (`#22c55e` active indicator dot and light validation/status border; `#166534` light ink and persisted chart stroke; `#4ade80` dark and on-carbon ink, border, live dot, and persisted chart stroke): Denotes active station status dots/badges (`Activa`), valid batch point-count stamps, async-confirmed node/batch notices, and persisted `"Guardado"` curves in `AdminMetricChartCard` (`#166534` stroke + `rgba(22, 101, 52, 0.10)` fill in light; `#4ade80` stroke + `rgba(74, 222, 128, 0.14)` fill in dark).

### Tertiary
- **Cartographic No-Data Neutral** (`#9ca3af` background with `#111827` high-contrast ink): Immutable across both light and dark themes. Reserved strictly for cells, WebGL tiles (`0.34` opacity), legend footer blocks, point-inspection banners, probes, tags, and rows where `has_data === false` or distance exceeds `R_max = 600 m` (`"No hay datos"`), plus `/map` playback historical-inactive indicator dots. Never used for inactive stations on `trujilloBasemap.ts`, inactive `/edit` directory rows, or placeholder skeletons.

### Neutral
- **Light-Table Bed** (`#dda126` base / `#c98c19` deep light; tightened in dark to `#15120e` base / `#0e0c09` deep with subtle `rgba(231, 175, 59, 0.1)` amber sheen): Continuous ambient workspace backdrop behind every slide sheet and the `/map` viewport shell.
- **Archival Slide Sheet & Sunken Well** (sheet `#f7f1e3` light / `#211c16` dark; sunken well `#efe6d2` light / `#2b241c` dark): Primary reading surface (slide sheets, `/map` legend panel, playback bar, point inspection drawer, layer `<select>`, MapLibre control buttons, `/edit` gate backdrop) and inset technical compartments (readout panes, metric registers, dossier boxes, chart canvas frames, empty chart wells, skeleton placeholders, `/edit` password field). Shipped dark surfaces tightened `#221D17` from the direction contract to `#211c16`.
- **Carbon Ink & Slate Anchors** (carbon `#191d20` light / `#100d0a` dark; slate `#29373e` light / `#332c24` dark): High-contrast headers for the graduated vertical rail, `/map` top toolbar, legend and drawer headers, clickable map node tags, play/pause button, active speed toggles, timeline scrubber thumb, pointer hover tooltip, IDW pane strips, PCB toolbar, formula bands, and modal header.
- **Typographic Inks & Hairline Rules** (primary ink `#181a1c` light / `#f5ecdb` dark; secondary ink `#433a2e` light / `#cbbca3` dark; muted carbon ink `#cfc3af` light / `#c2b298` dark; structural rules `#191d20` light / `#6b563a` dark; soft inner dividers `#c7b698` light / `#3d3224` dark): Ensures ≥4.5:1 legibility on warm surfaces across both themes.
- **Regulatory & Metric Status Overlays (Backend Data)** (`#22C55E` ECA Buena / normal bands, `#EAB308` ECA Moderada / EPA Moderate, `#F97316` Mala / Alto, `#EF4444` Peligrosa / Muy alto, `#3B82F6` Frío / Húmedo / Baja presión, `#00E400`–`#7E0023` EPA AQI 6-tier scale): Supplied by domain thresholds for heatmap fills, legend swatches, and contrast-computed drawer badges; never repurposed as general UI chrome.

### Named Rules
**The Warm-Polarity Parity Rule.** Both light and dark UI themes must remain anchored in warm amber, parchment, and umber charcoal (`#dda126`/`#f7f1e3` in light; `#15120e`/`#211c16` in dark); cool blue-gray or stark `#000000`/`#ffffff` UI surfaces are prohibited outside the smoked timetable slate, while ocean water fills (`#D5DFDC`/`#111A1C`) and regulatory status overlays remain confined to the geographic map canvas.
**The Honest Boundary Rule.** Neutral `#9ca3af` with `#111827` ink is reserved exclusively for missing coverage (`has_data === false`, `> 600 m`, `"No hay datos"`). Never repurpose `#9ca3af` for inactive UI tabs, placeholder skeletons, or inactive station markers (`#EFE6D2`/`#433A2E` light and `#2B241C`/`#6B563A` dark on the map; sunken-well tokens in `/edit`), and never display `0` for uncovered readings.
**The Signal Scarcity Rule.** Reserve Signal Vermilion (`#b83a28` / `#bf4128`) for primary actions, active probes, selected coordinate targets, and selected hardware blocks so urgent focus never competes with background ornamentation.

## Typography

**Display Font:** Barlow Condensed (fallback: sans-serif; self-hosted weights `600`, `700`, `800`)
**Body Font:** Archivo (fallback: sans-serif; self-hosted weights `400`, `500`, `600`, `700`)
**Label/Mono Font:** Azeret Mono (fallback: monospace; self-hosted weights `400`, `600`)

**Character:** All-caps condensed poster authority (`Barlow Condensed`) collides cleanly with high-legibility neo-grotesque prose (`Archivo`) and technical instrumentation monospacing (`Azeret Mono`), evoking printed 1970s transport timetables paired with modern environmental telemetry.

### Hierarchy
- **Display** (`800`, `clamp(2.55rem, 4.7vw, 4.15rem)`, `0.98` line-height, `-0.02em` tracking, uppercase): Main hero `h1` and large metric readouts (`clamp(2rem, 3.2vw, 2.75rem)` on `/` IDW probe; `2.75rem` desktop / `2.1rem` mobile on `/map` point inspection value). Carries the viewport immediately with no eyebrow or kicker above it.
- **Headline** (`800`, `clamp(1.9rem, 3.2vw, 2.75rem)`, `1.02` line-height, uppercase): Slide section `h2` titles (`02` through `06`) and modal `h2` header (`clamp(1.9rem, 3.2vw, 2.5rem)`).
- **Title** (`700`–`800`, `0.98rem–1.75rem` / representative `1.625rem`, `1.05` line-height, uppercase): Dossier block titles (`28px`), station detail headers (`26px`), `/map` carbon toolbar wordmark (`1.45rem` desktop / `1.15rem` mobile, `800`, `+0.03em` tracking), `/edit` `AdminTopBar` wordmark (`1.38rem` desktop / `1.15rem` mobile, `800`), `/edit` panel titles (`1.25rem`, `800`), inspection drawer title (`1.35rem`), `"NO HAY DATOS"` banner (`1.55rem`, `800`), legend, `/edit` section subheadings, and chart headers (`1.05rem` / `0.98rem`), formula/architecture titles (`22px–23px`), and Home header brand wordmark (`20px`, `800`, `+0.03em` tracking).
- **Body** (`400`–`600`, `1rem` base / `17px` hero lead / `15.5px` section summaries / `13px–14px` drawer, `/edit` descriptive name input, and technical notes / `11.5px–12.5px` legend rows, `/edit` directory names, and picker hints, `1.35–1.55` line-height): Constrained to `58ch` on hero leads and `65ch–68ch` on section summaries for effortless scanning.
- **Label** (`600`, `0.5625rem–0.8125rem` / primary button `12.5px` (`0.78125rem`) and `/edit` gate/mono inputs `13px`, `0.03em–0.08em` tracking, uppercase on controls/headers): Raked buttons, vertical rail ticks (`01–06`), timetable column headers, `/map` toolbar controls (`11px–11.5px`), `/edit` 24h Lima `<select>` (`12px`), CSV batch inputs (`11.5px`), mode/template tabs (`10.5px`), node code badges and playback/chart ticks (`9px–10px`), clock chips (`10.5px–12px`), JSON contract keys, variable chips, and colophon labels.

### Named Rules
**The Tabular Lock Rule.** Apply tabular numerals (`font-variant-numeric: tabular-nums`) to every clock time (`21:15 UTC-5`), pollutant readout, coordinate (`-8.111600, -79.028700`), distance checkpoint (`450 m`), node count, and stage code (`01–06`) so interactive state swaps never cause horizontal jitter.
**The Upright-Label Skew Rule.** Whenever a control container is skewed (`-9deg`), its inner text wrapper must apply the exact inverse skew (`9deg`) so monospace labels never distort diagonally.

## Layout

The landing architecture (`/`) uses a `1320px` max-width two-column slide rack centered inside the amber light-table bed (`28px clamp(18px, 4vw, 56px) 56px` padding). A sticky `86px` left graduated ruler rail anchors the `minmax(0, 1fr)` main slide column with a `24px` gutter. Inside the main column, a sticky top timetable bar (`top: 12px`, `z-index: 20`) sits above a vertical stack of six slide sheets separated by `32px` gaps, each padded with `clamp(24px, 3.2vw, 38px)` interior breathing room.

Internal splits vary deliberately per stage on `/` to avoid repetitive card grids:
- **Stage 01 Hero:** `1.06fr 0.94fr` split pairing left display copy and CTAs with the right-hand glass-slate timetable, capped along the bottom edge by a 5-column quick-preview strip (`02`–`06`).
- **Stage 02 Station Preview:** `0.95fr 1.05fr` split pairing a vertical station selector list with a 3-column × 4-row (`12`-variable) telemetry register, plus a wrapping canonical variable strip below.
- **Stage 03 IDW Probe:** `1.25fr 0.95fr` split pairing the 5-checkpoint SVG cross-section (`0–750 m`) with a 2-column backend contract field readout.
- **Stage 04 PCB Explorer:** `1.35fr 0.85fr` split pairing the zoom-safe horizontal-scrollable SVG PCB schematic and quick-chip bar with the sticky hardware BOM dossier.
- **Stage 05 Math & Stack:** Equal `repeat(2, 1fr)` ruled mathematical formula ledger above a 3-row architectural ledger (`240px minmax(0, 1fr)` tier column + 2-column specification grid).
- **Stage 06 IEEE Close:** `1.25fr 0.75fr` split pairing the IEEE modal trigger and author/map actions with a 4-row technical colophon table.

**Operative Cartographic Console (`/map`):**
The `/map` route occupies a locked full-viewport flex shell (`100svh`, `min-height: 540px`, `overflow: hidden`) stacking the solid carbon `MapTopToolbar` (`min-height: 56px`) and an optional `6px 16px` status banner above the full-bleed `flex: 1` `MapCanvas` stage (`min-height: 0`). Three operative docks float inside `MapStage`:
- **Bottom-Left Legend Dock:** `left: 12px; bottom: 94px` (`z-index: 10`).
- **Bottom 24h Playback Dock:** `left: 12px; right: 12px; bottom: 12px` (`z-index: 14`).
- **Right Point Inspection Drawer:** `top: 12px; right: 12px; bottom: 92px; width: min(392px, calc(100vw - 24px))` (`z-index: 15`).
- **Camera Safe Padding:** `fitCorridor` frames the Trujillo–Huanchaco corridor with asymmetric desktop bounds padding (`top: 48, bottom: 112, left: 268, right: 120` or `440` when the inspection drawer is open) and pans automatically when a selected coordinate falls within `484px` of the right viewport edge.

**Discreet Operator Gate & Workbench (`/edit`):**
- **Route Isolation & Gate (`FE-006`):** Replaced the placeholder with `EditView` at `/edit`, keeping zero public links to `/edit` from `/` or `/map`. Unauthenticated state renders strictly a centered single-screen grid (`min-height: 100svh; place-items: center; padding: 24px`) on `var(--surface-ivory)` containing only a `min(320px, 100%)` inline `<form>` (`8px` gap) with one `type="password"` input (no placeholder) and one arrow-submit button—zero headings, logos, subtitles, error hints, or offline banners. Silent rejection occurs on blank submissions or HTTP `401`/`403`/`429`; offline demo mode activates strictly when `:8080` is unreachable (network error, `502–504`, or non-JSON proxy/SPA fallback) and a non-empty passphrase is submitted.
- **Authenticated Workbench (`FE-007`):** Full-height flex shell (`min-height: 100dvh; overflow-x: clip`) on `var(--bed-bg)` stacking `AdminTopBar` (`min-height: 56px`), the conditional full-width `offline-demo-banner`, and a `1560px` max-width two-column split (`minmax(380px, 456px) minmax(0, 1fr)`, `18px` padding and gap, `align-items: start`) pairing `AdminNodePanel` (`AdminCoordPickerMap`) on the left with `AdminBatchPanel` (`AdminSeriesPreviewGrid` + `AdminMetricChartCard`) on the right.
- **Coordinate Picker Map Framing (`AdminCoordPickerMap`):** `230px` MapLibre GL canvas (`fitCorridor` bounds padding `top: 28, bottom: 28, left: 32, right: 24` desktop; `32, 38, 52, 20` below `380px`) with `NavigationControl` (zoom only) top-left, `"Centrar"` button top-right (`z-index: 4`), and bottom-left (`bottom: 8px; left: 8px; z-index: 4`) offline cartographic pill over the Pacific Ocean water corner.

**Responsive Breakpoints:**
- `≤ 1160px` (`/`): Top navigation horizontal raked tabs hide to preserve unobstructed header breathing room.
- `≤ 1100px` (`/map`): Topbar `America/Lima` timestamp chip hides while the bottom playback bar preserves the active timestamp.
- `≤ 1079px` (`/edit`): Two-pane workbench stacks to 1 column (`14px` padding and gap).
- `≤ 1024px` (`/`): PCB Explorer collapses from 2 columns to 1 vertical stack; board frame switches its right hairline divider to a bottom divider.
- `≤ 960px` (`/` + `/edit`): On `/`, left `86px` vertical rail hides, 3-column mobile stage strip appears below top bar, all multi-column splits stack to 1 column, tilted timetable schedule resets from `-2.5deg` to `0deg`, and bottom depth-preview strip drops to 2 columns; on `/edit`, the 12-variable CSV input grid shifts from 2 columns to 1.
- `≤ 900px` (`/map` + `/edit`): On `/map`, `MapLegend` narrows to `236px`, defaults to collapsed, repositions to `left: 10px; bottom: 128px` (`max-height: 40svh`), and hides when the inspection drawer opens; `PointInspectionDrawer` transforms into a bottom sheet (`left: 10px; right: 10px; bottom: 128px; max-height: calc(100% - 254px)`); `PlaybackBar` adopts `10px` edge insets. On `/edit`, `AdminTopBar` hides `"estaciones activas"` and `"Inicio"` labels.
- `≤ 820px` (`/edit`): `AdminSeriesPreviewGrid` 2-column chart card grid collapses to 1 column.
- `≤ 780px` (`/map`): `MapTopToolbar` switches to a 2-row grid (`minmax(0, 1fr) auto` brand/meta on row 1; full-width standard + layer selector on row 2) and collapses tool button text labels to `15px` icons; `PlaybackBar` switches to a 2-row grid (full-width scrubber on row 1; play/speed left and `"En vivo"` CTA right on row 2) and hides the step tag.
- `≤ 720px` (`/edit`): `AdminCoordPickerMap` canvas height compacts from `230px` to `196px`.
- `≤ 640px` (`/`): PCB SVG enforces `560px` minimum width inside an overflow-x scroll frame and reveals a mono scroll hint.
- `≤ 600px` (`/map` + `/edit`): On `/map`, status banner (`5px 10px`, `10.5px`) and offline basemap label (`3px 8px`, `9.5px`) compact cleanly; on `/edit`, `offline-demo-banner` pads `9px 12px`, `workbench-main` compacts to `10px` padding (`12px` gap), and `AdminTopBar` drops padding (`8px 12px`), shrinks its heading to `1.15rem`, and hides the theme label and brand divider.
- `≤ 580px` / `≤ 560px` (`/` + `/edit`): On `/`, top bar unsticks (`position: static`), wraps, and expands the primary map CTA to `100%` width; metric register shifts from 3 columns to 2; IDW checkpoint tabs shift from 5 columns to 3; contract field grid shifts from 2 columns to 1; depth-preview strip stacks to 1 column. On `/edit`, `AdminNodePanel` 2-column form fields grid shifts to 1 column.
- `≤ 520px` (`/map`): Topbar hides `Inicio` text, `"nodos activos"` word, and long standard prefixes/suffixes (`ECA /`, `(Perú)`, `(EPA)`), leaving compact `INCA` / `AQI` switches.

**Browser Surfaces:** Text selection uses Carbon Ink with Archival Ivory ink in light mode and Golden Amber with `#15120e` ink in dark mode; text input carets use Signal Vermilion; scrollbars measure `10px` with Golden Amber thumbs framed inside Carbon Ink tracks; keyboard focus rings render a crisp `2px solid` outline (`#b83a28` light / `#e6aa3c` dark) offset `3px` from the element edge.

## Elevation & Depth

**Zero Drop Shadows.** This system never uses blurred outer drop shadows (`box-shadow` offsets) to lift UI surfaces off the page. Depth is communicated entirely through four stacked tonal planes (`light-table bed` → `ivory/charcoal sheet` → `sunken well` → `solid carbon frame`), `1px` and `1.5px` structural hairlines, and an ambient `16s` diagonal light-table sheen sweeping behind the rack (`cubic-bezier(0.16, 1, 0.3, 1)`). On `/map` and `/edit`, `MapCanvas` and `AdminCoordPickerMap` explicitly override MapLibre's default control shadows (`box-shadow: none`) so floating docks (`MapLegend`, `PlaybackBar`, `PointInspectionDrawer`, `.hover-tooltip`) and picker overlays separate cleanly from the cartographic canvas through opaque ivory/carbon fills and `1.5px` structural borders, ordered along predictable z-index strata (`/map`: `0` canvas → `6` offline pill → `10` legend → `12` hover tooltip → `14` playback dock → `15` inspection drawer → `18` status banner → `20` top toolbar; `/edit` picker: `0` canvas → `1` node tags → `4` `"Centrar"` & bottom-left offline carto pill → `5` selected node tag rendered last in DOM → `6` hovered node tag → `20` `AdminTopBar`).

The single translucent layer is the opening timetable card on `/`, which uses tinted glass-slate (`rgba(25, 34, 38, 0.91)` light / `rgba(17, 14, 11, 0.88)` dark) with `backdrop-filter: blur(12px) saturate(118%)` and a `1px` top inner specular edge (`box-shadow: inset 0 1px 0`) so it floats visibly like a dark smoked-acetate timetable overlay on the ivory sheet. Modal overlays dim the viewport behind a heavy warm scrim (`rgba(24, 20, 14, 0.62)` light / `rgba(8, 6, 4, 0.78)` dark). On the offline geographic canvas itself, `.map-carto-label` and `.picker-carto-label` markers apply a tight `1px`–`2px` `text-shadow` cartographic halo solely to keep map lettering legible over arterial strokes.

Motion uses an exponential ease-out curve (`cubic-bezier(0.16, 1, 0.3, 1)`) across `0.14s–0.35s` theme and hover transitions (`0.14s–0.18s` across `/edit` controls and gate button), view-swap transitions (`opacity` + `8px` vertical shift), modal-sheet entrances (`14px` rise + `0.985` scale), `/map` drawer slides (`0.22s` `18px` horizontal desktop / `20px` vertical bottom-sheet mobile), IDW probe glides (`0.36s`), Chart.js curve updates (`180ms` on `/map`, suppressed to `'none'` during live scrubber point highlighting; `150ms` + `ResizeObserver` canvas resizes on `/edit`), loading skeleton pulses (`1.2s`), and dashed copper PCB trace currents (`0.9s linear infinite`). Under `@media (prefers-reduced-motion: reduce)`, smooth scrolling disables and all animations and transitions collapse to `0.01ms`.

### Named Rules
**The Zero-Shadow Hairline Rule.** Structural boundaries rely strictly on `1px`, `1.5px`, or `2px` hairline borders and tonal surface stepping; never add drop shadows to cards, buttons, floating map docks, or modals.
**The Authored Glass-Slate Exception.** Backdrop blur (`12px`, `saturate(118%)`) and inset top specular hairlines belong exclusively to the opening timetable card on `/`; do not apply glassmorphism to secondary panels, `/map` docks, or modals.

## Shapes

Every HTML interface container, slide sheet, `/map` dock, `/edit` workbench card and directory row, MapLibre control group, button, navigation tab, rail tick, chip, badge, select box, range scrubber thumb, input field, Chart.js tooltip (`cornerRadius: 0`), and modal adopts strict **`0px` square corners**, recalling die-cut paper cards and ruled engineering ledgers.

Kinetic movement is expressed through a **consistent `-9deg` forward skew** on key action and standard/preview switching controls (primary/secondary/carbon buttons, `/` top navigation tabs, IDW checkpoint tabs, PCB subsystem filter tabs, `/map` `ECA / INCA` ↔ `AQI (EPA)` tabs, `/map` `"En vivo"` raked CTA, and `/edit` `"Cerrar sesión"`, `"Nuevo nodo"`, `"Crear nodo"`/`"Guardar cambios"`, `"Guardar lote"`, and `"Combinado"` | `"Solo borrador"` | `"Solo guardado"` mode switches), paired with a `-(-9deg)` counter-skew on inner text spans. Operative dense controls inside `/map` and `/edit` (icon tool buttons, play/pause, `1x`/`2x`/`4x` speed selectors, legend toggles, `<select>` pickers, station directory items, `"Reintentar API"`, `"Ahora (Lima)"`, quick-fill preset chips, and inline delete confirm actions) remain unskewed `0deg` rectangles so packed instrumentation never collides diagonally. The opening glass-slate timetable on `/` receives a gentler `-2.5deg` architectural lean on desktop that straightens cleanly to `0deg` below `960px`.

Curved geometry is confined strictly inside data and hardware visualizations: the custom SVG PCB hardware canvas (`10px` radius on the green solder-mask board perimeter, `6px` on the dashed silkscreen border, `4px` on individual BOM component pads, plus circular mounting holes and IDW distance-probe dots) and WebGL/Chart.js cartographic point symbols (`nodes-outer-halo`, `nodes-core`, `selected-point-ring`, `selected-point-dot`, and the active 24h playback data marker).

### Named Rules
**The Sharp-Rectilinear Rule.** Keep `border-radius: 0px` across all HTML UI surfaces, MapLibre controls, range thumbs, tooltips, chips, and inputs; reserve curved corners exclusively for physical PCB board geometry inside SVG diagrams and circular geographic/chart coordinate markers.

## Components

Every component conveys crisp instrument precision: dense tabular typography, zero-radius boundaries, explicit hover/focus inversions, and immediate visual differentiation between active measurements and `#9ca3af` uncovered states.

### Buttons
- **Shape:** `-9deg` skewed parallelogram with `0px` corners and counter-skewed inner label wrapper for primary/standard actions; hovers lift `-2px` vertically along the skew axis (`0.22s cubic-bezier(0.16, 1, 0.3, 1)` on `/` and `/map`; compact `0.16s` transitions on `/edit`). Operative micro-controls (`.tool-btn`, `.play-btn`, `.close-btn`, `.speed-btn`, `.retry-api-btn`, `.now-btn`, `.center-map-btn`, `.preset-btn`, `.clear-btn`, `.danger-ghost-btn`, `.danger-solid-btn`) stay unskewed `0deg`.
- **Primary:** Signal vermilion background (`#b83a28` light / `#bf4128` dark), `#fff9f2` ink, `1.5px` vermilion border, `12px 22px` padding (`9px 16px` compact top-bar / `PlaybackBar` non-live `"En vivo"` variant; `9px 18px` on `/edit` `.raked-submit-btn` and `.raked-save-batch-btn`), inverting to deeper `#9d2f20` light / brighter `#d35035` dark on hover.
- **Secondary:** Raised well background (`#efe6d2` light / `#2b241c` dark), primary ink, `1.5px` structural border (`1px` on `/edit` `.secondary-link-btn` and `.preset-btn`), inverting cleanly to solid carbon (`#191d20` / `#100d0a`) with ivory ink on hover.
- **Carbon:** Solid carbon background (`#191d20` / `#100d0a`) with ivory ink, transitioning to signal vermilion on hover (used for closing secondary map links, `/map` `play-btn`, `/edit` `.discreet-submit` gate button, and `/edit` `.raked-new-btn` `"Nuevo nodo"` trigger).
- **Live Synced CTA (`/map`):** Raked `-9deg` `.live-btn.is-synced` switches from Signal Vermilion to Bottle Green (`#265d43` light / `#2f6e4f` dark) with ivory ink (`#f7f1e3` / `#f5ecdb`) when synchronized at the latest timestamp.
- **Operator Destructive & Utility Controls (`/edit`):** `.raked-logout-btn` pairs a `-9deg` transparent frame and `1px solid` vermilion border on carbon, filling vermilion on hover; `"Reintentar API"` pairs `#f7f1e3`/`#211c16` with `1.5px solid #181a1c` on the amber offline banner and inverts to carbon; 2-step inline deletion swaps `.danger-ghost-btn` (`Eliminar`) for `.danger-solid-btn` (`Confirmar baja`) beside `"Cancelar"`.
- **Theme & Map Utility Toggles:** Unskewed `0px` `1px`–`1.5px`-bordered buttons (`9px 12px` on `/`; `6px 10px` on `/map` and `/edit` carbon toolbars) pairing drawn `15px` SVG icons (`sun`, `moon`, `locate`) with labels that collapse gracefully to icons below `780px` (`/map`) or `600px` (`/edit`).

### Navigation
- **Top Timetable Bar (`/`):** Sticky ivory/charcoal bar (`10px 18px` padding, `1.5px` structural border) housing the `30px` carbon `AT` badge + condensed wordmark on the left, horizontal `-9deg` slide tabs in the center (hidden below `1160px`), and theme switch + primary map CTA on the right.
- **Operative Map Top Toolbar (`MapTopToolbar` on `/map`):** Full-width solid carbon header (`min-height: 56px; padding: 8px 16px; border-bottom: 1.5px solid var(--rule-color)`) housing the `Inicio` return link + condensed `AIRE TRUJILLO` heading + active node count pill on the left, `-9deg` `ECA / INCA (Perú)` ↔ `AQI (EPA)` switches + ivory `33px` layer `<select>` in the center, and amber `America/Lima` timestamp chip + `Centrar` + theme toggle on the right. Collapses into a clean 2-row CSS grid at `≤ 780px`.
- **Operator Workbench Header (`AdminTopBar` on `/edit`):** Full-width solid carbon header (`min-height: 56px; padding: 8px 18px; border-bottom: 1.5px solid var(--rule-color); z-index: 20`) pairing outbound `Mapa` + `Inicio` `0px` chips, a `1px × 22px` hairline divider, `AIRE TRUJILLO` (`1.38rem`), and the `"active/total estaciones activas"` pill on the left with theme toggle + `-9deg` `"Cerrar sesión"` button on the right.
- **Graduated Vertical Progress Rail (`/`):** Sticky `86px` carbon spine featuring a right-edge `5px`-wide `8px`-pitch amber measuring rule, `TRJ · HCH` / `UTC-5` stamps, and interactive `01–06` tick buttons that illuminate in solid golden amber (`#e7b043` / `#d99b26`) with dark `#181a1c` ink when active.
- **Mobile Stage Strip & Bottom Peek Strip (`/`):** Below `960px`, a 3-column unskewed carbon/ivory pill grid replaces the vertical rail; on the hero sheet footer, a 5-column (`2` cols tablet, `1` col mobile) ruled strip previews stages `02`–`06` and inverts to carbon on hover.

### Chips & Status Tags
- **Style:** Unskewed `0px` rectangular tags (`1px 5px` to `5px 12px` padding) in `Azeret Mono` `9.5px–12px` `600`.
- **Variants:** Amber timestamp tags (`21:15 UTC-5` and `/map` drawer node codes), golden-amber `"Virtual"` badges (`#181a1c` ink), bottle-green geospatial tags (`GRILLA 50 M · INCA`), `#9ca3af` + `#111827` no-data boundary tags (`CUT > 600 M`, `#9CA3AF · No hay datos`), contrast-computed regulatory category badges in `/map`, slate/charcoal node-count pills (`4 nodos activos` on `/map`; `active/total estaciones activas` with `#22c55e` dot on `/edit`), `"Activa"` (`#22c55e` border with `#166534` light / `#4ade80` dark ink) vs sunken-well `"Inactiva"` state tags, `"Guardado"` / `"Borrador"` series chips and `N pts` / `Error` badges on `/edit`, sharp `0px` offline basemap notices (top-left on `/map`; bottom-left ocean-corner on `/edit`), clickable carbon map node badges (`TRJ-01`, hovering vermilion; selected tag in vermilion at `z-index: 5` on `/edit`), and soft-ruled variable pills (`pm25 µg/m³`).

### Cards / Containers
- **Corner Style:** Strictly `0px` square edges throughout.
- **Background:** Ivory sheet (`#f7f1e3` / `#211c16`) for top-level slides, modals, `/map` floating docks (`MapLegend`, `PlaybackBar`, `PointInspectionDrawer`), `/edit` left/right workbench panels (`AdminNodePanel`, `AdminBatchPanel`), and MapLibre control buttons; raised well (`#efe6d2` / `#2b241c`) for nested inspection panels, readout wells, Chart.js frames, empty no-data chart wells, colophon tables, formula wells, `/edit` anchor/submit toolbars, `metric-input-card`, `metric-chart-card`, and dashed empty/loading boxes; carbon (`#191d20` / `#100d0a`) for headers, tooltips, and selected `/edit` station directory items (`.directory-item.is-selected`, matching `HomeView` carbon inversion with zero `border-left > 1px` stripe).
- **Shadow Strategy:** Zero outer shadows; internal separation via `1px` soft dividers (`#c7b698` / `#3d3224`) and `1.5px` outer rules (`#191d20` / `#6b563a`), except the `-2.5deg` tilted glass-slate timetable card (`backdrop-filter: blur(12px) saturate(118%)`, `inset 0 1px 0` specular top edge).

### Inputs / Fields
- **Layer Select (`MapTopToolbar`):** Unskewed `0px` ivory wrapper (`height: 33px` desktop / `30px` mobile, `min-width: 215px` desktop, `1px` hairline border) housing a native `<select>` styled in `Azeret Mono` `11.5px` `600` flanked by `14px` `layers` and `chevron-down` drawn SVGs; highlights the wrapper with a `2px solid var(--focus-ring)` outline on `:focus-within`.
- **24h Timeline Scrubber (`PlaybackBar`):** Custom `0px` `<input type="range">` (`8px` track height, `1px` structural border) filled dynamically with `var(--accent-vermilion)` up to `--progress` and `var(--surface-ivory-2)` beyond, driven by a `14px × 18px` carbon rectangular thumb.
- **Discreet Gate (`/edit`):** Centered `min(320px, 100%)` inline row (`8px` gap) pairing a `0px` sunken-well `type="password"` input (`#efe6d2` / `#2b241c`, `1.5px` structural border, `10px 12px` padding, `13px` `Azeret Mono`, `autocomplete="off"`, no placeholder, `2px` high-contrast focus ring offset `3px`) with a carbon submit button (`10px 16px` padding, `1.5px` border, vermilion fill on hover) bearing a drawn `15px` right-arrow SVG and zero public labels, error hints, or offline banners.
- **Workbench Fields & 24h Lima Anchor (`AdminNodePanel` + `AdminBatchPanel` on `/edit`):** Sunken-well `1.5px`-bordered directory search (`type="search"`), 2-segment `Activa` / `Inactiva` status switch (chosen state in solid carbon with `#4ade80` or golden-amber ink), `step="0.000001"` numeric `lat`/`lon` inputs synchronized bi-directionally with the MapLibre picker, `America/Lima` `-05:00` date field + native 288-slot 24h 5-minute `<select>` (`00:00` to `23:55` UTC-5, `12px` `Azeret Mono`, `chevron-down` `14px` SVG), and 12 canonical `METRIC_SPECS` comma-separated `tabular-nums` text inputs (index `0` newest -> steps back `-5 min`, max `288` points).

### Signature Components
- **IDW Spatial Coverage Probe (`/`):** Two-pane ledger pairing `-9deg` distance checkpoint tabs (`0 m`, `300 m`, `450 m`, `550 m`, `700 m`) and a scaled SVG radial cross-section (`0–750 m` showing `50 m` cell gridlines, `400 m` fade dashed marker, `600 m` cutoff wall, and animated vermilion probe glide) with a live backend contract panel displaying `has_data`, `label`, `current.confidence`, `current.value`, and `#9ca3af` null-contract alerts.
- **Interactive 11-Block PCB Explorer (`/`):** Full-width carbon-headed workstation containing subsystem filter tabs (`-9deg`), an interactive SVG board diagram on pine solder-mask (`#23332d` / `#1b2420`) with animated dashed copper traces (`0.9s linear infinite`) and clickable `4px`-radius pads, quick-access designator chips (`U1`, `U2`, `J1 · PM`, etc.), and a synchronized hardware BOM dossier listing connector, bus, quantity, and measured canonical variables.
- **Dynamic Cartographic Legend (`MapLegend` on `/map`):** `268px` (`236px` mobile) `0px` panel crowned by a collapsible carbon header (`1.05rem` display title + golden-amber unit badge + rotating chevron), listing active standard/variable swatches (`14px × 14px`), the `0–400 m Pleno` / `400–600 m Borde` decay bar, and the mandatory `#9ca3af` `"No hay datos · Sin cobertura · > 600 m"` footer block.
- **24h Temporal Playback Bar (`PlaybackBar` on `/map`):** Three-column operative transport strip pairing a carbon `24h`/`Pausa` button + `1x`/`2x`/`4x` speed switches (active speed in carbon + golden amber ink), the live/historical `America/Lima` timestamp + `8px` vermilion-progress range scrubber with start/mid/end `HH:MM` ticks, and the `-9deg` `"En vivo"` raked CTA.
- **Point Inspection Drawer & 24h Chart (`PointInspectionDrawer` on `/map`):** Right slide-over (`min(392px, calc(100vw - 24px))` desktop; bottom sheet below `900px`) featuring a carbon coordinate header, loading skeletons, explicit `#9ca3af` `"NO HAY DATOS"` banner + empty dashed chart well for uncovered coordinates, or covered readout well (`2.75rem` value, contrast-safe category badge, `7px` IDW confidence bar) above a `224px` (`150px` mobile) `Chart.js` 24-hour curve (`spanGaps: false`, `0px` carbon tooltip, playback-synced active point dot).
- **Trujillo–Huanchaco WebGL Map Canvas (`MapCanvas` on `/map`):** `MapLibre GL JS` canvas rendering `50 m` IDW polygon fills (`0.72` covered / `0.34` uncovered `#9ca3af`), warm-polarity offline-resilient coastal vector geometry (with sunken-well inactive station dots `#EFE6D2`/`#433A2E` light and `#2B241C`/`#6B563A` dark) beneath desaturated CARTO rasters, clickable carbon node code tags, vermilion coordinate target rings, shadow-free `0px` navigation controls, and pointer-fine carbon hover tooltips (`border: 1.5px solid var(--accent-amber-fill)`).
- **Offline Demo Session Banner (`EditView` on `/edit`):** Full-width golden-amber strip (`#e7b043` light / `#d99b26` dark, `#181a1c` ink, `1.5px` ochre/amber bottom rule) rendered strictly after authentication when `:8080` was unreachable, displaying `alert` (`15px`) + `"Sin conexión al backend :8080 — sesión de muestra local"` alongside the `0px` `"Reintentar API"` button (`refresh` `13px`).
- **Station Directory & Coordinate Picker (`AdminNodePanel` + `AdminCoordPickerMap` on `/edit`):** Left split workbench combining a filterable station directory (committed `HomeView` carbon inversion on `.is-selected`, zero `border-left > 1px` stripe), bi-directional numeric `lat`/`lon` inputs and `230px` (`196px` mobile) MapLibre click coordinate selector (`z-index: 5` + DOM-last ordering on `.picker-node-tag.is-selected`, bottom-left ocean placement for `.offline-tile-pill`, shared `trujilloBasemap.ts` inactive marker polarity), virtual node Create/Update, inline 2-step delete confirmation (`Eliminar` → `Confirmar baja` / `Cancelar`), and separated `nodeError` / async-confirmed `lastNodeNotice` alerts.
- **Batch Telemetry Workbench & Variable Preview Grid (`AdminBatchPanel` + `AdminSeriesPreviewGrid` + `AdminMetricChartCard` on `/edit`):** Right split workbench pairing the `America/Lima` `-05:00` date + `24h` 288-slot 5-minute `<select>` (`00:00..23:55`) + `"Ahora (Lima)"` clock bar, quick-fill templates (`Día típico Trujillo`, `Hora punta Av. Larco`, `Brisa costera Huanchaco`), 12 canonical `METRIC_SPECS` comma-input cards (newest-first `-5 min` stepping, max `288`), separated `batchError` / batch save result summary, `-9deg` `Combinado` | `Solo borrador` | `Solo guardado` toggles, and `146px` `Chart.js` variable cards (`ResizeObserver`, `cornerRadius: 0`, `spanGaps: false`, solid `#166534` light / `#4ade80` dark Saved series vs dashed `#B83A28` light / `#E6AA3C` dark Draft series).

## Do's and Don'ts

### Do:
- **Do** pair every light-mode surface change with its calibrated warm dark-mode token (`#dda126`/`#f7f1e3` ↔ `#15120e`/`#211c16`).
- **Do** enforce `0px` corners on every HTML UI container, button, tab, chip, modal, MapLibre control group, range thumb, Chart.js tooltip (`cornerRadius: 0`), and input outside SVG PCB substrate geometry and circular map/chart point markers.
- **Do** counter-skew inner text `<span>` elements (`9deg`) inside every `-9deg` raked control and apply `tabular-nums` to all numeric columns, coordinates, timestamps, and distance readouts.
- **Do** render `#9ca3af` (`#111827` ink) and the exact text `"No hay datos"` wherever `has_data === false` or distance exceeds `600 m`, keeping `spanGaps: false` on 24h series charts while reserving warm sunken-well tokens (`#EFE6D2`/`#433A2E` light, `#2B241C`/`#6B563A` dark) for inactive station markers.
- **Do** rely on `1px` / `1.5px` hairline borders and tonal surface contrast (`#f7f1e3` vs `#efe6d2` vs `#191d20`) rather than drop shadows across `/`, `/map`, and `/edit` (using carbon surface inversion rather than `border-left > 1px` stripes on selected directory rows).
- **Do** keep `/edit` unlinked from public routes, keep its unauthenticated gate strictly to one `type="password"` input and arrow submit button, activate offline demo mode only on non-empty passphrase submission when `:8080` is unreachable, and separate `nodeError`, `batchError`, and async-confirmed `lastNodeNotice` channels.
- **Do** clamp all transitions and keyframe animations to `0.01ms` under `@media (prefers-reduced-motion: reduce)`.

### Don't:
- **Don't** introduce cold blue-slate dark modes, neon cyan/purple gradients, or pure `#000000` / `#ffffff` UI surfaces outside cartographic water fills (`#D5DFDC` / `#111A1C`) and smoked timetable glass.
- **Don't** add outer drop shadows (`box-shadow` with Y/blur spread) to cards, floating map docks, tooltips, or MapLibre controls, add `border-left > 1px` stripes to selected rows/cards, or convert buttons and status tags into rounded pills.
- **Don't** place kickers or eyebrows above display or section headings; let condensed uppercase titles stand directly.
- **Don't** substitute `0`, `null`, `NaN`, or extrapolated green colors in uncovered zones (`> 600 m`), repurpose `#9ca3af` for inactive station dots/badges, show public navigation links to `/edit`, or expose offline demo banners on the unauthenticated `/edit` gate.
- **Don't** use glassmorphism (`backdrop-filter`) on cards other than the opening timetable slate on `/`, and don't replace drawn `1.8px`-stroke geometric SVG icons with emoji or text glyphs.

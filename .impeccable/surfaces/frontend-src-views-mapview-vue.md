---
version: 1
slug: "frontend-src-views-mapview-vue"
primary_target: "frontend/src/views/MapView.vue"
related_targets: ["frontend/src/App.vue","frontend/src/stores/map.ts","frontend/src/api/client.ts"]
---

# Surface Brief — Mapa Principal (`/map`) (`FE-002` a `FE-005`)

- **Scope & Mode:** Route `/map` in `frontend/` (`MapView.vue`, shell `App.vue`, `api/client.ts`, `stores/map.ts`, `data/legends.ts`, `components/map/*`). Mode: **Operate** (`code-led`, `buildPath: code`).
- **Audience & Job:** Evaluadores técnicos y ciudadanía de Trujillo / Huanchaco consultando en mapa WebGL (`MapLibre GL JS`) la distribución espacial actual e histórica de calidad del aire (**ECA / INCA Perú** predeterminado, conmutable a **AQI US EPA** y 12 variables individuales), verificando fronteras sin cobertura (`#9CA3AF`, `"No hay datos"`), inspeccionando series de 24h por punto (`(lat, lon)`) y reproduciendo evolución temporal en pasos de 5 minutos (`America/Lima`).
- **Primary Action & Proof:** Leer instantáneamente el mapa de calor interpolado en backend, conmutar estándares/capas sin latencia bloqueante, hacer clic en cualquier coordenada o estación para abrir el panel puntual con gráfico de 24h (o declaración explícita `"No hay datos"` fuera de cobertura) y controlar el playback de 24 horas (`Play/Pause`, `1x/2x/4x`, *scrubber*, `"En vivo"`).
- **Constraints:** Cero cálculos IDW o INCA/EPA en cliente; cero usos de `any`; cero enlaces públicos a `/edit`; respetar `0px` radio, bordes hairline y paleta cálida dual (`#f7f1e3`/`#191d20` claro, `#211c16`/`#100d0a` oscuro) sin glassmorphism en paneles secundarios ni sombras exteriores; `#9CA3AF` + `"No hay datos"` estricto para `has_data === false`. Banner explícito de respaldo de muestra únicamente cuando `:8080` esté inaccesible, con reintento manual a API real.
- **Signature Interaction & Motion Grammar:** Controles sesgados `-9deg` (con texto enderezado `9deg`) para acciones clave (`ECA / INCA` ↔ `AQI EPA`, regreso a `En vivo`), transiciones operacionales rápidas (`150–250ms` con `--ease-expo`), marcador puntual animado con anillo de estación activa, barra de reproducción inferior fija sobre lienzo completo y panel puntual deslizable derecho (escritorio) / hoja inferior (móvil).

## Direction contract

- **THESIS:** Mesa de control cartográfico operativa de pantalla completa donde el mapa WebGL ocupa el protagonismo absoluto, envuelto por barra superior en carbón, leyenda cromática dinámica a la izquierda, reproductor 24h fijo abajo y cajón deslizable de consulta puntual a la derecha (hoja inferior en móvil).
- **OWN-WORLD:** Hereda exactamente el mundo `rw-timetable-slide-rack` en clave **Operate**: barra carbón (`#191d20` claro / `#100d0a` oscuro), superficies instrumentales en marfil (`#f7f1e3` / `#211c16`) y pozo hundido (`#efe6d2` / `#2b241c`), acento bermellón (`#b83a28` / `#bf4128`) para puntos seleccionados e indicador activo, verde botella (`#265d43` / `#2f6e4f`) para estado en vivo, y gris neutro `#9ca3af` exclusivo para sin cobertura (`"No hay datos"`).
- **STORY:** El visitante entra desde `/` o directo a `/map`, ve inmediatamente Trujillo–Huanchaco con su malla de `50 m` y nodos activos sobre **ECA / INCA (Perú)**, cambia a **AQI (EPA)** o variables físicas con barra superior y leyenda sincronizada, hace clic en cualquier zona (cubierta o gris sin datos) para inspeccionar su serie de 24h en panel deslizable, y recorre las últimas 24 horas paso a paso con el *playback* inferior.
- **FIRST VIEWPORT:** Cabecera operativa superior (`52px–58px`) con retorno a `/`, selector `ECA / INCA` ↔ `AQI EPA`, selector de las 12 variables individuales, estampilla horaria `America/Lima` y selector claro/oscuro; debajo, lienzo `MapLibre GL JS` a ancho/alto completo con zoom/navegación, marcador de nodos activos, leyenda dinámica compacta abajo-izquierda y barra de reproducción temporal (`Play/Pause`, `1x/2x/4x`, scrubber 24h, `"En vivo"`) anclada abajo.
- **FORM:** Composición estructural en mapa a pantalla completa (`pinned by user UX confirmation`) sobre el sistema visual `rw-timetable-slide-rack` (`DESIGN.md`, `seed:6243dd066db22f12` en registro operativo; decisión de composición ratificada por usuario).
- **FINISH:** "unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, DESIGN.md, and every shipping raster carrying its provenance"

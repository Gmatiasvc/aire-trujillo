---
version: 1
slug: "frontend-src-views-homeview-vue"
primary_target: "frontend/src/views/HomeView.vue"
related_targets: ["frontend","frontend/src/App.vue"]
---

# Surface Brief — Home (`/`) (`FE-001`)

- **Scope & Mode:** Route `/` in `frontend/` (`HomeView.vue`, shell `App.vue`). Mode: **Persuade** (`code-led`, `buildPath: code`).
- **Audience & Job:** Evaluadores académicos/IEEE y ciudadanía de Trujillo / Huanchaco que necesitan pasar en segundos del acceso directo al mapa (`/map`) hacia la comprensión progresiva del sistema físico y computacional.
- **Primary Action & Proof:** Ir a `/map`, validar una vista previa demostrativa con conmutador `ECA / INCA (Perú)` vs `AQI (EPA)`, verificar la frontera cartográfica honesta (`400–600 m` hacia `#9CA3AF` `"No hay datos"`), inspeccionar los 11 bloques del PCB custom, revisar la arquitectura Go/PostGIS/MQTT y abrir la ficha técnica del informe IEEE + enlace `https://gmatiasvc.vercel.app`.
- **Constraints:** Paleta cálida tanto en modo claro como en modo oscuro con selector explícito; textos concisos y complejidad creciente en 6 niveles (`01`–`06`); animaciones de entrada, salida (`<Transition>`), y hover; cero enlaces públicos a `/edit`; prohibido usar `any`, inventar métricas reales o mostrar `0` en lugar de `"No hay datos"`.
- **Signature Interaction & Motion Grammar:** Riel lateral graduado `01–06` sincronizado al desplazamiento + haz de luz cálida sobre la mesa de trabajo + conmutación con transición de entrada/salida en la vista demostrativa, sonda interactiva de distancia IDW (`0–800 m`), placa PCB interactiva de 11 bloques con trazado de buses y ficha modal del informe IEEE.
- **Unresolved Decisions:** URL definitiva del paper IEEE pendiente (`Próximamente` con ficha modal local); nombre provisional **Aire Trujillo**.

## Direction contract

- **THESIS:** La portada opera como un tarjetero horario ambiental de 6 láminas progresivas sobre una mesa iluminada en ámbar cálido (en claro) y carbón ámbar (en oscuro), rechazando la landing genérica de tarjetas iguales con gradientes neón fríos.
- **OWN-WORLD:** Mesa de luz en ámbar dorado (`#DDA126` claro) / carbón umbrío (`#16130F` oscuro) con láminas en marfil (`#F7F1E3`) / pizarra cálida (`#221D17`), botones y pestañas sesgados (`-9deg` con texto enderezado), acentos bermellón (`#B83A28`) y verde botella (`#2C6B4F`), frontera neutra `#9CA3AF` para `"No hay datos"`, y tipografía condensada (`Barlow Condensed`) + grotesca técnica (`Archivo`) + monoespaciada tabular (`Azeret Mono`).
- **STORY:** El visitante entra por la necesidad cotidiana (abrir `/map`), comprueba datos demostrativos locales, entiende por qué la red jamás inventa valores fuera de `600 m`, explora los 11 bloques físicos del PCB y cierra verificando la arquitectura Go/PostGIS/MQTT y el informe IEEE.
- **FIRST VIEWPORT:** Barra superior en marfil/carbón con marca, anclas de progreso, conmutador de tema cálido claro/oscuro y botón directo a `/map`; debajo, riel vertical izquierdo `01–06` junto a la lámina principal dividida en titular condensado directo, resumen de dos líneas y botón primario `"IR AL MAPA INTERACTIVO →"` a la izquierda, y tabla horaria de pulso Trujillo–Huanchaco (incluyendo fila `"No hay datos"` `#9CA3AF`) a la derecha, dejando asomar la lámina `02` al pie.
- **FORM:** `rw-timetable-slide-rack` (Tarjetero de Horarios sobre Mesa Ámbar, retador competitivo adoptado por elección del usuario en la mesa de decisión; `seed:46908e12c545f558`).
- **FINISH:** "unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, DESIGN.md, and every shipping raster carrying its provenance"

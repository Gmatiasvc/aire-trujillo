# Registro Maestro de Features — Backend

| ID | Feature | Estado | Design Doc |
| :--- | :--- | :--- | :--- |
| BE-001 | Infraestructura base (`docker-compose.yml`: PostgreSQL 16 + PostGIS, Eclipse Mosquitto, Go API) | Implementado | [BE-001-infraestructura-base.md](./BE-001-infraestructura-base.md) |
| BE-002 | Autenticación de administrador por *passphrase* (hash + salt) | Implementado | [BE-002-auth-passphrase.md](./BE-002-auth-passphrase.md) |
| BE-003 | Gestión de nodos virtuales (CRUD, coordenadas `lat/lon`) y carga de series simuladas (5 min) | Implementado | [BE-003-nodos-virtuales-y-simulacion.md](./BE-003-nodos-virtuales-y-simulacion.md) |
| BE-004 | Ingesta MQTT desde Eclipse Mosquitto y persistencia en PostGIS | Implementado | [BE-004-ingesta-mqtt-postgis.md](./BE-004-ingesta-mqtt-postgis.md) |
| BE-005 | Motor de cálculo de índices: ECA / INCA (Perú) y AQI (EPA) | Implementado | [BE-005-indices-calidad-aire.md](./BE-005-indices-calidad-aire.md) |
| BE-006 | Interpolación espacial dinámica y manejo explícito de zonas sin datos | Implementado | [BE-006-interpolacion-espacial.md](./BE-006-interpolacion-espacial.md) |
| BE-007 | Endpoints REST para heatmap, consulta puntual interpolada y *playback* 24h | Implementado | [BE-007-api-rest.md](./BE-007-api-rest.md) |

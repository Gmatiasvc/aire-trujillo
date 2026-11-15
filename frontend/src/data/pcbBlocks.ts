import type { MetricKey } from '../types/telemetry'

export type PcbGroup = 'all' | 'sensors' | 'compute' | 'power'

export interface PcbBlock {
  id: string
  designator: string
  name: string
  subsystem: string
  group: Exclude<PcbGroup, 'all'>
  quantity: string
  bus: string
  connector: string
  functionSummary: string
  engineeringDetail: string
  metrics: MetricKey[]
  x: number
  y: number
  w: number
  h: number
  pinNodes: Array<[number, number]>
}

export const PCB_BLOCKS: PcbBlock[] = [
  {
    id: 'esp32-s3',
    designator: 'U1',
    name: 'ESP32-S3-WROOM-1-N16R8',
    subsystem: 'MCU Principal',
    group: 'compute',
    quantity: '1×',
    bus: 'I2C · I2S · SPI · UART',
    connector: 'Módulo SMD en placa base',
    functionSummary: 'Orquesta el muestreo cada 5–10 min, compensación y publicación MQTT.',
    engineeringDetail:
      'Ejecuta firmware C/ESP-IDF con adquisición sincronizada de los 12 parámetros, empaquetado JSON canónico y cola de respaldo offline ante pérdida de señal celular.',
    metrics: [],
    x: 245,
    y: 120,
    w: 170,
    h: 112,
    pinNodes: [
      [245, 176],
      [415, 176],
      [330, 120],
      [330, 232],
    ],
  },
  {
    id: 'sim7080g',
    designator: 'U2',
    name: 'SIMCom SIM7080G',
    subsystem: 'Módem LTE-M / MQTT',
    group: 'compute',
    quantity: '1×',
    bus: 'UART + Antena LTE',
    connector: 'Conector u.FL / RF',
    functionSummary: 'Transmite telemetría MQTT QoS 1 hacia el broker Eclipse Mosquitto.',
    engineeringDetail:
      'Publica tramas al tópico aire-trujillo/v1/nodes/{node_code}/telemetry con bajo consumo entre ráfagas de transmisión.',
    metrics: [],
    x: 455,
    y: 44,
    w: 168,
    h: 92,
    pinNodes: [
      [455, 90],
      [415, 152],
    ],
  },
  {
    id: 'sps30',
    designator: 'J1 · PM',
    name: 'Sensirion SPS30',
    subsystem: 'Material Particulado',
    group: 'sensors',
    quantity: '1×',
    bus: 'I2C / UART',
    connector: 'Header JST-ZHR-5',
    functionSummary: 'Mide fracciones ópticas PM2.5 y PM10 por dispersión láser.',
    engineeringDetail:
      'Conecta mediante Header JST-ZHR-5 para limpieza automática y lectura calibrada en µg/m³.',
    metrics: ['pm25', 'pm10'],
    x: 34,
    y: 38,
    w: 158,
    h: 82,
    pinNodes: [
      [192, 79],
      [245, 146],
    ],
  },
  {
    id: 'ads1115',
    designator: 'U3 · U4',
    name: '2× TI ADS1115',
    subsystem: 'Conversión Analógica (ADC)',
    group: 'sensors',
    quantity: '2× (8 ch)',
    bus: 'I2C (16-bit)',
    connector: 'Puente a Headers JST-XH-4',
    functionSummary: 'Digitaliza los 4 gases electroquímicos y canales de corrección térmica.',
    engineeringDetail:
      'Dos ADC de 16 bits (8 canales totales) capturan los 4 Winsen ZE03 más 1 canal libre de referencia cero por sensor.',
    metrics: ['co', 'no2', 'so2', 'h2s'],
    x: 245,
    y: 266,
    w: 170,
    h: 66,
    pinNodes: [
      [330, 266],
      [330, 232],
    ],
  },
  {
    id: 'ze03-array',
    designator: 'J2–J5',
    name: '4× Winsen ZE03',
    subsystem: 'Gases Electroquímicos',
    group: 'sensors',
    quantity: '4× módulos',
    bus: 'Analógico → ADS1115',
    connector: '4× Headers JST-XH-4',
    functionSummary: 'Sensores independientes para CO, NO2, SO2 y H2S con lectura compensada.',
    engineeringDetail:
      'Cada cabezal conecta vía JST-XH-4 y entrega tensión proporcional convertida por firmware a mg/m³ (CO) y µg/m³ (NO2, SO2, H2S).',
    metrics: ['co', 'no2', 'so2', 'h2s'],
    x: 34,
    y: 248,
    w: 168,
    h: 92,
    pinNodes: [
      [202, 294],
      [245, 294],
    ],
  },
  {
    id: 'scd41',
    designator: 'U5',
    name: 'Sensirion SCD41',
    subsystem: 'Dióxido de Carbono',
    group: 'sensors',
    quantity: '1×',
    bus: 'I2C',
    connector: 'Soldado en zona ambiental',
    functionSummary: 'Medición fotoacústica NDIR de CO2 urbano en partes por millón.',
    engineeringDetail:
      'Opera entre 400 y 10 000 ppm con compensación ambiental cruzada.',
    metrics: ['co2'],
    x: 34,
    y: 144,
    w: 158,
    h: 44,
    pinNodes: [
      [192, 166],
      [245, 166],
    ],
  },
  {
    id: 'sgp40',
    designator: 'U6',
    name: 'Sensirion SGP40',
    subsystem: 'Compuestos Orgánicos Volátiles',
    group: 'sensors',
    quantity: '1×',
    bus: 'I2C',
    connector: 'Bus I2C ambiental',
    functionSummary: 'Estima el índice relativo VOC (1–500) compensado por humedad.',
    engineeringDetail:
      'Calcula el índice VOC sin calibración manual usando temperatura y humedad del BME280.',
    metrics: ['voc_index'],
    x: 34,
    y: 196,
    w: 158,
    h: 42,
    pinNodes: [
      [192, 217],
      [245, 200],
    ],
  },
  {
    id: 'bme280',
    designator: 'U7',
    name: 'Bosch BME280',
    subsystem: 'Microclima Costero',
    group: 'sensors',
    quantity: '1×',
    bus: 'I2C',
    connector: 'Abertura en borde térmico',
    functionSummary: 'Registra temperatura (°C), humedad relativa (%) y presión (hPa).',
    engineeringDetail:
      'Aporta variables meteorológicas de la brisa de Huanchaco y compensación térmica para los gases electroquímicos.',
    metrics: ['temperature_c', 'humidity_pct', 'pressure_hpa'],
    x: 455,
    y: 156,
    w: 168,
    h: 72,
    pinNodes: [
      [455, 192],
      [415, 192],
    ],
  },
  {
    id: 'ics43434',
    designator: 'MK1',
    name: 'InvenSense ICS-43434',
    subsystem: 'Acústica (Ruido)',
    group: 'sensors',
    quantity: '1×',
    bus: 'I2S Digital',
    connector: 'Puerto acústico inferior',
    functionSummary: 'Micrófono MEMS digital para presión sonora ponderada en dBA.',
    engineeringDetail:
      'Captura audio digital directo por I2S sin etapa analógica ruidosa, promediado en dBA entre 20 y 130 dBA.',
    metrics: ['noise_dba'],
    x: 455,
    y: 244,
    w: 168,
    h: 52,
    pinNodes: [
      [455, 270],
      [415, 212],
    ],
  },
  {
    id: 'power-solar',
    designator: 'U8–U10',
    name: 'CN3791 + TPS61023',
    subsystem: 'Solar + TPS22919 Switch',
    group: 'power',
    quantity: '3× ICs',
    bus: 'MPPT + Boost 5V/3V3',
    connector: 'Entrada panel solar + Li-Ion',
    functionSummary: 'Carga solar MPPT (CN3791), boost TPS61023 y corte TPS22919.',
    engineeringDetail:
      'CN3791 gestiona batería de litio desde panel solar; TPS61023 y TPS22919 aíslan rieles de sensores entre ciclos de lectura.',
    metrics: [],
    x: 245,
    y: 38,
    w: 170,
    h: 64,
    pinNodes: [
      [330, 102],
      [330, 120],
    ],
  },
  {
    id: 'storage-usb',
    designator: 'J6 · J7',
    name: 'MicroSD + USB-C 16P',
    subsystem: 'Respaldo y Protección IP66',
    group: 'power',
    quantity: '1× + 1×',
    bus: 'SPI + USB 2.0',
    connector: 'Socket MicroSD + USB-C 16P',
    functionSummary: 'Buffering local offline y puerto de mantenimiento en caja IP66.',
    engineeringDetail:
      'Almacena lecturas en MicroSD cuando cae LTE-M; la placa lleva conformal coating contra corrosión marina en poste.',
    metrics: [],
    x: 455,
    y: 308,
    w: 168,
    h: 54,
    pinNodes: [
      [455, 335],
      [415, 224],
    ],
  },
]

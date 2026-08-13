# SLO: latencia de gostalgia-api

## SLI (indicador)

Fracción de requests HTTP a `gostalgia-api` cuya duración es `<= 200ms`,
medida sobre `gostalgia_http_request_duration_seconds_bucket{le="0.2"}`.

```promql
sum(rate(gostalgia_http_request_duration_seconds_bucket{le="0.2"}[<window>]))
  /
sum(rate(gostalgia_http_request_duration_seconds_count[<window>]))
```

El histograma tiene un bucket exacto en `0.2` (ver comentario en
`internal/infra/metrics/metrics.go`) a propósito: `histogram_quantile` y
los ratios por bucket solo son exactos en un límite que exista como
bucket real. Con los `prometheus.DefBuckets` de fábrica (que saltan de
`0.1` a `0.25`) esta cuenta hubiera salido interpolada — y mal.

## SLO (objetivo)

**99% de requests bajo 200ms**, medido sobre una ventana de 30 días.

¿Por qué 200ms y no otro número? Es un umbral razonable para una API
REST de lectura (búsqueda/listado de archivos) sin llamadas a
servicios externos lentos — no es un número derivado de un SLA real con
usuarios, es el punto de partida para tener *algo* medible y iterar.

## Error budget

`1 - SLO = 1%` de los requests en 30 días pueden tardar más de 200ms sin
que se incumpla el objetivo. Ese 1% es el "presupuesto" para desplegar,
tener degradaciones puntuales, picos de carga, etc., sin que cada
request lento dispare pánico.

## Alerting: multi-window burn rate

En vez de una sola alerta de umbral fijo, dos alertas basadas en
**velocidad de consumo del budget** (burn rate), siguiendo el approach
de Google SRE ([*Alerting on SLOs*](https://sre.google/workbook/alerting-on-slos/)):

```
burn_rate(window) = (1 - success_ratio(window)) / (1 - SLO)
                   = (1 - success_ratio(window)) / 0.01
```

Un `burn_rate` de 1 significa "gastando el budget exactamente al ritmo
sostenible para 30 días". Más que 1, se agota antes.

| Alerta     | Ventana larga | Ventana corta | Burn rate | Budget en la ventana | Se agota en | Severidad |
|------------|----------------|----------------|-----------|-----------------------|-------------|-----------|
| Fast burn  | 1h             | 5m             | > 14.4    | 2%                    | ~2 días     | page      |
| Slow burn  | 6h             | 30m            | > 6       | 5%                    | ~5 días     | ticket    |

Cada alerta exige que **ambas** ventanas (larga + corta) superen el
umbral a la vez — evita que un pico de 2 minutos que se resuelve solo
dispare una alerta de "page". Tabla de burn rates estándar tomada del
workbook de Google para un SLO del 99% con ventana de 30 días.

Definición completa (recording rules + alertas) en
`~/poor-mans-fury/observability/gostalgia-slo-rules.yaml` (`PrometheusRule`).

## Limitación conocida

`gostalgia-api` en este homelab recibe tráfico bajo/sintético (curls de
prueba, no usuarios reales constantes) — el mecanismo de burn-rate
alerting está bien implementado y es correcto, pero con tan poco
volumen de requests las ventanas cortas (5m/30m) van a tener ratios
ruidosos (pocas muestras). El valor de este ejercicio es la técnica
(cómo se define un SLO real y se alerta sobre el error budget, no solo
sobre un umbral), no la significancia estadística de las alertas en
este entorno particular.
